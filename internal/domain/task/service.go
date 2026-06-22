package task

import (
	"ariadm/internal/domain/config"
	"errors"
	"regexp"
	"time"
)

var urlPattern = regexp.MustCompile(`^(https?|ftp)://[^\s/$.?#].[^\s]*$`)

type TaskService struct {
	taskRepo   TaskRepository
	engine     DownloadEngine
	configRepo config.ConfigRepository
}

func NewTaskService(tr TaskRepository, de DownloadEngine, cr config.ConfigRepository) *TaskService {
	return &TaskService{
		taskRepo:   tr,
		engine:     de,
		configRepo: cr,
	}
}

func (s *TaskService) DownloadFile(url string) (*Task, error) {
	if !urlPattern.MatchString(url) {
		return nil, ErrInvalidURL
	}

	// 2. Get the current download destination path from configurations
	cfg, err := s.configRepo.Load()
	if err != nil {
		return nil, errors.New("failed to load configuration details")
	}

	// 3. Dispatch the URL request to the aria2c engine instance
	gid, err := s.engine.AddURI(url, cfg.DefaultDownloadPath)
	if err != nil {
		return nil, err
	}

	// 4. Construct the local tracking model
	newTask := &Task{
		ID:        "local_" + gid,
		GID:       gid,
		URL:       url,
		Status:    StatusDownloading,
		CreatedAt: time.Now(),
	}

	// 5. Persist the tracking entity to our database layers
	if err := s.taskRepo.Create(newTask); err != nil {
		return nil, err
	}

	return newTask, nil
}

func (s *TaskService) TogglePauseTask(id string) error {
	// 1. Fetch the target task from the tracking database
	t, err := s.taskRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 2. Evaluate the current task state machine status
	switch t.Status {
	case StatusDownloading:
		// Instruct aria2c engine to halt network operations
		if err := s.engine.Pause(t.GID); err != nil {
			return err
		}
		t.Status = StatusPaused

	case StatusPaused:
		// Instruct aria2c engine to fire up network segments again
		if err := s.engine.Unpause(t.GID); err != nil {
			return err
		}
		t.Status = StatusDownloading

	default:
		// Block transitions if the file has already completed downloading or errored out
		return ErrCannotTogglePause
	}

	// 3. Write back the refreshed state properties into SQLite
	return s.taskRepo.Update(t)
}
