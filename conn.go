package turbo

import (
	"github.com/gorilla/websocket"
)

type connection struct {
	// The id of this connection
	id uint64
	// The websocket connection.
	ws *websocket.Conn
	// Buffered channel of outbound messages.
	outbox chan []byte
	// Event subscriptions
	subs map[*SubSet]bool
}

func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		msgHub.inbox <- &RawMsg{Conn: c, Payload: message}
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.outbox {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

func (c *connection) kill() {
	msgHub.unregistration <- c
}
