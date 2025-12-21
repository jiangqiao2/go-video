package sse

// NewSyncMapHub is kept for benchmark/backwards compatibility.
// It currently returns the default Hub implementation, which uses
// sync.Map internally for subscriber storage.
func NewSyncMapHub() *Hub {
	return NewHub()
}
