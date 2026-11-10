package main

import (
	"embed"
	"log"
	"os"

	"ariadm/internal/domain/config"
	"ariadm/internal/domain/task"
	"ariadm/internal/infra/database"
	"ariadm/internal/infra/rpc"
	"ariadm/internal/ingress/httpserver"
	"ariadm/internal/ingress/wailsbridge"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/trayicon.png
var trayIcon []byte

const (
	appName  = "ariadm"
	rpcPort  = "6800"
	httpPort = "9999"
)

func main() {
	// Infrastructure layer.
	configRepo, err := database.NewJSONConfigRepository(appName)
	if err != nil {
		log.Fatalf("Failed to initialize config storage: %v", err)
	}

	taskRepo, err := database.NewSQLiteTaskRepository(appName, "downloads.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	rpcClient := rpc.NewAria2Client("http://127.0.0.1:" + rpcPort + "/jsonrpc")

	// Domain services.
	configService := config.NewConfigService(configRepo, rpcClient)
	taskService := task.NewTaskService(taskRepo, rpcClient, configRepo)

	// Ingress controllers.
	bridge := wailsbridge.NewWailsBridge(configService, taskService)

	localHTTPServer := httpserver.NewHTTPServer(httpPort, taskService)
	if err := localHTTPServer.Start(); err != nil {
		log.Printf("Warning: Extension listener failed to bind: %v", err)
	} else {
		log.Printf("Local HTTP Server is listening on http://127.0.0.1:%s", httpPort)
	}
	defer func() { _ = localHTTPServer.Stop() }()

	// Desktop shell.
	desktop := newShell(shellOptions{
		name:             appName,
		description:      "Aria2 download manager",
		singleInstanceID: "ariadm-lock-8f2k",
		trayIcon:         trayIcon,
		assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		services: []application.Service{
			application.NewService(NewApp()),
			application.NewService(bridge),
		},
	})

	if err := desktop.run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
