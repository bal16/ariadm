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
		Status:    StatusActive,
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
	case StatusActive:
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
		t.Status = StatusActive

	default:
		// Block transitions if the file has already completed downloading or errored out
		return ErrCannotTogglePause
	}

	// 3. Write back the refreshed state properties into SQLite
	return s.taskRepo.Update(t)
}

func (s *TaskService) GetAllTasks() ([]*Task, error) {
	return s.taskRepo.GetAll()
}

// SyncAndGetAllTasks fetches every task from the database, enriches each one with
// a live aria2c TellStatus call, persists any changed fields, then returns the
// updated list. The telemetry loop in the Wails bridge calls this every 500 ms.
func (s *TaskService) SyncAndGetAllTasks() ([]*Task, error) {
	tasks, err := s.taskRepo.GetAll()
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		// Only poll aria2c for tasks that are actively transferring data.
		// Paused tasks are in a user-controlled stable state; completed/errored
		// tasks are terminal. Polling these would be wasteful and could cause
		// stale aria2c data to overwrite deliberate DB state.
		if t.Status != StatusActive && t.Status != StatusWaiting {
			continue
		}

		live, err := s.engine.TellStatus(t.GID)
		if err != nil {
			// aria2c may not know about this GID (e.g. daemon restarted) — skip silently
			continue
		}

		// Map aria2c status string to our internal Status type
		newStatus := Status(live.Status) // e.g. "active", "paused", "complete", "error"
		if live.Status == "complete" {
			newStatus = StatusCompleted // use our canonical constant value
		}

		// Detect whether any field actually changed before writing to SQLite
		changed := t.TotalLength != live.TotalLength ||
			t.CompletedLength != live.CompletedLength ||
			t.Speed != live.DownloadSpeed ||
			t.Status != newStatus ||
			(live.FileName != "" && t.FileName != live.FileName)

		if changed {
			t.TotalLength = live.TotalLength
			t.CompletedLength = live.CompletedLength
			t.Speed = live.DownloadSpeed
			t.Status = newStatus
			if live.FileName != "" {
				t.FileName = live.FileName
			}
			// Best-effort write-back — don't fail the whole list if one update errors
			_ = s.taskRepo.Update(t)
		}
	}

	return tasks, nil
}
