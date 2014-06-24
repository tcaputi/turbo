package turbo

import (
	"encoding/json"
	"log"
	"strings"
)

type MsgHub struct {
	// Registered connections.
	connections map[*connection]bool // *NOTE: in go, its a popular pattern to emulate a set wih map[*]bool
	// Inbound messages from the connections.
	inbox chan *RawMsg
	// Register requests from the connections.
	registration chan *connection
	// Unregister requests from connections.
	unregistration chan *connection
}

var (
	msgHub = &MsgHub{
		inbox:          make(chan *RawMsg),
		registration:   make(chan *connection),
		unregistration: make(chan *connection),
		connections:    make(map[*connection]bool),
	}
)

func (hub *MsgHub) run() {
	for {
		select {
		// There is a connection 'c' in the registration queue
		case c := <-hub.registration:
			hub.connections[c] = true
		// There is a connection 'c' in the unregistration queue
		case c := <-hub.unregistration:
			delete(hub.connections, c)
			msgBus.unsubscribeAll(c)
			close(c.outbox)
			c.ws.Close()
			log.Printf("Connection #%d was killed.\n", c.id)
		// There is a message 'm' in the broadcast queue
		case m := <-hub.inbox:
			// Send the incoming message to the message router
			hub.route(m)
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
		log.Printf("Connection #%d subscribed to path: '%s', event: '%s'\n", conn.id, msg.Path, msg.Event)
		msgBus.subscribe(msg.Path, msg.Event, conn)
	case MSG_CMD_OFF:
		log.Printf("Connection #%d unsubscribed from path: '%s', event: '%s'\n", conn.id, msg.Path, msg.Event)
		msgBus.unsubscribe(msg.Path, msg.Event, conn)
	case MSG_CMD_SET:
		log.Printf("Connection #%d has set a new value to path: '%s'\n", conn.id, msg.Path)
		hub.handleSet(&msg, conn)
	default:
		log.Fatalf("Received msg with cmd '%s' which is unsupported\n", msg.Cmd)
	}
}

func (hub *MsgHub) handleSet(msg *Msg, conn *connection) {
	//
	// TODO run the db query representing the set
	// TODO db needs to tell us if it was a create or an update
	sendAck(conn, msg.Ack, nil, nil) // <- this will send errors from the db insert in future
	// Check if anyone is event subscribed to this
	hasValueSubs := msgBus.hasSubscribers(msg.Path, EVENT_TYPE_VALUE)
	hasChildChangedSubs := msgBus.hasSubscribers(msg.Path, EVENT_TYPE_CHILD_CHANGED)
	if !hasValueSubs && !hasChildChangedSubs {
		return
	}
	// This is what we are sending subscribers to a value related event
	evt := ValueChangeEvent{
		Type:  MSG_CMD_ON,
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
		// FIRE AWAY
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
		// FIRE AWAY
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
	response := MsgResponse{Type: MSG_CMD_ACK, Ack: ack, Result: result}
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
