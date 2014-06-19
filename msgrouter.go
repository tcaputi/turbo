package main

import (
    "encoding/json"
)

const (
    MSG_CMD_ON          = "on"
    MSG_CMD_OFF         = "off"
    MSG_CMD_SET         = "set"
    MSG_CMD_UPDATE      = "update"
    MSG_CMD_REMOVE      = "remove"
    MSG_CMD_TRANS_SET   = "transSet"
    MSG_CMD_PUSH        = "push"
    MSG_CMD_TRANS_GET   = "transGet"
    MSG_CMD_AUTH        = "auth"
    MSG_CMD_UNAUTH      = "unauth"
)

type Msg struct {
    Cmd         string
    Path        string
    EventType   string
    Revision    int
    Value       interface{}
    Ack         int
}

type MsgRouter struct {

}

func (mr *MsgRouter) route(payload byte[], conn *connection) {
    msg := &make(Msg)
    err := json.Unmarshal(payload, &msg)
    if err != nil return

    switch msg.Cmd {
        case MSG_CMD_ON:
            msgBus.subscribe(msg.Path, msg.EventType, conn)
        case MSG_CMD_OFF:
            msgBus.unsubscribe(msg.Path, msg.EventType, conn)
    }
}