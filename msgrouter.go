package main

import (
	"encoding/json"
	"log"
	"strings"
)

const (
	MSG_CMD_ON        = "on"
	MSG_CMD_OFF       = "off"
	MSG_CMD_SET       = "set"
	MSG_CMD_UPDATE    = "update"
	MSG_CMD_REMOVE    = "remove"
	MSG_CMD_TRANS_SET = "transSet"
	MSG_CMD_PUSH      = "push"
	MSG_CMD_TRANS_GET = "transGet"
	MSG_CMD_AUTH      = "auth"
	MSG_CMD_UNAUTH    = "unauth"

	MSG_RESP_TYPE_ACK = "ack"
	MSG_RESP_TYPE_ON  = MSG_CMD_ON
)

type Msg struct {
	Cmd      string          `json:"cmd"`
	Path     string          `json:"path"`
	Event    string          `json:"eventType"`
	Revision int             `json:"revision"`
	Value    json.RawMessage `json:"value"`
	Ack      int             `json:"ack"`
}

type ValueChangeEvent struct {
	Type  string           `json:"type"`
	Path  string           `json:"path"`
	Event string           `json:"eventType"`
	Value *json.RawMessage `json:"value"`
}

type MsgResponse struct {
	Type   string      `json:"type"`
	Error  string      `json:"err"`
	Result interface{} `json:"res"`
	Ack    int         `json:"ack"`
}

func hasParent(path string) bool {
	return path != "/"
}

func parent(path string) string {
	lastIndex := strings.LastIndex(path, "/")
	return path[0:lastIndex]
}

func sendAck(conn *connection, ack int, error *string, result interface{}) {
	response := MsgResponse{Type: MSG_RESP_TYPE_ACK, Ack: ack, Result: result}
	if error != nil {
		response.Error = *error
	}
	payload, err := json.Marshal(response)
	if err == nil {
		select {
		case conn.send <- payload:
		default:
			// TODO make this kill the conn
		}
	}
}

func routeRawMsg(rawMsg *rawMsg) {
	payload := rawMsg.payload
	conn := rawMsg.conn
	msg := Msg{}
	err := json.Unmarshal(payload, &msg)
	if err != nil {
		log.Fatalln("Msg router could not marshal json of an incoming msg", err)
		return
	}

	switch msg.Cmd {
	case MSG_CMD_ON:
		log.Printf("A connection subscribed to path: '%s', event: '%s'\n", msg.Path, msg.Event)
		msgBus.subscribe(msg.Path, msg.Event, conn)
	case MSG_CMD_OFF:
		log.Printf("A connection unsubscribed to path: '%s', event: '%s'\n", msg.Path, msg.Event)
		msgBus.unsubscribe(msg.Path, msg.Event, conn)
	case MSG_CMD_SET:
		log.Printf("A connection has set a new value to path: '%s'\n", msg.Path)
		// TODO run the db query
		msgValue, err := msg.Value.MarshalJSON()
		if err != nil {
			problem := "Couldn't marshal msg value json\n"
			log.Fatalln(problem, err)
			sendAck(conn, msg.Ack, &problem, nil)
			return
		}
		log.Println("Hey look! MSGVALUE: ", string(msgValue[:]))

		evt := ValueChangeEvent{Type: MSG_RESP_TYPE_ON, Event: EVENT_TYPE_VALUE, Path: msg.Path, Value: &(msg.Value)}
		evtJson, err := json.Marshal(evt)
		if err != nil {
			problem := "Couldn't marshal event json\n"
			log.Fatalln(problem, err)
			sendAck(conn, msg.Ack, &problem, nil)
			return
		}

		msgBus.publish(msg.Path, EVENT_TYPE_VALUE, evtJson)
		if hasParent(msg.Path) {
			// Redo this for parent event
			evt.Event = EVENT_TYPE_CHILD_CHANGED
			evtJson, err := json.Marshal(evt)
			if err != nil {
				problem := "Couldn't marshal msg value json\n"
				log.Fatalln(problem, err)
				sendAck(conn, msg.Ack, &problem, nil)
				return
			}

			msgBus.publish(parent(msg.Path), EVENT_TYPE_CHILD_CHANGED, evtJson)
			sendAck(conn, msg.Ack, nil, nil)
		} else {
			sendAck(conn, msg.Ack, nil, nil)
		}
	default:
		log.Fatalf("Received msg with cmd '%s' which is unsupported\n", msg.Cmd)
	}
}
