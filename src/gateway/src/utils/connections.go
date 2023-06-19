package gateway

import (
	"github.com/bengosborn/cue/utils"
	"github.com/gorilla/websocket"
)

type Connections struct {
	connections map[string]*websocket.Conn
	lock        *utils.ResourceLock
}

// Create a new connections struct
func NewConnections() *Connections {
	return &Connections{connections: make(map[string]*websocket.Conn), lock: utils.NewResourceLock()}
}

// Close all connections
func (c *Connections) Close() {
	for id, conn := range c.connections {
		c.lock.LockWrite(id)
		defer c.lock.UnlockWrite(id)

		conn.Close()
	}
}

// Add a new connection
func (c *Connections) Add(id string, conn *websocket.Conn) {
	c.lock.LockWrite(id)
	defer c.lock.UnlockWrite(id)

	c.connections[id] = conn
}

// Remove a connection
func (c *Connections) Remove(id string) {
	c.lock.LockWrite(id)
	defer c.lock.UnlockWrite(id)

	c.connections[id].Close()
	delete(c.connections, id)
}

// Check if a connection exists
func (c *Connections) Exists(id string) bool {
	c.lock.LockRead(id)
	defer c.lock.UnlockRead(id)

	_, ok := c.connections[id]

	return ok
}

// Apply a function to all connections
func (c *Connections) ForEach(fn func(string, *websocket.Conn) error) error {
	for id, conn := range c.connections {
		c.lock.LockRead(id)
		defer c.lock.UnlockRead(id)

		if err := fn(id, conn); err != nil {
			return err
		}
	}

	return nil
}

// Apply a function to a particular connection
func (c *Connections) Apply(id string, fn func(string, *websocket.Conn) error) (bool, error) {
	c.lock.LockRead(id)
	defer c.lock.UnlockRead(id)

	conn, ok := c.connections[id]

	if !ok {
		return false, nil
	}

	if err := fn(id, conn); err != nil {
		return false, err
	}

	return true, nil
}
