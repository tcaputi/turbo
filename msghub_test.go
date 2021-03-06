package turbo

import (
	"encoding/json"
	"testing"
)

func TestJoinPaths(t *testing.T) {
	hub := NewMsgHub(nil, nil)

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
	hub := NewMsgHub(bus, nil)
	conn := Conn{
		id:            1,
		ws:            nil,
		outbox:        make(chan []byte, 256),
		subscriptions: make(map[*map[*Conn]bool]bool),
		hub:           nil,
	}
	// Test error
	errStr := "This is an error"
	hub.sendAck(&conn, 1, &errStr, nil, 0)
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
	jsonErr, jsonVal := json.Marshal(testVal)
	if jsonErr != nil {
		t.Error("Could not serialize testVal", jsonErr)
		t.FailNow()
		return
	}
	hub.sendAck(&conn, 2, nil, jsonVal, 0)
	// With hash
	// hub.sendAck(&conn, 3, nil, jsonVal, hash(testVal))
	// TODO read through outbox to check the acks
}

func TestObjHash(t *testing.T) {
	// TODO check our hash actually fucking works
}
