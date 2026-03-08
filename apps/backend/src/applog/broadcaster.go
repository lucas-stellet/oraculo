// apps/backend/src/applog/broadcaster.go
package applog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const ringSize = 200

// entry is the JSON form of a log record sent over SSE.
type entry struct {
	Time  string `json:"time"`
	Level string `json:"level"`
	Msg   string `json:"msg"`
}

// Broadcaster implements slog.Handler. It writes formatted text to out
// (typically os.Stderr) and fans out JSON entries to SSE subscribers via a
// ring buffer for replay.
type Broadcaster struct {
	out  io.Writer
	mu   sync.Mutex
	ring [ringSize][]byte
	head int // next write position (mod ringSize)
	size int // number of valid entries (capped at ringSize)
	subs map[chan []byte]struct{}
}

// NewBroadcaster creates a Broadcaster that writes text to out and serves SSE.
func NewBroadcaster(out io.Writer) *Broadcaster {
	return &Broadcaster{
		out:  out,
		subs: make(map[chan []byte]struct{}),
	}
}

// Enabled implements slog.Handler. All levels are enabled.
func (b *Broadcaster) Enabled(_ context.Context, _ slog.Level) bool { return true }

// Handle implements slog.Handler. Formats the record, writes text to out,
// stores in ring buffer, and notifies all active SSE subscribers.
func (b *Broadcaster) Handle(_ context.Context, r slog.Record) error {
	e := entry{
		Time:  r.Time.UTC().Format(time.RFC3339),
		Level: r.Level.String(),
		Msg:   r.Message,
	}
	raw, _ := json.Marshal(e)
	fmt.Fprintf(b.out, "%s %s %s\n", e.Time, e.Level, e.Msg)

	b.mu.Lock()
	b.ring[b.head%ringSize] = raw
	b.head++
	if b.size < ringSize {
		b.size++
	}
	for ch := range b.subs {
		select {
		case ch <- raw:
		default: // slow subscriber: drop
		}
	}
	b.mu.Unlock()

	return nil
}

// WithAttrs implements slog.Handler (no-op: structured attrs not needed yet).
func (b *Broadcaster) WithAttrs(_ []slog.Attr) slog.Handler { return b }

// WithGroup implements slog.Handler (no-op).
func (b *Broadcaster) WithGroup(_ string) slog.Handler { return b }

// subscribe registers a new SSE subscriber and returns its delivery channel
// plus a snapshot of the ring buffer for replay.
func (b *Broadcaster) subscribe() (chan []byte, [][]byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan []byte, 32)
	b.subs[ch] = struct{}{}

	start := 0
	if b.size == ringSize {
		start = b.head % ringSize
	}
	replay := make([][]byte, b.size)
	for i := range b.size {
		replay[i] = b.ring[(start+i)%ringSize]
	}
	return ch, replay
}

// unsubscribe removes and closes a subscriber channel.
func (b *Broadcaster) unsubscribe(ch chan []byte) {
	b.mu.Lock()
	delete(b.subs, ch)
	b.mu.Unlock()
	close(ch)
}

// ServeSSE handles GET /logs as an SSE stream. Replays buffered entries to
// new subscribers, then streams live entries until the client disconnects.
func (b *Broadcaster) ServeSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, replay := b.subscribe()
	defer b.unsubscribe(ch)

	for _, msg := range replay {
		fmt.Fprintf(w, "data: %s\n\n", msg)
	}
	flusher.Flush()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
