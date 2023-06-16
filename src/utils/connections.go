package utils

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Connections struct {
	connections map[string]*websocket.Conn
	mutex       sync.Mutex
}

// Create a new connections struct
func NewConnections() *Connections {
	return &Connections{connections: make(map[string]*websocket.Conn)}
}

// Add a new connection
func (connections *Connections) Add(id string, conn *websocket.Conn) {
	connections.mutex.Lock()
	defer connections.mutex.Unlock()

	connections.connections[id] = conn
}

// Remove a connection
func (connections *Connections) Remove(id string) {
	connections.mutex.Lock()
	defer connections.mutex.Unlock()

	delete(connections.connections, id)
}

// Check if a connection exists
func (connections *Connections) Exists(id string) bool {
	connections.mutex.Lock()
	defer connections.mutex.Unlock()

	_, ok := connections.connections[id]

	return ok
}

// Apply a function to all connections
func (connections *Connections) ForEach(fn func(string, *websocket.Conn) error) error {
	connections.mutex.Lock()
	defer connections.mutex.Unlock()

	for id, conn := range connections.connections {
		if err := fn(id, conn); err != nil {
			return err
		}
	}

	return nil
}

// Apply a function to a particular connection
func (connections *Connections) Apply(id string, fn func(string, *websocket.Conn) error) (bool, error) {
	connections.mutex.Lock()
	defer connections.mutex.Unlock()

	conn, ok := connections.connections[id]

	if !ok {
		return false, nil
	}

	err := fn(id, conn)

	return true, err
}
