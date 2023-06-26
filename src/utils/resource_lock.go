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
	mutex sync.Map
}

// Create a new resource lock
func NewResourceLock() *ResourceLock {
	return &ResourceLock{mutex: sync.Map{}}
}

// Lock the mutex for reading
func (r *ResourceLock) LockRead(id string) {
	lock, _ := r.mutex.LoadOrStore(id, &sync.RWMutex{})
	lock.(*sync.RWMutex).RLock()
}

// Unlock the mutex for reading
func (r *ResourceLock) UnlockRead(id string) error {
	value, ok := r.mutex.Load(id)
	if !ok {
		return errors.New("no lock with this id exists")
	}
	value.(*sync.RWMutex).RUnlock()

	return nil
}

// Lock the mutex for writing
func (r *ResourceLock) LockWrite(id string) {
	lock, _ := r.mutex.LoadOrStore(id, &sync.RWMutex{})
	lock.(*sync.RWMutex).Lock()
}

// Unlock the mutex for writing
func (r *ResourceLock) UnlockWrite(id string) error {
	value, ok := r.mutex.Load(id)
	if !ok {
		return errors.New("no lock with this id exists")
	}
	value.(*sync.RWMutex).Unlock()

	return nil
}

type ResourceLockDistributed struct {
	redisClient     *redis.Client
	redisLockClient *redislock.Client
	lock            sync.Map
	ctx             context.Context
	cond            sync.Map
	ttl             time.Duration
	expiry          chan string
}

const (
	resourcePrefix      = "resource-lock:resource"
	resourceLockChannel = "resource-lock:channel"
)

// Create a new distributed resource lock
func NewResourceLockDistributed(ctx context.Context, redis *redis.Client, ttl time.Duration) (*ResourceLockDistributed, error) {
	redisLockClient := redislock.New(redis)

	r := &ResourceLockDistributed{ctx: ctx, redisClient: redis, redisLockClient: redisLockClient, lock: sync.Map{}, ttl: ttl, cond: sync.Map{}, expiry: make(chan string)}

	go func() {
		pubsub := redis.Subscribe(ctx, resourceLockChannel)
		ch := pubsub.Channel()

		for {
			select {
			case lockId := <-ch:
				value, ok := r.cond.Load(lockId.Payload)
				if !ok {
					continue
				}
				cond := value.(*sync.Cond)

				cond.L.Lock()
				cond.Broadcast()
				cond.L.Unlock()
			case lockId := <-r.expiry:
				go func() {
					time.Sleep(ttl)

					value, ok := r.cond.Load(lockId)
					if !ok {
						return
					}
					cond := value.(*sync.Cond)

					cond.L.Lock()
					cond.Broadcast()
					cond.L.Unlock()
				}()
			}
		}
	}()

	return r, nil
}

// Lock the resource
func (r *ResourceLockDistributed) Lock(id string) {
	value, _ := r.cond.LoadOrStore(id, sync.NewCond(&sync.Mutex{}))
	cond := value.(*sync.Cond)

	cond.L.Lock()
	defer cond.L.Unlock()

	for {
		redisLock, err := r.redisLockClient.Obtain(r.ctx, id, r.ttl, nil)

		if err != nil {
			r.expiry <- id
			cond.Wait()
			continue
		}

		r.lock.Store(id, redisLock)

		return
	}
}

// Unlock the resource and declare if it has been processed
func (r *ResourceLockDistributed) Unlock(id string, processed bool) error {
	value, _ := r.cond.LoadOrStore(id, sync.NewCond(&sync.Mutex{}))
	cond := value.(*sync.Cond)

	cond.L.Lock()
	defer cond.L.Unlock()

	value, ok := r.lock.Load(id)
	if !ok {
		return errors.New("no lock with this id exists")
	}
	redisLock := value.(*redislock.Lock)

	if processed {
		if err := r.redisClient.Set(r.ctx, helpers.FormatKey(resourcePrefix, id), "TRUE", r.ttl).Err(); err != nil {
			return err
		}
	}

	if err := redisLock.Release(r.ctx); err != nil {
		return err
	}
	r.redisClient.Publish(r.ctx, resourceLockChannel, id)

	r.cond.Delete(id)
	r.lock.Delete(id)

	return nil
}

// Return whether a resource has been processed
func (r *ResourceLockDistributed) IsProcessed(id string) (bool, error) {
	result, err := r.redisClient.Exists(r.ctx, helpers.FormatKey(resourcePrefix, id)).Result()
	if err != nil {
		return false, nil
	}

	return result == 1, nil
}
