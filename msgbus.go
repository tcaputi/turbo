package turbo

import (
	"log"
)

type SubSet struct {
	subs map[*connection]bool
}

type SubSwitch struct {
	value        *SubSet
	childAdded   *SubSet
	childChanged *SubSet
	childMoved   *SubSet
	childRemoved *SubSet
}

type MessageBus struct {
	subSwitches map[string]*SubSwitch
}

var msgBus = &MessageBus{
	subSwitches: make(map[string]*SubSwitch),
}

func (s *SubSwitch) subscribe(eventType string, conn *connection) {
	switch eventType {
	case EVENT_TYPE_VALUE:
		if s.value == nil {
			s.value = &SubSet{subs: make(map[*connection]bool)}
		}
		conn.subs[s.value] = true
		s.value.subs[conn] = true
	case EVENT_TYPE_CHILD_ADDED:
		if s.childAdded == nil {
			s.childAdded = &SubSet{subs: make(map[*connection]bool)}
		}
		conn.subs[s.childAdded] = true
		s.childAdded.subs[conn] = true
	case EVENT_TYPE_CHILD_CHANGED:
		if s.childChanged == nil {
			s.childChanged = &SubSet{subs: make(map[*connection]bool)}
		}
		conn.subs[s.childChanged] = true
		s.childChanged.subs[conn] = true
	case EVENT_TYPE_CHILD_MOVED:
		if s.childMoved == nil {
			s.childMoved = &SubSet{subs: make(map[*connection]bool)}
		}
		conn.subs[s.childMoved] = true
		s.childMoved.subs[conn] = true
	case EVENT_TYPE_CHILD_REMOVED:
		if s.childRemoved == nil {
			s.childRemoved = &SubSet{subs: make(map[*connection]bool)}
		}
		conn.subs[s.childRemoved] = true
		s.childRemoved.subs[conn] = true
	}
}

func (s *SubSwitch) unsubscribe(eventType string, conn *connection) {
	switch eventType {
	case EVENT_TYPE_VALUE:
		if s.value == nil {
			return
		}
		delete(s.value.subs, conn)
	case EVENT_TYPE_CHILD_ADDED:
		if s.childAdded == nil {
			return
		}
		delete(s.childAdded.subs, conn)
	case EVENT_TYPE_CHILD_CHANGED:
		if s.childChanged == nil {
			return
		}
		delete(s.childChanged.subs, conn)
	case EVENT_TYPE_CHILD_MOVED:
		if s.childMoved == nil {
			return
		}
		delete(s.childMoved.subs, conn)
	case EVENT_TYPE_CHILD_REMOVED:
		if s.childRemoved == nil {
			return
		}
		delete(s.childRemoved.subs, conn)
	}
}

func (s *SubSwitch) subscribers(eventType string) map[*connection]bool {
	switch eventType {
	case EVENT_TYPE_VALUE:
		if s.value == nil {
			return nil
		}
		return s.value.subs
	case EVENT_TYPE_CHILD_ADDED:
		if s.childAdded == nil {
			return nil
		}
		return s.childAdded.subs
	case EVENT_TYPE_CHILD_CHANGED:
		if s.childChanged == nil {
			return nil
		}
		return s.childChanged.subs
	case EVENT_TYPE_CHILD_MOVED:
		if s.childMoved == nil {
			return nil
		}
		return s.childMoved.subs
	case EVENT_TYPE_CHILD_REMOVED:
		if s.childRemoved == nil {
			return nil
		}
		return s.childRemoved.subs
	default:
		return nil
	}
}

func (mb *MessageBus) subscribe(path string, eventType string, conn *connection) {
	if mb.subSwitches == nil {
		mb.subSwitches = make(map[string]*SubSwitch)
	}
	if mb.subSwitches[path] == nil {
		mb.subSwitches[path] = &SubSwitch{}
	}
	mb.subSwitches[path].subscribe(eventType, conn)
}

func (mb *MessageBus) unsubscribe(path string, eventType string, conn *connection) {
	if mb.subSwitches == nil {
		return
	}
	if mb.subSwitches[path] == nil {
		return
	}
	mb.subSwitches[path].unsubscribe(eventType, conn)
}

func (mb *MessageBus) unsubscribeAll(conn *connection) {
	for subSet := range conn.subs {
		delete(subSet.subs, conn)
	}
}

func (mb *MessageBus) publish(path string, eventType string, msg []byte) {
	subSwitch := mb.subSwitches[path]
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

func (mb *MessageBus) hasSubscribers(path string, eventType string) bool {
	subSwitch := mb.subSwitches[path]
	if subSwitch == nil {
		return false
	}
	subs := subSwitch.subscribers(eventType)
	if subs == nil {
		return false
	}
	return len(subs) >= 0
}
