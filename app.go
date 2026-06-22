package main

import (
	"ariadm/internal/infra/daemon"
	"context"
	"log"
)

// App struct
type App struct {
	ctx    context.Context
	daemon *daemon.DaemonManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		daemon: daemon.NewDaemonManager("aria2c", "6800"),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
// OnStartup is called when Wails initializes
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	log.Println("Initializing aria2c daemon engine...")
	if err := a.daemon.Start(); err != nil {
		log.Printf("CRITICAL: Failed to launch download engine: %v\n", err)
	}

	// TODO: Handle if daemon already started or port already used
}

// OnShutdown is called when the application window closes
func (a *App) OnShutdown(ctx context.Context) {
	log.Println("Stopping download engine safely...")
	if err := a.daemon.Stop(); err != nil {
		log.Printf("Error shutting down daemon: %v\n", err)
	}
}
