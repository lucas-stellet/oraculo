package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// NewSPAMiddleware returns a Wails v3 middleware that handles SPA routing.
// It lets /wails/* pass through to the Wails runtime, then applies
// placeholder substitution and shell fallback for Next.js routes.
func NewSPAMiddleware(rawAssets embed.FS) application.Middleware {
	assets, _ := fs.Sub(rawAssets, "frontend/dist")
	fileServer := http.FileServer(http.FS(assets))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Let Wails runtime routes pass through (/wails/runtime.js, IPC, etc.)
			if strings.HasPrefix(r.URL.Path, "/wails") {
				next.ServeHTTP(w, r)
				return
			}

			fsPath := strings.TrimPrefix(r.URL.Path, "/")

			// 1. Try exact file match.
			if fsPath != "" {
				if info, err := fs.Stat(assets, fsPath); err == nil && !info.IsDir() {
					fileServer.ServeHTTP(w, r)
					return
				}
			}

			// 2. Try placeholder substitution for dynamic routes.
			if phPath := withPlaceholders(fsPath); phPath != fsPath {
				if info, err := fs.Stat(assets, phPath); err == nil && !info.IsDir() {
					r2 := r.Clone(r.Context())
					r2.URL.Path = "/" + phPath
					fileServer.ServeHTTP(w, r2)
					return
				}
			}

			// 3. Fallback: serve the SPA shell.
			isRSC := r.URL.Query().Get("_rsc") != ""
			shell := spaShell(r.URL.Path, isRSC)
			r2 := r.Clone(r.Context())
			r2.URL.Path = shell
			fileServer.ServeHTTP(w, r2)
		})
	}
}

// withPlaceholders replaces dynamic route segments with "__placeholder__".
// Known dynamic positions: epics/{id} at index 1, approvals/{approvalId} at index 3,
// and stories/{storyId} at index 3 when under /stories/.
func withPlaceholders(fsPath string) string {
	segs := strings.Split(strings.Trim(fsPath, "/"), "/")
	if len(segs) < 2 || segs[0] != "epics" {
		return fsPath
	}
	result := make([]string, len(segs))
	copy(result, segs)
	changed := false
	if result[1] != "__placeholder__" {
		result[1] = "__placeholder__"
		changed = true
	}
	if len(result) >= 4 && result[2] == "approvals" && result[3] != "__placeholder__" {
		result[3] = "__placeholder__"
		changed = true
	}
	if len(result) >= 4 && result[2] == "stories" && result[3] != "__placeholder__" {
		result[3] = "__placeholder__"
		changed = true
	}
	if !changed {
		return fsPath
	}
	return strings.Join(result, "/")
}

// spaShell maps a URL path to the most appropriate pre-rendered shell file.
func spaShell(urlPath string, isRSC bool) string {
	ext := ".html"
	if isRSC {
		ext = ".txt"
	}
	segs := splitPath(urlPath)
	n := len(segs)

	// /epics/{id}/approvals/{approvalId}/review
	if n >= 5 && segs[0] == "epics" && segs[2] == "approvals" && segs[4] == "review" {
		return "/epics/__placeholder__/approvals/__placeholder__/review" + ext
	}
	// /epics/{id}/approvals
	if n >= 3 && segs[0] == "epics" && segs[2] == "approvals" {
		return "/epics/__placeholder__/approvals" + ext
	}
	// /epics/{id}/stories/{storyId}
	if n >= 4 && segs[0] == "epics" && segs[2] == "stories" {
		return "/epics/__placeholder__/stories/__placeholder__" + ext
	}
	// /epics/{id}
	if n >= 2 && segs[0] == "epics" {
		return "/epics/__placeholder__" + ext
	}
	return "/"
}

func splitPath(p string) []string {
	var segs []string
	for _, s := range strings.Split(strings.Trim(p, "/"), "/") {
		if s != "" {
			segs = append(segs, s)
		}
	}
	return segs
}
