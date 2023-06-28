package gateway

import (
	"sync"

	"github.com/bengosborn/cue/utils"
	"github.com/gorilla/websocket"
)

type Connections struct {
	connections sync.Map
	lock        *utils.ResourceLock
}

// Create a new connections struct
func NewConnections() *Connections {
	return &Connections{connections: sync.Map{}, lock: utils.NewResourceLock()}
}

// Close all connections
func (c *Connections) Close() {
	c.connections.Range(func(key, value interface{}) bool {
		id := key.(string)

		c.lock.LockWrite(id)
		defer c.lock.UnlockWrite(id)

		conn := value.(*websocket.Conn)

		conn.Close()

		return true
	})
}

// Add a new connection
func (c *Connections) Add(id string, conn *websocket.Conn) {
	c.lock.LockWrite(id)
	defer c.lock.UnlockWrite(id)

	c.connections.Store(id, conn)
}

// Remove a connection
func (c *Connections) Remove(id string) {
	c.lock.LockWrite(id)
	defer c.lock.UnlockWrite(id)

	value, ok := c.connections.Load(id)
	if !ok {
		return
	}

	conn := value.(*websocket.Conn)
	conn.Close()

	c.connections.Delete(id)
}

// TODO This would be better to rewrite as a safe wrapper passing the message or returning a value which gets sent by the websocket without exposing the socket

// Apply a function to a particular connection - fn must be read only
func (c *Connections) RApply(id string, fn func(string, *websocket.Conn) error) (bool, error) {
	value, ok := c.connections.Load(id)
	if !ok {
		return false, nil
	}

	conn := value.(*websocket.Conn)

	c.lock.LockRead(id)
	defer c.lock.UnlockRead(id)

	if err := fn(id, conn); err != nil {
		return false, err
	}

	return true, nil
}

// Apply a function to a particular connection
func (c *Connections) Apply(id string, fn func(string, *websocket.Conn) error) (bool, error) {
	value, ok := c.connections.Load(id)
	if !ok {
		return false, nil
	}

	conn := value.(*websocket.Conn)

	c.lock.LockWrite(id)
	defer c.lock.UnlockWrite(id)

	if err := fn(id, conn); err != nil {
		return false, err
	}

	return true, nil
}
