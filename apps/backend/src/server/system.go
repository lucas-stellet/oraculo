package server

import (
	"net/http"
	"os"
	"os/exec"
	"strings"
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
