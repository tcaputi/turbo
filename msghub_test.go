package turbo

import (
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
