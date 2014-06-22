package main

const (
    EVENT_TYPE_VALUE            = "value"
    EVENT_TYPE_CHILD_ADDED      = "child_added"
    EVENT_TYPE_CHILD_CHANGED    = "child_changed"
    EVENT_TYPE_CHILD_MOVED      = "child_moved"
    EVENT_TYPE_CHILD_REMOVED    = "child_removed"
)

type SubSet struct {
    subs map[*connection]bool
}

type SubSwitch struct {
    value           *SubSet
    childAdded      *SubSet
    childChanged    *SubSet
    childMoved      *SubSet
    childRemoved    *SubSet
}

func (s *SubSwitch) subscribe(eventType string, conn *connection) {
    switch eventType {
        case EVENT_TYPE_VALUE:
            if s.value == nil { s.value = &SubSet{ subs: make(map[*connection]bool) } }
            conn.subs[s.value] = true
            s.value.subs[conn] = true
        case EVENT_TYPE_CHILD_ADDED:
            if s.childAdded == nil { s.childAdded = &SubSet{ subs: make(map[*connection]bool) } }
            conn.subs[s.childAdded] = true
            s.childAdded.subs[conn] = true
        case EVENT_TYPE_CHILD_CHANGED:
            if s.childChanged == nil { s.childChanged = &SubSet{ subs: make(map[*connection]bool) } }
            conn.subs[s.childChanged] = true
            s.childChanged.subs[conn] = true
        case EVENT_TYPE_CHILD_MOVED:
            if s.childMoved == nil { s.childMoved = &SubSet{ subs: make(map[*connection]bool) } }
            conn.subs[s.childMoved] = true
            s.childMoved.subs[conn] = true
        case EVENT_TYPE_CHILD_REMOVED:
            if s.childRemoved == nil { s.childRemoved = &SubSet{ subs: make(map[*connection]bool) } }
            conn.subs[s.childRemoved] = true
            s.childRemoved.subs[conn] = true
    }
}

func (s *SubSwitch) unsubscribe(eventType string, conn *connection) {
    switch eventType {
        case EVENT_TYPE_VALUE:
            if s.value == nil { return }
            delete(s.value.subs, conn)
        case EVENT_TYPE_CHILD_ADDED:
            if s.childAdded == nil { return }
            delete(s.childAdded.subs, conn)
        case EVENT_TYPE_CHILD_CHANGED:
            if s.childChanged == nil { return }
            delete(s.childChanged.subs, conn)
        case EVENT_TYPE_CHILD_MOVED:
            if s.childMoved == nil { return }
            delete(s.childMoved.subs, conn)
        case EVENT_TYPE_CHILD_REMOVED:
            if s.childRemoved == nil { return }
            delete(s.childRemoved.subs, conn)
    }
}

func (s *SubSwitch) subscribers(eventType string) map[*connection]bool {
    switch eventType {
        case EVENT_TYPE_VALUE:
            if s.value == nil { return nil }
            return s.value.subs
        case EVENT_TYPE_CHILD_ADDED:
            if s.childAdded == nil { return nil }
            return s.childAdded.subs
        case EVENT_TYPE_CHILD_CHANGED:
            if s.childChanged == nil { return nil }
            return s.childChanged.subs
        case EVENT_TYPE_CHILD_MOVED:
            if s.childMoved == nil { return nil }
            return s.childMoved.subs
        case EVENT_TYPE_CHILD_REMOVED:
            if s.childRemoved == nil { return nil }
            return s.childRemoved.subs
        default:
            return nil
    }
}

type MessageBus struct {
    managers map[string]*SubSwitch
}

func (mb *MessageBus) subscribe(path string, eventType string, conn *connection) {
    if mb.managers == nil {
        mb.managers = make(map[string]*SubSwitch)
    }
    if mb.managers[path] == nil {
        mb.managers[path] = &SubSwitch{}
    }
    mb.managers[path].subscribe(eventType, conn)
}

func (mb *MessageBus) unsubscribe(path string, eventType string, conn *connection) {
    if mb.managers == nil {
        return
    }
    if mb.managers[path] == nil {
        return
    }
    mb.managers[path].unsubscribe(eventType, conn)
}

func (mb *MessageBus) unsubscribeAll(conn *connection) {
    for subSet := range conn.subs {
        delete(subSet.subs, conn)
    }
}

func (mb *MessageBus) publish(path string, eventType string, msg []byte) {
    subSwitch := mb.managers[path]
    if subSwitch != nil {
        for sub := range subSwitch.subscribers(eventType) {
            select {
            case sub.send <- msg:
            default:
                // TODO - kill this particular connection since its dead
            }
        }
    }
}

var msgBus = &MessageBus { managers: make(map[string]*SubSwitch) }