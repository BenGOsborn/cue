package gateway

import (
	"sync"

	"github.com/bengosborn/cue/utils"
	"github.com/gorilla/websocket"
)

type Connections struct {
	connections sync.Map
	lock1       *utils.ResourceLock
	lock2       *utils.ResourceLock
}

// Create a new connections struct
func NewConnections() *Connections {
	return &Connections{connections: sync.Map{}, lock1: utils.NewResourceLock(), lock2: utils.NewResourceLock()}
}

// Close all connections
func (c *Connections) Close() {
	c.connections.Range(func(key, value interface{}) bool {
		id := key.(string)

		c.lock1.LockWrite(id)
		defer c.lock1.UnlockWrite(id)

		c.lock2.LockWrite(id)
		defer c.lock2.UnlockWrite(id)

		conn := value.(*websocket.Conn)

		conn.Close()

		return true
	})
}

// Add a new connection
func (c *Connections) Add(id string, conn *websocket.Conn) {
	c.lock1.LockWrite(id)
	defer c.lock1.UnlockWrite(id)

	c.lock2.LockWrite(id)
	defer c.lock2.UnlockWrite(id)

	c.connections.Store(id, conn)
}

// Remove a connection
func (c *Connections) Remove(id string) {
	c.lock1.LockWrite(id)
	defer c.lock1.UnlockWrite(id)

	c.lock2.LockWrite(id)
	defer c.lock2.UnlockWrite(id)

	value, ok := c.connections.Load(id)
	if !ok {
		return
	}

	conn := value.(*websocket.Conn)
	conn.Close()

	c.connections.Delete(id)
}

func (c *Connections) apply(id string, lock *utils.ResourceLock, fn func(string, *websocket.Conn) error) (bool, error) {
	value, ok := c.connections.Load(id)
	if !ok {
		return false, nil
	}

	conn := value.(*websocket.Conn)

	lock.LockWrite(id)
	defer lock.UnlockWrite(id)

	if err := fn(id, conn); err != nil {
		return false, err
	}

	return true, nil
}

// Apply a function to a particular connection for reads
func (c *Connections) RApply(id string, fn func(string, *websocket.Conn) error) (bool, error) {
	return c.apply(id, c.lock1, fn)
}

// Apply a function to a particular connection for writes
func (c *Connections) WApply(id string, fn func(string, *websocket.Conn) error) (bool, error) {
	return c.apply(id, c.lock2, fn)
}
