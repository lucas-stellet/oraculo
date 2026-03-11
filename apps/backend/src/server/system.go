package server

import (
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type SystemHandler struct {
	startedAt  time.Time
	binaryPath string
	version    string
}

func newSystemHandler(version string) *SystemHandler {
	binaryPath, _ := os.Executable()
	return &SystemHandler{
		startedAt:  time.Now(),
		binaryPath: binaryPath,
		version:    version,
	}
}

func (h *SystemHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	info, err := os.Stat(h.binaryPath)
	updateAvailable := false
	if err == nil {
		updateAvailable = info.ModTime().After(h.startedAt)
	}
	newVersion := ""
	if updateAvailable {
		newVersion = binaryVersion(h.binaryPath)
	}
	writeJSON(w, map[string]any{
		"update_available": updateAvailable,
		"started_at":       h.startedAt,
		"version":          h.version,
		"project_commit":   projectCommit(),
		"new_version":      newVersion,
	})
}

func (h *SystemHandler) handleRestart(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "restarting"})
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		env := append(os.Environ(), "ORACULO_NO_BROWSER=1")
		_ = syscall.Exec(h.binaryPath, []string{h.binaryPath, "start", "http"}, env)
	}()
}

// projectCommit returns the short git commit hash of the current working directory,
// or an empty string if git is unavailable or not a repository.
func projectCommit() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// binaryVersion runs "<binary> version" and returns its output.
func binaryVersion(binaryPath string) string {
	out, err := exec.Command(binaryPath, "version").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
