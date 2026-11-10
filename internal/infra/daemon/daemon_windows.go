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

// stopProcess terminates the process. Windows does not support SIGTERM, so the
// daemon is killed directly, matching the previous behaviour.
func stopProcess(process *os.Process) {
	_ = process.Kill()
}
