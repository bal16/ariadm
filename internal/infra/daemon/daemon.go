package daemon

import (
	"errors"
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
)

type DaemonManager struct {
	cmd         *exec.Cmd
	binaryPath  string
	rpcPort     string
	sessionPath string
}

func NewDaemonManager(binaryPath string, rpcPort string, sessionPath string) *DaemonManager {
	return &DaemonManager{
		binaryPath:  binaryPath,
		rpcPort:     rpcPort,
		sessionPath: sessionPath,
	}
}

// Start launches the headless aria2c daemon
func (dm *DaemonManager) Start() error {
	if dm.cmd != nil && dm.cmd.Process != nil {
		return errors.New("aria2c daemon is already running")
	}

	if dm.sessionPath != "" {
		if err := ensureSessionFile(dm.sessionPath); err != nil {
			return err
		}
	}

	dm.cmd = exec.Command(dm.binaryPath, dm.buildArgs()...)

	// Apply OS-specific attributes (defined in separate files)
	setPlatformAttributes(dm.cmd)

	if err := dm.cmd.Start(); err != nil {
		return err
	}

	return nil
}

// buildArgs prepares a secure local headless instance. When a session path is
// configured, aria2 saves unfinished downloads on exit and reloads them on the
// next start, preserving their GIDs and paused state.
func (dm *DaemonManager) buildArgs() []string {
	args := []string{
		"--enable-rpc",
		"--rpc-listen-all=false", // Secure: only listen to localhost
		"--rpc-listen-port=" + dm.rpcPort,
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
		return nil // Nothing to stop
	}

	cmd := dm.cmd
	dm.cmd = nil

	// The RPC shutdown works on every platform, including Windows where
	// TerminateProcess would not give aria2 a chance to save its session.
	dm.requestShutdown()

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

	return nil
}

// requestShutdown sends a best-effort aria2.shutdown call over RPC.
func (dm *DaemonManager) requestShutdown() {
	if dm.rpcPort == "" {
		return
	}

	payload := `{"jsonrpc":"2.0","id":"ariadm-shutdown","method":"aria2.shutdown","params":[]}`
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Post(
		"http://127.0.0.1:"+dm.rpcPort+"/jsonrpc",
		"application/json",
		strings.NewReader(payload),
	)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}
