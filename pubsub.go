package main

const (
    EVENT_TYPE_VALUE            = "value"
    EVENT_TYPE_CHILD_ADDED      = "child_added"
    EVENT_TYPE_CHILD_CHANGED    = "child_changed"
    EVENT_TYPE_CHILD_MOVED      = "child_moved"
    EVENT_TYPE_CHILD_REMOVED    = "child_removed"
)

type subSet struct {
    conns map[*connection]bool
}

type subManager struct {
    value           *subSet
    childAdded      *subSet
    childChanged    *subSet
    childMoved      *subSet
    childRemoved    *subSet
}

func (s *subManager) subscribe(eventType string, conn *connection) {
    switch eventType {
        case EVENT_TYPE_VALUE:
            if s.value == nil { s.value = &subSet{ conns: make(map[*connection]bool) } }
            conn.subs[s.value] = true
            s.value.conns[conn] = true
        case EVENT_TYPE_CHILD_ADDED:
            if s.childAdded == nil { s.childAdded = &subSet{ conns: make(map[*connection]bool) } }
            conn.subs[s.childAdded] = true
            s.childAdded.conns[conn] = true
        case EVENT_TYPE_CHILD_CHANGED:
            if s.childChanged == nil { s.childChanged = &subSet{ conns: make(map[*connection]bool) } }
            conn.subs[s.childChanged] = true
            s.childChanged.conns[conn] = true
        case EVENT_TYPE_CHILD_MOVED:
            if s.childMoved == nil { s.childMoved = &subSet{ conns: make(map[*connection]bool) } }
            conn.subs[s.childMoved] = true
            s.childMoved.conns[conn] = true
        case EVENT_TYPE_CHILD_REMOVED:
            if s.childRemoved == nil { s.childRemoved = &subSet{ conns: make(map[*connection]bool) } }
            conn.subs[s.childRemoved] = true
            s.childRemoved.conns[conn] = true
    }
}

func (s *subManager) unsubscribe(eventType string, conn *connection) {
    var subList map[*connection]bool
    switch eventType {
        case EVENT_TYPE_VALUE:
            if s.value == nil return
            delete(s.value.conns, conn)
        case EVENT_TYPE_CHILD_ADDED:
            if s.childAdded == nil return
            delete(s.childAdded.conns, conn)
        case EVENT_TYPE_CHILD_CHANGED:
            if s.childChanged == nil return
            delete(s.childChanged.conns, conn)
        case EVENT_TYPE_CHILD_MOVED:
            if s.childMoved == nil return
            delete(s.childMoved.conns, conn)
        case EVENT_TYPE_CHILD_REMOVED:
            if s.childRemoved == nil return
            delete(s.childRemoved.conns, conn)
    }
}

func (s *subManager) subscribers(eventType string) map[*connection]bool {
    switch eventType {
        case EVENT_TYPE_VALUE:
            if s.value == nil return nil
            return s.value.conns
        case EVENT_TYPE_CHILD_ADDED:
            if s.childAdded == nil return nil
            return s.childAdded.conns
        case EVENT_TYPE_CHILD_CHANGED:
            if s.childChanged == nil return nil
            return s.childChanged.conns
        case EVENT_TYPE_CHILD_MOVED:
            if s.childMoved == nil return nil
            return s.childMoved.conns
        case EVENT_TYPE_CHILD_REMOVED:
            if s.childRemoved == nil return nil
            return s.childRemoved.conns
        default:
            return nil
    }
}

type pubSubManager struct {
    managers map[string]*subManager
}

func (p *pubSubManager) subscribe(path string, eventType string, conn *connection) {
    if p.managers == nil {
        p.managers = make(map[string]*subManager)
    }
    if !p.managers[path] {
        p.managers[path] = &make(subManager)
    }
    p.managers[path].subscribe(eventType, conn)
}

func (p *pubSubManager) unsubscribe(path string, eventType string, conn *connection) {
    if p.managers == nil {
        return
    }
    if !p.managers[path] {
        return
    }
    p.managers[path].unsubscribe(eventType, conn)
}

func (p *pubSubManager) unsubscribeAll(conn *connection) {
    for subSet := range conn.subs {
        delete(subSet, conn)
    }
}

func (p *pubSubManager) publish(path string, eventType string, msg []byte) {
    if subManager := p.managers[path] {
        for sub := range subManager.subscribers(eventType) {
            sub.send <- msg
        }
    }
}

var pubSub = &pubSubManager {
    managers: make(map[string]*subManager)
}