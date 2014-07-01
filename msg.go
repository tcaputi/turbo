package turbo

import (
	"encoding/json"
)

const (
	MSG_CMD_ON        = 1
	MSG_CMD_OFF       = 2
	MSG_CMD_SET       = 3
	MSG_CMD_UPDATE    = 4
	MSG_CMD_REMOVE    = 5
	MSG_CMD_TRANS_SET = 6
	MSG_CMD_PUSH      = 7
	MSG_CMD_TRANS_GET = 8
	MSG_CMD_AUTH      = 9
	MSG_CMD_UNAUTH    = 10
	MSG_CMD_ACK       = 11

	EVENT_TYPE_VALUE         = 1
	EVENT_TYPE_CHILD_ADDED   = 2
	EVENT_TYPE_CHILD_CHANGED = 3
	EVENT_TYPE_CHILD_MOVED   = 4
	EVENT_TYPE_CHILD_REMOVED = 5
)

type Msg struct {
	Cmd   byte            `json:"cmd"`
	Path  string          `json:"path"`
	Event byte            `json:"eventType"`
	Value json.RawMessage `json:"value"`
	Ack   int             `json:"ack"`
	Hash  string          `json:"hash"`
}

type ValueEvent struct {
	Path  string           `json:"path"`
	Event byte             `json:"eventType"`
	Value *json.RawMessage `json:"value"`
}

type Ack struct {
	Type   byte        `json:"type"`
	Error  string      `json:"err"`
	Result interface{} `json:"res"`
	Ack    int         `json:"ack"`
	Hash   string      `json:"hash"`
}

type RawMsg struct {
	Payload []byte
	Conn    *connection
}
