package wailsbridge

import (
	"ariadm/internal/domain/config"
	"ariadm/internal/domain/task"
)

type WailsBridge struct {
	configService *config.ConfigService
	taskService   *task.TaskService
}

func NewWailsBridge(cs *config.ConfigService, ts *task.TaskService) *WailsBridge {
	return &WailsBridge{
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