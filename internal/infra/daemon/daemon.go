package daemon

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	shutdownTimeout    = 5 * time.Second
	sessionSaveSeconds = 30
	startTimeout       = 3 * time.Second
)

// DaemonManager owns the aria2c child process.
type DaemonManager struct {
	cmd         *exec.Cmd
	binaryPath  string
	ports       []int
	sessionPath string
	pidPath     string
	rpcPort     int

	selectPort func(ports []int, prober Prober, canBind func(int) bool) (int, []PortStatus, error)
	prober     Prober
}

func NewDaemonManager(binaryPath string, ports []int, sessionPath string, pidPath string) *DaemonManager {
	return &DaemonManager{
		binaryPath:  binaryPath,
		ports:       ports,
		sessionPath: sessionPath,
		pidPath:     pidPath,
		selectPort:  SelectEnginePort,
	}
}

// Start reclaims orphaned instances, launches aria2c on the first usable port
// and waits until its RPC endpoint answers. It returns the bound port.
func (dm *DaemonManager) Start() (int, error) {
	if dm.cmd != nil && dm.cmd.Process != nil {
		return dm.rpcPort, errors.New("aria2c daemon is already running")
	}

	if dm.sessionPath != "" {
		if err := ensureSessionFile(dm.sessionPath); err != nil {
			return 0, err
		}
	}

	prober := dm.prober
	if prober == nil {
		prober = newRPCProber(dm.sessionPath, defaultProbeTimeout)
	}
	selector := dm.selectPort
	if selector == nil {
		selector = SelectEnginePort
	}

	remaining := append([]int(nil), dm.ports...)
	var lastErr error

	for len(remaining) > 0 {
		port, _, err := selector(remaining, prober, canBindPort)
		if err != nil {
			if lastErr != nil {
				return 0, lastErr
			}
			return 0, err
		}

		cmd := exec.Command(dm.binaryPath, dm.buildArgs(port)...)
		setPlatformAttributes(cmd)

		if err := cmd.Start(); err != nil {
			lastErr = fmt.Errorf("failed to launch aria2c on port %d: %w", port, err)
			remaining = withoutPort(remaining, port)
			continue
		}

		if err := dm.waitForEngine(prober, port); err != nil {
			lastErr = err
			// Only a child this process just spawned may be killed.
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			removePIDFile(dm.pidPath)
			remaining = withoutPort(remaining, port)
			continue
		}

		dm.cmd = cmd
		dm.rpcPort = port
		// The PID file is diagnostics only; failing to write it must not
		// disable a working engine.
		_ = writePIDFile(dm.pidPath, cmd.Process.Pid)

		return port, nil
	}

	if lastErr == nil {
		lastErr = ErrNoPortAvailable
	}
	return 0, lastErr
}

func (dm *DaemonManager) waitForEngine(prober Prober, port int) error {
	deadline := time.Now().Add(startTimeout)
	for time.Now().Before(deadline) {
		if prober.Identify(port) == PortOurs {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("aria2c on port %d did not answer RPC within %s", port, startTimeout)
}

// buildArgs prepares a secure local headless instance. When a session path is
// configured, aria2 saves unfinished downloads on exit and reloads them on the
// next start, preserving their GIDs and paused state.
func (dm *DaemonManager) buildArgs(port int) []string {
	args := []string{
		"--enable-rpc",
		"--rpc-listen-all=false", // Secure: only listen to localhost
		"--rpc-listen-port=" + strconv.Itoa(port),
		"--rpc-max-request-size=2M",
		"--daemon=false", // Let Go manage the lifecycle directly
	}

	if dm.sessionPath != "" {
		args = append(args,
			"--save-session="+dm.sessionPath,
			"--save-session-interval="+strconv.Itoa(sessionSaveSeconds),
			"--input-file="+dm.sessionPath,
		)
	}

	return args
}

// ensureSessionFile creates an empty session file when it does not exist yet.
// aria2c refuses to start when --input-file points at a missing path.
func ensureSessionFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	// The session file contains download URLs, which may include credentials.
	return os.Chmod(path, 0600)
}

// Stop asks aria2c to terminate so it can flush its partially downloaded files,
// control files and session before exiting. Falls back to a hard kill if the
// daemon does not stop within the timeout.
func (dm *DaemonManager) Stop() error {
	if dm.cmd == nil || dm.cmd.Process == nil {
		removePIDFile(dm.pidPath)
		return nil // Nothing to stop
	}

	cmd := dm.cmd
	dm.cmd = nil

	// The RPC shutdown works on every platform, including Windows where
	// TerminateProcess would not give aria2 a chance to save its session.
	dm.requestShutdown(dm.rpcPort)

	// Unix gets SIGTERM as a fallback when RPC is unavailable. Windows is a
	// no-op here and relies on the RPC path or the timeout kill below.
	stopProcess(cmd.Process)

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
	case <-time.After(shutdownTimeout):
		_ = cmd.Process.Kill()
		<-done
	}

	removePIDFile(dm.pidPath)

	return nil
}

// requestShutdown sends a best-effort aria2.shutdown call over RPC.
func (dm *DaemonManager) requestShutdown(port int) {
	if port == 0 {
		return
	}

	payload := `{"jsonrpc":"2.0","id":"ariadm-shutdown","method":"aria2.shutdown","params":[]}`
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Post(
		fmt.Sprintf("http://127.0.0.1:%d/jsonrpc", port),
		"application/json",
		strings.NewReader(payload),
	)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}

// canBindPort reports whether a new process can still bind the port.
func canBindPort(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	_ = listener.Close()
	return true
}

func withoutPort(ports []int, port int) []int {
	filtered := make([]int, 0, len(ports))
	for _, candidate := range ports {
		if candidate != port {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

// writePIDFile records the child PID as diagnostic evidence. Ownership is still
// decided over RPC, so a stale file can never cause a foreign process to be
// touched.
func writePIDFile(path string, pid int) error {
	if path == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)+"\n"), 0600)
}

func removePIDFile(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}
