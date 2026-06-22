//go:build !windows
package daemon

import "os/exec"

func setPlatformAttributes(cmd *exec.Cmd) {
	// No window-hiding mechanisms are required on Unix architectures
}