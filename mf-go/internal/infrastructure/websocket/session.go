package websocket

import (
	"time"

	gorillaws "github.com/gorilla/websocket"
)

// Session manages the lifecycle of a single WebSocket connection.
type Session struct {
	ID       string
	Conn     *gorillaws.Conn
	Send     chan []byte
	Hub      *Hub
	OnAction func(clientID string, message []byte)
}

// Start begins read and write pumps for the session.
func (s *Session) Start(pingInterval time.Duration, unregister func()) {
	c := &client{
		id:   s.ID,
		conn: s.Conn,
		send: s.Send,
		hub:  s.Hub,
	}

	go c.writePump(pingInterval)
	go func() {
		defer unregister()
		c.readPump(func(cl *client, msg []byte) {
			if s.OnAction != nil {
				s.OnAction(cl.id, msg)
			}
		})
	}()
}
