package main

import (
	"ariadm/internal/domain/config"
	"ariadm/internal/domain/task"
	"ariadm/internal/infra/database"
	"ariadm/internal/infra/rpc"
	"ariadm/internal/ingress/httpserver"
	"ariadm/internal/ingress/wailsbridge"
	"embed"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

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
	var actuallyQuit atomic.Bool

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
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "ariadm-lock-8f2k",
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				// Re-show window if it was hidden in background
				if window, ok := application.Get().Window.GetByName("main"); ok {
					window.Show()
					window.Focus()
				}
			},
		},
	})

	mainWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            APP_NAME,
		Width:            1024,
		Height:           768,
		BackgroundColour: application.NewRGB(255, 255, 255),
		URL:              "/",
	})

	// Intercept the native close button so the frontend can ask for confirmation.
	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if actuallyQuit.Load() {
			return
		}
		e.Cancel()
		app.Event.Emit("app:request-close")
	})

	// The frontend emits this once the user confirms a full quit.
	app.Event.On("app:force-quit", func(event *application.CustomEvent) {
		actuallyQuit.Store(true)
		app.Quit()
	})

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
