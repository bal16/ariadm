package daemon

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildArgs_IncludesSessionPersistence(t *testing.T) {
	dm := &DaemonManager{rpcPort: "6800", sessionPath: "/tmp/ariadm/session.txt"}

	args := dm.buildArgs()

	assert.Contains(t, args, "--save-session=/tmp/ariadm/session.txt")
	assert.Contains(t, args, "--input-file=/tmp/ariadm/session.txt")
	assert.Contains(t, args, "--save-session-interval=30")
}

func TestBuildArgs_OmitsSessionPersistenceWithoutPath(t *testing.T) {
	dm := &DaemonManager{rpcPort: "6800"}

	args := dm.buildArgs()

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

	_, port, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	require.NoError(t, err)

	dm := &DaemonManager{rpcPort: port}
	dm.requestShutdown()

	assert.Equal(t, "aria2.shutdown", method)
}

func TestRequestShutdown_SkipsEmptyPort(t *testing.T) {
	dm := &DaemonManager{}
	assert.NotPanics(t, dm.requestShutdown)
}
