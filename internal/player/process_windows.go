//go:build windows

package player

import (
	"os/exec"
	"syscall"
)

// setupPlayerProcess configures the process for detached execution
func setupPlayerProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | syscall.DETACHED_PROCESS,
	}
}

// releasePlayerProcess handles post-start process management
func releasePlayerProcess(cmd *exec.Cmd) error {
	// Windows doesn't need explicit process release
	return nil
}
