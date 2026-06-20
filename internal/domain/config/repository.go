package config

type ConfigRepository interface {
	Load() (*AppConfig, error)
	Save(cfg *AppConfig) error
}

type EngineClient interface {
	ChangeGlobalOption(options map[string]string) error
}