package config

import "runtime"

type AppConfig struct {
	DefaultDownloadPath    string `json:"default_download_path"`
	SpeedLimit             int64  `json:"speed_limit"` // in bytes per second (0 means unlimited)
	MaxConcurrentTasks     int    `json:"max_concurrent_tasks"`
	ConnectionsPerDownload int    `json:"connections_per_download"`
	MinimizeToTray         bool   `json:"minimize_to_tray"`
	ShowInAppMenu          bool   `json:"show_in_app_menu"`
}

// NewDefaultConfig returns a fallback configuration if the file doesn't exist
func NewDefaultConfig() *AppConfig {
	return &AppConfig{
		DefaultDownloadPath:    "", // Will be resolved by OS specific downloads folder later
		SpeedLimit:             0,  // Unlimited
		MaxConcurrentTasks:     3,
		ConnectionsPerDownload: 4,
		MinimizeToTray:         true,
		ShowInAppMenu:          runtime.GOOS == "linux",
	}
}
