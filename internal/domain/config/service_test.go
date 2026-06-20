package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 1. Manual Mock Implementation
type ConfigRepositoryMock struct {
	mock.Mock
}

func (m *ConfigRepositoryMock) Load() (*AppConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AppConfig), args.Error(1)
}

func (m *ConfigRepositoryMock) Save(cfg *AppConfig) error {
	args := m.Called(cfg)
	return args.Error(0)
}

type EngineClientMock struct {
	mock.Mock
}

func (m *EngineClientMock) ChangeGlobalOption(options map[string]string) error {
	args := m.Called(options)
	return args.Error(0)
}

// 2. The Test Case (RED Phase)
func TestGetConfig_Success(t *testing.T) {
	repoMock := new(ConfigRepositoryMock)
	engineMock := new(EngineClientMock)

	dummyConfig := &AppConfig{DefaultDownloadPath: "/downloads/test"}

	// Define behavior: When Load is called, return dummyConfig and nil error
	repoMock.On("Load").Return(dummyConfig, nil)

	service := NewConfigService(repoMock, engineMock)
	res, err := service.GetConfig()

	assert.NoError(t, err)
	assert.Equal(t, "/downloads/test", res.DefaultDownloadPath)
	repoMock.AssertExpectations(t)
}

func TestGetConfig_FallbackToDefault(t *testing.T) {
	repoMock := new(ConfigRepositoryMock)
	engineMock := new(EngineClientMock)

	// Simulate repository returning an error
	repoMock.On("Load").Return(nil, errors.New("file not found"))

	service := NewConfigService(repoMock, engineMock)
	res, err := service.GetConfig()

	assert.NoError(t, err)                     // Service should NOT crash; it should handle it gracefully
	assert.Equal(t, 3, res.MaxConcurrentTasks) // Should match default config values
	repoMock.AssertExpectations(t)
}

func TestUpdateSettings_Success(t *testing.T) {
	repoMock := new(ConfigRepositoryMock)
	engineMock := new(EngineClientMock)

	newConfig := &AppConfig{
		DefaultDownloadPath: "/new/path",
		SpeedLimit:          512000, // 500 KB/s
		MaxConcurrentTasks:  5,
	}

	// Expectation 1: Save to JSON file
	repoMock.On("Save", newConfig).Return(nil)

	// Expectation 2: Tell aria2c to change options live
	expectedOptions := map[string]string{
		"max-overall-download-limit": "512000",
		"max-concurrent-downloads":   "5",
	}
	engineMock.On("ChangeGlobalOption", expectedOptions).Return(nil)

	service := NewConfigService(repoMock, engineMock)
	err := service.UpdateSettings(newConfig)

	assert.NoError(t, err)
	repoMock.AssertExpectations(t)
	engineMock.AssertExpectations(t)
}
