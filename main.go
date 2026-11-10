package main

import (
	"ariadm/internal/domain/config"
	"ariadm/internal/domain/task"
	"ariadm/internal/infra/database"
	"ariadm/internal/infra/rpc"
	"ariadm/internal/ingress/httpserver"
	"ariadm/internal/ingress/tray"
	"ariadm/internal/ingress/wailsbridge"
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/trayicon.png
var trayIcon []byte

// TODO: Handle if port already used
var (
	APP_NAME = "ariadm"
	RPC_PORT = "6800"
	PORT     = "9999"
)

var (
	appInstance       *application.App
	mainWindow        *application.WebviewWindow
	trayController    *tray.Controller
	windowMu          sync.Mutex
	actuallyQuit      atomic.Bool
	backgroundClosing atomic.Bool
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

// trayApplication adapts the Wails app so quitting from the tray follows the
// same path as the frontend's "Quit Completely" action.
type trayApplication struct {
	app *application.App
}

func (a trayApplication) Quit() {
	actuallyQuit.Store(true)
	a.app.Quit()
}

func createMainWindow(app *application.App) *application.WebviewWindow {
	w := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            APP_NAME,
		Width:            1024,
		Height:           768,
		BackgroundColour: application.NewRGB(255, 255, 255),
		URL:              "/",
	})

	w.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if actuallyQuit.Load() || backgroundClosing.Load() {
			return
		}
		e.Cancel()
		app.Event.Emit("app:request-close")
	})

	return w
}

func currentMainWindow() *application.WebviewWindow {
	windowMu.Lock()
	defer windowMu.Unlock()
	return mainWindow
}

// newMainWindow recreates the main window after background mode released it.
func newMainWindow() *application.WebviewWindow {
	windowMu.Lock()
	defer windowMu.Unlock()

	if appInstance == nil {
		return nil
	}
	mainWindow = createMainWindow(appInstance)
	return mainWindow
}

// enterBackground destroys the webview so its memory is released, while the
// download engine keeps running. The window is recreated on demand.
func enterBackground() {
	windowMu.Lock()
	closed := mainWindow != nil
	if closed {
		backgroundClosing.Store(true)
		mainWindow.Close()
		backgroundClosing.Store(false)
		mainWindow = nil
	}
	windowMu.Unlock()

	if !closed || runtime.GOOS != "linux" {
		return
	}

	// WebKitGTK keeps its web/network processes alive after the window is
	// destroyed. Kill them explicitly so background mode frees RAM.
	killWebkitChildren()
	time.Sleep(200 * time.Millisecond)
}

// setupTray installs the system tray icon and context menu.
func setupTray(app *application.App, controller *tray.Controller) *application.SystemTray {
	systray := app.SystemTray.New()
	systray.SetIcon(trayIcon)
	systray.SetTooltip(APP_NAME + " - download manager")
	systray.OnClick(controller.Toggle)

	menu := app.NewMenu()
	for _, item := range tray.Menu() {
		if item.Label == "" {
			menu.AddSeparator()
			continue
		}
		action := item.Action
		menu.Add(item.Label).OnClick(func(*application.Context) {
			controller.Handle(action)
		})
	}
	systray.SetMenu(menu)

	return systray
}

// killWebkitChildren terminates WebKitGTK helper processes that survive
// window destruction, so background mode actually frees their memory.
func killWebkitChildren() {
	pid := os.Getpid()
	// Collect all descendant PIDs.
	var collect func(int) []int
	collect = func(parent int) []int {
		var out []int
		entries, _ := os.ReadDir("/proc")
		for _, e := range entries {
			id, err := strconv.Atoi(e.Name())
			if err != nil {
				continue
			}
			data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", id))
			if err != nil {
				continue
			}
			fields := strings.Fields(string(data))
			if len(fields) < 4 {
				continue
			}
			ppid, err := strconv.Atoi(fields[3])
			if err != nil || ppid != parent {
				continue
			}
			out = append(out, id)
			out = append(out, collect(id)...)
		}
		return out
	}

	for _, id := range collect(pid) {
		comm, _ := os.ReadFile(fmt.Sprintf("/proc/%d/comm", id))
		name := strings.TrimSpace(string(comm))
		if strings.Contains(name, "WebKit") {
			if p, err := os.FindProcess(id); err == nil {
				_ = p.Kill()
			}
		}
	}
}

func main() {
	// 1. Setup Infrastructure Layers
	configRepo, err := database.NewJSONConfigRepository(APP_NAME)
	if err != nil {
		log.Fatalf("Failed to initialize config storage: %v", err)
	}

	taskRepo, err := database.NewSQLiteTaskRepository(APP_NAME, "downloads.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	rpcClient := rpc.NewAria2Client("http://127.0.0.1:" + RPC_PORT + "/jsonrpc")

	// 2. Setup Core Domain Services (Injecting Dependencies)
	configService := config.NewConfigService(configRepo, rpcClient)
	taskService := task.NewTaskService(taskRepo, rpcClient, configRepo)

	// 3. Setup Ingress Controllers
	bridge := wailsbridge.NewWailsBridge(configService, taskService)

	localHTTPServer := httpserver.NewHTTPServer(PORT, taskService)
	if err := localHTTPServer.Start(); err != nil {
		log.Printf("Warning: Extension listener failed to bind: %v", err)
	} else {
		log.Printf("Local HTTP Server is listening on http://127.0.0.1:%s", PORT)
	}

	defer func() { _ = localHTTPServer.Stop() }()

	// 4. Launch the Wails Desktop Shell Application Window
	appService := NewApp()

	app := application.New(application.Options{
		Name:        APP_NAME,
		Description: "Aria2 download manager",
		Services: []application.Service{
			application.NewService(appService),
			application.NewService(bridge),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
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
			UniqueID: "ariadm-lock-8f2k",
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				if trayController != nil {
					trayController.Show()
				}
			},
		},
	})

	appInstance = app

	trayController = tray.NewController(tray.Config{
		App: trayApplication{app: app},
		Window: func() tray.Window {
			window := currentMainWindow()
			if window == nil {
				return nil
			}
			return trayWindow{window: window}
		},
		CreateWindow: func() tray.Window {
			window := newMainWindow()
			if window == nil {
				return nil
			}
			return trayWindow{window: window}
		},
		HideWindow: enterBackground,
	})

	// The frontend emits this once the user confirms a full quit.
	app.Event.On("app:force-quit", func(event *application.CustomEvent) {
		actuallyQuit.Store(true)
		app.Quit()
	})

	// The frontend asks to keep running in the background: drop the webview
	// (freeing its memory) while keeping the download engine alive.
	app.Event.On("app:background", func(event *application.CustomEvent) {
		enterBackground()
	})

	setupTray(app, trayController)

	mainWindow = newMainWindow()

	// Gracefully handle terminal shutdown signals.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		println("\nTerminal shutdown signal detected. Exiting gracefully...")
		actuallyQuit.Store(true)
		app.Quit()
	}()

	if err := app.Run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
