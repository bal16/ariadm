package wailsbridge_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bal16/ariadm/internal/domain/config"
	"github.com/bal16/ariadm/internal/domain/task"
	"github.com/bal16/ariadm/internal/infra/database"
	"github.com/bal16/ariadm/internal/ingress/wailsbridge"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockDownloadEngine struct{ mock.Mock }

func (m *MockDownloadEngine) AddURI(url, path, fileName string) (string, error) {
	args := m.Called(url, path)
	return args.String(0), args.Error(1)
}
func (m *MockDownloadEngine) ResumeURI(url, path, fileName string) (string, error) {
	args := m.Called(url, path)
	return args.String(0), args.Error(1)
}
func (m *MockDownloadEngine) Pause(gid string) error   { return m.Called(gid).Error(0) }
func (m *MockDownloadEngine) Unpause(gid string) error { return m.Called(gid).Error(0) }
func (m *MockDownloadEngine) PauseAll() error          { return m.Called().Error(0) }
func (m *MockDownloadEngine) UnpauseAll() error        { return m.Called().Error(0) }
func (m *MockDownloadEngine) Ping() error              { return m.Called().Error(0) }
func (m *MockDownloadEngine) Remove(gid string) error  { return m.Called(gid).Error(0) }
func (m *MockDownloadEngine) RemoveDownloadResult(gid string) error {
	return m.Called(gid).Error(0)
}
func (m *MockDownloadEngine) TellStatus(gid string) (*task.Aria2Status, error) {
	args := m.Called(gid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*task.Aria2Status), args.Error(1)
}
func (m *MockDownloadEngine) ListDownloads() ([]*task.Aria2Status, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*task.Aria2Status), args.Error(1)
}
func (m *MockDownloadEngine) ChangeGlobalOption(options map[string]string) error {
	return m.Called(options).Error(0)
}

type MockConfigRepository struct{ mock.Mock }

func (m *MockConfigRepository) Load() (*config.AppConfig, error) {
	args := m.Called()
	return args.Get(0).(*config.AppConfig), args.Error(1)
}
func (m *MockConfigRepository) Save(cfg *config.AppConfig) error { return m.Called(cfg).Error(0) }

func setupIntegrationTest(t *testing.T) (*wailsbridge.WailsBridge, *MockDownloadEngine, *MockConfigRepository, func()) {
	database.IsDev = true
	testAppName := "ariadm_test_suite"
	testDBFile := "bridge_integration.db"

	// Track true path utilizing domain file logic to avoid orphan files
	appDir, err := database.ResolveAppDir(testAppName)
	require.NoError(t, err)
	fullPath := filepath.Join(appDir, testDBFile)

	// Clean out any lingering databases from previous executions
	_ = os.Remove(fullPath)

	taskRepo, err := database.NewSQLiteTaskRepository(testAppName, testDBFile)
	require.NoError(t, err)

	mockEngine := new(MockDownloadEngine)
	mockConfigRepo := new(MockConfigRepository)

	mockConfig := &config.AppConfig{DefaultDownloadPath: "/tmp/downloads"}
	mockConfigRepo.On("Load").Return(mockConfig, nil)

	configService := config.NewConfigService(mockConfigRepo, mockEngine)
	taskService := task.NewTaskService(taskRepo, mockEngine, mockConfigRepo)

	bridge := wailsbridge.NewWailsBridge(configService, taskService)

	teardown := func() {
		_ = os.Remove(fullPath)
	}

	return bridge, mockEngine, mockConfigRepo, teardown
}

func TestWailsBridge_DownloadAndList_Integration(t *testing.T) {
	bridge, mockEngine, _, teardown := setupIntegrationTest(t)
	defer teardown()

	targetURL := "https://releases.cachyos.org/desktop/cachyos-desktop-linux.iso"
	mockedGID := "aria2_generated_gid_12345"

	mockEngine.On("AddURI", targetURL, "/tmp/downloads").Return(mockedGID, nil)

	// SyncAndGetAllTasks will call TellStatus for each stored task
	mockEngine.On("TellStatus", mockedGID).Return(&task.Aria2Status{
		GID:             mockedGID,
		Status:          "active",
		TotalLength:     104857600, // 100 MB
		CompletedLength: 52428800,  // 50 MB
		DownloadSpeed:   5242880,   // 5 MB/s
		FileName:        "cachyos-desktop-linux.iso",
	}, nil)

	// --- STEP 1: TEST TRIGGERING DOWNLOAD ---
	createdTask, err := bridge.TriggerNewDownload(targetURL)
	require.NoError(t, err) // 👈 Stops right here if database errors are thrown
	require.NotNil(t, createdTask)
	assert.Equal(t, mockedGID, createdTask.GID)
	assert.Equal(t, string(task.StatusActive), string(createdTask.Status))

	// --- STEP 2: TEST FETCHING LIVE QUEUE ---
	activeQueue, err := bridge.GetTasks()
	require.NoError(t, err)
	assert.Len(t, activeQueue, 1)
	assert.Equal(t, createdTask.ID, activeQueue[0].ID)
}

func TestWailsBridge_TogglePauseState_Integration(t *testing.T) {
	bridge, mockEngine, _, teardown := setupIntegrationTest(t)
	defer teardown()

	targetURL := "https://files.minecraft.net/modpack.zip"
	mockedGID := "aria2_gid_9999"

	mockEngine.On("AddURI", targetURL, "/tmp/downloads").Return(mockedGID, nil)
	mockEngine.On("Pause", mockedGID).Return(nil)
	// SyncAndGetAllTasks will call TellStatus once (while the task is still active).
	// After ToggleTaskPauseState writes "paused" to SQLite, GetTasks skips TellStatus
	// because paused tasks are in a user-controlled stable state.
	mockEngine.On("TellStatus", mockedGID).Return(&task.Aria2Status{
		GID:    mockedGID,
		Status: "active",
	}, nil)

	createdTask, err := bridge.TriggerNewDownload(targetURL)
	require.NoError(t, err)
	require.NotNil(t, createdTask)

	err = bridge.ToggleTaskPauseState(createdTask.ID)
	require.NoError(t, err)

	updatedQueue, err := bridge.GetTasks()
	require.NoError(t, err)
	assert.Equal(t, "paused", string(updatedQueue[0].Status))
}

func TestWailsBridge_PauseResumeAll_Integration(t *testing.T) {
	bridge, mockEngine, _, teardown := setupIntegrationTest(t)
	defer teardown()

	targetURL := "https://files.minecraft.net/modpack.zip"
	mockedGID := "aria2_gid_bulk_1"

	mockEngine.On("AddURI", targetURL, "/tmp/downloads").Return(mockedGID, nil)
	mockEngine.On("PauseAll").Return(nil)
	mockEngine.On("UnpauseAll").Return(nil)
	mockEngine.On("TellStatus", mockedGID).Return(&task.Aria2Status{
		GID:    mockedGID,
		Status: "active",
	}, nil)

	createdTask, err := bridge.TriggerNewDownload(targetURL)
	require.NoError(t, err)
	require.NotNil(t, createdTask)

	require.NoError(t, bridge.PauseAllTasks())

	pausedQueue, err := bridge.GetTasks()
	require.NoError(t, err)
	require.Len(t, pausedQueue, 1)
	assert.Equal(t, "paused", string(pausedQueue[0].Status))

	require.NoError(t, bridge.ResumeAllTasks())

	activeQueue, err := bridge.GetTasks()
	require.NoError(t, err)
	require.Len(t, activeQueue, 1)
	assert.Equal(t, "active", string(activeQueue[0].Status))
}
