package websocket

import (
	"sync"
	"time"

	gorillaws "github.com/gorilla/websocket"
	"github.com/masterfabric-go/masterfabric/internal/domain/realtime/model"
	realtimeService "github.com/masterfabric-go/masterfabric/internal/domain/realtime/service"
)

// client represents a single WebSocket connection.
type client struct {
	id     string
	info   realtimeService.ClientInfo
	conn   *gorillaws.Conn
	send   chan []byte
	rooms  map[model.RoomKey]struct{}
	hub    *Hub
	mu     sync.Mutex
	closed bool
}

// writePump sends messages from the send channel to the WebSocket connection.
func (c *client) writePump(pingInterval time.Duration) {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}
			if !ok {
				_ = c.conn.WriteMessage(gorillaws.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(gorillaws.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}
			if err := c.conn.WriteMessage(gorillaws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump reads messages from the WebSocket and dispatches client actions.
func (c *client) readPump(onMessage func(*client, []byte)) {
	defer func() {
		c.hub.removeClient(c)
	}()

	c.conn.SetReadLimit(4096)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		if onMessage != nil {
			onMessage(c, message)
		}
	}
}

func (c *client) subscribe(room model.RoomKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rooms == nil {
		c.rooms = make(map[model.RoomKey]struct{})
	}
	c.rooms[room] = struct{}{}
}

func (c *client) unsubscribe(room model.RoomKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.rooms, room)
}

func (c *client) isSubscribed(room model.RoomKey) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.rooms[room]
	return ok
}

func (c *client) closeSend() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.closed = true
		close(c.send)
	}
}
