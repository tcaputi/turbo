package turbo

import (
	"encoding/json"
	"log"
	"strings"
)

type MsgHub struct {
	// Registered connections.
	connections map[uint64]*connection
	// Register requests from the connections.
	registration chan *connection
	// Unregister requests from connections.
	unregistration chan *connection
}

var (
	msgHub = &MsgHub{
		registration:   make(chan *connection),
		unregistration: make(chan *connection),
		connections:    make(map[uint64]*connection),
	}
)

func (hub *MsgHub) run() {
	for {
		select {
		// There is a connection 'c' in the registration queue
		case conn := <-hub.registration:
			hub.connections[conn.id] = conn
		// There is a connection 'c' in the unregistration queue
		case conn := <-hub.unregistration:
			delete(hub.connections, conn.id)
			msgBus.unsubscribeAll(conn)
			close(conn.outbox)
			conn.ws.Close()
			log.Printf("Connection #%d was killed.\n", conn.id)
		}
	}
}

func (hub *MsgHub) route(rawMsg *RawMsg) {
	payload := rawMsg.Payload
	conn := rawMsg.Conn

	msg := Msg{}
	err := json.Unmarshal(payload, &msg)
	if err != nil {
		log.Fatalln("Msg router could not marshal json of an incoming msg", err)
		return
	}

	switch msg.Cmd {
	case MSG_CMD_ON:
		log.Printf("Connection #%d subscribed to: '%s', event: '%d'\n", conn.id, msg.Path, msg.Event)
		msgBus.subscribe(msg.Path, msg.Event, conn)
	case MSG_CMD_OFF:
		log.Printf("Connection #%d unsubscribed from: '%s', event: '%d'\n", conn.id, msg.Path, msg.Event)
		msgBus.unsubscribe(msg.Path, msg.Event, conn)
	case MSG_CMD_SET:
		log.Printf("Connection #%d has set a new value to path: '%s'\n", conn.id, msg.Path)
		go hub.handleSet(&msg, conn)
	default:
		log.Fatalf("Connection #%d submitted a message with cmd #%d which is unsupported\n", conn.id, msg.Cmd)
	}
}

func (hub *MsgHub) handleSet(msg *Msg, conn *connection) {
	var hasValueSubs bool
	var hasChildChangedSubs bool
	var parentPath string

	if hasParent(msg.Path) {
		parentPath = parentOf(msg.Path)
	}
	//
	// TODO run the db query representing the set
	// TODO db needs to tell us if it was a create or an update
	sendAck(conn, msg.Ack, nil, nil) // <- this will send errors from the db insert in future
	// Check if anyone is event subscribed to this
	hasValueSubs = msgBus.hasSubscribers(msg.Path, EVENT_TYPE_VALUE)
	if parentPath != "" {
		hasChildChangedSubs = msgBus.hasSubscribers(msg.Path, EVENT_TYPE_CHILD_CHANGED)
	}
	// Do not continue if there are no subscribers
	if !hasValueSubs && !hasChildChangedSubs {
		return
	}
	// This is what we are sending subscribers to a value related event
	evt := ValueEvent{
		Path:  msg.Path,
		Value: &(msg.Value),
	}
	// Send the event to value listeners
	if hasValueSubs {
		// Set the event type; jsonify
		evt.Event = EVENT_TYPE_VALUE
		evtJson, err := json.Marshal(evt)
		if err != nil {
			problem := "Couldn't marshal msg value json\n"
			log.Fatalln(problem, err)
			return
		}
		msgBus.publish(msg.Path, EVENT_TYPE_VALUE, evtJson)
	}
	// Send the event to child changed listeners
	if hasChildChangedSubs && hasParent(msg.Path) {
		// Set the event type, parent path; jsonify
		evt.Event = EVENT_TYPE_CHILD_CHANGED
		evt.Path = parentOf(msg.Path)
		evtJson, err := json.Marshal(evt)
		if err != nil {
			problem := "Couldn't marshal msg value json\n"
			log.Fatalln(problem, err)
			return
		}
		msgBus.publish(msg.Path, EVENT_TYPE_CHILD_CHANGED, evtJson)
	}
}

func hasParent(path string) bool {
	return path != "/"
}

func parentOf(path string) string {
	lastIndex := strings.LastIndex(path, "/")
	return path[0:lastIndex]
}

func sendAck(conn *connection, ack int, error *string, result interface{}) {
	response := Ack{
		Type:   MSG_CMD_ACK,
		Ack:    ack,
		Result: result,
	}
	if error != nil {
		response.Error = *error
	}
	payload, err := json.Marshal(response)
	if err == nil {
		select {
		case conn.outbox <- payload:
		default:
			defer conn.kill()
		}
	}
}
