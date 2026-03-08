package ws

import (
	"context"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

type client struct {
	conn *websocket.Conn
	send chan []byte
}

// Hub manages WebSocket connections and broadcasts messages.
type Hub struct {
	mu        sync.Mutex
	clients   map[*client]struct{}
	broadcast chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*client]struct{}),
		broadcast: make(chan []byte, 64),
	}
}

// Run processes broadcasts until ctx is cancelled.
func (h *Hub) Run(ctx context.Context) error {
	for {
		select {
		case msg := <-h.broadcast:
			h.mu.Lock()
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					// Client too slow, drop message
				}
			}
			h.mu.Unlock()
		case <-ctx.Done():
			h.mu.Lock()
			for c := range h.clients {
				c.conn.Close(websocket.StatusGoingAway, "server shutdown")
				close(c.send)
				delete(h.clients, c)
			}
			h.mu.Unlock()
			return nil
		}
	}
}

// Broadcast queues a message. Non-blocking: drops if buffer is full.
func (h *Hub) Broadcast(msg []byte) {
	select {
	case h.broadcast <- msg:
	default:
	}
}

// ServeWS upgrades an HTTP connection to WebSocket and blocks until the
// connection is closed. Blocking is required so that the HTTP server keeps
// the underlying TCP connection alive while the WebSocket session is active.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}

	c := &client{
		conn: conn,
		send: make(chan []byte, 16),
	}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
		conn.CloseNow()
	}()

	// connCtx is cancelled when the client disconnects (read side closes).
	connCtx := conn.CloseRead(r.Context())

	// Write loop: deliver messages until the connection closes.
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			if err := conn.Write(r.Context(), websocket.MessageText, msg); err != nil {
				return
			}
		case <-connCtx.Done():
			return
		case <-r.Context().Done():
			return
		}
	}
}
