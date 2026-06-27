package task

import (
	"ariadm/internal/domain/config"
	"errors"
	"log"
	"os"
	"regexp"
	"strings"
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
	gid, err := s.engine.AddURI(url, cfg.DefaultDownloadPath, "")
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

	// Fetch download directory layout in case we need to heal an orphaned task tracking block
	cfg, err := s.configRepo.Load()
	if err != nil {
		return err
	}

	// 2. Evaluate the current task state machine status
	switch t.Status {
	case StatusActive:
		// Instruct aria2c engine to halt network operations
		if err := s.engine.Pause(t.GID); err != nil {
			// If the daemon was reset and the GID is already gone, just force local synchronization
			if strings.Contains(err.Error(), "is not found") {
				t.Status = StatusPaused
				break
			}
			return err
		}
		t.Status = StatusPaused

	case StatusPaused:
		// Instruct aria2c engine to fire up network segments again
		if err := s.engine.Unpause(t.GID); err != nil {
			// 🎯 ON-THE-FLY HEALING: Handle daemon session wipe or dropped GID environments
			if strings.Contains(err.Error(), "is not found") {
				// Re-inject link with the verified filename to attach to the existing .aria2 chunk file
				newGID, rpcErr := s.engine.AddURI(t.URL, cfg.DefaultDownloadPath, t.FileName)
				if rpcErr != nil {
					return rpcErr
				}

				// Assign the fresh daemon tracking key and push the state machine back to active
				t.GID = newGID
				t.Status = StatusActive
				break
			}
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

// DeleteTask removes the task from the database, tells aria2c to drop it, and optionally deletes physical files
func (s *TaskService) DeleteTask(id string, deleteFiles bool) error {
	t, err := s.taskRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 1. If physical deletion is requested, try to get the file paths before removing from aria2c
	var filesToDelete []string
	if deleteFiles {
		live, err := s.engine.TellStatus(t.GID)
		if err == nil && len(live.Files) > 0 {
			// aria2c knows about the files
			filesToDelete = live.Files
		} else if t.FileName != "" {
			// fallback: guess the path based on the default config if aria2c forgot it
			if cfg, err := s.configRepo.Load(); err == nil {
				importPath := cfg.DefaultDownloadPath + "/" + t.FileName
				filesToDelete = []string{importPath}
			}
		}
	}

	// 2. Notify aria2c to drop the download — ignore errors if the daemon
	//    was restarted and no longer knows about this GID.
	switch t.Status {
	case StatusActive, StatusWaiting, StatusPaused:
		// Still in aria2c's active queue: force-stop and remove it
		_ = s.engine.Remove(t.GID)
	case StatusCompleted, StatusError:
		// In aria2c's result list: purge the finished entry from memory
		_ = s.engine.RemoveDownloadResult(t.GID)
	}

	// 3. Perform physical deletion if requested
	if deleteFiles {
		for _, f := range filesToDelete {
			_ = os.Remove(f)
		}
	}

	// 4. Always delete from SQLite regardless of aria2c's response
	return s.taskRepo.Delete(id)
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

// ReconcileSessionTasks inspects the local database tracking queue on startup and
// automatically heals or re-injects tasks that were dropped from active daemon memory.
func (s *TaskService) ReconcileSessionTasks() error {
	tasks, err := s.taskRepo.GetAll()
	if err != nil {
		return err
	}

	// downloadDir := s.configProvider.GetDefaultDownloadPath()
	cfg, err := s.configRepo.Load()
	if err != nil {
		return errors.New("failed to load configuration details")
	}

	for _, t := range tasks {
		// We only care about tracking states that should actively exist in a runtime queue
		if t.Status == StatusActive || t.Status == StatusPaused {
			_, err := s.engine.TellStatus(t.GID)
			if err != nil {
				// GID is dead or missing from the current aria2c process session context
				// Re-inject the resource URL. Aria2c automatically picks up existing files natively.
				log.Printf("%v", t)
				newGID, rpcErr := s.engine.AddURI(t.URL, cfg.DefaultDownloadPath, t.FileName)
				if rpcErr != nil {
					// If the engine fails completely, flag the task tracking row locally as an error
					log.Printf("error happen %v", rpcErr)
					t.Status = StatusError
					_ = s.taskRepo.Update(t)
					continue
				}

				// Map the freshly generated runtime identifier back to the system entity record
				t.GID = newGID
				err = s.taskRepo.Update(t)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
