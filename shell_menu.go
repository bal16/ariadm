package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// installMenu builds the native application menu. The native menu is always
// enabled so desktop environments that support global menus (for example
// Noctalia/KDE/GNOME extensions) can access it through DBusMenu.
func (s *shell) installMenu() {
	menu := s.app.NewMenu()

	fileMenu := menu.AddSubmenu("File")
	fileMenu.Add("New Download").SetAccelerator("Ctrl+N").OnClick(func(ctx *application.Context) {
		if window := s.currentWindow(); window == nil {
			s.trayController.Show()
		}
		s.app.Event.Emit("menu:new-task")
	})
	fileMenu.Add("Open Download Folder").OnClick(func(ctx *application.Context) {
		go s.openDownloadFolder()
	})
	fileMenu.Add("Preferences").SetAccelerator("Ctrl+P").OnClick(func(ctx *application.Context) {
		if window := s.currentWindow(); window == nil {
			s.trayController.Show()
		}
		s.app.Event.Emit("menu:preferences")
	})
	fileMenu.AddSeparator()
	fileMenu.Add("Quit").SetAccelerator("Ctrl+Q").OnClick(func(ctx *application.Context) {
		s.Quit()
	})

	downloadsMenu := menu.AddSubmenu("Downloads")
	downloadsMenu.Add("Pause All").OnClick(func(ctx *application.Context) {
		if s.taskService != nil {
			_ = s.taskService.PauseAllTasks()
		}
	})
	downloadsMenu.Add("Resume All").OnClick(func(ctx *application.Context) {
		if s.taskService != nil {
			_ = s.taskService.ResumeAllTasks()
		}
	})

	viewMenu := menu.AddSubmenu("View")
	viewMenu.Add("Light Theme").OnClick(func(ctx *application.Context) {
		s.app.Event.Emit("theme:set", "light")
	})
	viewMenu.Add("Dark Theme").OnClick(func(ctx *application.Context) {
		s.app.Event.Emit("theme:set", "dark")
	})
	viewMenu.Add("System Theme").OnClick(func(ctx *application.Context) {
		s.app.Event.Emit("theme:set", "system")
	})
	viewMenu.AddSeparator()
	viewMenu.Add("Toggle In-App Menu").OnClick(func(ctx *application.Context) {
		go s.toggleInAppMenu()
	})

	helpMenu := menu.AddSubmenu("Help")
	helpMenu.Add("Keyboard Shortcuts").OnClick(func(ctx *application.Context) {
		if window := s.currentWindow(); window == nil {
			s.trayController.Show()
		}
		s.app.Event.Emit("menu:shortcuts")
	})
	helpMenu.Add("About AriaDM").OnClick(func(ctx *application.Context) {
		if window := s.currentWindow(); window == nil {
			s.trayController.Show()
		}
		s.app.Event.Emit("menu:about")
	})

	s.app.Menu.Set(menu)
}

func (s *shell) openDownloadFolder() {
	if s.configService == nil {
		return
	}
	cfg, err := s.configService.GetConfig()
	if err != nil || cfg == nil || cfg.DefaultDownloadPath == "" {
		return
	}
	if s.bridge != nil {
		_ = s.bridge.OpenPath(cfg.DefaultDownloadPath)
	}
}

func (s *shell) toggleInAppMenu() {
	if s.configService == nil {
		return
	}
	cfg, err := s.configService.GetConfig()
	if err != nil || cfg == nil {
		return
	}
	cfg.ShowInAppMenu = !cfg.ShowInAppMenu
	if err := s.configService.UpdateSettings(cfg); err != nil {
		return
	}
	s.app.Event.Emit("config:updated", cfg)
}
