package database_test

import (
	"ariadm/internal/domain/task"
	"ariadm/internal/infra/database"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestSQLiteTaskRepository_Integration(t *testing.T) {
	// 1. Setup: Specify a dedicated test database path
	testDBPath := "integration_test.db"

	// Ensure a clean environment before starting, and remove the file after the test is done
	os.Remove(testDBPath)
	defer os.Remove(testDBPath)

	// 2. Initialize the SQLiteTaskRepository
	repo, err := database.NewSQLiteTaskRepository(testDBPath)
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	// Prepare dummy data with truncated time (Truncate) to microseconds
	// because SQLite rounds the time string when stored.
	now := time.Now().UTC().Truncate(time.Second)
	dummyTask := &task.Task{
		ID:              "local_test_123",
		GID:             "aria2_gid_test_123",
		URL:             "https://example.com/linux-distro.iso",
		FileName:        "linux-distro.iso",
		TotalLength:     2147483648, // 2 GB
		CompletedLength: 0,
		Speed:           0,
		Status:          task.StatusDownloading,
		CreatedAt:       now,
	}

	// 3. TEST EXECUTION: [CREATE]
	err = repo.Create(dummyTask)
	assert.NoError(t, err)

	// 4. TEST EXECUTION: [GET BY ID]
	fetchedByID, err := repo.GetByID(dummyTask.ID)
	assert.NoError(t, err)
	assert.Equal(t, dummyTask.ID, fetchedByID.ID)
	assert.Equal(t, dummyTask.GID, fetchedByID.GID)
	assert.Equal(t, dummyTask.URL, fetchedByID.URL)
	assert.Equal(t, dummyTask.FileName, fetchedByID.FileName)
	assert.Equal(t, dummyTask.TotalLength, fetchedByID.TotalLength)
	assert.Equal(t, task.StatusDownloading, fetchedByID.Status)

	// 5. TEST EXECUTION: [UPDATE]
	dummyTask.CompletedLength = 1073741824 // 1 GB downloaded
	dummyTask.Speed = 5242880              // 5 MB/s
	dummyTask.Status = task.StatusPaused   // Change status to Paused

	err = repo.Update(dummyTask)
	assert.NoError(t, err)

	// Verify that the changes are reflected in the database
	updatedTask, err := repo.GetByID(dummyTask.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1073741824), updatedTask.CompletedLength)
	assert.Equal(t, int64(5242880), updatedTask.Speed)
	assert.Equal(t, task.StatusPaused, updatedTask.Status)

	// 6. TEST EXECUTION: [GET BY GID]
	fetchedByGID, err := repo.GetByGID(dummyTask.GID)
	assert.NoError(t, err)
	assert.Equal(t, dummyTask.ID, fetchedByGID.ID)
	assert.Equal(t, string(task.StatusPaused), string(fetchedByGID.Status))
}
