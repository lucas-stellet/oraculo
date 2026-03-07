// Package server provides the HTTP server for Oraculo.
package server

import "embed"

// DashboardAssets contains the embedded Next.js static files.
//go:embed dashboard_assets
var DashboardAssets embed.FS
