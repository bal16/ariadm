package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bal16/ariadm/internal/appstate"
	"github.com/bal16/ariadm/internal/domain/config"
	"github.com/bal16/ariadm/internal/domain/task"
	"github.com/bal16/ariadm/internal/infra/daemon"
	"github.com/bal16/ariadm/internal/infra/database"
	"github.com/bal16/ariadm/internal/infra/rpc"
	"github.com/bal16/ariadm/internal/ingress/httpserver"
	"github.com/bal16/ariadm/internal/ingress/wailsbridge"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/trayicon.png
var trayIcon []byte

//go:embed build/trayicon-dark.png
var trayIconDark []byte

//go:embed build/trayicon-template.png
var trayIconTemplate []byte

const (
	appName = "ariadm"

	rpcPortStart  = 6800
	rpcPortCount  = 10
	httpPortStart = 9999
	httpPortCount = 10
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

	appDir, err := database.ResolveAppDir(appName)
	if err != nil {
		log.Fatalf("Failed to resolve application data directory: %v", err)
	}

	rpcPorts := portRange(rpcPortStart, rpcPortCount)
	pidPath := filepath.Join(appDir, "daemon.pid")
	sessionPath := filepath.Join(appDir, "session.txt")

	hub := appstate.NewHub(func(status appstate.Status) {
		if app := application.Get(); app != nil {
			app.Event.Emit("runtime:status", status)
		}
	})

	rpcClient := rpc.NewAria2Client(fmt.Sprintf("http://127.0.0.1:%d/jsonrpc", rpcPorts[0]))

	// Domain services.
	configService := config.NewConfigService(configRepo, rpcClient)
	taskService := task.NewTaskService(taskRepo, rpcClient, configRepo)

	// Ingress controllers.
	localHTTPServer := httpserver.NewHTTPServer(portRange(httpPortStart, httpPortCount), taskService)
	if port, err := localHTTPServer.Start(); err != nil {
		hub.SetExtension(appstate.ExtensionUnavailable, 0, err.Error())
		log.Printf("Warning: Extension listener failed to bind: %v", err)
	} else {
		hub.SetExtension(appstate.ExtensionListening, port, "")
		log.Printf("Local HTTP Server is listening on http://127.0.0.1:%d", port)
	}
	defer func() { _ = localHTTPServer.Stop() }()

	daemonManager := daemon.NewDaemonManager("aria2c", rpcPorts, sessionPath, pidPath)
	bridge := wailsbridge.NewWailsBridge(configService, taskService, hub)

	// Desktop shell.
	desktop := newShell(shellOptions{
		name:             appName,
		description:      "Aria2 download manager",
		singleInstanceID: "ariadm-lock-8f2k",
		trayIcon:         trayIcon,
		trayIconDark:     trayIconDark,
		trayIconTemplate: trayIconTemplate,
		assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		services: []application.Service{
			application.NewService(NewApp(daemonManager, rpcClient, hub)),
			application.NewService(bridge),
		},
	})

	if err := desktop.run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}

// portRange builds the bounded candidate list for a service port.
func portRange(start, count int) []int {
	ports := make([]int, 0, count)
	for port := start; port < start+count; port++ {
		ports = append(ports, port)
	}
	return ports
}
