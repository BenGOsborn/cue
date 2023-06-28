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
	User      string    `json:"user"`
	Lat       float32   `json:"lat"`
	Long      float32   `json:"long"`
	Timestamp time.Time `json:"timestamp"`
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
	ttl        time.Duration
}

const (
	statePrefix = "location:stage"
)

// Make a new location structure
func NewLocation(ctx context.Context, id string, ttl time.Duration, redis *redis.Client, lock *utils.ResourceLockDistributed) *Location {
	return &Location{ctx: ctx, Location: &sync.Map{}, User: &sync.Map{}, redis: redis, lock: lock, EventStack: list.New(), id: id, ttl: ttl}
}

// Add a new user
func (l *Location) upsert(user string, lat float32, long float32, timestamp time.Time) error {
	// Get the partition
	partition, err := NewPartitionFromCoords(lat, long)
	if err != nil {
		return err
	}

	// Remove the user from the previous partition if they already exist
	value, ok := l.User.Load(user)
	if ok {
		prev := value.(*Partition)

		value, ok := l.Location.Load(prev.Encoded)
		if ok {
			partitionUsers := value.(map[string]*UserData)

			if userData, ok := partitionUsers[user]; !ok || userData.Timestamp.After(timestamp) {
				return nil
			}

			delete(partitionUsers, user)
		}
	}

	// Add the user to the partition
	l.User.Store(user, partition)
	value, _ = l.Location.LoadOrStore(partition.Encoded, make(map[string]*UserData))
	partitionUsers := value.(map[string]*UserData)

	partitionUsers[user] = &UserData{User: user, Lat: lat, Long: long, Timestamp: timestamp}

	return nil
}

// Public method for upsert which locks
func (l *Location) Upsert(user string, lat float32, long float32) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	timestamp := time.Now()

	if err := l.upsert(user, lat, long, timestamp); err != nil {
		return err
	}

	l.EventStack.PushFront(&UserData{User: user, Timestamp: timestamp, Lat: lat, Long: long})

	return nil
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
			if usr != user && time.Now().Before(usrData.Timestamp.Add(l.ttl)) {
				users = append(users, usr)
			}
		}
	}

	return users, nil
}

// Get the data for a given user
func (l *Location) get(user string) (*UserData, error) {
	value, ok := l.User.Load(user)
	if !ok {
		return nil, errors.New("user does not exist")
	}
	partition := value.(*Partition)

	value, ok = l.Location.Load(partition.Encoded)
	if !ok {
		return nil, errors.New("user does not exist")
	}
	users := value.(map[string]*UserData)

	userData, ok := users[user]
	if !ok {
		panic("user exists but not in partition")
	}

	return userData, nil
}

// Public method to get the user with locks
func (l *Location) Get(user string) (*UserData, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.get(user)
}

// Merge changes between two lists
func (l *Location) merge(merge *Location) {
	seen := make(map[string]bool)
	temp := list.New()

	// Sync local changes with shared state
	for l.EventStack.Len() > 0 || merge.EventStack.Len() > 0 {
		// Choose the most recent event
		var event *UserData
		var value *list.Element
		var location *Location

		value1 := l.EventStack.Front()
		value2 := merge.EventStack.Front()

		if value1 == nil {
			event = value2.Value.(*UserData)
			value = value2
			location = merge
		} else if value2 == nil {
			event = value1.Value.(*UserData)
			value = value1
			location = l
		} else {
			event1 := value1.Value.(*UserData)
			event2 := value2.Value.(*UserData)

			if event1.Timestamp.After(event2.Timestamp) {
				event = event1
				value = value1
				location = l
			} else {
				event = event2
				value = value2
				location = merge
			}
		}

		// Ensure no duplicate processing
		location.EventStack.Remove(value)

		if _, ok := seen[event.User]; ok {
			continue
		}

		seen[event.User] = true

		// Add event to both locations
		if time.Now().Before(event.Timestamp.Add(l.ttl)) {
			l.upsert(event.User, event.Lat, event.Long, event.Timestamp)
			merge.upsert(event.User, event.Lat, event.Long, event.Timestamp)

			temp.PushFront(event)
		}
	}

	// Push temp stack elements back
	for temp.Len() > 0 {
		value := temp.Front()
		event := value.Value.(*UserData)
		temp.Remove(value)

		l.EventStack.PushFront(event)
		merge.EventStack.PushFront(&UserData{User: event.User, Timestamp: event.Timestamp, Lat: event.Lat, Long: event.Long})
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
	EventStack *[]*UserData                     `json:"eventStack"`
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
		EventStack: func() *[]*UserData {
			result := make([]*UserData, l.EventStack.Len())
			i := 0

			for e := l.EventStack.Front(); e != nil; e = e.Next() {
				result[i] = e.Value.(*UserData)
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
