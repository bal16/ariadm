package wailsbridge

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
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

// themeSyncCallback is stored outside the struct so the Wails binding generator
// does not try to serialize a function-typed field.
var themeSyncCallback func(string)

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

// TriggerNewDownloadWithOptions passes a web URL down into the download manager
// pipelines with optional destination, file name, and connection count overrides.
func (b *WailsBridge) TriggerNewDownloadWithOptions(url, downloadPath, fileName string, connections int) (*task.Task, error) {
	return b.taskService.DownloadFileWithOptions(url, downloadPath, fileName, connections)
}

// SelectDownloadFolder opens a directory picker and returns the chosen path.
func (b *WailsBridge) SelectDownloadFolder() (string, error) {
	app := application.Get()
	if app == nil {
		return "", fmt.Errorf("application is not available")
	}

	path, err := app.Dialog.OpenFile().
		SetTitle("Select Download Folder").
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	return path, nil
}

// OpenPath opens a file or folder in the operating system's default handler.
func (b *WailsBridge) OpenPath(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	if app := application.Get(); app != nil {
		return app.Browser.OpenFile(path)
	}

	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start()
	case "windows":
		return exec.Command("explorer", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
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

// SyncTheme updates the native menu's selected theme indicator.
func (b *WailsBridge) SyncTheme(mode string) {
	if themeSyncCallback != nil {
		themeSyncCallback(mode)
	}
}

// SetThemeSync registers the callback used to update native menu theme state.
func (b *WailsBridge) SetThemeSync(fn func(string)) {
	themeSyncCallback = fn
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
