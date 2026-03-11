package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/lucas/oraculo/apps/backend/src/spa"
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
			if phPath := spa.WithPlaceholders(fsPath); phPath != fsPath {
				if info, err := fs.Stat(assets, phPath); err == nil && !info.IsDir() {
					r2 := r.Clone(r.Context())
					r2.URL.Path = "/" + phPath
					fileServer.ServeHTTP(w, r2)
					return
				}
			}

			// 3. Fallback: serve the SPA shell.
			isRSC := r.URL.Query().Get("_rsc") != ""
			shell := spa.Shell(r.URL.Path, isRSC)
			r2 := r.Clone(r.Context())
			r2.URL.Path = shell
			fileServer.ServeHTTP(w, r2)
		})
	}
}
