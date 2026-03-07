// Package server provides the HTTP server for Oraculo.
package server

import (
	"embed"
	"io/fs"
)

// dashboardAssets contains the raw embedded Next.js static files.
//go:embed dashboard_assets
var dashboardAssets embed.FS

// DashboardAssets is the embedded filesystem with the "dashboard_assets" prefix stripped,
// so that files are served at their expected paths (e.g. /index.html, /_next/...).
var DashboardAssets, _ = fs.Sub(dashboardAssets, "dashboard_assets")
