package sse

import (
	"sync"
)

// Event represents a server-sent notification event payload.
// Type is used as SSE "event:" name, Data is an arbitrary JSON-serialisable body.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// Hub keeps in-memory subscribers for each user.
// This is process-local and intended for single-instance or dev environments.
type Hub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan Event]struct{}
}

// NewHub constructs a Hub.
func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string]map[chan Event]struct{}),
	}
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

	h.mu.Lock()
	if h.subscribers == nil {
		h.subscribers = make(map[string]map[chan Event]struct{})
	}
	if h.subscribers[userUUID] == nil {
		h.subscribers[userUUID] = make(map[chan Event]struct{})
	}
	h.subscribers[userUUID][ch] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		subs := h.subscribers[userUUID]
		if subs != nil {
			if _, ok := subs[ch]; ok {
				delete(subs, ch)
				close(ch)
			}
			if len(subs) == 0 {
				delete(h.subscribers, userUUID)
			}
		}
	}

	return ch, unsubscribe
}

// Publish sends an event to all subscribers of the given user.
// Slow consumers are skipped to avoid blocking producer code.
func (h *Hub) Publish(userUUID string, ev Event) {
	h.mu.RLock()
	subs := h.subscribers[userUUID]
	if len(subs) == 0 {
		h.mu.RUnlock()
		return
	}
	// copy keys to avoid holding lock while sending
	chans := make([]chan Event, 0, len(subs))
	for ch := range subs {
		chans = append(chans, ch)
	}
	h.mu.RUnlock()

	for _, ch := range chans {
		select {
		case ch <- ev:
		default:
			// drop if subscriber is slow
		}
	}
}
