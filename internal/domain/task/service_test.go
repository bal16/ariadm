package task_test

import (
	"ariadm/internal/domain/config"
	"ariadm/internal/domain/task"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 1. Manual Mocks
type TaskRepositoryMock struct{ mock.Mock }

func (m *TaskRepositoryMock) Create(t *task.Task) error { return m.Called(t).Error(0) }
func (m *TaskRepositoryMock) GetByID(id string) (*task.Task, error) {
	args := m.Called(id)
	return args.Get(0).(*task.Task), args.Error(1)
}
func (m *TaskRepositoryMock) GetByGID(gid string) (*task.Task, error) {
	args := m.Called(gid)
	return args.Get(0).(*task.Task), args.Error(1)
}
func (m *TaskRepositoryMock) Update(t *task.Task) error { return m.Called(t).Error(0) }

type DownloadEngineMock struct{ mock.Mock }

func (m *DownloadEngineMock) AddURI(url, path string) (string, error) {
	args := m.Called(url, path)
	return args.String(0), args.Error(1)
}
func (m *DownloadEngineMock) Pause(gid string) error   { return m.Called(gid).Error(0) }
func (m *DownloadEngineMock) Unpause(gid string) error { return m.Called(gid).Error(0) }
func (m *DownloadEngineMock) TellStatus(gid string) (*task.Aria2Status, error) {
	args := m.Called(gid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*task.Aria2Status), args.Error(1)
}

type ConfigRepositoryMock struct{ mock.Mock }

func (m *ConfigRepositoryMock) Load() (*config.AppConfig, error) {
	args := m.Called()
	return args.Get(0).(*config.AppConfig), args.Error(1)
}
func (m *ConfigRepositoryMock) Save(cfg *config.AppConfig) error { return m.Called(cfg).Error(0) }

func (m *TaskRepositoryMock) GetAll() ([]*task.Task, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*task.Task), args.Error(1)
}

// 2. Test Case (RED Phase)
func TestDownloadFile_Success(t *testing.T) {
	taskRepo := new(TaskRepositoryMock)
	engine := new(DownloadEngineMock)
	configRepo := new(ConfigRepositoryMock)

	targetURL := "https://example.com/file.zip"
	mockConfig := &config.AppConfig{DefaultDownloadPath: "/downloads"}
	expectedGID := "aria2_gid_999"

	// Mock expectations
	configRepo.On("Load").Return(mockConfig, nil)
	engine.On("AddURI", targetURL, "/downloads").Return(expectedGID, nil)

	// We check if it attempts to save a task with correct properties to the DB
	taskRepo.On("Create", mock.MatchedBy(func(t *task.Task) bool {
		return t.URL == targetURL && t.GID == expectedGID && t.Status == task.StatusActive
	})).Return(nil)

	service := task.NewTaskService(taskRepo, engine, configRepo)
	res, err := service.DownloadFile(targetURL)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, expectedGID, res.GID)

	taskRepo.AssertExpectations(t)
	engine.AssertExpectations(t)
	configRepo.AssertExpectations(t)
}

func TestDownloadFile_InvalidURL(t *testing.T) {
	taskRepo := new(TaskRepositoryMock)
	engine := new(DownloadEngineMock)
	configRepo := new(ConfigRepositoryMock)

	invalidURL := "ftp-malformed://missing-proper-structure"

	service := task.NewTaskService(taskRepo, engine, configRepo)
	res, err := service.DownloadFile(invalidURL)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "invalid url format")

	// Verify that the dependencies were NEVER called
	configRepo.AssertNotCalled(t, "Load")
	engine.AssertNotCalled(t, "AddURI", mock.Anything, mock.Anything)
	taskRepo.AssertNotCalled(t, "Create", mock.Anything)
}

// internal/domain/task/service_test.go

func TestTogglePauseTask_ToPaused(t *testing.T) {
	taskRepo := new(TaskRepositoryMock)
	engine := new(DownloadEngineMock)
	configRepo := new(ConfigRepositoryMock)

	taskID := "local_123"
	aria2GID := "aria2_123"

	existingTask := &task.Task{
		ID:     taskID,
		GID:    aria2GID,
		Status: task.StatusActive,
	}

	// 1. Expect service to fetch the current task state
	taskRepo.On("GetByID", taskID).Return(existingTask, nil)

	// 2. Expect engine to pause the task via aria2c GID
	engine.On("Pause", aria2GID).Return(nil)

	// 3. Expect database to store the updated "paused" status
	taskRepo.On("Update", mock.MatchedBy(func(t *task.Task) bool {
		return t.ID == taskID && t.Status == task.StatusPaused
	})).Return(nil)

	service := task.NewTaskService(taskRepo, engine, configRepo)

	// --- THIS WILL CAUSE A COMPILE ERROR (RED) ---
	// TogglePauseTask does not exist yet
	err := service.TogglePauseTask(taskID)
	assert.NoError(t, err)

	taskRepo.AssertExpectations(t)
	engine.AssertExpectations(t)
}

func TestTogglePauseTask_ToResume(t *testing.T) {
	taskRepo := new(TaskRepositoryMock)
	engine := new(DownloadEngineMock)
	configRepo := new(ConfigRepositoryMock)

	taskID := "local_456"
	aria2GID := "aria2_456"

	existingTask := &task.Task{
		ID:     taskID,
		GID:    aria2GID,
		Status: task.StatusPaused,
	}

	taskRepo.On("GetByID", taskID).Return(existingTask, nil)
	engine.On("Unpause", aria2GID).Return(nil) // Should call Unpause when currently Paused
	taskRepo.On("Update", mock.MatchedBy(func(t *task.Task) bool {
		return t.ID == taskID && t.Status == task.StatusActive
	})).Return(nil)

	service := task.NewTaskService(taskRepo, engine, configRepo)
	err := service.TogglePauseTask(taskID)
	assert.NoError(t, err)
}

func TestTogglePauseTask_InvalidState(t *testing.T) {
	taskRepo := new(TaskRepositoryMock)
	engine := new(DownloadEngineMock)
	configRepo := new(ConfigRepositoryMock)

	taskID := "local_789"
	existingTask := &task.Task{ID: taskID, Status: task.StatusCompleted}

	taskRepo.On("GetByID", taskID).Return(existingTask, nil)

	service := task.NewTaskService(taskRepo, engine, configRepo)
	err := service.TogglePauseTask(taskID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot toggle pause")

	// Engine and Update should NEVER be touched for completed items
	engine.AssertNotCalled(t, "Pause", mock.Anything)
	taskRepo.AssertNotCalled(t, "Update", mock.Anything)
}
