package main

import (
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// installMenu builds the native application menu. The menu is rendered by Wails
// as a GtkPopoverMenuBar inside the window; GTK4 does not export it over D-Bus,
// so a global-menu bar (such as Noctalia's) will not see it.
func (s *shell) installMenu() {
	menu := s.app.NewMenu()
	s.menu = menu

	// macOS uses a global application menu for About, Preferences and Quit.
	if runtime.GOOS == "darwin" {
		menu.AddRole(application.AppMenu)
	}

	// File menu
	fileMenu := menu.AddSubmenu("File")
	fileMenu.Add("New Download").SetAccelerator("CmdOrCtrl+N").OnClick(func(ctx *application.Context) {
		go s.showWindowThenEmit("menu:new-task")
	})
	fileMenu.Add("Open Download Folder").OnClick(func(ctx *application.Context) {
		go s.openDownloadFolder()
	})
	if runtime.GOOS != "darwin" {
		fileMenu.AddSeparator()
		fileMenu.Add("Preferences").SetAccelerator("CmdOrCtrl+,").OnClick(func(ctx *application.Context) {
			go s.showWindowThenEmit("menu:preferences")
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
	downloadsMenu.Add("Pause All").SetAccelerator("CmdOrCtrl+Shift+P").OnClick(func(ctx *application.Context) {
		if s.taskService != nil {
			_ = s.taskService.PauseAllTasks()
		}
	})
	downloadsMenu.Add("Resume All").SetAccelerator("CmdOrCtrl+Shift+R").OnClick(func(ctx *application.Context) {
		if s.taskService != nil {
			_ = s.taskService.ResumeAllTasks()
		}
	})

	// View menu
	viewMenu := menu.AddSubmenu("View")
	s.viewMenu = viewMenu
	s.themeRadios = map[string]*application.MenuItem{}

	mode := "system"
	if s.configService != nil {
		if cfg, err := s.configService.GetConfig(); err == nil && cfg != nil && cfg.ThemePreference != "" {
			mode = cfg.ThemePreference
		}
	}

	lightItem := viewMenu.AddCheckbox("Light Theme", mode == "light").SetAccelerator("CmdOrCtrl+Shift+L")
	darkItem := viewMenu.AddCheckbox("Dark Theme", mode == "dark").SetAccelerator("CmdOrCtrl+Shift+D")
	systemItem := viewMenu.AddCheckbox("System Theme", mode == "system").SetAccelerator("CmdOrCtrl+Shift+S")
	s.themeRadios["light"] = lightItem
	s.themeRadios["dark"] = darkItem
	s.themeRadios["system"] = systemItem

	lightItem.OnClick(func(ctx *application.Context) {
		for _, item := range s.themeRadios {
			item.SetChecked(false)
		}
		lightItem.SetChecked(true)
		application.InvokeSync(func() {
			if s.menu != nil {
				s.menu.Update()
			}
		})
		s.app.Event.Emit("theme:set", "light")
	})
	darkItem.OnClick(func(ctx *application.Context) {
		for _, item := range s.themeRadios {
			item.SetChecked(false)
		}
		darkItem.SetChecked(true)
		application.InvokeSync(func() {
			if s.menu != nil {
				s.menu.Update()
			}
		})
		s.app.Event.Emit("theme:set", "dark")
	})
	systemItem.OnClick(func(ctx *application.Context) {
		for _, item := range s.themeRadios {
			item.SetChecked(false)
		}
		systemItem.SetChecked(true)
		application.InvokeSync(func() {
			if s.menu != nil {
				s.menu.Update()
			}
		})
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
			go s.showWindowThenEmit("menu:shortcuts")
		})
		helpMenu.AddSeparator()
		helpMenu.Add("About AriaDM").OnClick(func(ctx *application.Context) {
			go s.showWindowThenEmit("menu:about")
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
		application.InvokeSync(func() {
			if s.menu != nil {
				s.menu.Update()
			}
		})
	})
}

func (s *shell) SyncTheme(mode string) {
	if mode == "" {
		mode = "system"
	}
	for key, item := range s.themeRadios {
		item.SetChecked(key == mode)
	}
	application.InvokeSync(func() {
		if s.menu != nil {
			s.menu.Update()
		}
	})
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

// showWindowThenEmit ensures the main window is visible and mounted before
// emitting a menu event, so the frontend cannot miss it.
func (s *shell) showWindowThenEmit(eventName string) {
	if window := s.currentWindow(); window != nil {
		window.Show()
		window.Focus()
		s.app.Event.Emit(eventName)
		return
	}

	window := s.newWindow()
	window.OnWindowEvent(events.Common.WindowShow, func(e *application.WindowEvent) {
		s.app.Event.Emit(eventName)
	})
	window.Show()
	window.Focus()
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
