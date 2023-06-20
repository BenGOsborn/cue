package utils

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

type ResourceLock struct {
	mutex map[string]*sync.RWMutex
}

// Create a new resource lock
func NewResourceLock() *ResourceLock {
	return &ResourceLock{mutex: make(map[string]*sync.RWMutex)}
}

// Lock the mutex for reading
func (r *ResourceLock) LockRead(id string) error {
	lock, ok := r.mutex[id]

	if !ok {
		lock = &sync.RWMutex{}
		r.mutex[id] = lock
	}

	lock.RLock()

	return nil
}

// Unlock the mutex for reading
func (r *ResourceLock) UnlockRead(id string) error {
	lock, ok := r.mutex[id]

	if !ok {
		return errors.New("no lock with this id exists")
	}

	lock.RUnlock()

	return nil
}

// Lock the mutex for writing
func (r *ResourceLock) LockWrite(id string) error {
	lock, ok := r.mutex[id]

	if !ok {
		lock = &sync.RWMutex{}
		r.mutex[id] = lock
	}

	lock.Lock()

	return nil
}

// Unlock the mutex for writing
func (r *ResourceLock) UnlockWrite(id string) error {
	lock, ok := r.mutex[id]

	if !ok {
		return errors.New("no lock with this id exists")
	}

	lock.Unlock()

	return nil
}

type ResourceLockDistributed struct {
	mutex  sync.Mutex
	client *redislock.Client
	lock   map[string]*redislock.Lock
	ctx    context.Context
	ttl    time.Duration
}

// Create a new distributed resource lock
func NewResourceLockDistributed(ctx context.Context, redisUrl string, ttl time.Duration) (*ResourceLockDistributed, error) {
	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}
	redisClient := redis.NewClient(opt)
	redisLockClient := redislock.New(redisClient)

	return &ResourceLockDistributed{ctx: ctx, client: redisLockClient, lock: make(map[string]*redislock.Lock), ttl: ttl}, nil
}

// Lock the mutex
func (r *ResourceLockDistributed) Lock(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	redisLock, err := r.client.Obtain(r.ctx, id, r.ttl, nil)
	if err != nil {
		return err
	}

	r.lock[id] = redisLock

	return nil
}

// Unlock the mutex
func (r *ResourceLockDistributed) Unlock(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	redisLock, ok := r.lock[id]
	if !ok {
		return errors.New("no lock with this id exists")
	}

	return redisLock.Release(r.ctx)
}
