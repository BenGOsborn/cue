package utils

import (
	"errors"
	"sync"
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
