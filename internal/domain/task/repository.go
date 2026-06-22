package task

// TaskRepository handles database operations (SQLite)
type TaskRepository interface {
	Create(t *Task) error
	GetByID(id string) (*Task, error)
	GetByGID(gid string) (*Task, error)
	Update(t *Task) error
}

// DownloadEngine handles task manipulation commands sent to aria2c
type DownloadEngine interface {
	AddURI(url string, downloadPath string) (string, error) // Returns GID
	Pause(gid string) error
	Unpause(gid string) error
}
