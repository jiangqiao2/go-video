package sse

import "sync"

// Event represents a server-sent notification event payload.
// Type is used as SSE "event:" name, Data is an arbitrary JSON-serialisable body.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// Hub keeps in-memory SSE subscribers grouped by user.
// This hub is process-local and intended for single-instance or dev environments.
// Internally it uses sync.Map to minimise lock contention at high scale.
type Hub struct {
	// subscribers maps user UUID -> *sync.Map representing a set of channels.
	subscribers sync.Map // map[string]*sync.Map
}

// NewHub constructs a Hub.
func NewHub() *Hub {
	return &Hub{}
}

var defaultHub = NewHub()

// DefaultHub exposes the process-global hub.
func DefaultHub() *Hub {
	return defaultHub
}

// Subscribe registers a user-specific subscriber and returns a channel
// plus an unsubscribe function that should be called on disconnect.
func (h *Hub) Subscribe(userUUID string) (<-chan Event, func()) {
	ch := make(chan Event, 16)

	// Lazily create the inner set for this user.
	v, _ := h.subscribers.LoadOrStore(userUUID, &sync.Map{})
	inner := v.(*sync.Map)
	inner.Store(ch, struct{}{})

	unsubscribe := func() {
		inner.Delete(ch)
		close(ch)
		// Note: we intentionally do not remove empty inner maps from
		// the outer subscribers map to keep implementation simple.
	}

	return ch, unsubscribe
}

// Publish sends an event to all subscribers of the given user.
// Slow consumers are skipped to avoid blocking producer code.
func (h *Hub) Publish(userUUID string, ev Event) {
	v, ok := h.subscribers.Load(userUUID)
	if !ok {
		return
	}
	inner := v.(*sync.Map)

	inner.Range(func(key, _ interface{}) bool {
		ch, ok := key.(chan Event)
		if !ok {
			return true
		}
		select {
		case ch <- ev:
		default:
			// drop if subscriber is slow
		}
		return true
	})
}
