package database_test

import (
	"ariadm/internal/domain/task"
	"ariadm/internal/infra/database"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestSQLiteTaskRepository_Integration(t *testing.T) {
	database.IsDev = true

	testAppName := "ariadm_test"
	testDBPath := "integration_test.db"

	// Resolve the absolute file target to clean up stale test runs cleanly
	appDir, err := database.ResolveAppDir(testAppName)
	require.NoError(t, err)
	fullPath := filepath.Join(appDir, testDBPath)

	// Clean before and after run execution
	_ = os.Remove(fullPath)
	defer func() { _ = os.Remove(fullPath) }()

	repo, err := database.NewSQLiteTaskRepository(testAppName, testDBPath)
	require.NoError(t, err)
	require.NotNil(t, repo)

	now := time.Now().UTC().Truncate(time.Second)
	dummyTask := &task.Task{
		ID:              "local_test_123",
		GID:             "aria2_gid_test_123",
		URL:             "https://example.com/linux-distro.iso",
		FileName:        "linux-distro.iso",
		TotalLength:     2147483648,
		CompletedLength: 0,
		Speed:           0,
		Status:          task.StatusActive, // Matches aria2c "active" status
		CreatedAt:       now,
	}

	// 3. CREATE
	err = repo.Create(dummyTask)
	require.NoError(t, err) // Stop right here if insertion fails

	// 4. GET BY ID
	fetchedByID, err := repo.GetByID(dummyTask.ID)
	require.NoError(t, err)
	assert.Equal(t, dummyTask.ID, fetchedByID.ID)
	assert.Equal(t, dummyTask.GID, fetchedByID.GID)
	assert.Equal(t, string(task.StatusActive), string(fetchedByID.Status))

	// 5. UPDATE
	dummyTask.CompletedLength = 1073741824
	dummyTask.Speed = 5242880
	dummyTask.Status = task.StatusPaused

	err = repo.Update(dummyTask)
	require.NoError(t, err)

	updatedTask, err := repo.GetByID(dummyTask.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1073741824), updatedTask.CompletedLength)
	assert.Equal(t, int64(5242880), updatedTask.Speed)
	assert.Equal(t, task.StatusPaused, updatedTask.Status)

	// 6. GET BY GID
	fetchedByGID, err := repo.GetByGID(dummyTask.GID)
	require.NoError(t, err)
	assert.Equal(t, dummyTask.ID, fetchedByGID.ID)

	// 7. UPDATE GID
	// Re-injection and reconciliation assign a fresh daemon GID to an existing
	// task. Update must persist it, otherwise telemetry keeps polling a GID that
	// the engine no longer knows and the task stays "active" at 0 B/s.
	oldGID := dummyTask.GID
	dummyTask.GID = "aria2_gid_updated_456"

	err = repo.Update(dummyTask)
	require.NoError(t, err)

	updatedByNewGID, err := repo.GetByGID(dummyTask.GID)
	require.NoError(t, err)
	assert.Equal(t, dummyTask.ID, updatedByNewGID.ID)

	_, err = repo.GetByGID(oldGID)
	assert.Error(t, err, "the stale GID must no longer resolve to the task")
}
