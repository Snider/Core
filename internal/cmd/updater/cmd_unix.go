//go:build !windows

package updater

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

// spawnWatcher spawns a background process that watches for the current process
// to exit, then restarts the binary with --version to confirm the update.
func spawnWatcher() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	pid := os.Getpid()

	// Spawn: core update --watch-pid=<pid>
	cmd := exec.Command(executable, "update", "--watch-pid", strconv.Itoa(pid))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Detach from parent process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd.Start()
}

// watchAndRestart waits for the given PID to exit, then restarts the binary.
func watchAndRestart(pid int) error {
	// Wait for the parent process to die
	for isProcessRunning(pid) {

		time.Sleep(100 * time.Millisecond)
	}

	// Small delay to ensure file handle is released
	time.Sleep(200 * time.Millisecond)

	// Get executable path
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	// Use exec to replace this process
	return syscall.Exec(executable, []string{executable, "--version"}, os.Environ())
}

// isProcessRunning checks if a process with the given PID is still running.
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds, so we need to send signal 0
	// to check if the process actually exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
