// apps/backend/src/applog/broadcaster_test.go
package applog_test

import (
	"bufio"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/applog"
)

// TestBroadcaster_LiveEntry verifies that a log entry emitted after a subscriber
// connects is delivered over SSE.
func TestBroadcaster_LiveEntry(t *testing.T) {
	b := applog.NewBroadcaster(io.Discard)
	logger := slog.New(b)

	srv := httptest.NewServer(http.HandlerFunc(b.ServeSSE))
	defer srv.Close()

	// http.Get blocks until response headers are received. By that point the
	// handler has already called subscribe(), so the channel is registered.
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET /logs: %v", err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("Content-Type = %q, want text/event-stream", ct)
	}

	logger.Info("mcp.connected")

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			if !strings.Contains(line, "mcp.connected") {
				t.Errorf("SSE line %q does not contain mcp.connected", line)
			}
			if !strings.Contains(line, `"level":"INFO"`) {
				t.Errorf("SSE line %q does not contain INFO level", line)
			}
			return
		}
	}
	t.Fatal("no SSE data line received")
}

// TestBroadcaster_ReplayOnConnect verifies that entries emitted before a
// subscriber connects are replayed immediately on connection.
func TestBroadcaster_ReplayOnConnect(t *testing.T) {
	b := applog.NewBroadcaster(io.Discard)
	logger := slog.New(b)

	// Emit before any subscriber exists — goes into ring buffer only.
	logger.Info("server.started")

	srv := httptest.NewServer(http.HandlerFunc(b.ServeSSE))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET /logs: %v", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "server.started") {
			return
		}
	}
	t.Error("expected replay of server.started on connect")
}
