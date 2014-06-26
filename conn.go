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
	subs map[*SubscriberSet]bool
}

func (conn *connection) reader() {
	for {
		_, message, err := conn.ws.ReadMessage()

		if err != nil {
			break
		}

		msgHub.route(&RawMsg{
			Conn:    conn,
			Payload: message,
		})
	}
	conn.ws.Close()
}

func (conn *connection) writer() {
	for message := range conn.outbox {
		err := conn.ws.WriteMessage(websocket.TextMessage, message)

		if err != nil {
			break
		}
	}
	conn.ws.Close()
}

func (conn *connection) kill() {
	msgHub.unregistration <- conn
}
