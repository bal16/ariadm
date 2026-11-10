package wailsbridge

import (
	"context"
	"log"
	"time"

	"github.com/bal16/ariadm/internal/appstate"
	"github.com/bal16/ariadm/internal/domain/config"
	"github.com/bal16/ariadm/internal/domain/task"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type WailsBridge struct {
	ctx           context.Context
	cancel        context.CancelFunc
	configService *config.ConfigService
	taskService   *task.TaskService
	hub           *appstate.Hub
}

func NewWailsBridge(cs *config.ConfigService, ts *task.TaskService, hub *appstate.Hub) *WailsBridge {
	return &WailsBridge{
		configService: cs,
		taskService:   ts,
		hub:           hub,
	}
}

// GetRuntimeStatus returns the current engine and extension state. The frontend
// reads it on mount because background mode recreates the webview and events
// published before that point are lost.
func (b *WailsBridge) GetRuntimeStatus() appstate.Status {
	return b.hub.Snapshot()
}

// GetApplicationConfig fetches the current local settings values
func (b *WailsBridge) GetApplicationConfig() (*config.AppConfig, error) {
	return b.configService.GetConfig()
}

// SaveApplicationConfig saves configuration updates and syncs them live to aria2c
func (b *WailsBridge) SaveApplicationConfig(cfg *config.AppConfig) error {
	return b.configService.UpdateSettings(cfg)
}

// TriggerNewDownload passes a web URL down into the download manager pipelines
func (b *WailsBridge) TriggerNewDownload(url string) (*task.Task, error) {
	return b.taskService.DownloadFile(url)
}

// ToggleTaskPauseState flips the network engine status for a single task queue item
func (b *WailsBridge) ToggleTaskPauseState(taskID string) error {
	err := b.taskService.TogglePauseTask(taskID)

	log.Printf("ToggleTaskPauseState called for taskID=%s, error=%v", taskID, err)

	return err
}

// PauseAllTasks halts every active and waiting download in the queue
func (b *WailsBridge) PauseAllTasks() error {
	return b.taskService.PauseAllTasks()
}

// ResumeAllTasks resumes every paused download in the queue
func (b *WailsBridge) ResumeAllTasks() error {
	return b.taskService.ResumeAllTasks()
}

// DeleteTask removes a download from aria2c's queue and wipes its record from the local database
func (b *WailsBridge) DeleteTask(taskID string, deleteFiles bool) error {
	return b.taskService.DeleteTask(taskID, deleteFiles)
}

// ServiceStartup is invoked automatically by the Wails runtime engine.
func (b *WailsBridge) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	b.ctx, b.cancel = context.WithCancel(ctx)

	go func() {
		if err := b.taskService.WaitForEngine(5 * time.Second); err != nil {
			log.Printf("Error: Aria2c daemon RPC socket failed to initialize within timeout: %v", err)
			b.hub.SetEngineHealth(appstate.EngineDisconnected)
			return
		}
		if err := b.taskService.ReconcileSessionTasks(); err != nil {
			// Log the initialization warning, but don't crash the app
			log.Println("Warning: Cross-session reconciliation encountered an issue:", err.Error())
		}
		b.startTelemetryLoop()
	}()

	return nil
}

// ServiceShutdown stops the telemetry loop when the application quits.
func (b *WailsBridge) ServiceShutdown() error {
	if b.cancel != nil {
		b.cancel()
	}
	return nil
}

func (b *WailsBridge) startTelemetryLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			// Fetch fresh tasks and sync live aria2c progress data into SQLite
			tasks, err := b.taskService.SyncAndGetAllTasks()
			if err != nil {
				b.hub.SetEngineHealth(appstate.EngineDisconnected)
				continue
			}

			// Broadcast live array bursts straight onto the SolidJS frontend
			b.emit("tasks:update", tasks)
			b.hub.SetEngineHealth(appstate.EngineRunning)
		}
	}
}

// emit publishes an application event to the frontend, if the app is running.
func (b *WailsBridge) emit(name string, data ...any) {
	app := application.Get()
	if app == nil {
		return
	}
	app.Event.Emit(name, data...)
}

// GetTasks queries, enriches with live aria2c data, and returns all task records
func (b *WailsBridge) GetTasks() ([]*task.Task, error) {
	return b.taskService.SyncAndGetAllTasks()
}
