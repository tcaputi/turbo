package turbo

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"testing"
)

func TestJoinPaths(t *testing.T) {
	hub := &MsgHub{
		registration:   make(chan *Conn),
		unregistration: make(chan *Conn),
		connections:    make(map[uint64]*Conn),
	}

	str1 := hub.joinPaths("/", "/dfdf/dsfsdf/ds")
	str2 := hub.joinPaths("/234/45/", "/dfdf/dsfsdf/ds")
	str3 := hub.joinPaths("/234/45", "dfdf/dsfsdf/ds")

	if str1 != "/dfdf/dsfsdf/ds" {
		t.Error("The path join with str1 failed", str1)
	}
	if str2 != "/234/45/dfdf/dsfsdf/ds" {
		t.Error("The path join with str2 failed", str2)
	}
	if str3 != "/234/45/dfdf/dsfsdf/ds" {
		t.Error("The path join with str3 failed", str3)
	}
}

func TestSendAck(t *testing.T) {
	bus := NewMsgBus()
	hub := NewMsgHub(bus)
	conn := Conn{
		id:            1,
		ws:            nil,
		outbox:        make(chan []byte, 256),
		subscriptions: make(map[*map[*Conn]bool]bool),
		hub:           nil,
	}
	// Test error
	errStr := "This is an error"
	hub.sendAck(&conn, 1, &errStr, nil, "")
	// Test regular w/ empty hash
	testVal := map[string]interface{}{
		"key1": "value1",
		"key2": 2,
		"key3": map[string]interface{}{
			"subkey1": "value1",
			"subkey2": "value2",
		},
		"key4": [...]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	jsonVal, jsonErr := json.Marshal(testVal)
	if jsonErr != nil {
		t.Error("Could not serialize testVal", jsonErr)
		t.FailNow()
		return
	}
	hub.sendAck(&conn, 2, nil, jsonVal, "")
	// With hash
	gob.Register(map[string]interface{}{})
	gob.Register([10]int{})
	hashErr, hashVal := (&hub).hashify(testVal)
	if hashErr != nil {
		t.Error("Could not hash testVal", hashErr)
		t.FailNow()
	}
	hub.sendAck(&conn, 3, nil, jsonVal, string(hashVal[:]))

	// TODO read through outbox to check the acks
	for i := 0; i < 3; i++ {
		select {
		case msg := <-conn.outbox:
			fmt.Println("Msg", i, "came out with", string(msg[:]))
		}
	}
}

func TestObjHash(t *testing.T) {
	// TODO check our hash actually fucking works
}
