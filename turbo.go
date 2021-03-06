package turbo

import (
	"log"
	"net/http"
)

type Turbo struct {
	bus *MsgBus
	hub *MsgHub
}

func (t *Turbo) Handler(res http.ResponseWriter, req *http.Request) {
	conn, err := NewConn(t.hub, res, req)
	if err != nil {
		log.Println("Could not setup incoming Conn", err)
		return
	}
	t.hub.registerConn(conn)
	defer t.hub.unregisterConn(conn)
	go conn.writer()
	conn.reader() // Left outside go routine to block
}

// TODO this should take config
func New(connectionString, dbName, colName string) (error, *Turbo) {
	bus := NewMsgBus()
	db, err := NewDatabase(connectionString, dbName, colName)
	if err != nil {
		return err, nil
	}
	hub := NewMsgHub(bus, db)
	turbo := Turbo{
		bus: bus,
		hub: hub,
	}
	// Run the hub
	go hub.listen()

	return nil, &turbo
}
