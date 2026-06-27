package wailsbridge

import (
	"ariadm/internal/domain/config"
	"ariadm/internal/domain/task"
	"context"
	"errors"
	"log"
	"net"
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
	err := b.taskService.TogglePauseTask(taskID)

	log.Printf("ToggleTaskPauseState called for taskID=%s, error=%v", taskID, err)

	return err
}

// DeleteTask removes a download from aria2c's queue and wipes its record from the local database
func (b *WailsBridge) DeleteTask(taskID string, deleteFiles bool) error {
	return b.taskService.DeleteTask(taskID, deleteFiles)
}

// OnStartup is invoked automatically by the Wails runtime engine engine
func (b *WailsBridge) OnStartup(ctx context.Context) {
	b.ctx = ctx

	go func() {
		targetPort := "127.0.0.1:6800"
		if err := b.waitForDaemon(targetPort, 5*time.Second); err != nil {
			log.Printf("Error: Aria2c daemon RPC socket failed to initialize within timeout: %v", err)
			runtime.EventsEmit(b.ctx, "engine:status", "disconnected")
			return
		}
		if err := b.taskService.ReconcileSessionTasks(); err != nil {
			// Log the initialization warning, but don't crash the app
			log.Println("Warning: Cross-session reconciliation encountered an issue:", err.Error())
		}
		b.startTelemetryLoop()
	}()

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

// waitForDaemon polls the network address until it accepts connections or hits the timeout threshold
func (b *WailsBridge) waitForDaemon(target string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Attempt a low-overhead raw TCP connection probe
		conn, err := net.DialTimeout("tcp", target, 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil // Daemon port is active and listening!
		}

		// Back-off briefly before retrying
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("timeout reached waiting for port %s to open", target)

	return errors.New("timeout reached waiting for daemon to start")
}
