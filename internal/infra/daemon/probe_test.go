package daemon

import (
	"encoding/json"
	"net"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startRPCStub serves a JSON-RPC endpoint and returns its port. The handler
// receives the method name and returns the raw result payload.
func startRPCStub(t *testing.T, result func(method string) (string, bool)) (int, *http.Server) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Method string `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)

		payload, ok := result(request.Method)
		if !ok {
			http.Error(w, "unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"ariadm-probe","result":` + payload + `}`))
	})}

	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { _ = server.Close() })

	return listener.Addr().(*net.TCPAddr).Port, server
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())

	return port
}

func TestIdentify_FreeWhenNothingListens(t *testing.T) {
	prober := newRPCProber(filepath.Join(t.TempDir(), "session.txt"), 200*time.Millisecond)

	assert.Equal(t, PortFree, prober.Identify(freePort(t)))
}

func TestIdentify_OursWhenSessionMatches(t *testing.T) {
	sessionPath := filepath.Join(t.TempDir(), "session.txt")

	port, _ := startRPCStub(t, func(method string) (string, bool) {
		switch method {
		case "aria2.getVersion":
			return `{"version":"1.37.0"}`, true
		case "aria2.getGlobalOption":
			return `{"save-session":` + quote(sessionPath) + `}`, true
		default:
			return "", false
		}
	})

	prober := newRPCProber(sessionPath, 200*time.Millisecond)

	assert.Equal(t, PortOurs, prober.Identify(port))
}

func TestIdentify_ForeignWhenSessionDiffers(t *testing.T) {
	sessionPath := filepath.Join(t.TempDir(), "session.txt")

	port, _ := startRPCStub(t, func(method string) (string, bool) {
		switch method {
		case "aria2.getVersion":
			return `{"version":"1.37.0"}`, true
		case "aria2.getGlobalOption":
			return `{"save-session":"/home/user/other-session.txt"}`, true
		default:
			return "", false
		}
	})

	prober := newRPCProber(sessionPath, 200*time.Millisecond)

	assert.Equal(t, PortForeign, prober.Identify(port))
}

func TestIdentify_ForeignWhenResponseIsNotJSON(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json"))
	})}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { _ = server.Close() })

	prober := newRPCProber(filepath.Join(t.TempDir(), "session.txt"), 200*time.Millisecond)

	assert.Equal(t, PortForeign, prober.Identify(listener.Addr().(*net.TCPAddr).Port))
}

func TestReclaim_ShutsDownThenWaitsForPortToClose(t *testing.T) {
	sessionPath := filepath.Join(t.TempDir(), "session.txt")

	var server *http.Server
	var port int
	port, server = startRPCStub(t, func(method string) (string, bool) {
		switch method {
		case "aria2.getVersion":
			return `{"version":"1.37.0"}`, true
		case "aria2.getGlobalOption":
			return `{"save-session":` + quote(sessionPath) + `}`, true
		case "aria2.shutdown":
			go func() {
				time.Sleep(20 * time.Millisecond)
				_ = server.Close()
			}()
			return `"OK"`, true
		default:
			return "", false
		}
	})

	prober := newRPCProber(sessionPath, 200*time.Millisecond)
	require.Equal(t, PortOurs, prober.Identify(port))

	require.NoError(t, prober.Reclaim(port))
	assert.Equal(t, PortFree, prober.Identify(port))
}

func TestReclaim_ReturnsErrorWhenShutdownFails(t *testing.T) {
	sessionPath := filepath.Join(t.TempDir(), "session.txt")

	port, _ := startRPCStub(t, func(method string) (string, bool) {
		if method == "aria2.shutdown" {
			return "", false
		}
		return `{"version":"1.37.0"}`, true
	})

	prober := newRPCProber(sessionPath, 200*time.Millisecond)

	assert.Error(t, prober.Reclaim(port))
}

func quote(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}
