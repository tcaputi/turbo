package turbo

import (
	"encoding/json"
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
	MSG_CMD_ACK       = "ack"

	EVENT_TYPE_VALUE         = "value"
	EVENT_TYPE_CHILD_ADDED   = "child_added"
	EVENT_TYPE_CHILD_CHANGED = "child_changed"
	EVENT_TYPE_CHILD_MOVED   = "child_moved"
	EVENT_TYPE_CHILD_REMOVED = "child_removed"
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

type RawMsg struct {
	Payload []byte
	Conn    *connection
}
