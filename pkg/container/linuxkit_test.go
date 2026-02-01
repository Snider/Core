package container

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockHypervisor is a mock implementation for testing.
type MockHypervisor struct {
	name         string
	available    bool
	buildErr     error
	lastImage    string
	lastOpts     *HypervisorOptions
	commandToRun string
}

func NewMockHypervisor() *MockHypervisor {
	return &MockHypervisor{
		name:         "mock",
		available:    true,
		commandToRun: "echo",
	}
}

func (m *MockHypervisor) Name() string {
	return m.name
}

func (m *MockHypervisor) Available() bool {
	return m.available
}

func (m *MockHypervisor) BuildCommand(ctx context.Context, image string, opts *HypervisorOptions) (*exec.Cmd, error) {
	m.lastImage = image
	m.lastOpts = opts
	if m.buildErr != nil {
		return nil, m.buildErr
	}
	// Return a simple command that exits quickly
	return exec.CommandContext(ctx, m.commandToRun, "test"), nil
}

// newTestManager creates a LinuxKitManager with mock hypervisor for testing.
// Uses manual temp directory management to avoid race conditions with t.TempDir cleanup.
func newTestManager(t *testing.T) (*LinuxKitManager, *MockHypervisor, string) {
	tmpDir, err := os.MkdirTemp("", "linuxkit-test-*")
	require.NoError(t, err)

	// Manual cleanup that handles race conditions with state file writes
	t.Cleanup(func() {
		// Give any pending file operations time to complete
		time.Sleep(10 * time.Millisecond)
		_ = os.RemoveAll(tmpDir)
	})

	statePath := filepath.Join(tmpDir, "containers.json")

	state, err := LoadState(statePath)
	require.NoError(t, err)

	mock := NewMockHypervisor()
	manager := NewLinuxKitManagerWithHypervisor(state, mock)

	return manager, mock, tmpDir
}

func TestNewLinuxKitManagerWithHypervisor_Good(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "containers.json")
	state, _ := LoadState(statePath)
	mock := NewMockHypervisor()

	manager := NewLinuxKitManagerWithHypervisor(state, mock)

	assert.NotNil(t, manager)
	assert.Equal(t, state, manager.State())
	assert.Equal(t, mock, manager.Hypervisor())
}

func TestLinuxKitManager_Run_Good_Detached(t *testing.T) {
	manager, mock, tmpDir := newTestManager(t)

	// Create a test image file
	imagePath := filepath.Join(tmpDir, "test.iso")
	err := os.WriteFile(imagePath, []byte("fake image"), 0644)
	require.NoError(t, err)

	// Use a command that runs briefly then exits
	mock.commandToRun = "sleep"

	ctx := context.Background()
	opts := RunOptions{
		Name:   "test-vm",
		Detach: true,
		Memory: 512,
		CPUs:   2,
	}

	container, err := manager.Run(ctx, imagePath, opts)
	require.NoError(t, err)

	assert.NotEmpty(t, container.ID)
	assert.Equal(t, "test-vm", container.Name)
	assert.Equal(t, imagePath, container.Image)
	assert.Equal(t, StatusRunning, container.Status)
	assert.Greater(t, container.PID, 0)
	assert.Equal(t, 512, container.Memory)
	assert.Equal(t, 2, container.CPUs)

	// Verify hypervisor was called with correct options
	assert.Equal(t, imagePath, mock.lastImage)
	assert.Equal(t, 512, mock.lastOpts.Memory)
	assert.Equal(t, 2, mock.lastOpts.CPUs)

	// Clean up - stop the container
	time.Sleep(100 * time.Millisecond)
}

