package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type wsMonitor struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func newWSMonitor() *wsMonitor {
	return &wsMonitor{cancels: make(map[string]context.CancelFunc)}
}

func (m *wsMonitor) Connect(project string, port int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.cancels[project]; ok {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancels[project] = cancel

	go m.listen(ctx, project, port)
}

func (m *wsMonitor) Disconnect(project string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cancel, ok := m.cancels[project]; ok {
		cancel()
		delete(m.cancels, project)
	}
}

func (m *wsMonitor) DisconnectAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cancel := range m.cancels {
		cancel()
	}
	m.cancels = make(map[string]context.CancelFunc)
}

func (m *wsMonitor) listen(ctx context.Context, project string, port int) {
	url := fmt.Sprintf("ws://localhost:%d/ws", port)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, _, err := websocket.Dial(ctx, url, nil)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		m.readLoop(ctx, conn, project)
		conn.CloseNow()
	}
}

type wsEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data,omitempty"`
}

func (m *wsMonitor) readLoop(ctx context.Context, conn *websocket.Conn, project string) {
	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			return
		}

		var evt wsEvent
		if err := json.Unmarshal(msg, &evt); err != nil {
			continue
		}

		switch evt.Event {
		case "approval_requested":
			notifyApprovalPending(project)
		case "story_completed":
			notifyStoryCompleted(project, "")
		case "epic_completed":
			notifyEpicCompleted(project, "")
		}
	}
}
