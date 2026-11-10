package httpserver_test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/bal16/ariadm/internal/ingress/httpserver"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// freePort reserves an ephemeral port and releases it immediately.
func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())

	return port
}

func TestStart_UsesNextPortWhenFirstIsBusy(t *testing.T) {
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = busy.Close() }()

	busyPort := busy.Addr().(*net.TCPAddr).Port
	nextPort := freePort(t)

	server := httpserver.NewHTTPServer([]int{busyPort, nextPort}, nil)
	port, err := server.Start()
	require.NoError(t, err)
	defer func() { _ = server.Stop() }()

	assert.Equal(t, nextPort, port)
	assert.Equal(t, nextPort, server.Port())
}

func TestStart_ReturnsErrorWhenRangeIsFull(t *testing.T) {
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = busy.Close() }()

	server := httpserver.NewHTTPServer([]int{busy.Addr().(*net.TCPAddr).Port}, nil)

	port, err := server.Start()

	require.Error(t, err)
	assert.Equal(t, 0, port)
	assert.Contains(t, err.Error(), "no usable HTTP port")
}

func TestHealth_ReportsAppIdentity(t *testing.T) {
	server := httpserver.NewHTTPServer([]int{freePort(t)}, nil)
	port, err := server.Start()
	require.NoError(t, err)
	defer func() { _ = server.Stop() }()

	response, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()

	require.Equal(t, http.StatusOK, response.StatusCode)

	var payload map[string]string
	require.NoError(t, json.NewDecoder(response.Body).Decode(&payload))
	assert.Equal(t, "ok", payload["status"])
	assert.Equal(t, "ariadm", payload["app"])
}

func TestStop_ClosesListener(t *testing.T) {
	server := httpserver.NewHTTPServer([]int{freePort(t)}, nil)
	port, err := server.Start()
	require.NoError(t, err)

	require.NoError(t, server.Stop())

	assert.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
		if err != nil {
			return true
		}
		_ = conn.Close()
		return false
	}, time.Second, 20*time.Millisecond)
}
