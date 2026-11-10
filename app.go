package main

import (
	"context"
	"errors"
	"log"

	"github.com/bal16/ariadm/internal/appstate"
	"github.com/bal16/ariadm/internal/infra/daemon"
	"github.com/bal16/ariadm/internal/infra/rpc"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// App owns the download engine process and publishes its state.
type App struct {
	daemon *daemon.DaemonManager
	client *rpc.Aria2Client
	hub    *appstate.Hub
}

// NewApp creates a new App application struct
func NewApp(dm *daemon.DaemonManager, client *rpc.Aria2Client, hub *appstate.Hub) *App {
	return &App{
		daemon: dm,
		client: client,
		hub:    hub,
	}
}

// ServiceStartup is called by Wails when the application starts.
func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	log.Println("Initializing aria2c daemon engine...")

	port, err := a.daemon.Start()
	if err != nil {
		state := appstate.EngineFailed
		if errors.Is(err, daemon.ErrNoPortAvailable) {
			state = appstate.EnginePortConflict
		}
		a.hub.SetEngine(state, 0, err.Error())
		log.Printf("CRITICAL: Failed to launch download engine: %v", err)
		return nil
	}

	a.client.SetPort(port)
	a.hub.SetEngine(appstate.EngineRunning, port, "")
	log.Printf("Download engine is listening on 127.0.0.1:%d", port)

	return nil
}

// ServiceShutdown is called by Wails when the application is shutting down.
func (a *App) ServiceShutdown() error {
	log.Println("Stopping download engine safely...")
	if err := a.daemon.Stop(); err != nil {
		log.Printf("Error shutting down daemon: %v", err)
	}
	a.hub.SetEngine(appstate.EngineStopped, 0, "")
	return nil
}
