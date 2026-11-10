//go:build windows

package daemon

import (
	"os"
	"os/exec"
	"syscall"
)

func setPlatformAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true, // Hides the cmd.exe window on Windows environments
	}
}

// stopProcess is a no-op on Windows: TerminateProcess gives aria2c no chance to
// flush its session. Stop asks aria2 to shut down over RPC first and falls back
// to Kill after the shutdown timeout.
func stopProcess(process *os.Process) {
	_ = process
}
