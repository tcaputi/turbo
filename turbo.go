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

type Turbo struct {
}

func (t *Turbo) Handler(res http.ResponseWriter, req *http.Request) {
	ws, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println("Could not upgrade incoming connection", err)
		return
	}
	conn := &connection{
		id:     connectionId(),
		outbox: make(chan []byte, 256),
		ws:     ws,
		subs:   make(map[*SubscriberSet]bool),
	}
	msgHub.registration <- conn
	defer conn.kill() // Kill the connection on exit
	go conn.writer()
	conn.reader() // Left outside go routine to stall execution
}

// TODO this should take config
func New() (error, *Turbo) {
	go msgHub.run()
	return nil, &Turbo{}
}

func connectionId() uint64 {
	var newId uint64

	connectionIdMutex.Lock()
	connectionIdCounter += 1
	newId = connectionIdCounter
	connectionIdMutex.Unlock()

	return newId
}
