package task

// TaskRepository handles database operations (SQLite)
type TaskRepository interface {
	Create(t *Task) error
	GetByID(id string) (*Task, error)
	GetByGID(gid string) (*Task, error)
	Update(t *Task) error
	GetAll() ([]*Task, error)
}

// Aria2Status is the live progress snapshot returned by aria2c for a single download
type Aria2Status struct {
	GID             string
	Status          string // "active", "waiting", "paused", "error", "complete", "removed"
	TotalLength     int64
	CompletedLength int64
	DownloadSpeed   int64
	FileName        string // Extracted from files[0].path
}

// DownloadEngine handles task manipulation commands sent to aria2c
type DownloadEngine interface {
	AddURI(url string, downloadPath string) (string, error) // Returns GID
	Pause(gid string) error
	Unpause(gid string) error
	TellStatus(gid string) (*Aria2Status, error) // Fetch live progress snapshot for one GID
}
