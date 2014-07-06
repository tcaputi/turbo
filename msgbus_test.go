package turbo

import (
	"sync"
	"testing"
)

func TestSubscribe(t *testing.T) {
	bus := NewMsgBus()
	conn := Conn{
		id:            1,
		ws:            nil,
		outbox:        make(chan []byte),
		subscriptions: make(map[*map[*Conn]bool]bool),
		hub:           nil,
	}
	paths := map[string]byte{
		"/a/b/c":       EVENT_TYPE_VALUE,
		"/a/b/c/d":     EVENT_TYPE_CHILD_ADDED,
		"/a/b":         EVENT_TYPE_CHILD_CHANGED,
		"/a/b/c/d/e/f": EVENT_TYPE_CHILD_REMOVED,
		"/1/2/3":       EVENT_TYPE_CHILD_MOVED,
	}

	for path, evt := range paths {
		bus.subscribe(evt, path, &conn)
	}

	if len(bus.evtMaps) != 5 {
		t.Error("The bus had the wrong number of event maps", len(bus.evtMaps))
	}

	var evtMap [EVENT_TYPES]*map[*Conn]bool
	var connSet map[*Conn]bool
	var node *PathTreeNode

	for path, evt := range paths {
		node = bus.pathTree.get(path)
		if node == nil {
			t.Error("Path tree didn't have", path)
			continue
		}

		if _, exists := bus.evtMaps[node]; !exists {
			t.Error("Event map was not created for", path)
			continue
		} else {
			evtMap = *(bus.evtMaps[node])
		}

		if evtMap[evt] == nil {
			t.Error("Event was not allocated in event map for", path)
			continue
		} else {
			connSet = *(evtMap[evt])
		}

		if _, exists := connSet[&conn]; !exists {
			t.Error("Event map did not register for", path)
		}
	}
}

func TestPublish(t *testing.T) {
	bus := NewMsgBus()

	conn1 := Conn{
		id:            1,
		ws:            nil,
		outbox:        make(chan []byte, 256),
		subscriptions: make(map[*map[*Conn]bool]bool),
		hub:           nil,
	}
	conn2 := Conn{
		id:            2,
		ws:            nil,
		outbox:        make(chan []byte, 256),
		subscriptions: make(map[*map[*Conn]bool]bool),
		hub:           nil,
	}

	paths1 := map[string]byte{
		"/a/b/c":       EVENT_TYPE_VALUE,
		"/a/b/c/d":     EVENT_TYPE_CHILD_ADDED,
		"/a/b":         EVENT_TYPE_CHILD_CHANGED,
		"/a/b/c/d/e/f": EVENT_TYPE_CHILD_REMOVED,
		"/1/2/3":       EVENT_TYPE_CHILD_MOVED,
	}
	paths2 := map[string]byte{
		"/a/b/c":   EVENT_TYPE_VALUE,
		"/x/y":     EVENT_TYPE_CHILD_ADDED,
		"/a/b":     EVENT_TYPE_CHILD_CHANGED,
		"/w/x/y/z": EVENT_TYPE_CHILD_REMOVED,
		"/1/2/3":   EVENT_TYPE_CHILD_MOVED,
	}

	for path, evt := range paths1 {
		bus.subscribe(evt, path, &conn1)
	}

	for path, evt := range paths2 {
		bus.subscribe(evt, path, &conn2)
	}

	bus.publish(EVENT_TYPE_VALUE, "/a/b/c", []byte("test1"))
	bus.publish(EVENT_TYPE_CHILD_ADDED, "/a/b/c/d", []byte("test2"))
	bus.publish(EVENT_TYPE_CHILD_ADDED, "/x/y", []byte("test3"))
	bus.publish(EVENT_TYPE_CHILD_CHANGED, "/a/b", []byte("test4"))
	bus.publish(EVENT_TYPE_CHILD_REMOVED, "/a/b/c/d/e/f", []byte("test5"))
	bus.publish(EVENT_TYPE_CHILD_REMOVED, "/w/x/y/z", []byte("test6"))
	bus.publish(EVENT_TYPE_CHILD_MOVED, "/1/2/3", []byte("test7"))

	for i := 0; i < 5; i++ {
		select {
		case <-conn1.outbox:
		default:
			t.Error("Outbox closed on conn1")
		}
	}

	for i := 0; i < 5; i++ {
		select {
		case <-conn2.outbox:
		default:
			t.Error("Outbox closed on conn2")
		}
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := NewMsgBus()
	conn := Conn{
		id:            1,
		ws:            nil,
		outbox:        make(chan []byte),
		subscriptions: make(map[*map[*Conn]bool]bool),
		hub:           nil,
	}
	bus.subscribe(EVENT_TYPE_VALUE, "/a", &conn)
	bus.subscribe(EVENT_TYPE_VALUE, "/a/b", &conn)
	bus.subscribe(EVENT_TYPE_VALUE, "/a/b/c", &conn)
	bus.subscribe(EVENT_TYPE_VALUE, "/a/b/c/d", &conn)
	bus.subscribe(EVENT_TYPE_VALUE, "/a/b/c/d/e", &conn)

	bus.unsubscribe(EVENT_TYPE_VALUE, "/a", &conn)
	bus.unsubscribe(EVENT_TYPE_VALUE, "/a/b", &conn)
	bus.unsubscribe(EVENT_TYPE_VALUE, "/a/b/c", &conn)
	bus.unsubscribe(EVENT_TYPE_VALUE, "/a/b/c/d", &conn)
	bus.unsubscribe(EVENT_TYPE_VALUE, "/a/b/c/d/e", &conn)

	if bus.hasSubscribers(EVENT_TYPE_VALUE, "/a") {
		t.Error("/a had subscribers")
	}
	if bus.hasSubscribers(EVENT_TYPE_VALUE, "/a/b") {
		t.Error("/a/b had subscribers")
	}
	if bus.hasSubscribers(EVENT_TYPE_VALUE, "/a/b/c") {
		t.Error("/a/b/c had subscribers")
	}
	if bus.hasSubscribers(EVENT_TYPE_VALUE, "/a/b/c/d") {
		t.Error("/a/b/c/d had subscribers")
	}
	if bus.hasSubscribers(EVENT_TYPE_VALUE, "/a/b/c//d/e") {
		t.Error("/a/b/c/d/e had subscribers")
	}
}

func TestSubscribeConcurrency(t *testing.T) {
	var wg sync.WaitGroup
	bus := NewMsgBus()

	wg.Add(20)
	for i := 0; i < 20; i++ {
		conn := Conn{
			id:            uint64(i + 1),
			ws:            nil,
			outbox:        make(chan []byte, 256),
			subscriptions: make(map[*map[*Conn]bool]bool),
			hub:           nil,
		}
		go (func() {
			bus.subscribe(EVENT_TYPE_VALUE, "/a/b/c", &conn)
			wg.Done()
		})()
	}

	wg.Wait()
	node := bus.pathTree.get("/a/b/c")
	if node != nil {
		evtMap := *(bus.evtMaps[node])
		connSet := *(evtMap[EVENT_TYPE_VALUE])
		subs := len(connSet)
		if subs != 20 {
			t.Error("Some subscriptions were droppped! Recorded subscriptions:", subs)
		}
	} else {
		t.Error("Node was nil")
	}
}
