package database

import (
	"os"
	"path/filepath"
)

// IsDev can be set via ldflags, e.g., -ldflags "-X ariadm/internal/infra/database.IsDev=true"
// Or we can fall back to checking an environment variable.
var IsDev = os.Getenv("APP_ENV") == "development"

// ResolveAppDir returns the current working directory in dev mode,
// and the system user config directory (~/.config/appName) in production mode.
func ResolveAppDir(appName string) (string, error) {
	if IsDev {
		// Save directly in the current working directory during development
		return "./generated", nil
	}

	// Production / standard build mode: use ~/.config or %APPDATA%
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, appName), nil
}
