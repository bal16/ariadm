// Package tray maps operating-system system-tray actions onto the desktop
// shell: showing, hiding and quitting the application.
package tray

import "sync"

// Action is a command exposed by the system-tray menu.
type Action int

const (
	// ActionShow brings the main window to the foreground, recreating it when
	// it was released earlier (for example by background mode).
	ActionShow Action = iota
	// ActionHide releases the main window while keeping the application
	// resident in the system tray.
	ActionHide
	// ActionQuit terminates the application and its background services.
	ActionQuit
)

// String returns a stable identifier for the action.
func (a Action) String() string {
	switch a {
	case ActionShow:
		return "show"
	case ActionHide:
		return "hide"
	case ActionQuit:
		return "quit"
	default:
		return "unknown"
	}
}

// Item is a single system-tray menu entry. An Item with an empty Label is a
// visual separator.
type Item struct {
	Label  string
	Action Action
}

// Menu returns the tray menu layout in display order.
func Menu() []Item {
	return []Item{
		{Label: "Show AriaDM", Action: ActionShow},
		{Label: "Hide to Tray", Action: ActionHide},
		{},
		{Label: "Quit", Action: ActionQuit},
	}
}

// Window is the subset of the desktop window the tray controls.
type Window interface {
	Show()
	Hide()
	Focus()
	IsVisible() bool
}

// Application is the subset of the desktop application the tray controls.
type Application interface {
	Quit()
}

// Config wires a Controller to the desktop shell. Window and CreateWindow are
// called on the UI thread. HideWindow is optional; when it is set it replaces
// Window.Hide, which lets the host release the webview instead of merely
// hiding it.
type Config struct {
	App          Application
	Window       func() Window
	CreateWindow func() Window
	HideWindow   func()
}

// Controller executes system-tray actions against the desktop shell.
type Controller struct {
	cfg Config
	mu  sync.Mutex
}

// NewController returns a Controller bound to cfg.
func NewController(cfg Config) *Controller {
	return &Controller{cfg: cfg}
}

// Handle executes action.
func (c *Controller) Handle(action Action) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch action {
	case ActionShow:
		c.show()
	case ActionHide:
		c.hide()
	case ActionQuit:
		c.quit()
	}
}

// Show brings the main window to the foreground. A window released earlier is
// recreated first, so the tray can restore the UI after background mode.
func (c *Controller) Show() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.show()
}

// Hide releases the main window. The application stays resident so downloads
// keep running in the background.
func (c *Controller) Hide() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hide()
}

// Toggle hides a visible window and shows a hidden or missing one.
func (c *Controller) Toggle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	window := c.window()
	if window != nil && window.IsVisible() {
		c.hide()
		return
	}
	c.showWindow(window)
}

func (c *Controller) show() {
	c.showWindow(c.window())
}

func (c *Controller) showWindow(window Window) {
	if window == nil {
		if c.cfg.CreateWindow == nil {
			return
		}
		window = c.cfg.CreateWindow()
	}
	if window == nil {
		return
	}
	window.Show()
	window.Focus()
}

func (c *Controller) hide() {
	if c.cfg.HideWindow != nil {
		c.cfg.HideWindow()
		return
	}
	if window := c.window(); window != nil {
		window.Hide()
	}
}

func (c *Controller) quit() {
	if c.cfg.App != nil {
		c.cfg.App.Quit()
	}
}

func (c *Controller) window() Window {
	if c.cfg.Window == nil {
		return nil
	}
	return c.cfg.Window()
}
