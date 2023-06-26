package utils

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/bengosborn/cue/helpers"
	"github.com/bengosborn/cue/utils"
	"github.com/redis/go-redis/v9"
)

type UserData struct {
	Lat       float32   `json:"lat"`
	Long      float32   `json:"long"`
	Timestamp time.Time `json:"timestamp"`
}

type stackEvent int

const (
	eventUpsert stackEvent = iota
	eventDelete
)

type stackNode struct {
	Event stackEvent `json:"event"`
	User  string     `json:"user"`
}

type Location struct {
	ctx        context.Context                `json:"-"`
	redis      *redis.Client                  `json:"-"`
	lock       *utils.ResourceLockDistributed `json:"-"`
	mutex      sync.RWMutex                   `json:"-"`
	Location   sync.Map                       `json:"location"`
	User       sync.Map                       `json:"user"`
	EventStack *list.List                     `json:"eventStack"`
}

const (
	timeout  = 5 * time.Minute
	stateKey = "location:stage"
)

// **** I need to use this with a distributed lock and redis...
// 1. When reading, we need to look at redis values and the values we have locally and choose the one with the most recent timestamp
// 2. When writing, we will have a method to sync the local data with what is in the redis database, where it will only update the partition with the latest timestamp

// Make a new location structure
func NewLocation(ctx context.Context, redis *redis.Client, lock *utils.ResourceLockDistributed) *Location {
	return &Location{ctx: ctx, Location: sync.Map{}, User: sync.Map{}, redis: redis, lock: lock, EventStack: list.New()}
}

// Add a new user
func (l *Location) upsert(user string, lat float32, long float32, timestamp time.Time) error {
	// Remove the user from the previous partition if they already exist
	l.remove(user)

	// Add the user to the partition
	partition, err := NewPartitionFromCoords(lat, long)
	if err != nil {
		return err
	}

	l.User.Store(user, partition)
	value, _ := l.Location.LoadOrStore(partition.Encoded, make(map[string]*UserData))
	partitionUsers := value.(map[string]*UserData)

	partitionUsers[user] = &UserData{Lat: lat, Long: long, Timestamp: timestamp}

	return nil
}

// Public method for upsert which records event and locks
func (l *Location) Upsert(user string, lat float32, long float32) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	timestamp := time.Now()

	if err := l.upsert(user, lat, long, timestamp); err != nil {
		return err
	}

	l.EventStack.PushFront(&stackNode{Event: eventUpsert, User: user})

	return nil
}

// Remove a user
func (l *Location) remove(user string) bool {
	// Remove user
	l.User.Delete(user)

	value, ok := l.User.Load(user)
	if ok {
		prev := value.(*Partition)

		value, ok := l.Location.Load(prev.Encoded)
		if ok {
			partitionUsers := value.(map[string]*UserData)
			delete(partitionUsers, user)

			return true
		}
	}

	return false
}

// Public method for remove which records event and locks
func (l *Location) Remove(user string) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.remove(user) {
		l.EventStack.PushFront(&stackNode{Event: eventDelete, User: user})
		return true
	}

	return false
}

// Lookup a users data
func (l *Location) get(user string) (*UserData, bool) {
	value, ok := l.User.Load(user)
	if !ok {
		return nil, false
	}
	userPartition := value.(*Partition)

	value, ok = l.Location.Load(userPartition.Encoded)
	if !ok {
		panic("user exists but not within partition")
	}
	partitionUsers := value.(map[string]*UserData)

	userData, ok := partitionUsers[user]

	return userData, ok
}

// Public method for get with locks
func (l *Location) Get(user string) (*UserData, bool) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.get(user)
}

// Get nearby users
func (l *Location) Nearby(user string, radius int) ([]string, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	// Get the partition for the user
	value, ok := l.User.Load(user)
	if !ok {
		return nil, errors.New("user does not exist")
	}
	userPartition := value.(*Partition)

	// Get the nearby partitions and find all users
	partitions, err := userPartition.Nearby(radius)
	if err != nil {
		return nil, err
	}

	users := make([]string, 0)
	for _, partition := range *partitions {
		value, ok := l.Location.Load(partition.Encoded)
		if !ok {
			continue
		}
		partitionUsers := value.(map[string]*UserData)

		for usr, usrData := range partitionUsers {
			if usr != user && time.Now().Before(usrData.Timestamp.Add(timeout)) {
				users = append(users, usr)
			}
		}
	}

	return users, nil
}

// Merge changes of another list into the current
func (l *Location) merge(merge *Location) {
	seen := make(map[string]bool)
	temp := list.New()

	// Sync local changes with shared state
	for merge.EventStack.Len() > 0 {
		value := merge.EventStack.Front()
		event := value.Value.(*stackNode)
		merge.EventStack.Remove(value)

		if _, ok := seen[event.User]; ok {
			continue
		}

		if event.Event == eventDelete {
			l.remove(event.User)
			temp.PushFront(event)
		} else if event.Event == eventUpsert {
			mergeUserData, ok := merge.get(event.User)
			if !ok {
				panic("user does not exist")
			}

			userData, ok := l.get(event.User)
			if !ok || userData.Timestamp.Before(mergeUserData.Timestamp) && time.Now().Before(userData.Timestamp.Add(timeout)) {
				l.upsert(event.User, mergeUserData.Lat, mergeUserData.Long, mergeUserData.Timestamp)
				temp.PushFront(event)
			}
		} else {
			panic("invalid event type")
		}

		seen[event.User] = true
	}

	// Push temp stack elements back
	for temp.Len() > 0 {
		value := temp.Front()
		event := value.Value.(*stackNode)
		temp.Remove(value)

		merge.EventStack.PushFront(event)
	}
}

// Public method for merge which locks
func (l *Location) Merge(merge *Location) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	merge.mutex.Lock()
	defer merge.mutex.Unlock()

	l.merge(merge)
}

// Sync local changes
func (l *Location) Sync() error {
	key := helpers.FormatKey(stateKey, "lock")
	l.lock.Lock(key)
	defer l.lock.Unlock(key, false)

	l.mutex.Lock()
	defer l.mutex.Unlock()

	// **** Clearly it seems that the sync map does not serialize well

	data, err := l.redis.Get(l.ctx, stateKey).Result()
	if err == nil {
		staged := &Location{}
		json.Unmarshal([]byte(data), staged)

		l.merge(staged)
		staged.merge(l)

	} else if err != redis.Nil {
		return err
	}

	// Push changes locally to redis
	locationData, err := json.Marshal(l)
	if err != nil {
		return err
	}
	l.redis.Set(l.ctx, stateKey, locationData, 0)

	return nil
}
