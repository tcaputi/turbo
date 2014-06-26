package turbo

import (
	"log"
)

type SubscriberSet struct {
	subs map[*connection]bool
}

type EventMap struct {
	value        *SubscriberSet
	childAdded   *SubscriberSet
	childChanged *SubscriberSet
	childMoved   *SubscriberSet
	childRemoved *SubscriberSet
}

type MessageBus struct {
	pathMap map[string]*EventMap
}

var (
	msgBus = &MessageBus{
		pathMap: make(map[string]*EventMap),
	}
)

func (evtMap *EventMap) subscribe(eventType byte, conn *connection) {
	switch eventType {
	case EVENT_TYPE_VALUE:
		if evtMap.value == nil {
			evtMap.value = &SubscriberSet{subs: make(map[*connection]bool)}
		}
		conn.subs[evtMap.value] = true
		evtMap.value.subs[conn] = true
	case EVENT_TYPE_CHILD_ADDED:
		if evtMap.childAdded == nil {
			evtMap.childAdded = &SubscriberSet{subs: make(map[*connection]bool)}
		}
		conn.subs[evtMap.childAdded] = true
		evtMap.childAdded.subs[conn] = true
	case EVENT_TYPE_CHILD_CHANGED:
		if evtMap.childChanged == nil {
			evtMap.childChanged = &SubscriberSet{subs: make(map[*connection]bool)}
		}
		conn.subs[evtMap.childChanged] = true
		evtMap.childChanged.subs[conn] = true
	case EVENT_TYPE_CHILD_MOVED:
		if evtMap.childMoved == nil {
			evtMap.childMoved = &SubscriberSet{subs: make(map[*connection]bool)}
		}
		conn.subs[evtMap.childMoved] = true
		evtMap.childMoved.subs[conn] = true
	case EVENT_TYPE_CHILD_REMOVED:
		if evtMap.childRemoved == nil {
			evtMap.childRemoved = &SubscriberSet{subs: make(map[*connection]bool)}
		}
		conn.subs[evtMap.childRemoved] = true
		evtMap.childRemoved.subs[conn] = true
	}
}

func (evtMap *EventMap) unsubscribe(eventType byte, conn *connection) {
	switch eventType {
	case EVENT_TYPE_VALUE:
		if evtMap.value == nil {
			return
		}
		delete(evtMap.value.subs, conn)
	case EVENT_TYPE_CHILD_ADDED:
		if evtMap.childAdded == nil {
			return
		}
		delete(evtMap.childAdded.subs, conn)
	case EVENT_TYPE_CHILD_CHANGED:
		if evtMap.childChanged == nil {
			return
		}
		delete(evtMap.childChanged.subs, conn)
	case EVENT_TYPE_CHILD_MOVED:
		if evtMap.childMoved == nil {
			return
		}
		delete(evtMap.childMoved.subs, conn)
	case EVENT_TYPE_CHILD_REMOVED:
		if evtMap.childRemoved == nil {
			return
		}
		delete(evtMap.childRemoved.subs, conn)
	}
}

func (evtMap *EventMap) subscribers(eventType byte) map[*connection]bool {
	switch eventType {
	case EVENT_TYPE_VALUE:
		if evtMap.value == nil {
			return nil
		}
		return evtMap.value.subs
	case EVENT_TYPE_CHILD_ADDED:
		if evtMap.childAdded == nil {
			return nil
		}
		return evtMap.childAdded.subs
	case EVENT_TYPE_CHILD_CHANGED:
		if evtMap.childChanged == nil {
			return nil
		}
		return evtMap.childChanged.subs
	case EVENT_TYPE_CHILD_MOVED:
		if evtMap.childMoved == nil {
			return nil
		}
		return evtMap.childMoved.subs
	case EVENT_TYPE_CHILD_REMOVED:
		if evtMap.childRemoved == nil {
			return nil
		}
		return evtMap.childRemoved.subs
	default:
		return nil
	}
}

func (mb *MessageBus) subscribe(path string, eventType byte, conn *connection) {
	if mb.pathMap == nil {
		mb.pathMap = make(map[string]*EventMap)
	}
	if mb.pathMap[path] == nil {
		mb.pathMap[path] = &EventMap{}
	}
	mb.pathMap[path].subscribe(eventType, conn)
}

func (mb *MessageBus) unsubscribe(path string, eventType byte, conn *connection) {
	if mb.pathMap == nil {
		return
	}
	if mb.pathMap[path] == nil {
		return
	}
	mb.pathMap[path].unsubscribe(eventType, conn)
}

func (mb *MessageBus) unsubscribeAll(conn *connection) {
	for subSet := range conn.subs {
		delete(subSet.subs, conn)
	}
}

func (mb *MessageBus) publish(path string, eventType byte, msg []byte) {
	subSwitch := mb.pathMap[path]
	if subSwitch != nil {
		for sub := range subSwitch.subscribers(eventType) {
			select {
			case sub.outbox <- msg:
			default:
				defer sub.kill()
			}
		}
	} else {
		log.Println("Could not find a sub switch for path", path)
	}
}

func (mb *MessageBus) hasSubscribers(path string, eventType byte) bool {
	subSwitch := mb.pathMap[path]
	if subSwitch == nil {
		return false
	}
	subs := subSwitch.subscribers(eventType)
	if subs == nil {
		return false
	}
	return len(subs) >= 0
}
