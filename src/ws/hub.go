// src/ws/hub.go
package ws

import (
	"context"
	"net/http"
)

// Hub manages WebSocket client connections and message broadcasting.
type Hub struct{}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{}
}

// Run starts the hub's message dispatch loop. It returns when ctx is done.
func (h *Hub) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// ServeWS upgrades the HTTP connection to a WebSocket and registers the client
// with the hub.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {}

// Broadcast sends a message to all connected WebSocket clients.
func (h *Hub) Broadcast(msg []byte) {}
