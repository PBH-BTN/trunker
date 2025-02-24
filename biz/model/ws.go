package model

import (
	"sync"

	"github.com/hertz-contrib/websocket"
)

type Conn struct {
	conn *websocket.Conn
	send chan []byte
	m    sync.Mutex
}

func NewConn(conn *websocket.Conn) *Conn {
	return &Conn{
		conn,
		make(chan []byte),
		sync.Mutex{},
	}
}

func (c *Conn) WriteJSON(v any) error {
	c.m.Lock()
	defer c.m.Unlock()
	return c.conn.WriteJSON(v)
}

func (c *Conn) Close() error {
	c.m.Lock()
	defer c.m.Unlock()
	return c.conn.Close()
}
