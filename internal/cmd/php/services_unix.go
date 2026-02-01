//go:build unix

package php

import (
	"os/exec"
	"syscall"
)

// setSysProcAttr sets Unix-specific process attributes for clean process group handling.
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// signalProcessGroup sends a signal to the process group.
// On Unix, this uses negative PID to signal the entire group.
func signalProcessGroup(cmd *exec.Cmd, sig syscall.Signal) error {
	if cmd.Process == nil {
		return nil
	}

	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		return syscall.Kill(-pgid, sig)
	}

	// Fallback to signaling just the process
	return cmd.Process.Signal(sig)
}

// termSignal returns SIGTERM for Unix.
func termSignal() syscall.Signal {
	return syscall.SIGTERM
}

// killSignal returns SIGKILL for Unix.
func killSignal() syscall.Signal {
	return syscall.SIGKILL
}
