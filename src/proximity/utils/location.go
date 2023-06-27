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
	id         string
	ctx        context.Context
	redis      *redis.Client
	lock       *utils.ResourceLockDistributed
	mutex      sync.RWMutex
	Location   *sync.Map
	User       *sync.Map
	EventStack *list.List
}

const (
	timeout     = 5 * time.Minute
	statePrefix = "location:stage"
)

// Make a new location structure
func NewLocation(ctx context.Context, redis *redis.Client, lock *utils.ResourceLockDistributed, id string) *Location {
	return &Location{ctx: ctx, Location: &sync.Map{}, User: &sync.Map{}, redis: redis, lock: lock, EventStack: list.New(), id: id}
}

// **** We do need the events to be a part of this, which means we cannot use the remove function within this

// Add a new user
func (l *Location) upsert(user string, lat float32, long float32, timestamp time.Time) error {
	// Get the partition
	partition, err := NewPartitionFromCoords(lat, long)
	if err != nil {
		return err
	}

	// Remove the user from the previous partition if they already exist
	l.User.Delete(user)

	value, ok := l.User.Load(user)
	if ok {
		prev := value.(*Partition)

		value, ok := l.Location.Load(prev.Encoded)
		if ok {
			partitionUsers := value.(map[string]*UserData)
			delete(partitionUsers, user)
		}
	}

	// Add the user to the partition
	l.User.Store(user, partition)
	value, _ = l.Location.LoadOrStore(partition.Encoded, make(map[string]*UserData))
	partitionUsers := value.(map[string]*UserData)

	partitionUsers[user] = &UserData{Lat: lat, Long: long, Timestamp: timestamp}

	l.EventStack.PushFront(&stackNode{Event: eventUpsert, User: user})

	return nil
}

// Public method for upsert which locks
func (l *Location) Upsert(user string, lat float32, long float32) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := l.upsert(user, lat, long, time.Now()); err != nil {
		return err
	}

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

	l.EventStack.PushFront(&stackNode{Event: eventDelete, User: user})

	return false
}

// Public method for remove which records event and locks
func (l *Location) Remove(user string) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.remove(user)
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
	stateKey := helpers.FormatKey(statePrefix, l.id)

	l.lock.Lock(l.id)
	defer l.lock.Unlock(l.id, true)

	l.mutex.Lock()
	defer l.mutex.Unlock()

	data, err := l.redis.Get(l.ctx, stateKey).Result()
	if err == nil {
		staged := &Location{}
		if err := json.Unmarshal([]byte(data), staged); err != nil {
			return err
		}

		staged.merge(l)
		l.merge(staged)

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

type temp struct {
	Location   *map[string]map[string]*UserData `json:"location"`
	User       *map[string]*Partition           `json:"user"`
	EventStack *[]*stackNode                    `json:"eventStack"`
}

func (l *Location) MarshalJSON() ([]byte, error) {
	return json.Marshal(&temp{
		Location: func() *map[string]map[string]*UserData {
			m := make(map[string]map[string]*UserData)

			l.Location.Range(func(key, value interface{}) bool {
				location := key.(string)

				m[location] = value.(map[string]*UserData)
				return true
			})

			return &m
		}(),
		User: func() *map[string]*Partition {
			m := make(map[string]*Partition)

			l.User.Range(func(key, value interface{}) bool {
				user := key.(string)
				partition := value.(*Partition)

				m[user] = partition

				return true
			})

			return &m
		}(),
		EventStack: func() *[]*stackNode {
			result := make([]*stackNode, l.EventStack.Len())
			i := 0

			for e := l.EventStack.Front(); e != nil; e = e.Next() {
				result[i] = e.Value.(*stackNode)
				i += 1
			}

			return &result
		}(),
	})
}

func (l *Location) UnmarshalJSON(data []byte) error {
	tmp := &temp{}
	if err := json.Unmarshal(data, tmp); err != nil {
		return err
	}

	// Update the event stack
	eventStack := list.New()
	for _, value := range *tmp.EventStack {
		eventStack.PushBack(value)
	}
	l.EventStack = eventStack

	// Update the user
	userSyncMap := &sync.Map{}
	for key, value := range *tmp.User {
		userSyncMap.Store(key, value)
	}
	l.User = userSyncMap

	// Update the location
	locationSyncMap := &sync.Map{}
	for key, value := range *tmp.Location {
		locationSyncMap.Store(key, value)
	}
	l.Location = locationSyncMap

	return nil
}
