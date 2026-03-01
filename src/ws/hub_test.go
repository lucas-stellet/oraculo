package ws_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/lucas/oraculo/src/ws"
)

func TestHub_BroadcastToClient(t *testing.T) {
	hub := ws.NewHub()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start hub in background
	go hub.Run(ctx)

	// Start test HTTP server with WS endpoint
	srv := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer srv.Close()

	// Connect a WS client
	wsURL := "ws" + srv.URL[4:] // http -> ws
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.CloseNow()

	// Give client time to register
	time.Sleep(50 * time.Millisecond)

	// Broadcast a message
	msg := map[string]string{"type": "test", "payload": "hello"}
	data, _ := json.Marshal(msg)
	hub.Broadcast(data)

	// Read message from client
	_, received, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	var got map[string]string
	json.Unmarshal(received, &got)
	if got["type"] != "test" {
		t.Errorf("type = %q, want %q", got["type"], "test")
	}
}

func TestHub_BroadcastNonBlocking(t *testing.T) {
	hub := ws.NewHub()
	// Don't start Run — broadcast should not block even without consumers
	hub.Broadcast([]byte("should not block"))
	// If we reach here, the test passes (non-blocking)
}
