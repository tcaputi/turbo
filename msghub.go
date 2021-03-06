package turbo

import (
	"encoding/json"
	"log"
	"strings"
)

type MsgHub struct {
	// Registered connections.
	connections map[uint64]*Conn
	// Register requests from the connections.
	registration chan *Conn
	// Unregister requests from connections.
	unregistration chan *Conn
	// Message bus reference
	bus *MsgBus
	// The database
	db *Database
	// Locker for transactions
	locker *Locker
}

func NewMsgHub(bus *MsgBus, db *Database) *MsgHub {
	hub := MsgHub{
		registration:   make(chan *Conn),
		unregistration: make(chan *Conn),
		connections:    make(map[uint64]*Conn),
		bus:            bus,
		db:             db,
		locker:         NewLocker(),
	}
	return &hub
}

func (hub *MsgHub) listen() {
	for {
		select {
		// There is a Conn 'c' in the registration queue
		case conn := <-hub.registration:
			hub.connections[conn.id] = conn
			log.Printf("Connection #%d connected.\n", conn.id)
		// There is a Conn 'c' in the unregistration queue
		case conn := <-hub.unregistration:
			delete(hub.connections, conn.id)
			hub.bus.unsubscribeAll(conn)
			close(conn.outbox)
			conn.ws.Close()
			log.Printf("Connection #%d was killed.\n", conn.id)
		}
	}
}

func (hub *MsgHub) registerConn(conn *Conn) {
	hub.registration <- conn
}

func (hub *MsgHub) unregisterConn(conn *Conn) {
	hub.unregistration <- conn
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
		hub.bus.subscribe(msg.Event, msg.Path, conn)
	case MSG_CMD_OFF:
		log.Printf("Connection #%d unsubscribed from: '%s', event: '%d'\n", conn.id, msg.Path, msg.Event)
		hub.bus.unsubscribe(msg.Event, msg.Path, conn)
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
	case MSG_CMD_PUSH:
		log.Printf("Connection #%d has done a push on path: '%s'\n", conn.id, msg.Path)
		// go hub.handlePush(&msg, conn)
	case MSG_CMD_TRANS_GET:
		log.Printf("Connection #%d has done trans-get on path: '%s'\n", conn.id, msg.Path)
		go hub.handleTransGet(&msg, conn)
	case MSG_CMD_AUTH:
		log.Printf("Connection #%d has done an auth on path: '%s'\n", conn.id, msg.Path)
		// go hub.handleAuth(&msg, conn)
	case MSG_CMD_UNAUTH:
		log.Printf("Connection #%d has done an unauth on path: '%s'\n", conn.id, msg.Path)
		// go hub.handleUnauth(&msg, conn)

	default:
		log.Fatalf("Connection #%d submitted a message with cmd #%d which is unsupported\n", conn.id, msg.Cmd)
	}
}

// TODO: db should delete, then set new value
func (hub *MsgHub) handleSet(msg *Msg, conn *Conn) {
	var unmarshalledValue interface{}
	jsonErr := json.Unmarshal(msg.Data, &unmarshalledValue)
	if jsonErr != nil {
		errStr := jsonErr.Error()
		hub.sendAck(conn, msg.Ack, &errStr, nil, 0)
	} else {
		log.Println("Now setting value to path ", msg.Path)
		// Notify all listeners of recursive value change
		hub.publishAndDestroy(msg.Path)
		// Set the new value
		setErr := hub.db.set(msg.Path, unmarshalledValue)
		if setErr != nil {
			log.Fatalln("Couldn't set node value", setErr)
			errStr := setErr.Error()
			hub.sendAck(conn, msg.Ack, &errStr, nil, 0)
		} else {
			hub.sendAck(conn, msg.Ack, nil, nil, 0)
			hub.publishValueEvent(msg.Path, &msg.Data, conn)
		}
	}
}

