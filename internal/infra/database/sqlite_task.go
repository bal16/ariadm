package database

import (
	"ariadm/internal/domain/task"
	"database/sql"
	"time"
)

type SQLiteTaskRepository struct {
	db *sql.DB
}

func NewSQLiteTaskRepository(dbPath string) (*SQLiteTaskRepository, error) {
	// 1. Open connection using the pure Go sqlite driver
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// 2. Initialize the schema migration automatically
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		gid TEXT NOT nil UNIQUE,
		url TEXT NOT nil,
		file_name TEXT,
		total_length INTEGER DEFAULT 0,
		completed_length INTEGER DEFAULT 0,
		speed INTEGER DEFAULT 0,
		status TEXT NOT nil,
		created_at DATETIME NOT nil
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
