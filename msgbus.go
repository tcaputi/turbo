package turbo

import (
	"sync"
)

type MsgBus struct {
	evtMaps   map[*PathTreeNode]*[EVENT_TYPES]*map[*Conn]bool
	pathTree  *PathTree
	pathLocks map[string]*sync.Mutex
}

func NewMsgBus() *MsgBus {
	bus := MsgBus{
		evtMaps:   make(map[*PathTreeNode]*[EVENT_TYPES]*map[*Conn]bool),
		pathTree:  NewPathTree(),
		pathLocks: make(map[string]*sync.Mutex),
	}
	return &bus
}

func (bus *MsgBus) subscribe(evt byte, path string, conn *Conn) {
	var evtMap [EVENT_TYPES]*map[*Conn]bool
	var connSet map[*Conn]bool
	var node *PathTreeNode

	node = bus.pathTree.put(path)
	if _, exists := bus.evtMaps[node]; !exists {
		var lock *sync.Mutex
		if bus.pathLocks[path] != nil {
			lock = bus.pathLocks[path]
		} else {
			lock = &sync.Mutex{}
			bus.pathLocks[path] = lock
		}
		lock.Lock()

		if _, exists := bus.evtMaps[node]; !exists {
			evtMap = [EVENT_TYPES]*map[*Conn]bool{}
			bus.evtMaps[node] = &evtMap
		}
		delete(bus.pathLocks, path)

		lock.Unlock()
	} else {
		evtMap = *(bus.evtMaps[node])
	}

	if evtMap[evt] == nil {
		var lock *sync.Mutex
		if bus.pathLocks[path] != nil {
			lock = bus.pathLocks[path]
		} else {
			lock = &sync.Mutex{}
			bus.pathLocks[path] = lock
		}
		lock.Lock()

		connSet = make(map[*Conn]bool)
		evtMap[evt] = &connSet
		delete(bus.pathLocks, path)

		lock.Unlock()
	} else {
		connSet = *(evtMap[evt])
	}

	connSet[conn] = true
	conn.subscriptions[&connSet] = true
}

func (bus *MsgBus) unsubscribe(evt byte, path string, conn *Conn) {
	var evtMap [EVENT_TYPES]*map[*Conn]bool
	var connSet map[*Conn]bool
	var node *PathTreeNode

	node = bus.pathTree.get(path)
	if node == nil {
		return
	}

	if _, exists := bus.evtMaps[node]; !exists {
		return
	} else {
		evtMap = *(bus.evtMaps[node])
	}

	if evtMap[evt] == nil {
		return
	} else {
		connSet = *(evtMap[evt])
	}

	if _, exists := connSet[conn]; !exists {
		return
	} else {
		delete(connSet, conn)
		delete(conn.subscriptions, &connSet)
	}
}

func (bus *MsgBus) publish(evt byte, path string, msg []byte) {
	var evtMap [EVENT_TYPES]*map[*Conn]bool
	var connSet map[*Conn]bool
	var node *PathTreeNode

	node = bus.pathTree.get(path)
	if node == nil {
		return
	}

	if _, exists := bus.evtMaps[node]; !exists {
		return
	} else {
		evtMap = *(bus.evtMaps[node])
	}

	if evtMap[evt] == nil {
		return
	} else {
		connSet = *(evtMap[evt])
	}

	for conn, _ := range connSet {
		conn.outbox <- msg
	}
}

func (bus *MsgBus) unsubscribeAll(conn *Conn) {
	for subscription := range conn.subscriptions {
		delete(*subscription, conn)
		delete(conn.subscriptions, subscription)
	}
}

func (bus *MsgBus) hasSubscribers(evt byte, path string) bool {
	var evtMap [EVENT_TYPES]*map[*Conn]bool
	var node *PathTreeNode

	node = bus.pathTree.get(path)
	if node == nil {
		return false
	}

	if _, exists := bus.evtMaps[node]; !exists {
		return false
	} else {
		evtMap = *(bus.evtMaps[node])
	}

	if evtMap[evt] == nil {
		return false
	} else {
		return len(*evtMap[evt]) > 0
	}
}