// TODO add "remove with null" support
func (hub *MsgHub) handleUpdate(msg *Msg, conn *Conn) {
	if msg.DataMap == nil {
		return
	}
	propertyMap := make(map[string]json.RawMessage)
	json.Unmarshal(msg.DataMap, &propertyMap)
	responses := make(chan *error, 256)
	for property, value := range propertyMap {
		go (func(path string, val json.RawMessage) {
			var unmarshalledValue interface{}
			newPath := hub.joinPaths(msg.Path, path)
			jsonErr := json.Unmarshal(val, &unmarshalledValue)
			if jsonErr != nil {
				log.Fatalln("Couldn't marshal new value json", jsonErr)
				responses <- &jsonErr
				return // ಠ_ಠ
			} else {
				hub.locker.lock(newPath)
				setErr := hub.db.set(newPath, unmarshalledValue)
				hub.locker.unlock(newPath)
				if setErr != nil {
					log.Fatalln("Couldn't set node value", setErr)
					responses <- &setErr
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
					hub.sendAck(conn, msg.Ack, nil, nil, 0)
				} else {
					hub.sendAck(conn, msg.Ack, &problems, nil, 0)
				}
				close(responses)
				return
			} else {
				i = i + 1
			}
		}
	}
}

func (hub *MsgHub) handleRemove(msg *Msg, conn *Conn) {
	// Remove children first
	node := hub.bus.pathTree.get(msg.Path)
	if node == nil {
		err := "Path does not exist"
		hub.sendAck(conn, msg.Ack, &err, nil, 0)
		return
	}
	// Depth first traversal of path
	hub.publishAndDestroy(node.path)
	hub.locker.lock(node.path)
	setErr := hub.db.set(node.path, nil)
	hub.locker.unlock(node.path)
	if setErr != nil {
		log.Fatalln("Couldn't set node value", setErr)
		errStr := setErr.Error()
		hub.sendAck(conn, msg.Ack, &errStr, nil, 0)
	} else {
		hub.sendAck(conn, msg.Ack, nil, nil, 0)
	}
}

func (hub *MsgHub) handleTransSet(msg *Msg, conn *Conn) {
	hub.locker.lock(msg.Path)
	err, value, rev := hub.db.get(msg.Path)
	if err != nil {
		errStr := err.Error()
		hub.sendAck(conn, msg.Ack, &errStr, nil, 0)
	}

	// compare revisions
	if msg.Revision == rev {
		hub.locker.unlock(msg.Path)
		hub.handleSet(msg, conn)
	} else {
		hub.locker.unlock(msg.Path)
		errStr := MSG_ERR_TRANS_CONFLICT
		hub.sendAck(conn, msg.Ack, &errStr, value, 0)
	}
}

func (hub *MsgHub) handleTransGet(msg *Msg, conn *Conn) {
	err, val, rev := hub.db.get(msg.Path)

	if err != nil {
		errStr := err.Error()
		hub.sendAck(conn, msg.Ack, &errStr, nil, 0)
	} else {
		hub.sendAck(conn, msg.Ack, nil, val, rev)
	}
}

func (hub *MsgHub) publishAndDestroy(path string) {
	node := hub.bus.pathTree.get(path)

	if node != nil {
		node.cascade(func(child *PathTreeNode) {
			evt := ValueEvent{}
			evt.Event = EVENT_TYPE_VALUE
			evt.Data = nil
			evt.Path = child.path
			evtJson, jsonErr := json.Marshal(evt)
			if jsonErr != nil {
				log.Fatalln("Couldn't marshal event json", jsonErr)
			} else {
				hub.bus.publish(EVENT_TYPE_VALUE, child.path, evtJson)
			}
			// Check any parents for the child removed
			if child.hasImmediateParent() {
				// We need to get the child value
				getErr, childVal, _ := hub.db.get(child.path)
				if getErr != nil {
					log.Fatalln("Couldn't fetch node value", getErr)
					return
				}
				evt.Event = EVENT_TYPE_CHILD_REMOVED
				evt.Data = childVal
				evtJson, jsonErr = json.Marshal(evt)
				if jsonErr != nil {
					log.Fatalln("Couldn't marshal event json", jsonErr)
				} else {
					hub.bus.publish(EVENT_TYPE_CHILD_REMOVED, child.parent.path, evtJson)
				}
			}
		})
	}
}

func (hub *MsgHub) publishValueEvent(path string, value *json.RawMessage, conn *Conn) {
	var hasValueSubs bool
	var hasChildChangedSubs bool
	var parentPath string

	if hub.hasParent(path) {
		parentPath = hub.parentOf(path)
	}
	// Check if anyone is event subscribed to this
	hasValueSubs = hub.bus.hasSubscribers(EVENT_TYPE_VALUE, path)
	if parentPath != "" {
		hasChildChangedSubs = hub.bus.hasSubscribers(EVENT_TYPE_CHILD_CHANGED, path)
	}
	// Do not continue if there are no subscribers
	if !hasValueSubs && !hasChildChangedSubs {
		return
	}
	// This is what we are sending subscribers to a value related event
	evt := ValueEvent{
		Path: path,
		Data: value,
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
		hub.bus.publish(EVENT_TYPE_VALUE, path, evtJson)
	}
	// Send the event to child changed listeners
	if hasChildChangedSubs && hub.hasParent(path) {
		// Set the event type, parent path; jsonify
		if value == nil {
			evt.Event = EVENT_TYPE_CHILD_REMOVED
		} else {
			evt.Event = EVENT_TYPE_CHILD_CHANGED
		}
		evt.Path = hub.parentOf(path)
		evtJson, err := json.Marshal(evt)
		if err != nil {
			problem := "Couldn't marshal msg value json\n"
			log.Fatalln(problem, err)
			return
		}
		hub.bus.publish(EVENT_TYPE_CHILD_CHANGED, path, evtJson)
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

func (hub *MsgHub) sendAck(conn *Conn, ack int, errString *string, result interface{}, rev int) {
	response := Ack{
		Type:     MSG_CMD_ACK,
		Ack:      ack,
		Data:     result,
		Revision: rev,
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
			hub.unregisterConn(conn)
		}
	}
}
