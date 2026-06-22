package database

import (
	"ariadm/internal/domain/config"
	"encoding/json"
	"os"
	"path/filepath"
)

type JSONConfigRepository struct {
	appName  string
	filePath string
}

func NewJSONConfigRepository(appName string) (*JSONConfigRepository, error) {
	// Dynamically resolve the application directory based on mode
	appDir, err := ResolveAppDir(appName)
	if err != nil {
		return nil, err
	}

	repo := &JSONConfigRepository{
		appName:  appName,
		filePath: filepath.Join(appDir, "config.json"),
	}

	repo.Load()

	return repo, nil
}

func (r *JSONConfigRepository) Load() (*config.AppConfig, error) {
	// 1. If the configuration file does not exist, initialize a default one immediately
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		defaultCfg := config.NewDefaultConfig()

		// Attempt to populate an OS-native downloads path as the default folder
		homeDir, errDir := os.UserHomeDir()
		if errDir == nil {
			defaultCfg.DefaultDownloadPath = filepath.Join(homeDir, "Downloads")
		}

		if errSave := r.Save(defaultCfg); errSave != nil {
			return nil, errSave
		}
		return defaultCfg, nil
	}

	// 2. Read the existing file bytes
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, err
	}

	// 3. Parse JSON strings into our domain struct
	var cfg config.AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save marshals and writes configuration updates down to the hard drive
func (r *JSONConfigRepository) Save(cfg *config.AppConfig) error {
	// 1. Ensure the parent directory tree exists before writing the file
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 2. Format the JSON data cleanly for human readability
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// 3. Write atomic bytes securely with standard read/write permissions
	return os.WriteFile(r.filePath, data, 0644)
}
