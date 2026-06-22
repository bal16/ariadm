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

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

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
		OnStartup:        app.OnStartup,
		Bind: []interface{}{
			app,
			bridge,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
