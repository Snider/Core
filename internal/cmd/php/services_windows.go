//go:build windows

package php

import (
	"os"
	"os/exec"
)

// setSysProcAttr sets Windows-specific process attributes.
// Windows doesn't support Setpgid, so this is a no-op.
func setSysProcAttr(cmd *exec.Cmd) {
	// No-op on Windows - process groups work differently
}

// signalProcessGroup sends a termination signal to the process.
// On Windows, we can only signal the main process, not a group.
func signalProcessGroup(cmd *exec.Cmd, sig os.Signal) error {
	if cmd.Process == nil {
		return nil
	}

	return cmd.Process.Signal(sig)
}

// termSignal returns os.Interrupt for Windows (closest to SIGTERM).
func termSignal() os.Signal {
	return os.Interrupt
}

// killSignal returns os.Kill for Windows.
func killSignal() os.Signal {
	return os.Kill
}
