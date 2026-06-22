package database

import (
	"ariadm/internal/domain/task"
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteTaskRepository struct {
	db *sql.DB
}

func NewSQLiteTaskRepository(appName string, dbPath string) (*SQLiteTaskRepository, error) {
	appDir, err := ResolveAppDir(appName)
	if err != nil {
		return nil, err
	}

	// Create the directory path if it doesn't exist yet (important for prod)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, err
	}

	fullDBPath := filepath.Join(appDir, dbPath)

	// Open connection using the pure Go sqlite driver
	db, err := sql.Open("sqlite", fullDBPath)
	if err != nil {
		return nil, err
	}

	// 2. Initialize the schema migration automatically
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		gid TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL,
		file_name TEXT,
		total_length INTEGER DEFAULT 0,
		completed_length INTEGER DEFAULT 0,
		speed INTEGER DEFAULT 0,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &SQLiteTaskRepository{db: db}, nil
}

// Create inserts a brand new download tracking entry
func (r *SQLiteTaskRepository) Create(t *task.Task) error {
	query := `INSERT INTO tasks (id, gid, url, file_name, total_length, completed_length, speed, status, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, t.ID, t.GID, t.URL, t.FileName, t.TotalLength, t.CompletedLength, t.Speed, string(t.Status), t.CreatedAt)
	return err
}

// GetByID fetches a task records using our internal localized ID keys
func (r *SQLiteTaskRepository) GetByID(id string) (*task.Task, error) {
	query := `SELECT id, gid, url, file_name, total_length, completed_length, speed, status, created_at FROM tasks WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var t task.Task
	var statusStr string
	var createdAtStr string

	err := row.Scan(&t.ID, &t.GID, &t.URL, &t.FileName, &t.TotalLength, &t.CompletedLength, &t.Speed, &statusStr, &createdAtStr)
	if err != nil {
		return nil, err
	}

	t.Status = task.Status(statusStr)
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr) // Standard SQLite datetime string parse fallback
	return &t, nil
}

// GetByGID fetches a task records using aria2c's generated unique identifier
func (r *SQLiteTaskRepository) GetByGID(gid string) (*task.Task, error) {
	query := `SELECT id, gid, url, file_name, total_length, completed_length, speed, status, created_at FROM tasks WHERE gid = ?`
	row := r.db.QueryRow(query, gid)

	var t task.Task
	var statusStr string
	var createdAtStr string

	err := row.Scan(&t.ID, &t.GID, &t.URL, &t.FileName, &t.TotalLength, &t.CompletedLength, &t.Speed, &statusStr, &createdAtStr)
	if err != nil {
		return nil, err
	}

	t.Status = task.Status(statusStr)
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	return &t, nil
}

// Update rewrites mutable attributes (progress meters, speeds, download status changes)
func (r *SQLiteTaskRepository) Update(t *task.Task) error {
	query := `UPDATE tasks SET file_name = ?, total_length = ?, completed_length = ?, speed = ?, status = ? WHERE id = ?`

	_, err := r.db.Exec(query, t.FileName, t.TotalLength, t.CompletedLength, t.Speed, string(t.Status), t.ID)
	return err
}

func (r *SQLiteTaskRepository) GetAll() ([]*task.Task, error) {
	query := `SELECT id, gid, url, file_name, total_length, completed_length, speed, status, created_at FROM tasks ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*task.Task
	for rows.Next() {
		var t task.Task
		var statusStr string
		var createdAtStr string

		err := rows.Scan(&t.ID, &t.GID, &t.URL, &t.FileName, &t.TotalLength, &t.CompletedLength, &t.Speed, &statusStr, &createdAtStr)
		if err != nil {
			return nil, err
		}

		t.Status = task.Status(statusStr)
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		tasks = append(tasks, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
