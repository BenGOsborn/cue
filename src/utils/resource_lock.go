package utils

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bengosborn/cue/helpers"
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
	mutex           sync.Mutex
	redisClient     *redis.Client
	redisLockClient *redislock.Client
	lock            map[string]*redislock.Lock
	ctx             context.Context
	ttl             time.Duration
}

const (
	resourcePrefix = "resource"
)

// Create a new distributed resource lock
func NewResourceLockDistributed(ctx context.Context, redis *redis.Client, ttl time.Duration) (*ResourceLockDistributed, error) {
	redisLockClient := redislock.New(redis)

	return &ResourceLockDistributed{ctx: ctx, redisClient: redis, redisLockClient: redisLockClient, lock: make(map[string]*redislock.Lock), ttl: ttl}, nil
}

// Lock the resource
func (r *ResourceLockDistributed) Lock(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	redisLock, err := r.redisLockClient.Obtain(r.ctx, id, r.ttl, nil)
	if err != nil {
		return err
	}

	r.lock[id] = redisLock

	return nil
}

// Unlock the resource and declare if it has been processed
func (r *ResourceLockDistributed) Unlock(id string, processed bool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	redisLock, ok := r.lock[id]
	if !ok {
		return errors.New("no lock with this id exists")
	}

	if processed {
		if err := r.redisClient.Set(r.ctx, helpers.FormatKey(resourcePrefix, id), "TRUE", r.ttl).Err(); err != nil {
			return err
		}
	}

	return redisLock.Release(r.ctx)
}

// Return whether a resource has been processed
func (r *ResourceLockDistributed) IsProcessed(id string) (bool, error) {
	result, err := r.redisClient.Exists(r.ctx, helpers.FormatKey(resourcePrefix, id)).Result()
	if err != nil {
		return false, nil
	}

	return result == 1, nil
}
