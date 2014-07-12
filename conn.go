package turbo

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

const (
	UPGRADER_READ_BUF_SIZE  = 1024
	UPGRADER_WRITE_BUF_SIZE = 1024
)

var (
	connectionIdCounter uint64
	connectionIdMutex   = &sync.Mutex{}
	upgrader            = &websocket.Upgrader{
		ReadBufferSize:  UPGRADER_READ_BUF_SIZE,
		WriteBufferSize: UPGRADER_WRITE_BUF_SIZE,
	}
)

type Conn struct {
	// The id of this Conn
	id uint64
	// The websocket Conn.
	ws *websocket.Conn
	// Buffered channel of outbound messages.
	outbox chan []byte
	// Event subscriptions
	subscriptions map[*map[*Conn]bool]bool
	// Hub reference
	hub *MsgHub
}

func NewConn(hub *MsgHub, res http.ResponseWriter, req *http.Request) (*Conn, error) {
	ws, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println("Could not upgrade incoming Conn", err)
		return nil, err
	}
	conn := Conn{
		id:            newConnId(),
		outbox:        make(chan []byte, 256),
		ws:            ws,
		subscriptions: make(map[*map[*Conn]bool]bool),
		hub:           hub,
	}
	return &conn, nil
}

func (conn *Conn) reader() {
	for {
		_, message, err := conn.ws.ReadMessage()

		if err != nil {
			break
		}

		conn.hub.route(&RawMsg{
			Conn:    conn,
			Payload: message,
		})
	}
	conn.ws.Close()
}

func (conn *Conn) writer() {
	for message := range conn.outbox {
		err := conn.ws.WriteMessage(websocket.TextMessage, message)

		if err != nil {
			break
		}
	}
	conn.ws.Close()
}

func newConnId() uint64 {
	var newId uint64

	connectionIdMutex.Lock()
	connectionIdCounter += 1
	newId = connectionIdCounter
	connectionIdMutex.Unlock()

	return newId
}
