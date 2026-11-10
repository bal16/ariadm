package daemon

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildArgs_IncludesSessionPersistence(t *testing.T) {
	dm := &DaemonManager{sessionPath: "/tmp/ariadm/session.txt"}

	args := dm.buildArgs(6800)

	assert.Contains(t, args, "--save-session=/tmp/ariadm/session.txt")
	assert.Contains(t, args, "--input-file=/tmp/ariadm/session.txt")
	assert.Contains(t, args, "--save-session-interval=30")
}

func TestBuildArgs_IncludesSelectedPort(t *testing.T) {
	dm := &DaemonManager{}

	args := dm.buildArgs(6803)

	assert.Contains(t, args, "--rpc-listen-port=6803")
}

func TestBuildArgs_OmitsSessionPersistenceWithoutPath(t *testing.T) {
	dm := &DaemonManager{}

	args := dm.buildArgs(6800)

	for _, arg := range args {
		assert.NotContains(t, arg, "save-session")
		assert.NotContains(t, arg, "input-file")
	}
}

func TestEnsureSessionFile_CreatesPrivateEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "session.txt")

	require.NoError(t, ensureSessionFile(path))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Empty(t, content)
}

func TestRequestShutdown_CallsAria2Shutdown(t *testing.T) {
	var method string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Method string `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		method = request.Method

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"ariadm-shutdown","result":"OK"}`))
	}))
	defer server.Close()

	_, portValue, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	require.NoError(t, err)
	port, err := strconv.Atoi(portValue)
	require.NoError(t, err)

	dm := &DaemonManager{}
	dm.requestShutdown(port)

	assert.Equal(t, "aria2.shutdown", method)
}

func TestRequestShutdown_SkipsEmptyPort(t *testing.T) {
	dm := &DaemonManager{}
	assert.NotPanics(t, func() { dm.requestShutdown(0) })
}

func TestStart_ReturnsSelectionError(t *testing.T) {
	dm := &DaemonManager{
		binaryPath:  "aria2c",
		ports:       []int{6800},
		sessionPath: filepath.Join(t.TempDir(), "session.txt"),
		prober:      &fakeProber{},
		selectPort: func([]int, Prober, func(int) bool) (int, []PortStatus, error) {
			return 0, []PortStatus{{Port: 6800, Kind: PortForeign}}, ErrNoPortAvailable
		},
	}

	port, err := dm.Start()

	assert.Equal(t, 0, port)
	assert.ErrorIs(t, err, ErrNoPortAvailable)
}

func TestPIDFile_WriteAndRemove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "daemon.pid")

	require.NoError(t, writePIDFile(path, 4321))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "4321\n", string(content))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	removePIDFile(path)
	assert.NoFileExists(t, path)
}

func TestPIDFile_SkipsEmptyPath(t *testing.T) {
	require.NoError(t, writePIDFile("", 1))
	assert.NotPanics(t, func() { removePIDFile("") })
}
