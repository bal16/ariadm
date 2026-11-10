package daemon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	defaultProbeTimeout   = 500 * time.Millisecond
	defaultReclaimTimeout = 5 * time.Second
	reclaimPollInterval   = 100 * time.Millisecond
)

// rpcProber identifies aria2 daemons over JSON-RPC. Ownership is decided from
// getGlobalOption.save-session, which works the same on every platform and
// keeps a user's manually started aria2c safe from being stopped.
type rpcProber struct {
	sessionPath    string
	timeout        time.Duration
	reclaimTimeout time.Duration
	client         *http.Client
}

func newRPCProber(sessionPath string, timeout time.Duration) *rpcProber {
	return &rpcProber{
		sessionPath:    normalizePath(sessionPath),
		timeout:        timeout,
		reclaimTimeout: defaultReclaimTimeout,
		client:         &http.Client{Timeout: timeout},
	}
}

// Identify classifies a port as free, ours, or foreign.
func (p *rpcProber) Identify(port int) PortKind {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), p.timeout)
	if err != nil {
		return PortFree
	}
	_ = conn.Close()

	if _, err := p.call(port, "aria2.getVersion", nil); err != nil {
		return PortForeign
	}

	rawOptions, err := p.call(port, "aria2.getGlobalOption", nil)
	if err != nil {
		return PortForeign
	}

	var options map[string]string
	if err := json.Unmarshal(rawOptions, &options); err != nil {
		return PortForeign
	}

	if samePath(options["save-session"], p.sessionPath) {
		return PortOurs
	}
	return PortForeign
}

// Reclaim asks a daemon we own to shut down and waits until the port is free.
func (p *rpcProber) Reclaim(port int) error {
	if _, err := p.call(port, "aria2.shutdown", nil); err != nil {
		return fmt.Errorf("failed to shut down aria2 on port %d: %w", port, err)
	}

	deadline := time.Now().Add(p.reclaimTimeout)
	for time.Now().Before(deadline) {
		if p.Identify(port) != PortOurs {
			return nil
		}
		time.Sleep(reclaimPollInterval)
	}

	return fmt.Errorf("aria2 on port %d did not shut down within %s", port, p.reclaimTimeout)
}

func (p *rpcProber) call(port int, method string, params []interface{}) (json.RawMessage, error) {
	if params == nil {
		params = []interface{}{}
	}

	payload, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "ariadm-probe",
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Post(
		fmt.Sprintf("http://127.0.0.1:%d/jsonrpc", port),
		"application/json",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	if envelope.Error != nil {
		return nil, fmt.Errorf("rpc error: %s", envelope.Error.Message)
	}

	return envelope.Result, nil
}

// normalizePath makes session paths comparable regardless of how aria2 echoes
// them back. Symlinks are resolved when the path exists.
func normalizePath(path string) string {
	if path == "" {
		return ""
	}

	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return filepath.Clean(path)
}

func samePath(a, b string) bool {
	if a == "" || b == "" {
		return false
	}

	left := normalizePath(a)
	right := normalizePath(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
