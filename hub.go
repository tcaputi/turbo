package main

type hub struct {
	// Registered connections.
	connections map[*connection]bool	// *NOTE: in go, its a popular pattern to emulate a set wih map[*]bool
	// Inbound messages from the connections.
	broadcast chan []byte
	// Register requests from the connections.
	register chan *connection
	// Unregister requests from connections.
	unregister chan *connection
}

var h = hub {
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}

func (h *hub) run() {
	for {
		select {
		// There is a connection 'c' in the registration queue
		case c := <-h.register:
			h.connections[c] = true
		// There is a connection 'c' in the unregistration queue
		case c := <-h.unregister:
			delete(h.connections, c)	// Remove this connection from the connection map
			close(c.send)				// Kill this connection's outbox
		// There is a message 'm' in the broadcast queue
		case m := <-h.broadcast:
			// Put the message in the outboxes of all the connections
			for c := range h.connections {
				// TODO: figure out how this whole block operates
				select {
				case c.send <- m:
				default:
					// Unregister this connection right off the bat
					delete(h.connections, c)
					close(c.send)
				}
			}
		}
	}
}