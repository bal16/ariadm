package wailsbridge_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"ariadm/internal/domain/config"
	"ariadm/internal/domain/task"
	"ariadm/internal/infra/database"
	"ariadm/internal/infra/rpc"
	"ariadm/internal/ingress/wailsbridge"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ConfigMockForE2E satisfies the configuration repository without hitting actual user profiles
type ConfigMockForE2E struct{}

func (m *ConfigMockForE2E) Load() (*config.AppConfig, error) {
	return &config.AppConfig{
		DefaultDownloadPath: os.TempDir(),
		SpeedLimit:          0,
		MaxConcurrentTasks:  3,
		MinimizeToTray:      false,
	}, nil
}
func (m *ConfigMockForE2E) Save(cfg *config.AppConfig) error { return nil }

func setupTrueE2EContext(t *testing.T) (*wailsbridge.WailsBridge, func()) {
	database.IsDev = true
	testAppName := "ariadm_e2e_suite"
	testDBFile := "e2e_backend_live.db"
	testRPCPort := "6888"
	testRPCURL := "http://127.0.0.1:" + testRPCPort + "/jsonrpc"

	// 1. Programmatically spin up a live, isolated aria2c daemon process
	daemonCmd := exec.Command("aria2c",
		"--enable-rpc=true",
		"--rpc-listen-all=false",
		"--rpc-listen-port="+testRPCPort,
		"--quiet=true",
	)

	err := daemonCmd.Start()
	require.NoError(t, err, "Failed to initialize native aria2c system daemon dependency executable")

	// Allow a brief stabilization window for the local daemon to bind the network socket
	time.Sleep(300 * time.Millisecond)

	// 2. Initialize real SQLite infrastructure layers
	appDir, err := database.ResolveAppDir(testAppName)
	require.NoError(t, err)
	fullDBPath := filepath.Join(appDir, testDBFile)
	os.Remove(fullDBPath) // Wipe dirty historical databases

	taskRepo, err := database.NewSQLiteTaskRepository(testAppName, testDBFile)
	require.NoError(t, err)

	// 3. Initialize real JSON-RPC client targeting our test port
	realRPCClient := rpc.NewAria2Client(testRPCURL)

	// 4. Assemble the real domain services and controllers
	mockConfigRepo := &ConfigMockForE2E{}
	configService := config.NewConfigService(mockConfigRepo, realRPCClient)
	taskService := task.NewTaskService(taskRepo, realRPCClient, mockConfigRepo)
	bridge := wailsbridge.NewWailsBridge(configService, taskService)

	// 5. Define teardown routine to clean up the environment completely
	teardown := func() {
		// Stop the daemon process
		if daemonCmd.Process != nil {
			daemonCmd.Process.Kill()
			daemonCmd.Wait()
		}
		// Clear out the SQLite database file
		os.Remove(fullDBPath)
	}

	return bridge, teardown
}

func TestWailsBridge_E2E_SubmitAndPause_Lifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping heavy system environment E2E test loops")
	}

	bridge, teardown := setupTrueE2EContext(t)
	defer teardown()

	// Use a reliable, lightweight static asset to test real engine scheduling
	targetURL := "https://httpbin.org/bytes/1024"

	// --- 1. TEST REAL TRIGGER DOWNLOAD (Hits aria2c + inserts into SQLite) ---
	createdTask, err := bridge.TriggerNewDownload(targetURL)
	require.NoError(t, err)
	require.NotNil(t, createdTask)
	assert.NotEmpty(t, createdTask.GID, "Aria2 engine must generate and yield a valid tracking GID string")
	assert.Equal(t, string(task.StatusActive), string(createdTask.Status))

	// Verify the database recorded the task immediately
	queueAfterInsert, err := bridge.GetTasks()
	require.NoError(t, err)
	require.Len(t, queueAfterInsert, 1)
	assert.Equal(t, createdTask.ID, queueAfterInsert[0].ID)

	// --- 2. TEST REAL PAUSE OPERATIONS (Sends JSON-RPC command to process + updates SQLite) ---
	err = bridge.ToggleTaskPauseState(createdTask.ID)
	require.NoError(t, err)

	// Pull records out of the database to verify the state update persisted successfully
	updatedQueue, err := bridge.GetTasks()
	require.NoError(t, err)
	assert.Equal(t, "paused", string(updatedQueue[0].Status))
}

func TestWailsBridge_E2E_DownloadProperties_Verification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping heavy system environment E2E test loops")
	}

	bridge, teardown := setupTrueE2EContext(t)
	defer teardown()

	// 1. Target URL pointing to the Solid UI schema file configuration
	targetURL := "https://www.solid-ui.com/schema.json"

	// 2. Trigger the download through the Wails Bridge
	createdTask, err := bridge.TriggerNewDownload(targetURL)
	require.NoError(t, err)
	require.NotNil(t, createdTask)

	// Give the aria2 daemon a brief moment (1 second) to perform the HTTP handshake,
	// parse Content-Length headers, and determine the destination file properties.
	time.Sleep(1 * time.Second)

	// 3. Fetch the updated queue from the repository
	queue, err := bridge.GetTasks()
	require.NoError(t, err)
	require.NotEmpty(t, queue, "The download queue tracking matrix should not be empty")

	// 4. Evaluate the captured download properties
	targetTask := queue[0]

	assert.Equal(t, targetURL, targetTask.URL, "The stored URL property must match the input destination exactly")
	assert.NotEmpty(t, targetTask.GID, "The engine must have assigned a valid Aria2 GID string property")
	assert.True(t,
		string(targetTask.Status) == string(task.StatusActive) || string(targetTask.Status) == string(task.StatusCompleted),
		"Initial task state must be 'active' or 'complete', got: %s", targetTask.Status)

	// When Phase 7's core syncer calls `aria2.tellStatus`, these tracking fields
	// will be populated automatically from the running process bytes stream:
	t.Logf("Verified properties for GID [%s]: URL=%s, CreatedAt=%v",
		targetTask.GID, targetTask.URL, targetTask.CreatedAt)
}
