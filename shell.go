package main

import (
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"ariadm/internal/ingress/tray"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// shellOptions configures the desktop shell.
type shellOptions struct {
	name             string
	description      string
	singleInstanceID string
	assets           application.AssetOptions
	services         []application.Service
	trayIcon         []byte
	trayIconDark     []byte
	trayIconTemplate []byte
}

// shell owns the Wails application, the main window and the system tray. It
// serializes window lifecycle operations so tray actions, background mode and
// single-instance relaunches cannot race.
type shell struct {
	opts shellOptions

	app            *application.App
	window         *application.WebviewWindow
	trayController *tray.Controller

	windowMu          sync.Mutex
	actuallyQuit      atomic.Bool
	backgroundClosing atomic.Bool
}

func newShell(opts shellOptions) *shell {
	return &shell{opts: opts}
}

// run builds the Wails application, installs the tray and blocks until the
// application exits.
func (s *shell) run() error {
	s.app = application.New(application.Options{
		Name:        s.opts.name,
		Description: s.opts.description,
		Services:    s.opts.services,
		Assets:      s.opts.assets,
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		Windows: application.WindowsOptions{
			DisableQuitOnLastWindowClosed: true,
		},
		Linux: application.LinuxOptions{
			DisableQuitOnLastWindowClosed: true,
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: s.opts.singleInstanceID,
			OnSecondInstanceLaunch: func(application.SecondInstanceData) {
				if s.trayController != nil {
					s.trayController.Show()
				}
			},
		},
	})

	s.trayController = tray.NewController(tray.Config{
		App:          s,
		Window:       s.currentTrayWindow,
		CreateWindow: s.createTrayWindow,
		HideWindow:   s.enterBackground,
	})

	// The frontend emits this once the user confirms a full quit.
	s.app.Event.On("app:force-quit", func(*application.CustomEvent) {
		s.Quit()
	})

	// The frontend asks to keep running in the background: drop the webview
	// (freeing its memory) while keeping the download engine alive.
	s.app.Event.On("app:background", func(*application.CustomEvent) {
		s.enterBackground()
	})

	s.installTray()
	s.newWindow()
	s.handleShutdownSignals()

	return s.app.Run()
}

// Quit marks the application as terminating and stops it. It implements
// tray.Application so the tray menu uses the same shutdown path as the
// frontend's "Quit Completely" action.
func (s *shell) Quit() {
	s.actuallyQuit.Store(true)
	s.app.Quit()
}

func (s *shell) createWindow() *application.WebviewWindow {
	window := s.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            s.opts.name,
		Width:            1024,
		Height:           768,
		BackgroundColour: application.NewRGB(255, 255, 255),
		URL:              "/",
	})

	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if s.actuallyQuit.Load() || s.backgroundClosing.Load() {
			return
		}
		e.Cancel()
		s.app.Event.Emit("app:request-close")
	})

	return window
}

// newWindow creates a window and makes it the current one.
func (s *shell) newWindow() *application.WebviewWindow {
	s.windowMu.Lock()
	defer s.windowMu.Unlock()

	s.window = s.createWindow()
	return s.window
}

func (s *shell) currentWindow() *application.WebviewWindow {
	s.windowMu.Lock()
	defer s.windowMu.Unlock()
	return s.window
}

// enterBackground destroys the webview so its memory is released, while the
// download engine keeps running. The window is recreated on demand.
func (s *shell) enterBackground() {
	s.windowMu.Lock()
	closed := s.window != nil
	if closed {
		s.backgroundClosing.Store(true)
		s.window.Close()
		s.backgroundClosing.Store(false)
		s.window = nil
	}
	s.windowMu.Unlock()

	if !closed {
		return
	}
	cleanupWebviewProcesses()
}

func (s *shell) handleShutdownSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		println("\nTerminal shutdown signal detected. Exiting gracefully...")
		s.Quit()
	}()
}
