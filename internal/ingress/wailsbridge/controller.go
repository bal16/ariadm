package wailsbridge

import (
	"ariadm/internal/domain/config"
	"ariadm/internal/domain/task"
	"context"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type WailsBridge struct {
	ctx           context.Context
	configService *config.ConfigService
	taskService   *task.TaskService
}

func NewWailsBridge(cs *config.ConfigService, ts *task.TaskService) *WailsBridge {
	return &WailsBridge{
		ctx:           context.Background(),
		configService: cs,
		taskService:   ts,
	}
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
	return b.taskService.TogglePauseTask(taskID)
}

// OnStartup is invoked automatically by the Wails runtime engine engine
func (b *WailsBridge) OnStartup(ctx context.Context) {
	b.ctx = ctx
	go b.startTelemetryLoop() // 👈 Spin up the concurrent background engine ticker
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
				runtime.EventsEmit(b.ctx, "engine:status", "disconnected")
				continue
			}

			// Broadcast live array bursts straight onto the SolidJS frontend
			runtime.EventsEmit(b.ctx, "tasks:update", tasks)
			runtime.EventsEmit(b.ctx, "engine:status", "running")
		}
	}
}

// GetTasks queries, enriches with live aria2c data, and returns all task records
func (b *WailsBridge) GetTasks() ([]*task.Task, error) {
	return b.taskService.SyncAndGetAllTasks()
}
