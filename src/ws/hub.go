// Package ws provides a WebSocket hub for broadcasting messages to connected clients.
package ws

import (
	"context"
	"net/http"
	"sync"
)

// Hub manages a set of connected WebSocket clients and broadcasts messages to all of them.
type Hub struct {
	mu       sync.RWMutex
	clients  map[*client]struct{}
	broadcast chan []byte
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients:  make(map[*client]struct{}),
		broadcast: make(chan []byte, 256),
	}
}

// Run processes broadcast messages until ctx is cancelled.
// It should be called in a goroutine.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-h.broadcast:
			if !ok {
				return
			}
			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					// Drop the message if the client's buffer is full.
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends msg to all connected WebSocket clients.
func (h *Hub) Broadcast(msg []byte) {
	select {
	case h.broadcast <- msg:
	default:
		// Drop if broadcast channel is full.
	}
}

// ServeWS upgrades the HTTP connection to a WebSocket and registers the client with the hub.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	c, err := newClient(h, w, r)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	go c.writePump()
	go func() {
		c.readPump()
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
	}()
}

// ClientCount returns the number of currently connected WebSocket clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