func TestLinuxKitManager_Run_Good_DefaultValues(t *testing.T) {
	manager, mock, tmpDir := newTestManager(t)

	imagePath := filepath.Join(tmpDir, "test.qcow2")
	err := os.WriteFile(imagePath, []byte("fake image"), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	opts := RunOptions{Detach: true}

	container, err := manager.Run(ctx, imagePath, opts)
	require.NoError(t, err)

	// Check defaults were applied
	assert.Equal(t, 1024, mock.lastOpts.Memory)
	assert.Equal(t, 1, mock.lastOpts.CPUs)
	assert.Equal(t, 2222, mock.lastOpts.SSHPort)

	// Name should default to first 8 chars of ID
	assert.Equal(t, container.ID[:8], container.Name)

	// Wait for the mock process to complete to avoid temp dir cleanup issues
	time.Sleep(50 * time.Millisecond)
}

func TestLinuxKitManager_Run_Bad_ImageNotFound(t *testing.T) {
	manager, _, _ := newTestManager(t)

	ctx := context.Background()
	opts := RunOptions{Detach: true}

	_, err := manager.Run(ctx, "/nonexistent/image.iso", opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "image not found")
}

func TestLinuxKitManager_Run_Bad_UnsupportedFormat(t *testing.T) {
	manager, _, tmpDir := newTestManager(t)

	imagePath := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(imagePath, []byte("not an image"), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	opts := RunOptions{Detach: true}

	_, err = manager.Run(ctx, imagePath, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported image format")
}

func TestLinuxKitManager_Stop_Good(t *testing.T) {
	manager, _, _ := newTestManager(t)

	// Add a fake running container with a non-existent PID
	// The Stop function should handle this gracefully
	container := &Container{
		ID:        "abc12345",
		Status:    StatusRunning,
		PID:       999999, // Non-existent PID
		StartedAt: time.Now(),
	}
	manager.State().Add(container)

	ctx := context.Background()
	err := manager.Stop(ctx, "abc12345")

	// Stop should succeed (process doesn't exist, so container is marked stopped)
	assert.NoError(t, err)

	// Verify the container status was updated
	c, ok := manager.State().Get("abc12345")
	assert.True(t, ok)
	assert.Equal(t, StatusStopped, c.Status)
}

func TestLinuxKitManager_Stop_Bad_NotFound(t *testing.T) {
	manager, _, _ := newTestManager(t)

	ctx := context.Background()
	err := manager.Stop(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container not found")
}

func TestLinuxKitManager_Stop_Bad_NotRunning(t *testing.T) {
	manager, _, tmpDir := newTestManager(t)
	statePath := filepath.Join(tmpDir, "containers.json")
	state, _ := LoadState(statePath)
	manager = NewLinuxKitManagerWithHypervisor(state, NewMockHypervisor())

	container := &Container{
		ID:     "abc12345",
		Status: StatusStopped,
	}
	state.Add(container)

	ctx := context.Background()
	err := manager.Stop(ctx, "abc12345")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestLinuxKitManager_List_Good(t *testing.T) {
	manager, _, tmpDir := newTestManager(t)
	statePath := filepath.Join(tmpDir, "containers.json")
	state, _ := LoadState(statePath)
	manager = NewLinuxKitManagerWithHypervisor(state, NewMockHypervisor())

	state.Add(&Container{ID: "aaa11111", Status: StatusStopped})
	state.Add(&Container{ID: "bbb22222", Status: StatusStopped})

	ctx := context.Background()
	containers, err := manager.List(ctx)

	require.NoError(t, err)
	assert.Len(t, containers, 2)
}

func TestLinuxKitManager_List_Good_VerifiesRunningStatus(t *testing.T) {
	manager, _, tmpDir := newTestManager(t)
	statePath := filepath.Join(tmpDir, "containers.json")
	state, _ := LoadState(statePath)
	manager = NewLinuxKitManagerWithHypervisor(state, NewMockHypervisor())

	// Add a "running" container with a fake PID that doesn't exist
	state.Add(&Container{
		ID:     "abc12345",
		Status: StatusRunning,
		PID:    999999, // PID that almost certainly doesn't exist
	})

	ctx := context.Background()
	containers, err := manager.List(ctx)

	require.NoError(t, err)
	assert.Len(t, containers, 1)
	// Status should have been updated to stopped since PID doesn't exist
	assert.Equal(t, StatusStopped, containers[0].Status)
}

func TestLinuxKitManager_Logs_Good(t *testing.T) {
	manager, _, tmpDir := newTestManager(t)

	// Create a log file manually
	logsDir := filepath.Join(tmpDir, "logs")
	os.MkdirAll(logsDir, 0755)

	container := &Container{ID: "abc12345"}
	manager.State().Add(container)

	// Override the default logs dir for testing by creating the log file
	// at the expected location
	logContent := "test log content\nline 2\n"
	logPath, _ := LogPath("abc12345")
	os.MkdirAll(filepath.Dir(logPath), 0755)
	os.WriteFile(logPath, []byte(logContent), 0644)

	ctx := context.Background()
	reader, err := manager.Logs(ctx, "abc12345", false)

	require.NoError(t, err)
	defer reader.Close()

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	assert.Equal(t, logContent, string(buf[:n]))
}

func TestLinuxKitManager_Logs_Bad_NotFound(t *testing.T) {
	manager, _, _ := newTestManager(t)

	ctx := context.Background()
	_, err := manager.Logs(ctx, "nonexistent", false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container not found")
}

func TestLinuxKitManager_Logs_Bad_NoLogFile(t *testing.T) {
	manager, _, _ := newTestManager(t)

	// Use a unique ID that won't have a log file
	uniqueID, _ := GenerateID()
	container := &Container{ID: uniqueID}
	manager.State().Add(container)

	ctx := context.Background()
	reader, err := manager.Logs(ctx, uniqueID, false)

	// If logs existed somehow, clean up the reader
	if reader != nil {
		reader.Close()
	}

	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "no logs available")
	}
}

func TestLinuxKitManager_Exec_Bad_NotFound(t *testing.T) {
	manager, _, _ := newTestManager(t)

	ctx := context.Background()
	err := manager.Exec(ctx, "nonexistent", []string{"ls"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container not found")
}

func TestLinuxKitManager_Exec_Bad_NotRunning(t *testing.T) {
	manager, _, _ := newTestManager(t)

	container := &Container{ID: "abc12345", Status: StatusStopped}
	manager.State().Add(container)

	ctx := context.Background()
	err := manager.Exec(ctx, "abc12345", []string{"ls"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestDetectImageFormat_Good(t *testing.T) {
	tests := []struct {
		path   string
		format ImageFormat
	}{
		{"/path/to/image.iso", FormatISO},
		{"/path/to/image.ISO", FormatISO},
		{"/path/to/image.qcow2", FormatQCOW2},
		{"/path/to/image.QCOW2", FormatQCOW2},
		{"/path/to/image.vmdk", FormatVMDK},
		{"/path/to/image.raw", FormatRaw},
		{"/path/to/image.img", FormatRaw},
		{"image.iso", FormatISO},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.format, DetectImageFormat(tt.path))
		})
	}
}

func TestDetectImageFormat_Bad_Unknown(t *testing.T) {
	tests := []string{
		"/path/to/image.txt",
		"/path/to/image",
		"noextension",
		"/path/to/image.docx",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			assert.Equal(t, FormatUnknown, DetectImageFormat(path))
		})
	}
}

func TestQemuHypervisor_Name_Good(t *testing.T) {
	q := NewQemuHypervisor()
	assert.Equal(t, "qemu", q.Name())
}

func TestQemuHypervisor_BuildCommand_Good(t *testing.T) {
	q := NewQemuHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{
		Memory:  2048,
		CPUs:    4,
		SSHPort: 2222,
		Ports:   map[int]int{8080: 80},
		Detach:  true,
	}

	cmd, err := q.BuildCommand(ctx, "/path/to/image.iso", opts)
	require.NoError(t, err)
	assert.NotNil(t, cmd)

	// Check command path
	assert.Contains(t, cmd.Path, "qemu")

	// Check that args contain expected values
	args := cmd.Args
	assert.Contains(t, args, "-m")
	assert.Contains(t, args, "2048")
	assert.Contains(t, args, "-smp")
	assert.Contains(t, args, "4")
	assert.Contains(t, args, "-nographic")
}


func TestLinuxKitManager_Logs_Good_Follow(t *testing.T) {
	manager, _, _ := newTestManager(t)

	// Create a unique container ID
	uniqueID, _ := GenerateID()
	container := &Container{ID: uniqueID}
	manager.State().Add(container)

	// Create a log file at the expected location
	logPath, err := LogPath(uniqueID)
	require.NoError(t, err)
	os.MkdirAll(filepath.Dir(logPath), 0755)

	// Write initial content
	err = os.WriteFile(logPath, []byte("initial log content\n"), 0644)
	require.NoError(t, err)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Get the follow reader
	reader, err := manager.Logs(ctx, uniqueID, true)
	require.NoError(t, err)

	// Cancel the context to stop the follow
	cancel()

	// Read should return EOF after context cancellation
	buf := make([]byte, 1024)
	_, readErr := reader.Read(buf)
	// After context cancel, Read should return EOF
	assert.Equal(t, "EOF", readErr.Error())

	// Close the reader
	err = reader.Close()
	assert.NoError(t, err)
}

func TestFollowReader_Read_Good_WithData(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create log file with content
	content := "test log line 1\ntest log line 2\n"
	err := os.WriteFile(logPath, []byte(content), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	reader, err := newFollowReader(ctx, logPath)
	require.NoError(t, err)
	defer reader.Close()

	// The followReader seeks to end, so we need to append more content
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644)
	require.NoError(t, err)
	_, err = f.WriteString("new line\n")
	require.NoError(t, err)
	f.Close()

	// Give the reader time to poll
	time.Sleep(150 * time.Millisecond)

	buf := make([]byte, 1024)
	n, err := reader.Read(buf)
	if err == nil {
		assert.Greater(t, n, 0)
	}
}

func TestFollowReader_Read_Good_ContextCancel(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create log file
	err := os.WriteFile(logPath, []byte("initial content\n"), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	reader, err := newFollowReader(ctx, logPath)
	require.NoError(t, err)

	// Cancel the context
	cancel()

	// Read should return EOF
	buf := make([]byte, 1024)
	_, readErr := reader.Read(buf)
	assert.Equal(t, "EOF", readErr.Error())

	reader.Close()
}

func TestFollowReader_Close_Good(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	err := os.WriteFile(logPath, []byte("content\n"), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	reader, err := newFollowReader(ctx, logPath)
	require.NoError(t, err)

	err = reader.Close()
	assert.NoError(t, err)

	// Reading after close should fail or return EOF
	buf := make([]byte, 1024)
	_, readErr := reader.Read(buf)
	assert.Error(t, readErr)
}

func TestNewFollowReader_Bad_FileNotFound(t *testing.T) {
	ctx := context.Background()
	_, err := newFollowReader(ctx, "/nonexistent/path/to/file.log")

	assert.Error(t, err)
}

func TestLinuxKitManager_Run_Bad_BuildCommandError(t *testing.T) {
	manager, mock, tmpDir := newTestManager(t)

	// Create a test image file
	imagePath := filepath.Join(tmpDir, "test.iso")
	err := os.WriteFile(imagePath, []byte("fake image"), 0644)
	require.NoError(t, err)

	// Configure mock to return an error
	mock.buildErr = assert.AnError

	ctx := context.Background()
	opts := RunOptions{Detach: true}

	_, err = manager.Run(ctx, imagePath, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build hypervisor command")
}

func TestLinuxKitManager_Run_Good_Foreground(t *testing.T) {
	manager, mock, tmpDir := newTestManager(t)

	// Create a test image file
	imagePath := filepath.Join(tmpDir, "test.iso")
	err := os.WriteFile(imagePath, []byte("fake image"), 0644)
	require.NoError(t, err)

	// Use echo which exits quickly
	mock.commandToRun = "echo"

	ctx := context.Background()
	opts := RunOptions{
		Name:   "test-foreground",
		Detach: false, // Run in foreground
		Memory: 512,
		CPUs:   1,
	}

	container, err := manager.Run(ctx, imagePath, opts)
	require.NoError(t, err)

	assert.NotEmpty(t, container.ID)
	assert.Equal(t, "test-foreground", container.Name)
	// Foreground process should have completed
	assert.Equal(t, StatusStopped, container.Status)
}

func TestLinuxKitManager_Stop_Good_ContextCancelled(t *testing.T) {
	manager, mock, tmpDir := newTestManager(t)

	// Create a test image file
	imagePath := filepath.Join(tmpDir, "test.iso")
	err := os.WriteFile(imagePath, []byte("fake image"), 0644)
	require.NoError(t, err)

	// Use a command that takes a long time
	mock.commandToRun = "sleep"

	// Start a container
	ctx := context.Background()
	opts := RunOptions{
		Name:   "test-cancel",
		Detach: true,
	}

	container, err := manager.Run(ctx, imagePath, opts)
	require.NoError(t, err)

	// Ensure cleanup happens regardless of test outcome
	t.Cleanup(func() {
		_ = manager.Stop(context.Background(), container.ID)
	})

	// Create a context that's already cancelled
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	// Stop with cancelled context
	err = manager.Stop(cancelCtx, container.ID)
	// Should return context error
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestIsProcessRunning_Good_ExistingProcess(t *testing.T) {
	// Use our own PID which definitely exists
	running := isProcessRunning(os.Getpid())
	assert.True(t, running)
}

func TestIsProcessRunning_Bad_NonexistentProcess(t *testing.T) {
	// Use a PID that almost certainly doesn't exist
	running := isProcessRunning(999999)
	assert.False(t, running)
}

func TestLinuxKitManager_Run_Good_WithPortsAndVolumes(t *testing.T) {
	manager, mock, tmpDir := newTestManager(t)

	imagePath := filepath.Join(tmpDir, "test.iso")
	err := os.WriteFile(imagePath, []byte("fake image"), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	opts := RunOptions{
		Name:    "test-ports",
		Detach:  true,
		Memory:  512,
		CPUs:    1,
		SSHPort: 2223,
		Ports:   map[int]int{8080: 80, 443: 443},
		Volumes: map[string]string{"/host/data": "/container/data"},
	}

	container, err := manager.Run(ctx, imagePath, opts)
	require.NoError(t, err)

	assert.NotEmpty(t, container.ID)
	assert.Equal(t, map[int]int{8080: 80, 443: 443}, container.Ports)
	assert.Equal(t, 2223, mock.lastOpts.SSHPort)
	assert.Equal(t, map[string]string{"/host/data": "/container/data"}, mock.lastOpts.Volumes)

	time.Sleep(50 * time.Millisecond)
}

func TestFollowReader_Read_Good_ReaderError(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create log file
	err := os.WriteFile(logPath, []byte("content\n"), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	reader, err := newFollowReader(ctx, logPath)
	require.NoError(t, err)

	// Close the underlying file to cause read errors
	reader.file.Close()

	// Read should return an error
	buf := make([]byte, 1024)
	_, readErr := reader.Read(buf)
	assert.Error(t, readErr)
}

func TestLinuxKitManager_Run_Bad_StartError(t *testing.T) {
	manager, mock, tmpDir := newTestManager(t)

	imagePath := filepath.Join(tmpDir, "test.iso")
	err := os.WriteFile(imagePath, []byte("fake image"), 0644)
	require.NoError(t, err)

	// Use a command that doesn't exist to cause Start() to fail
	mock.commandToRun = "/nonexistent/command/that/does/not/exist"

	ctx := context.Background()
	opts := RunOptions{
		Name:   "test-start-error",
		Detach: true,
	}

	_, err = manager.Run(ctx, imagePath, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start VM")
}

func TestLinuxKitManager_Run_Bad_ForegroundStartError(t *testing.T) {
	manager, mock, tmpDir := newTestManager(t)

	imagePath := filepath.Join(tmpDir, "test.iso")
	err := os.WriteFile(imagePath, []byte("fake image"), 0644)
	require.NoError(t, err)

	// Use a command that doesn't exist to cause Start() to fail
	mock.commandToRun = "/nonexistent/command/that/does/not/exist"

	ctx := context.Background()
	opts := RunOptions{
		Name:   "test-foreground-error",
		Detach: false,
	}

	_, err = manager.Run(ctx, imagePath, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start VM")
}

func TestLinuxKitManager_Run_Good_ForegroundWithError(t *testing.T) {
	manager, mock, tmpDir := newTestManager(t)

	imagePath := filepath.Join(tmpDir, "test.iso")
	err := os.WriteFile(imagePath, []byte("fake image"), 0644)
	require.NoError(t, err)

	// Use a command that exits with error
	mock.commandToRun = "false" // false command exits with code 1

	ctx := context.Background()
	opts := RunOptions{
		Name:   "test-foreground-exit-error",
		Detach: false,
	}

	container, err := manager.Run(ctx, imagePath, opts)
	require.NoError(t, err) // Run itself should succeed

	// Container should be in error state since process exited with error
	assert.Equal(t, StatusError, container.Status)
}

func TestLinuxKitManager_Stop_Good_ProcessExitedWhileRunning(t *testing.T) {
	manager, _, _ := newTestManager(t)

	// Add a "running" container with a process that has already exited
	// This simulates the race condition where process exits between status check
	// and signal send
	container := &Container{
		ID:        "test1234",
		Status:    StatusRunning,
		PID:       999999, // Non-existent PID
		StartedAt: time.Now(),
	}
	manager.State().Add(container)

	ctx := context.Background()
	err := manager.Stop(ctx, "test1234")

	// Stop should succeed gracefully
	assert.NoError(t, err)

	// Container should be stopped
	c, ok := manager.State().Get("test1234")
	assert.True(t, ok)
	assert.Equal(t, StatusStopped, c.Status)
}
