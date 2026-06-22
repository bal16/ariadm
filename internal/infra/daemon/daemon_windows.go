//go:build windows

package daemon

import (
	"os/exec"
	"syscall"
)

func setPlatformAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true, // Hides the cmd.exe window on Windows environments
	}
}
