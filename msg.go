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

	EVENT_TYPE_VALUE         = 0
	EVENT_TYPE_CHILD_ADDED   = 1
	EVENT_TYPE_CHILD_CHANGED = 2
	EVENT_TYPE_CHILD_MOVED   = 3
	EVENT_TYPE_CHILD_REMOVED = 4
	EVENT_TYPES              = 5

	MSG_ERR_TRANS_CONFLICT = "conflict"
)

type Msg struct {
	Cmd       byte            `json:"cmd"`
	Path      string          `json:"path"`
	Event     byte            `json:"eventType"`
	Deltas    json.RawMessage `json:"deltas"`
	DeltasMap json.RawMessage `json:"deltasMap"`
	Ack       int             `json:"ack"`
	Revision  int             `json:"revision"`
}

type ValueEvent struct {
	Path   string           `json:"path"`
	Event  byte             `json:"eventType"`
	Deltas *json.RawMessage `json:"deltas"`
}

type Ack struct {
	Type     byte        `json:"type"`
	Error    string      `json:"err"`
	Deltas   interface{} `json:"deltas"`
	Value    interface{} `json:"value"`
	Ack      int         `json:"ack"`
	Revision int         `json:"revision"`
}

type RawMsg struct {
	Payload []byte
	Conn    *Conn
}
