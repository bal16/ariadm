package main

import (
	"ariadm/internal/domain/config"
	"ariadm/internal/domain/task"
	"ariadm/internal/infra/database"
	"ariadm/internal/infra/rpc"
	"ariadm/internal/ingress/httpserver"
	"ariadm/internal/ingress/wailsbridge"
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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
	}
	defer localHTTPServer.Stop()

	// 4. Launch the Wails Desktop Shell Application Window
	app := NewApp()
	actuallyQuit := false

	// Create application with options
	err = wails.Run(&options.App{
		Title:  APP_NAME,
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		// OnStartup:        app.OnStartup,
		OnStartup: func(ctx context.Context) {
			app.OnStartup(ctx)    // 1. Fire up your aria2c daemon engine first
			bridge.OnStartup(ctx) // 2. Pass context to the bridge to launch the telemetry ticker

			runtime.EventsOn(ctx, "app:force-quit", func(optionalData ...interface{}) {
				actuallyQuit = true
				runtime.Quit(ctx)
			})
		},
		OnShutdown: app.OnShutdown,
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			if actuallyQuit {
				return false // Allow the app to close
			}
			// Emit event to frontend to show quit confirmation dialog
			runtime.EventsEmit(ctx, "app:request-close")
			return true // Prevent closing directly
		},

		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "ariadm-unique-lock-8f2k",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				// Re-show window if it was hidden in background
				runtime.WindowShow(app.ctx)
			},
		},
		Bind: []interface{}{
			app,
			bridge,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
