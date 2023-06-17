package gateway

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

type Connections struct {
	connections map[string]*websocket.Conn
	mutex       map[string]*sync.RWMutex
}

// Create a new connections struct
func NewConnections() *Connections {
	return &Connections{connections: make(map[string]*websocket.Conn), mutex: make(map[string]*sync.RWMutex)}
}

// Lock the mutex for reading
func (connections *Connections) lockRead(id string) error {
	lock, ok := connections.mutex[id]

	if !ok {
		lock = &sync.RWMutex{}
		connections.mutex[id] = lock
	}

	lock.RLock()

	return nil
}

// Unlock the mutex for reading
func (connections *Connections) unlockRead(id string) error {
	lock, ok := connections.mutex[id]

	if !ok {
		return errors.New("No lock with this id exists.")
	}

	lock.RUnlock()

	return nil
}

// Lock the mutex for writing
func (connections *Connections) lockWrite(id string) error {
	lock, ok := connections.mutex[id]

	if !ok {
		lock = &sync.RWMutex{}
		connections.mutex[id] = lock
	}

	lock.Lock()

	return nil
}

// Unlock the mutex for writing
func (connections *Connections) unlockWrite(id string) error {
	lock, ok := connections.mutex[id]

	if !ok {
		return errors.New("No lock with this id exists.")
	}

	lock.Unlock()

	return nil
}

// Add a new connection
func (connections *Connections) Add(id string, conn *websocket.Conn) {
	connections.lockWrite(id)
	defer connections.unlockWrite(id)

	connections.connections[id] = conn
}

// Remove a connection
func (connections *Connections) Remove(id string) {
	connections.lockWrite(id)
	defer connections.unlockWrite(id)

	delete(connections.connections, id)
}

// Check if a connection exists
func (connections *Connections) Exists(id string) bool {
	connections.lockRead(id)
	defer connections.unlockRead(id)

	_, ok := connections.connections[id]

	return ok
}

// Apply a function to all connections
func (connections *Connections) ForEach(fn func(string, *websocket.Conn) error) error {
	for id, conn := range connections.connections {
		connections.lockRead(id)
		defer connections.unlockRead(id)

		if err := fn(id, conn); err != nil {
			return err
		}
	}

	return nil
}

// Apply a function to a particular connection
func (connections *Connections) Apply(id string, fn func(string, *websocket.Conn) error) (bool, error) {
	connections.lockRead(id)
	defer connections.unlockRead(id)

	conn, ok := connections.connections[id]

	if !ok {
		return false, nil
	}

	err := fn(id, conn)

	return true, err
}
