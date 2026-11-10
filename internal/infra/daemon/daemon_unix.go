//go:build !windows

package daemon

import (
	"os"
	"os/exec"
	"syscall"
)

func setPlatformAttributes(cmd *exec.Cmd) {
	// No window-hiding mechanisms are required on Unix architectures
}

// stopProcess sends SIGTERM so aria2c can flush its download control files.
func stopProcess(process *os.Process) {
	_ = process.Signal(syscall.SIGTERM)
}
