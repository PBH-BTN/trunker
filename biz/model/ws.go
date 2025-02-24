package model

import (
	"sync"

	"github.com/hertz-contrib/websocket"
)

type Conn struct {
	conn          *websocket.Conn
	send          chan []byte
	m             sync.Mutex
	CloseCallback func()
}

func NewConn(conn *websocket.Conn) *Conn {
	return &Conn{
		conn,
		make(chan []byte),
		sync.Mutex{},
		nil,
	}
}

func (c *Conn) WriteJSON(v any) error {
	c.m.Lock()
	defer c.m.Unlock()
	return c.conn.WriteJSON(v)
}

func (c *Conn) WriteString(s []byte) error {
	c.m.Lock()
	defer c.m.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, s)
}

func (c *Conn) Close() error {
	c.m.Lock()
	defer c.m.Unlock()
	if c.CloseCallback != nil {
		c.CloseCallback()
	}
	return c.conn.Close()
}
