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
func newTestManager(t *testing.T) (*LinuxKitManager, *MockHypervisor, string) {
	tmpDir := t.TempDir()
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

func TestQemuHypervisor_BuildCommand_Bad_UnknownFormat(t *testing.T) {
	q := NewQemuHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{Memory: 1024, CPUs: 1}

	_, err := q.BuildCommand(ctx, "/path/to/image.txt", opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown image format")
}

func TestHyperkitHypervisor_Name_Good(t *testing.T) {
	h := NewHyperkitHypervisor()
	assert.Equal(t, "hyperkit", h.Name())
}

func TestHyperkitHypervisor_BuildCommand_Good(t *testing.T) {
	h := NewHyperkitHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{
		Memory:  1024,
		CPUs:    2,
		SSHPort: 2222,
	}

	cmd, err := h.BuildCommand(ctx, "/path/to/image.iso", opts)
	require.NoError(t, err)
	assert.NotNil(t, cmd)

	args := cmd.Args
	assert.Contains(t, args, "-m")
	assert.Contains(t, args, "1024M")
	assert.Contains(t, args, "-c")
	assert.Contains(t, args, "2")
}

func TestHyperkitHypervisor_BuildCommand_Bad_UnknownFormat(t *testing.T) {
	h := NewHyperkitHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{Memory: 1024, CPUs: 1}

	_, err := h.BuildCommand(ctx, "/path/to/image.unknown", opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown image format")
}

func TestGetHypervisor_Bad_Unknown(t *testing.T) {
	_, err := GetHypervisor("unknown-hypervisor")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown hypervisor")
}
