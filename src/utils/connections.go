package utils

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

type Connections struct {
	connections map[string]*websocket.Conn
	mutex       map[string]*sync.Mutex
}

// Create a new connections struct
func NewConnections() *Connections {
	return &Connections{connections: make(map[string]*websocket.Conn), mutex: make(map[string]*sync.Mutex)}
}

// Lock the mutex
func (connections *Connections) lock(id string) error {
	lock, ok := connections.mutex[id]

	if !ok {
		return errors.New("No lock with this id exists.")
	}

	lock.Lock()

	return nil
}

// Unlock the mutex
func (connections *Connections) unlock(id string) error {
	lock, ok := connections.mutex[id]

	if !ok {
		return errors.New("No lock with this id exists.")
	}

	lock.Unlock()

	return nil
}

// Add a new connection
func (connections *Connections) Add(id string, conn *websocket.Conn) {
	connections.lock(id)
	defer connections.unlock(id)

	connections.connections[id] = conn
}

// Remove a connection
func (connections *Connections) Remove(id string) {
	connections.lock(id)
	defer connections.unlock(id)

	delete(connections.connections, id)
}

// Check if a connection exists
func (connections *Connections) Exists(id string) bool {
	connections.lock(id)
	defer connections.unlock(id)

	_, ok := connections.connections[id]

	return ok
}

// Apply a function to all connections
func (connections *Connections) ForEach(fn func(string, *websocket.Conn) error) error {
	for id, conn := range connections.connections {
		connections.lock(id)
		defer connections.unlock(id)

		if err := fn(id, conn); err != nil {
			return err
		}
	}

	return nil
}

// Apply a function to a particular connection
func (connections *Connections) Apply(id string, fn func(string, *websocket.Conn) error) (bool, error) {
	connections.lock(id)
	defer connections.unlock(id)

	conn, ok := connections.connections[id]

	if !ok {
		return false, nil
	}

	err := fn(id, conn)

	return true, err
}
