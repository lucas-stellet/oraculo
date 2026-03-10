package server

import (
	"net/http"
	"os"
	"syscall"
	"time"
)

type SystemHandler struct {
	startedAt  time.Time
	binaryPath string
}

func newSystemHandler() *SystemHandler {
	binaryPath, _ := os.Executable()
	return &SystemHandler{
		startedAt:  time.Now(),
		binaryPath: binaryPath,
	}
}

func (h *SystemHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	info, err := os.Stat(h.binaryPath)
	updateAvailable := false
	if err == nil {
		updateAvailable = info.ModTime().After(h.startedAt)
	}
	writeJSON(w, map[string]any{
		"update_available": updateAvailable,
		"started_at":       h.startedAt,
	})
}

func (h *SystemHandler) handleRestart(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "restarting"})
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = syscall.Exec(h.binaryPath, []string{h.binaryPath, "start", "http", "--no-browser"}, os.Environ())
	}()
}
