package main

import (
	"ariadm/internal/infra/daemon"
	"context"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// App struct
type App struct {
	daemon *daemon.DaemonManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		daemon: daemon.NewDaemonManager("aria2c", "6800"),
	}
}

// ServiceStartup is called by Wails when the application starts.
func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	log.Println("Initializing aria2c daemon engine...")
	if err := a.daemon.Start(); err != nil {
		log.Printf("CRITICAL: Failed to launch download engine: %v\n", err)
	}

	// TODO: Handle if daemon already started or port already used
	return nil
}

// ServiceShutdown is called by Wails when the application is shutting down.
func (a *App) ServiceShutdown() error {
	log.Println("Stopping download engine safely...")
	if err := a.daemon.Stop(); err != nil {
		log.Printf("Error shutting down daemon: %v\n", err)
	}
	return nil
}
