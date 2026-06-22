package daemon

import (
	"errors"
	"os/exec"
)

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

// Stop safely terminates the backend aria2c process
func (dm *DaemonManager) Stop() error {
	if dm.cmd == nil || dm.cmd.Process == nil {
		return nil // Nothing to stop
	}

	// Kill the process immediately
	if err := dm.cmd.Process.Kill(); err != nil {
		return err
	}

	// Wait for the process to release resources
	_, _ = dm.cmd.Process.Wait()
	dm.cmd = nil
	return nil
}
