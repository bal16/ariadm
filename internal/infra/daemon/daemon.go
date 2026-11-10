package daemon

import (
	"errors"
	"os/exec"
	"time"
)

const shutdownTimeout = 5 * time.Second

type DaemonManager struct {
	cmd        *exec.Cmd
	binaryPath string
	rpcPort    string
}

func NewDaemonManager(binaryPath string, rpcPort string) *DaemonManager {
	return &DaemonManager{
		binaryPath: binaryPath,
		rpcPort:    rpcPort,
	}
}

// Start launches the headless aria2c daemon
func (dm *DaemonManager) Start() error {
	if dm.cmd != nil && dm.cmd.Process != nil {
		return errors.New("aria2c daemon is already running")
	}

	// Prepare arguments for a secure local headless instance
	args := []string{
		"--enable-rpc",
		"--rpc-listen-all=false", // Secure: only listen to localhost
		"--rpc-listen-port=" + dm.rpcPort,
		"--rpc-max-request-size=2M",
		"--daemon=false", // Let Go manage the lifecycle directly
	}

	dm.cmd = exec.Command(dm.binaryPath, args...)

	// Apply OS-specific attributes (defined in separate files)
	setPlatformAttributes(dm.cmd)

	if err := dm.cmd.Start(); err != nil {
		return err
	}

	return nil
}

// Stop asks aria2c to terminate so it can flush its partially downloaded files
// and control files before exiting. Falls back to a hard kill if the daemon
// does not stop within the timeout.
func (dm *DaemonManager) Stop() error {
	if dm.cmd == nil || dm.cmd.Process == nil {
		return nil // Nothing to stop
	}

	cmd := dm.cmd
	dm.cmd = nil

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
