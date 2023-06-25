package utils

import (
	"context"
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
func NewLocation(ctx context.Context, redis *redis.Client, lock *utils.ResourceLockDistributed) *Location {
	return &Location{location: sync.Map{}, user: sync.Map{}, mutex: sync.RWMutex{}, redis: redis, lock: lock}
}

// Add a new user
func (u *Location) Upsert(user string, lat float32, long float32) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// **** We do know that the partition will indeed be a single map -> do something with this information

	// Remove the user from the previous partition if they already exist
	prev, ok := u.user[user]
	if ok {
		delete(u.location[prev.encoded], user)
	}

	// Add the user to the partition
	partition, err := NewPartitionFromCoords(lat, long)
	if err != nil {
		return err
	}

	u.user[user] = partition
	if _, ok := u.location[partition.encoded]; !ok {
		u.location[partition.encoded] = make(map[string]*UserData)
	}
	u.location[partition.encoded][user] = &UserData{lat: lat, long: long, timestamp: time.Now()}

	return nil
}

// // Remove a user
// func (u *Location) Remove(user string) error {
// 	u.mutex.Lock()
// 	defer u.mutex.Unlock()

// 	// Remove user
// 	prev, ok := u.user[user]
// 	if !ok {
// 		return errors.New("user does not exist")
// 	}

// 	delete(u.user, user)
// 	delete(u.location[prev.encoded], user)

// 	return nil
// }

// // Get nearby users
// func (u *Location) Nearby(user string, radius int) ([]string, error) {
// 	u.mutex.RLock()
// 	defer u.mutex.RUnlock()

// 	// Get the partition for the user
// 	userPartition, ok := u.user[user]
// 	if !ok {
// 		return nil, errors.New("user does not exist")
// 	}

// 	// Get the nearby partitions and find all users
// 	partitions, err := userPartition.Nearby(radius)
// 	if err != nil {
// 		return nil, err
// 	}

// 	users := make([]string, 0)
// 	for _, partition := range *partitions {
// 		for usr := range u.location[partition.encoded] {
// 			if usr != user {
// 				users = append(users, usr)
// 			}
// 		}
// 	}

// 	return users, nil
// }
