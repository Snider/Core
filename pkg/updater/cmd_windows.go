//go:build windows

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

	// On Windows, use CREATE_NEW_PROCESS_GROUP to detach
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	return cmd.Start()
}

// watchAndRestart waits for the given PID to exit, then restarts the binary.
func watchAndRestart(pid int) error {
	// Wait for the parent process to die
	for {
		if !isProcessRunning(pid) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Small delay to ensure file handle is released
	time.Sleep(500 * time.Millisecond)

	// Get executable path
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	// On Windows, spawn new process and exit
	cmd := exec.Command(executable, "--version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

// isProcessRunning checks if a process with the given PID is still running.
func isProcessRunning(pid int) bool {
	// On Windows, try to open the process with query rights
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	syscall.CloseHandle(handle)
	return true
}
