package main

import (
	"runtime"

	"ariadm/internal/ingress/tray"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// trayWindow adapts a Wails webview window to the tray's narrow Window
// interface.
type trayWindow struct {
	window *application.WebviewWindow
}

func (w trayWindow) Show()           { w.window.Show() }
func (w trayWindow) Hide()           { w.window.Hide() }
func (w trayWindow) Focus()          { w.window.Focus() }
func (w trayWindow) IsVisible() bool { return w.window.IsVisible() }

// currentTrayWindow exposes the current window to the tray controller, or nil
// when background mode released it.
func (s *shell) currentTrayWindow() tray.Window {
	window := s.currentWindow()
	if window == nil {
		return nil
	}
	return trayWindow{window: window}
}

// createTrayWindow recreates the main window for the tray's Show action.
func (s *shell) createTrayWindow() tray.Window {
	return trayWindow{window: s.newWindow()}
}

// installTray adds the system tray icon and context menu.
func (s *shell) installTray() {
	systray := s.app.SystemTray.New()
	if runtime.GOOS == "darwin" {
		// Template icons adapt automatically to light/dark menu bar.
		systray.SetTemplateIcon(s.opts.trayIconTemplate)
	} else {
		systray.SetIcon(s.opts.trayIcon)
		if len(s.opts.trayIconDark) > 0 {
			systray.SetDarkModeIcon(s.opts.trayIconDark)
		}
	}
	systray.SetTooltip(s.opts.name + " - download manager")
	systray.OnClick(s.trayController.Toggle)

	menu := s.app.NewMenu()
	for _, item := range tray.Menu() {
		if item.Label == "" {
			menu.AddSeparator()
			continue
		}
		action := item.Action
		menu.Add(item.Label).OnClick(func(*application.Context) {
			s.trayController.Handle(action)
		})
	}
	systray.SetMenu(menu)
}
