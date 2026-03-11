package main

import (
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

func setupTray(app *application.App, launcherWindow *application.WebviewWindow) {
	menu := application.NewMenu()
	menu.Add("Open Oraculo").OnClick(func(ctx *application.Context) {
		launcherWindow.Show()
		launcherWindow.Focus()
	})
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	tray := app.SystemTray.New()
	tray.SetTooltip("Oraculo")
	tray.SetMenu(menu)
	if runtime.GOOS == "darwin" {
		tray.SetTemplateIcon(trayTemplateData)
	} else {
		tray.SetIcon(trayIconData)
	}
	tray.AttachWindow(launcherWindow)

	// Hide the window instead of closing it when the user clicks the close button.
	launcherWindow.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		launcherWindow.Hide()
		event.Cancel()
	})

	// On macOS, clicking the dock icon should show the launcher window.
	if runtime.GOOS == "darwin" {
		app.Event.OnApplicationEvent(events.Mac.ApplicationShouldHandleReopen, func(event *application.ApplicationEvent) {
			if !launcherWindow.IsVisible() {
				launcherWindow.Show()
				launcherWindow.Focus()
			}
		})
	}
}
