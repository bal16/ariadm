package main

import (
	"ariadm/internal/domain/config"
	"ariadm/internal/domain/task"
	"ariadm/internal/infra/database"
	"ariadm/internal/infra/rpc"
	"ariadm/internal/ingress/httpserver"
	"ariadm/internal/ingress/wailsbridge"
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

// TODO: Handle if port already used
var (
	APP_NAME = "ariadm"
	RPC_PORT = "6800"
	PORT     = "9999"
)

var (
	appInstance       *application.App
	mainWindow        *application.WebviewWindow
	actuallyQuit      atomic.Bool
	backgroundClosing atomic.Bool
)

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
				if mainWindow == nil {
					if appInstance != nil {
						mainWindow = createMainWindow(appInstance)
					}
				}
				if mainWindow != nil {
					mainWindow.Show()
					mainWindow.Focus()
				}
			},
		},
	})

	appInstance = app

	// The frontend emits this once the user confirms a full quit.
	app.Event.On("app:force-quit", func(event *application.CustomEvent) {
		actuallyQuit.Store(true)
		app.Quit()
	})

	// The frontend asks to keep running in the background: drop the webview
	// (freeing its memory) while keeping the download engine alive.
	app.Event.On("app:background", func(event *application.CustomEvent) {
		if mainWindow != nil {
			backgroundClosing.Store(true)
			mainWindow.Close()
			backgroundClosing.Store(false)
			mainWindow = nil
		}
		// WebKitGTK keeps its web/network processes alive after the window
		// is destroyed. Kill them explicitly so background mode frees RAM.
		if runtime.GOOS == "linux" {
			killWebkitChildren()
			time.Sleep(200 * time.Millisecond)
		}
	})

	mainWindow = createMainWindow(app)

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
