package main

import (
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// installMenu builds the native application menu. The native menu is always
// enabled so desktop environments that support global menus (for example
// Noctalia/KDE/GNOME extensions) can access it through DBusMenu.
func (s *shell) installMenu() {
	menu := s.app.NewMenu()

	// macOS uses a global application menu for About, Preferences and Quit.
	if runtime.GOOS == "darwin" {
		menu.AddRole(application.AppMenu)
	}

	// File menu
	fileMenu := menu.AddSubmenu("File")
	fileMenu.Add("New Download").SetAccelerator("CmdOrCtrl+N").OnClick(func(ctx *application.Context) {
		if window := s.currentWindow(); window == nil {
			s.trayController.Show()
		}
		s.app.Event.Emit("menu:new-task")
	})
	fileMenu.Add("Open Download Folder").OnClick(func(ctx *application.Context) {
		go s.openDownloadFolder()
	})
	if runtime.GOOS != "darwin" {
		fileMenu.AddSeparator()
		fileMenu.Add("Preferences").SetAccelerator("CmdOrCtrl+P").OnClick(func(ctx *application.Context) {
			if window := s.currentWindow(); window == nil {
				s.trayController.Show()
			}
			s.app.Event.Emit("menu:preferences")
		})
		fileMenu.AddSeparator()
		fileMenu.Add("Quit").SetAccelerator("CmdOrCtrl+Q").OnClick(func(ctx *application.Context) {
			s.Quit()
		})
	}

	// Standard edit operations for text fields.
	menu.AddRole(application.EditMenu)

	// Downloads menu
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

	// View menu
	viewMenu := menu.AddSubmenu("View")
	s.viewMenu = viewMenu
	s.themeRadios = map[string]*application.MenuItem{}

	lightItem := viewMenu.AddRadio("Light Theme", false)
	darkItem := viewMenu.AddRadio("Dark Theme", false)
	systemItem := viewMenu.AddRadio("System Theme", true)
	s.themeRadios["light"] = lightItem
	s.themeRadios["dark"] = darkItem
	s.themeRadios["system"] = systemItem

	lightItem.OnClick(func(ctx *application.Context) {
		s.app.Event.Emit("theme:set", "light")
	})
	darkItem.OnClick(func(ctx *application.Context) {
		s.app.Event.Emit("theme:set", "dark")
	})
	systemItem.OnClick(func(ctx *application.Context) {
		s.app.Event.Emit("theme:set", "system")
	})

	viewMenu.AddSeparator()
	viewMenu.Add("Toggle In-App Menu").OnClick(func(ctx *application.Context) {
		go s.toggleInAppMenu()
	})

	// Window management
	menu.AddRole(application.WindowMenu)

	// Help menu
	helpMenu := menu.AddSubmenu("Help")
	if runtime.GOOS != "darwin" {
		helpMenu.Add("Keyboard Shortcuts").OnClick(func(ctx *application.Context) {
			if window := s.currentWindow(); window == nil {
				s.trayController.Show()
			}
			s.app.Event.Emit("menu:shortcuts")
		})
		helpMenu.AddSeparator()
		helpMenu.Add("About AriaDM").OnClick(func(ctx *application.Context) {
			if window := s.currentWindow(); window == nil {
				s.trayController.Show()
			}
			s.app.Event.Emit("menu:about")
		})
	}

	s.app.Menu.Set(menu)

	// Keep native radio selection in sync when the theme changes from the UI.
	s.app.Event.On("theme:changed", func(e *application.CustomEvent) {
		mode := ""
		switch v := e.Data.(type) {
		case string:
			mode = v
		case []interface{}:
			if len(v) > 0 {
				if str, ok := v[0].(string); ok {
					mode = str
				}
			}
		}
		if mode == "" {
			mode = "system"
		}
		for key, item := range s.themeRadios {
			item.SetChecked(key == mode)
		}
		if s.viewMenu != nil {
			s.viewMenu.Update()
		}
	})
}

func (s *shell) SyncTheme(mode string) {
	if mode == "" {
		mode = "system"
	}
	for key, item := range s.themeRadios {
		item.SetChecked(key == mode)
	}
	if s.viewMenu != nil {
		s.viewMenu.Update()
	}
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
