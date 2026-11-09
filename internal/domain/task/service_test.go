package task_test

import (
	"ariadm/internal/domain/config"
	"ariadm/internal/domain/task"
	"errors"
	"testing"
	"time"

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
func (m *TaskRepositoryMock) Delete(id string) error    { return m.Called(id).Error(0) }

type DownloadEngineMock struct{ mock.Mock }

func (m *DownloadEngineMock) AddURI(url, path string, fileName string) (string, error) {
	args := m.Called(url, path)
	return args.String(0), args.Error(1)
}
func (m *DownloadEngineMock) Pause(gid string) error   { return m.Called(gid).Error(0) }
func (m *DownloadEngineMock) Unpause(gid string) error { return m.Called(gid).Error(0) }
func (m *DownloadEngineMock) Remove(gid string) error  { return m.Called(gid).Error(0) }
func (m *DownloadEngineMock) RemoveDownloadResult(gid string) error {
	return m.Called(gid).Error(0)
}
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

	// configRepo.Load is called internally by TogglePauseTask for the healing path
	configRepo.On("Load").Return(&config.AppConfig{DefaultDownloadPath: "/downloads"}, nil)

	// 2. Expect engine to pause the task via aria2c GID
	engine.On("Pause", aria2GID).Return(nil)

	// 3. Expect database to store the updated "paused" status
	taskRepo.On("Update", mock.MatchedBy(func(t *task.Task) bool {
		return t.ID == taskID && t.Status == task.StatusPaused
	})).Return(nil)

	service := task.NewTaskService(taskRepo, engine, configRepo)
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

	// configRepo.Load is called internally by TogglePauseTask for the healing path
	configRepo.On("Load").Return(&config.AppConfig{DefaultDownloadPath: "/downloads"}, nil)

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

	// configRepo.Load is called before the switch statement, even for terminal states
	configRepo.On("Load").Return(&config.AppConfig{DefaultDownloadPath: "/downloads"}, nil)

	service := task.NewTaskService(taskRepo, engine, configRepo)
	err := service.TogglePauseTask(taskID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot toggle pause")

	// Engine and Update should NEVER be touched for completed items
	engine.AssertNotCalled(t, "Pause", mock.Anything)
	taskRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestDeleteTask_ActiveTask(t *testing.T) {
	taskRepo := new(TaskRepositoryMock)
	engine := new(DownloadEngineMock)
	configRepo := new(ConfigRepositoryMock)

	taskID := "local_del_active"
	gid := "aria2_del_active"

	existingTask := &task.Task{ID: taskID, GID: gid, Status: task.StatusActive}

	// Service must: fetch the task, call aria2.remove, then delete from SQLite
	taskRepo.On("GetByID", taskID).Return(existingTask, nil)
	engine.On("Remove", gid).Return(nil)
	taskRepo.On("Delete", taskID).Return(nil)

	service := task.NewTaskService(taskRepo, engine, configRepo)
	err := service.DeleteTask(taskID, false)

	assert.NoError(t, err)
	taskRepo.AssertExpectations(t)
	engine.AssertExpectations(t)
	// RemoveDownloadResult must NOT be called for an active task
	engine.AssertNotCalled(t, "RemoveDownloadResult", mock.Anything)
}

func TestDeleteTask_CompletedTask(t *testing.T) {
	taskRepo := new(TaskRepositoryMock)
	engine := new(DownloadEngineMock)
	configRepo := new(ConfigRepositoryMock)

	taskID := "local_del_complete"
	gid := "aria2_del_complete"

	existingTask := &task.Task{ID: taskID, GID: gid, Status: task.StatusCompleted}

	// Service must: fetch the task, call aria2.removeDownloadResult, then delete from SQLite
	taskRepo.On("GetByID", taskID).Return(existingTask, nil)
	engine.On("RemoveDownloadResult", gid).Return(nil)
	taskRepo.On("Delete", taskID).Return(nil)

	service := task.NewTaskService(taskRepo, engine, configRepo)
	err := service.DeleteTask(taskID, false)

	assert.NoError(t, err)
	taskRepo.AssertExpectations(t)
	engine.AssertExpectations(t)
	// Remove must NOT be called for a completed task
	engine.AssertNotCalled(t, "Remove", mock.Anything)
}

func TestTaskService_ReconcileSessionTasks_OrphanRecovery(t *testing.T) {
	mockRepo := new(TaskRepositoryMock)
	mockEngine := new(DownloadEngineMock) // Reusing the mock engine interface from previous iterations
	mockConfig := new(ConfigRepositoryMock)

	service := task.NewTaskService(mockRepo, mockEngine, mockConfig)

	// Setup: two tasks that are active/paused and should be in aria2c's queue
	orphanActive := &task.Task{
		ID:     "local_orphan_1",
		GID:    "stale_gid_1",
		URL:    "https://example.com/file1.zip",
		Status: task.StatusActive,
	}
	orphanPaused := &task.Task{
		ID:     "local_orphan_2",
		GID:    "stale_gid_2",
		URL:    "https://example.com/file2.zip",
		Status: task.StatusPaused,
	}

	mockConfig.On("Load").Return(&config.AppConfig{DefaultDownloadPath: "/downloads"}, nil)

	// Both tasks are "orphaned" — aria2c no longer knows about them
	mockRepo.On("GetAll").Return([]*task.Task{orphanActive, orphanPaused}, nil)
	mockEngine.On("TellStatus", "stale_gid_1").Return(nil, errors.New("GID stale_gid_1 is not found"))
	mockEngine.On("TellStatus", "stale_gid_2").Return(nil, errors.New("GID stale_gid_2 is not found"))

	// Engine should re-inject both URLs so the downloads can resume
	mockEngine.On("AddURI", orphanActive.URL, "/downloads").Return("new_gid_1", nil)
	mockEngine.On("AddURI", orphanPaused.URL, "/downloads").Return("new_gid_2", nil)

	// Repo should be updated with the new GIDs
	mockRepo.On("Update", mock.MatchedBy(func(t *task.Task) bool {
		return t.ID == "local_orphan_1" && t.GID == "new_gid_1"
	})).Return(nil)
	mockRepo.On("Update", mock.MatchedBy(func(t *task.Task) bool {
		return t.ID == "local_orphan_2" && t.GID == "new_gid_2"
	})).Return(nil)

	err := service.ReconcileSessionTasks()

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
	mockConfig.AssertExpectations(t)
}

func TestTaskService_ReconcileSessionTasks_EngineFailure(t *testing.T) {
	mockRepo := new(TaskRepositoryMock)
	mockEngine := new(DownloadEngineMock)
	mockConfig := new(ConfigRepositoryMock)

	service := task.NewTaskService(mockRepo, mockEngine, mockConfig)

	orphanTask := &task.Task{
		ID:     "local_fail",
		GID:    "stale_gid_fail",
		URL:    "https://example.com/fail.zip",
		Status: task.StatusActive,
	}

	mockConfig.On("Load").Return(&config.AppConfig{DefaultDownloadPath: "/downloads"}, nil)
	mockRepo.On("GetAll").Return([]*task.Task{orphanTask}, nil)

	// aria2c does not know about this GID
	mockEngine.On("TellStatus", "stale_gid_fail").Return(nil, errors.New("GID stale_gid_fail is not found"))

	// Re-injection also fails (e.g. aria2c is down)
	mockEngine.On("AddURI", orphanTask.URL, "/downloads").Return("", errors.New("connection refused"))

	// Task should be marked as error in the DB
	mockRepo.On("Update", mock.MatchedBy(func(t *task.Task) bool {
		return t.ID == "local_fail" && t.Status == task.StatusError
	})).Return(nil)

	err := service.ReconcileSessionTasks()

	// The function should not propagate this error — it logs and continues
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestTaskService_ReconcileSessionTasks_SkipsTerminalStates(t *testing.T) {
	mockRepo := new(TaskRepositoryMock)
	mockEngine := new(DownloadEngineMock)
	mockConfig := new(ConfigRepositoryMock)

	service := task.NewTaskService(mockRepo, mockEngine, mockConfig)

	completedTask := &task.Task{
		ID:        "local_done",
		GID:       "done_gid",
		Status:    task.StatusCompleted,
		CreatedAt: time.Now(),
	}
	errorTask := &task.Task{
		ID:        "local_err",
		GID:       "error_gid",
		Status:    task.StatusError,
		CreatedAt: time.Now(),
	}

	mockConfig.On("Load").Return(&config.AppConfig{DefaultDownloadPath: "/downloads"}, nil)
	mockRepo.On("GetAll").Return([]*task.Task{completedTask, errorTask}, nil)

	err := service.ReconcileSessionTasks()

	assert.NoError(t, err)

	// TellStatus must NEVER be called for terminal states
	mockEngine.AssertNotCalled(t, "TellStatus", mock.Anything)
}
