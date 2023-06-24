package utils

import (
	"errors"
	"sync"
	"time"
)

type UserData struct {
	lat       float32
	long      float32
	timestamp time.Time
}

type Location struct {
	lock     sync.RWMutex
	location map[string]map[string]*UserData // partition encoded -> user -> user data
	user     map[string]*Partition           // user -> partition
}

// Make a new location structure
func NewLocation() *Location {
	return &Location{location: make(map[string]map[string]*UserData), user: make(map[string]*Partition), lock: sync.RWMutex{}}
}

// Add a new user
func (u *Location) Upsert(user string, lat float32, long float32) error {
	u.lock.Lock()
	defer u.lock.Unlock()

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

// Remove a user
func (u *Location) Remove(user string) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	// Remove user
	prev, ok := u.user[user]
	if !ok {
		return errors.New("user does not exist")
	}

	delete(u.user, user)
	delete(u.location[prev.encoded], user)

	return nil
}

// Get nearby users
func (u *Location) Nearby(user string, radius int) ([]string, error) {
	u.lock.RLock()
	defer u.lock.RUnlock()

	// Get the partition for the user
	userPartition, ok := u.user[user]
	if !ok {
		return nil, errors.New("user does not exist")
	}

	// Get the nearby partitions and find all users
	partitions, err := userPartition.Nearby(radius)
	if err != nil {
		return nil, err
	}

	users := make([]string, 0)
	for _, partition := range *partitions {
		for usr := range u.location[partition.encoded] {
			if usr != user {
				users = append(users, usr)
			}
		}
	}

	return users, nil
}
