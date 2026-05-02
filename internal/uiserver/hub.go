package uiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Hub struct {
	mu      sync.Mutex
	clients map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: map[chan []byte]struct{}{}}
}

// Broadcast serializes evt and pushes to all subscribers. Returns the
// number of clients delivered to.
func (h *Hub) Broadcast(evt EventEnvelope) int {
	if _, ok := evt["ts"]; !ok {
		evt["ts"] = time.Now().UnixMilli()
	}
	body, err := json.Marshal(evt)
	if err != nil {
		fmt.Printf("[hub] marshal: %v\n", err)
		return 0
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	delivered := 0
	for c := range h.clients {
		select {
		case c <- body:
			delivered++
		default:
			// drop on slow consumer
		}
	}
	return delivered
}

// ServeHTTP is the /api/events handler. Writes SSE frames until the
// client disconnects. Each frame is one `data: <json>\n\n` block.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable proxy buffering

	ch := make(chan []byte, 64)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, ch)
		h.mu.Unlock()
	}()

	// Initial hello so the client knows the stream is live.
	fmt.Fprintf(w, "data: {\"kind\":\"hello\",\"ts\":%d}\n\n", time.Now().UnixMilli())
	flusher.Flush()

	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case body := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", body)
			flusher.Flush()
		case <-keepalive.C:
			// SSE comment frame; clients ignore it but it keeps proxies
			// from killing idle connections.
			fmt.Fprintf(w, ": keepalive %d\n\n", time.Now().Unix())
			flusher.Flush()
		}
	}
}
