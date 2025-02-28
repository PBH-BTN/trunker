package model

import (
	"sync"
	"sync/atomic"

	"github.com/hertz-contrib/websocket"
)

type Conn struct {
	conn          *websocket.Conn
	m             *sync.Mutex
	CloseCallback func()
	closed        *atomic.Bool
}

func NewConn(conn *websocket.Conn) *Conn {
	closed := &atomic.Bool{}
	closed.Store(false)
	return &Conn{
		conn,
		&sync.Mutex{},
		nil,
		closed,
	}
}

func (c *Conn) WriteJSON(v any) error {
	if c.closed.Load() {
		return nil
	}
	c.m.Lock()
	defer c.m.Unlock()
	return c.conn.WriteJSON(v)
}

func (c *Conn) WriteString(s []byte) error {
	if c.closed.Load() {
		return nil
	}
	c.m.Lock()
	defer c.m.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, s)
}

func (c *Conn) Close() error {
	if c.closed.Load() {
		return nil
	}
	c.m.Lock()
	defer c.m.Unlock()
	if c.CloseCallback != nil {
		c.CloseCallback()
	}
	c.closed.Store(true)
	return c.conn.Close()
}
