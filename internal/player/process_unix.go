//go:build !windows

package player

import (
	"os/exec"
	"syscall"
)

// setupPlayerProcess configures the process for detached execution
func setupPlayerProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// releasePlayerProcess handles post-start process management
func releasePlayerProcess(cmd *exec.Cmd) error {
	if cmd.Process != nil {
		return cmd.Process.Release()
	}
	return nil
}
