package canvas

import (
	"sync"
	"sync/atomic"
)

// Dispatcher manages event subscribers for a figure canvas.
type Dispatcher struct {
	next     atomic.Int64
	mu       sync.RWMutex
	handlers map[EventType]map[ConnectionID]Handler
}

// Connect registers a handler for one event type.
func (d *Dispatcher) Connect(eventType EventType, handler Handler) ConnectionID {
	if handler == nil {
		return 0
	}

	id := nextConnectionID(&d.next)
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.handlers == nil {
		d.handlers = make(map[EventType]map[ConnectionID]Handler)
	}
	if d.handlers[eventType] == nil {
		d.handlers[eventType] = make(map[ConnectionID]Handler)
	}
	d.handlers[eventType][id] = handler
	return id
}

// Disconnect removes a registered handler.
func (d *Dispatcher) Disconnect(id ConnectionID) {
	if id == 0 {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	for eventType, handlers := range d.handlers {
		if _, ok := handlers[id]; !ok {
			continue
		}
		delete(handlers, id)
		if len(handlers) == 0 {
			delete(d.handlers, eventType)
		}
		return
	}
}

// Emit dispatches an event to all handlers registered for its type.
func (d *Dispatcher) Emit(event Event) error {
	d.mu.RLock()
	handlers := make([]Handler, 0, len(d.handlers[event.Type]))
	for _, handler := range d.handlers[event.Type] {
		handlers = append(handlers, handler)
	}
	d.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			return err
		}
	}
	return nil
}
