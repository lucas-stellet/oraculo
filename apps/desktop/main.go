package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "Oraculo",
		Description: "Oraculo Desktop — multi-project launcher",
		Services: []application.Service{
			application.NewService(NewLauncherService()),
			application.NewService(NewServerService()),
		},
		Assets: application.AssetOptions{
			Handler:    application.BundledAssetFileServer(assets),
			Middleware: NewSPAMiddleware(assets),
		},
		Mac: application.MacOptions{
			ActivationPolicy:                                application.ActivationPolicyRegular,
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		Windows: application.WindowsOptions{
			DisableQuitOnLastWindowClosed: true,
		},
	})

	launcherWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Oraculo",
		Name:   "launcher",
		Width:  1024,
		Height: 700,
		URL:    "/",
		Mac: application.MacWindow{
			TitleBar:                application.MacTitleBarHidden,
			InvisibleTitleBarHeight: 28,
		},
	})

	setupTray(app, launcherWindow)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
