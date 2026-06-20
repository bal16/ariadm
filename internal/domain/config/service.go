package config

import "strconv"

type ConfigService struct {
	repo   ConfigRepository
	engine EngineClient
}

func NewConfigService(repo ConfigRepository, engine EngineClient) *ConfigService {
	return &ConfigService{
		repo:   repo,
		engine: engine,
	}
}

func (s *ConfigService) GetConfig() (*AppConfig, error) {
	cfg, err := s.repo.Load()
	if err != nil {
		// Fallback to default if there's an error loading
		return NewDefaultConfig(), nil
	}
	return cfg, nil
}

func (s *ConfigService) UpdateSettings(cfg *AppConfig) error {
	// 1. Persist to local JSON file
	if err := s.repo.Save(cfg); err != nil {
		return err
	}

	// 2. Map domain configuration to aria2c RPC parameters
	aria2Options := map[string]string{
		"max-overall-download-limit": strconv.FormatInt(cfg.SpeedLimit, 10),
		"max-concurrent-downloads":   strconv.Itoa(cfg.MaxConcurrentTasks),
	}

	// 3. Update the running aria2c daemon live
	if err := s.engine.ChangeGlobalOption(aria2Options); err != nil {
		// Log error or handle it, but we can return it for the sake of the transaction
		return err
	}

	return nil
}
