package utils

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/bengosborn/cue/utils"
	"github.com/redis/go-redis/v9"
)

type UserData struct {
	Lat       float32   `json:"lat"`
	Long      float32   `json:"long"`
	Timestamp time.Time `json:"timestamp"`
}

type Location struct {
	ctx      context.Context                `json:"-"`
	redis    *redis.Client                  `json:"-"`
	lock     *utils.ResourceLockDistributed `json:"-"`
	mutex    sync.RWMutex                   `json:"-"`
	Location sync.Map                       `json:"location"` // partition encoded -> user -> user data
	User     sync.Map                       `json:"user"`     // user -> partition
}

const (
	timeout  = 5 * time.Minute
	stateKey = "location:stage"
)

// **** I need to use this with a distributed lock and redis...
// 1. When reading, we need to look at redis values and the values we have locally and choose the one with the most recent timestamp
// 2. When writing, we will have a method to sync the local data with what is in the redis database, where it will only update the partition with the latest timestamp

// Make a new location structure
// func NewLocation(ctx context.Context, redis *redis.Client, lock *utils.ResourceLockDistributed) *Location {
func NewLocation() *Location {
	// return &Location{location: sync.Map{}, user: sync.Map{}, mutex: sync.RWMutex{}, redis: redis, lock: lock}
	return &Location{}
}

// Add a new user
func (l *Location) Upsert(user string, lat float32, long float32) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Remove the user from the previous partition if they already exist
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
	partition, err := NewPartitionFromCoords(lat, long)
	if err != nil {
		return err
	}

	l.User.Store(user, partition)
	value, _ = l.Location.LoadOrStore(partition.Encoded, make(map[string]*UserData))
	partitionUsers := value.(map[string]*UserData)

	partitionUsers[user] = &UserData{Lat: lat, Long: long, Timestamp: time.Now()}

	return nil
}

// Remove a user
func (l *Location) Remove(user string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Remove user
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

// Sync local changes
func (l *Location) Sync() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.lock.Lock(stateKey)
	defer l.lock.Unlock(stateKey, false)

	// Pull data from redis into the local changes
	data, err := l.redis.Get(l.ctx, stateKey).Result()
	if err == nil {
		staged := Location{}
		json.Unmarshal([]byte(data), &staged)

		// Compare the states and update the local state according to timestamps

		// **** This includes looking at the existing partition and user data (what users are in and which users are out)

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
