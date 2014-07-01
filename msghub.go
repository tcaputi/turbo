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
		log.Printf("Connection #%d has set a value to path: '%s'\n", conn.id, msg.Path)
		go hub.handleSet(&msg, conn)
	case MSG_CMD_UPDATE:
		log.Printf("Connection #%d has updated path: '%s'\n", conn.id, msg.Path)
		go hub.handleUpdate(&msg, conn)
	case MSG_CMD_REMOVE:
		log.Printf("Connection #%d has removed path: '%s'\n", conn.id, msg.Path)
		go hub.handleRemove(&msg, conn)
	case MSG_CMD_TRANS_SET:
		log.Printf("Connection #%d has done trans-set on path: '%s'\n", conn.id, msg.Path)
		go hub.handleTransSet(&msg, conn)
	case MSG_CMD_TRANS_GET:
		log.Printf("Connection #%d has done trans-get on path: '%s'\n", conn.id, msg.Path)
		go hub.handleTransGet(&msg, conn)
	default:
		log.Fatalf("Connection #%d submitted a message with cmd #%d which is unsupported\n", conn.id, msg.Cmd)
	}
}

func (hub *MsgHub) handleSet(msg *Msg, conn *connection) {
	var unmarshalledValue interface{}
	jsonErr := json.Unmarshal(msg.Value, &unmarshalledValue)
	if jsonErr != nil {
		errStr := jsonErr.Error()
		hub.sendAck(conn, msg.Ack, &errStr, nil, "")
	} else {
		log.Println("Now setting value to path ", msg.Path)
		err := database.set(msg.Path, unmarshalledValue)
		if err != nil {
			errStr := err.Error()
			hub.sendAck(conn, msg.Ack, &errStr, nil, "")
		} else {
			hub.sendAck(conn, msg.Ack, nil, nil, "")
			hub.publishValueEvent(msg.Path, &msg.Value, conn)
		}
	}
}

// TODO add "remove with null" support
func (hub *MsgHub) handleUpdate(msg *Msg, conn *connection) {
	if msg.Value == nil {
		return
	}
	propertyMap := make(map[string]json.RawMessage)
	json.Unmarshal(msg.Value, &propertyMap)
	responses := make(chan *error)
	for property, value := range propertyMap {
		go (func(path string, val json.RawMessage) {
			var unmarshalledValue interface{}

			newPath := hub.joinPaths(msg.Path, path)
			jsonErr := json.Unmarshal(val, &unmarshalledValue)
			if jsonErr != nil {
				responses <- &jsonErr
				return // ಠ_ಠ
			} else {
				err := database.set(path, unmarshalledValue)
				if err != nil {
					responses <- &err
					return
				}
			}
			responses <- nil
			hub.publishValueEvent(newPath, &val, conn)
		})(property, value)
	}
	// Collect the callbacks
	i := 1
	problems := ""
	for {
		select {
		case err := <-responses:
			if err != nil {
				problems += (*err).Error() + "\n"
			}

			if i == len(propertyMap) {
				// We're done - send the response
				if problems == "" {
					hub.sendAck(conn, msg.Ack, nil, nil, "")
				} else {
					hub.sendAck(conn, msg.Ack, &problems, nil, "")
				}
				close(responses)
				return
			} else {
				i = i + 1
			}
		}
	}
}

func (hub *MsgHub) handleRemove(msg *Msg, conn *connection) {
}

func (hub *MsgHub) handleTransSet(msg *Msg, conn *connection) {
	// db get
	err, val := database.get(msg.Path)
	if err != nil {
		errStr := err.Error()
		hub.sendAck(conn, msg.Ack, &errStr, nil, "")
	}

	if val != nil {
		// grab le hash
		err, currValHash := hash(val)
		if err != nil {
			errStr := err.Error()
			hub.sendAck(conn, msg.Ack, &errStr, nil, "")
		}
		// compare le hashes
		if msg.Hash == string(currValHash[:]) {
			hub.handleSet(msg, conn) // actually
			// this should work dafaq
		} else {
			errStr := "conflict"
			hub.sendAck(conn, msg.Ack, &errStr, nil, "")
		}
	} else {
		// hashing dont make sense if val is nil
		// this is fine for now
		// no - we send nil
		hub.handleSet(msg, conn)
	}
}

func (hub *MsgHub) handleTransGet(msg *Msg, conn *connection) {
	// db get
	err, val := database.get(msg.Path)
	if err != nil {
		errStr := err.Error()
		hub.sendAck(conn, msg.Ack, &errStr, nil, "")
		return
	}
	if val == nil {
		hub.sendAck(conn, msg.Ack, nil, nil, "")
		return
	}
	// grab le hash
	log.Println("Now hashing val:", val)
	err, currValHash := hash(val)
	if err != nil {
		errStr := err.Error()
		hub.sendAck(conn, msg.Ack, &errStr, nil, "")
	}
	// send the value with the hash
	hub.sendAck(conn, msg.Ack, nil, val, string(currValHash[:]))
}

func (hub *MsgHub) publishValueEvent(path string, value *json.RawMessage, conn *connection) {
	var hasValueSubs bool
	var hasChildChangedSubs bool
	var parentPath string

	if hub.hasParent(path) {
		parentPath = hub.parentOf(path)
	}
	// Check if anyone is event subscribed to this
	hasValueSubs = msgBus.hasSubscribers(path, EVENT_TYPE_VALUE)
	if parentPath != "" {
		hasChildChangedSubs = msgBus.hasSubscribers(path, EVENT_TYPE_CHILD_CHANGED)
	}
	// Do not continue if there are no subscribers
	if !hasValueSubs && !hasChildChangedSubs {
		return
	}
	// This is what we are sending subscribers to a value related event
	evt := ValueEvent{
		Path:  path,
		Value: value,
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
		msgBus.publish(path, EVENT_TYPE_VALUE, evtJson)
	}
	// Send the event to child changed listeners
	if hasChildChangedSubs && hub.hasParent(path) {
		// Set the event type, parent path; jsonify
		evt.Event = EVENT_TYPE_CHILD_CHANGED
		evt.Path = hub.parentOf(path)
		evtJson, err := json.Marshal(evt)
		if err != nil {
			problem := "Couldn't marshal msg value json\n"
			log.Fatalln(problem, err)
			return
		}
		msgBus.publish(path, EVENT_TYPE_CHILD_CHANGED, evtJson)
	}
}

func (hub *MsgHub) hasParent(path string) bool {
	return path != "/"
}

func (hub *MsgHub) parentOf(path string) string {
	lastIndex := strings.LastIndex(path, "/")
	return path[0:lastIndex]
}

func (hub *MsgHub) joinPaths(base string, extension string) string {
	if !strings.HasSuffix(base, "/") {
		base = base + "/"
	}
	if strings.HasPrefix(extension, "/") {
		if len(extension) > 1 {
			extension = extension[1:]
		} else {
			extension = ""
		}
	}
	return base + extension
}

func (hub *MsgHub) sendAck(conn *connection, ack int, errString *string, result interface{}, hash string) {
	response := Ack{
		Type:   MSG_CMD_ACK,
		Ack:    ack,
		Result: result,
	}
	if hash != "" {
		// Strings cant be nil in go
		// YES. I KNOW.
		response.Hash = hash
	}
	if errString != nil {
		log.Println("Sending problem back to client in ack form:", *errString)
		response.Error = *errString
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
