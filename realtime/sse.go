// Package realtime provides server-push streaming. SSE is the dependency-free
// core; WebSocket and gRPC ship as plugins. Publish events from controllers,
// actions, or hooks and they stream to connected clients.
package realtime

import (
	"fmt"
	"net/http"
	"sync"
)

// Broker fans out messages to all connected SSE clients.
type Broker struct {
	mu      sync.RWMutex
	clients map[chan string]struct{}
}

// NewBroker creates an SSE broker.
func NewBroker() *Broker { return &Broker{clients: map[chan string]struct{}{}} }

// Publish sends a named event with a data payload to all clients.
func (b *Broker) Publish(event, data string) {
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)
	b.mu.RLock()
	for c := range b.clients {
		select {
		case c <- msg:
		default: // drop for slow clients
		}
	}
	b.mu.RUnlock()
}

// Handler is the SSE endpoint. Mount it (e.g. at /events) and clients connect
// with an EventSource.
func (b *Broker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := make(chan string, 16)
		b.mu.Lock()
		b.clients[ch] = struct{}{}
		b.mu.Unlock()
		defer func() {
			b.mu.Lock()
			delete(b.clients, ch)
			b.mu.Unlock()
		}()

		for {
			select {
			case <-r.Context().Done():
				return
			case msg := <-ch:
				fmt.Fprint(w, msg)
				flusher.Flush()
			}
		}
	}
}
