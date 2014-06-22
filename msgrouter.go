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
)

type Msg struct {
	Cmd       string          `json:"cmd"`
	Path      string          `json:"path,omitempty"`
	EventType string          `json:"eventType,omitempty"`
	Revision  int             `json:"revision,omitempty"`
	Value     json.RawMessage `json:"value,omitempty"`
	Ack       int             `json:"ack,omitempty"`
}

type ValueChangeEvent struct {
	Path  string `json:"path"`
	Value []byte `json:"value"`
}

type MsgResponse struct {
	Type   string      `json:"type"`
	Error  string      `json:"err,omitempty"`
	Result interface{} `json:"res,omitempty"`
	Ack    int         `json:"ack,omitempty"`
}

func hasParent(path string) bool {
	return path == "/"
}

func parent(path string) string {
	lastIndex := strings.LastIndex(path, "/")
	return path[0:lastIndex]
}

func sendAck(conn *connection, ack int, error string, result interface{}) {
	response := MsgResponse{Type: MSG_RESP_TYPE_ACK, Ack: ack, Result: result}
	if error != "" {
		response.Error = error
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
		log.Fatalln("Msg router could npot marshal json of an incoming msg", err)
		return
	}

	switch msg.Cmd {
	case MSG_CMD_ON:
		msgBus.subscribe(msg.Path, msg.EventType, conn)
	case MSG_CMD_OFF:
		msgBus.unsubscribe(msg.Path, msg.EventType, conn)
	case MSG_CMD_SET:
		// TODO run the db query
		event, err := json.Marshal(ValueChangeEvent{msg.Path, msg.Value})
		if err == nil {
			msgBus.publish(msg.Path, EVENT_TYPE_VALUE, event)
			if hasParent(msg.Path) {
				msgBus.publish(parent(msg.Path), EVENT_TYPE_CHILD_CHANGED, event)
			}
		} else {
			log.Fatalf("Couldn't marshal value change event for set\n", msg.Cmd)
		}
	default:
		log.Fatalf("Received msg with cmd '%s' which is unsupported\n", msg.Cmd)
	}
}
