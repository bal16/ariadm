package task

import (
	"errors"
	"time"
)

type Status string

var (
	ErrInvalidURL        = errors.New("invalid url format: must use http, https, or ftp protocols")
	ErrCannotTogglePause = errors.New("cannot toggle pause: task is already completed or ended in an error status")
)

const (
	StatusWaiting   Status = "waiting"
	StatusActive    Status = "active"    // Maps to aria2c "active" status
	StatusPaused    Status = "paused"
	StatusCompleted Status = "complete"  // aria2c uses "complete" not "completed"
	StatusError     Status = "error"
)

type Task struct {
	ID              string    `json:"id"`
	GID             string    `json:"gid"` // aria2c unique download ID
	URL             string    `json:"url"`
	FileName        string    `json:"file_name"`
	TotalLength     int64     `json:"total_length"`
	CompletedLength int64     `json:"completed_length"`
	Speed           int64     `json:"speed"`
	Status          Status    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}
