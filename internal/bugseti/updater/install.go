// Package updater provides auto-update functionality for BugSETI.
package updater

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

// InstallResult contains the result of an installation.
type InstallResult struct {
	Success       bool   `json:"success"`
	OldPath       string `json:"oldPath"`
	NewPath       string `json:"newPath"`
	BackupPath    string `json:"backupPath"`
	RestartNeeded bool   `json:"restartNeeded"`
	Error         string `json:"error,omitempty"`
}

// Installer handles installing updates and restarting the application.
type Installer struct {
	executablePath string
}

// NewInstaller creates a new installer.
func NewInstaller() (*Installer, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks to get the real path
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable path: %w", err)
	}

	return &Installer{
		executablePath: execPath,
	}, nil
}

// Install replaces the current binary with the new one.
func (i *Installer) Install(newBinaryPath string) (*InstallResult, error) {
	result := &InstallResult{
		OldPath:       i.executablePath,
		NewPath:       newBinaryPath,
		RestartNeeded: true,
	}

	// Verify the new binary exists and is executable
	if _, err := os.Stat(newBinaryPath); err != nil {
		result.Error = fmt.Sprintf("new binary not found: %v", err)
		return result, fmt.Errorf("new binary not found: %w", err)
	}

	// Create backup of current binary
	backupPath := i.executablePath + ".bak"
	result.BackupPath = backupPath

	// Platform-specific installation
	var err error
	switch runtime.GOOS {
	case "windows":
		err = i.installWindows(newBinaryPath, backupPath)
	default:
		err = i.installUnix(newBinaryPath, backupPath)
	}

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	return result, nil
}

// installUnix performs the installation on Unix-like systems.
func (i *Installer) installUnix(newBinaryPath, backupPath string) error {
	// Remove old backup if exists
	os.Remove(backupPath)

	// Rename current binary to backup
	if err := os.Rename(i.executablePath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Copy new binary to target location
	// We use copy instead of rename in case they're on different filesystems
	if err := copyFile(newBinaryPath, i.executablePath); err != nil {
		// Try to restore backup
		os.Rename(backupPath, i.executablePath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(i.executablePath, 0755); err != nil {
		// Try to restore backup
		os.Remove(i.executablePath)
		os.Rename(backupPath, i.executablePath)
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	return nil
}

// installWindows performs the installation on Windows.
// On Windows, we can't replace a running executable, so we use a different approach:
// 1. Rename current executable to .old
// 2. Copy new executable to target location
// 3. On next start, clean up the .old file
func (i *Installer) installWindows(newBinaryPath, backupPath string) error {
	// Remove old backup if exists
	os.Remove(backupPath)

	// On Windows, we can rename the running executable
	if err := os.Rename(i.executablePath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Copy new binary to target location
	if err := copyFile(newBinaryPath, i.executablePath); err != nil {
		// Try to restore backup
		os.Rename(backupPath, i.executablePath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	return nil
}

// Restart restarts the application with the new binary.
func (i *Installer) Restart() error {
	args := os.Args
	env := os.Environ()

	switch runtime.GOOS {
	case "windows":
		return i.restartWindows(args, env)
	default:
		return i.restartUnix(args, env)
	}
}

// restartUnix restarts the application on Unix-like systems using exec.
func (i *Installer) restartUnix(args []string, env []string) error {
	// Use syscall.Exec to replace the current process
	// This is the cleanest way to restart on Unix
	return syscall.Exec(i.executablePath, args, env)
}

// restartWindows restarts the application on Windows.
func (i *Installer) restartWindows(args []string, env []string) error {
	// On Windows, we can't use exec to replace the process
	// Instead, we start a new process and exit the current one
	cmd := exec.Command(i.executablePath, args[1:]...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start new process: %w", err)
	}

	// Exit current process
	os.Exit(0)
	return nil // Never reached
}

// RestartLater schedules a restart for when the app next starts.
// This is useful when the user wants to continue working and restart later.
func (i *Installer) RestartLater() error {
	// Create a marker file that indicates a restart is pending
	markerPath := filepath.Join(filepath.Dir(i.executablePath), ".bugseti-restart-pending")
	return os.WriteFile(markerPath, []byte("restart"), 0644)
}

// CheckPendingRestart checks if a restart was scheduled.
func (i *Installer) CheckPendingRestart() bool {
	markerPath := filepath.Join(filepath.Dir(i.executablePath), ".bugseti-restart-pending")
	_, err := os.Stat(markerPath)
	return err == nil
}

// ClearPendingRestart clears the pending restart marker.
func (i *Installer) ClearPendingRestart() error {
	markerPath := filepath.Join(filepath.Dir(i.executablePath), ".bugseti-restart-pending")
	return os.Remove(markerPath)
}

// CleanupBackup removes the backup binary after a successful update.
func (i *Installer) CleanupBackup() error {
	backupPath := i.executablePath + ".bak"
	if _, err := os.Stat(backupPath); err == nil {
		return os.Remove(backupPath)
	}
	return nil
}

// Rollback restores the previous version from backup.
func (i *Installer) Rollback() error {
	backupPath := i.executablePath + ".bak"

	// Check if backup exists
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	// Remove current binary
	if err := os.Remove(i.executablePath); err != nil {
		return fmt.Errorf("failed to remove current binary: %w", err)
	}

	// Restore backup
	if err := os.Rename(backupPath, i.executablePath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// GetExecutablePath returns the path to the current executable.
func (i *Installer) GetExecutablePath() string {
	return i.executablePath
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Get source file info for permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

// CanSelfUpdate checks if the application has permission to update itself.
func CanSelfUpdate() bool {
	execPath, err := os.Executable()
	if err != nil {
		return false
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return false
	}

	// Check if we can write to the executable's directory
	dir := filepath.Dir(execPath)
	testFile := filepath.Join(dir, ".bugseti-update-test")

	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testFile)

	return true
}

// NeedsElevation returns true if the update requires elevated privileges.
func NeedsElevation() bool {
	return !CanSelfUpdate()
}
