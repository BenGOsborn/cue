package utils

import (
	"errors"
	"sync"
	"time"

	"github.com/bengosborn/cue/utils"
	"github.com/redis/go-redis/v9"
)

type UserData struct {
	lat       float32
	long      float32
	timestamp time.Time
}

type Location struct {
	redis    *redis.Client
	lock     *utils.ResourceLockDistributed
	mutex    sync.RWMutex
	location sync.Map // partition encoded -> user -> user data
	user     sync.Map // user -> partition
}

const (
	timeout = 5 * time.Minute
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
func (u *Location) Upsert(user string, lat float32, long float32) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// Remove the user from the previous partition if they already exist
	value, ok := u.user.Load(user)
	if ok {
		prev := value.(*Partition)

		value, ok := u.location.Load(prev.encoded)
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

	u.user.Store(user, partition)
	value, _ = u.location.LoadOrStore(partition.encoded, make(map[string]*UserData))
	partitionUsers := value.(map[string]*UserData)

	partitionUsers[user] = &UserData{lat: lat, long: long, timestamp: time.Now()}

	return nil
}

// Remove a user
func (u *Location) Remove(user string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// Remove user
	u.user.Delete(user)

	value, ok := u.user.Load(user)
	if ok {
		prev := value.(*Partition)

		value, ok := u.location.Load(prev.encoded)
		if ok {
			partitionUsers := value.(map[string]*UserData)

			delete(partitionUsers, user)
		}
	}
}

// Get nearby users
func (u *Location) Nearby(user string, radius int) ([]string, error) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	// Get the partition for the user
	value, ok := u.user.Load(user)
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
		value, ok := u.location.Load(partition.encoded)
		if !ok {
			continue
		}
		partitionUsers := value.(map[string]*UserData)

		for usr, usrData := range partitionUsers {
			if usr != user && time.Now().Before(usrData.timestamp.Add(timeout)) {
				users = append(users, usr)
			}
		}
	}

	return users, nil
}
