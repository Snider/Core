package devops

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/host-uk/core/pkg/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageName(t *testing.T) {
	name := ImageName()
	assert.Contains(t, name, "core-devops-")
	assert.Contains(t, name, runtime.GOOS)
	assert.Contains(t, name, runtime.GOARCH)
	assert.True(t, (name[len(name)-6:] == ".qcow2"))
}

func TestImagesDir(t *testing.T) {
	t.Run("default directory", func(t *testing.T) {
		// Unset env if it exists
		orig := os.Getenv("CORE_IMAGES_DIR")
		os.Unsetenv("CORE_IMAGES_DIR")
		defer os.Setenv("CORE_IMAGES_DIR", orig)

		dir, err := ImagesDir()
		assert.NoError(t, err)
		assert.Contains(t, dir, ".core/images")
	})

	t.Run("environment override", func(t *testing.T) {
		customDir := "/tmp/custom-images"
		t.Setenv("CORE_IMAGES_DIR", customDir)

		dir, err := ImagesDir()
		assert.NoError(t, err)
		assert.Equal(t, customDir, dir)
	})
}

func TestImagePath(t *testing.T) {
	customDir := "/tmp/images"
	t.Setenv("CORE_IMAGES_DIR", customDir)

	path, err := ImagePath()
	assert.NoError(t, err)
	expected := filepath.Join(customDir, ImageName())
	assert.Equal(t, expected, path)
}

func TestDefaultBootOptions(t *testing.T) {
	opts := DefaultBootOptions()
	assert.Equal(t, 4096, opts.Memory)
	assert.Equal(t, 2, opts.CPUs)
	assert.Equal(t, "core-dev", opts.Name)
	assert.False(t, opts.Fresh)
}

func TestIsInstalled_Bad(t *testing.T) {
	t.Run("returns false for non-existent image", func(t *testing.T) {
		// Point to a temp directory that is empty
		tempDir := t.TempDir()
		t.Setenv("CORE_IMAGES_DIR", tempDir)

		// Create devops instance manually to avoid loading real config/images
		d := &DevOps{}
		assert.False(t, d.IsInstalled())
	})
}

func TestIsInstalled_Good(t *testing.T) {
	t.Run("returns true when image exists", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv("CORE_IMAGES_DIR", tempDir)

		// Create the image file
		imagePath := filepath.Join(tempDir, ImageName())
		err := os.WriteFile(imagePath, []byte("fake image data"), 0644)
		require.NoError(t, err)

		d := &DevOps{}
		assert.True(t, d.IsInstalled())
	})
}

type mockHypervisor struct{}

func (m *mockHypervisor) Name() string    { return "mock" }
func (m *mockHypervisor) Available() bool { return true }
func (m *mockHypervisor) BuildCommand(ctx context.Context, image string, opts *container.HypervisorOptions) (*exec.Cmd, error) {
	return exec.Command("true"), nil
}

func TestDevOps_Status_Good(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	// Setup mock container manager
	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	// Add a fake running container
	c := &container.Container{
		ID:        "test-id",
		Name:      "core-dev",
		Status:    container.StatusRunning,
		PID:       os.Getpid(), // Use our own PID so isProcessRunning returns true
		StartedAt: time.Now().Add(-time.Hour),
		Memory:    2048,
		CPUs:      4,
	}
	err = state.Add(c)
	require.NoError(t, err)

	status, err := d.Status(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.Running)
	assert.Equal(t, "test-id", status.ContainerID)
	assert.Equal(t, 2048, status.Memory)
	assert.Equal(t, 4, status.CPUs)
}

func TestDevOps_Status_Good_NotInstalled(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	status, err := d.Status(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.False(t, status.Installed)
	assert.False(t, status.Running)
	assert.Equal(t, 2222, status.SSHPort)
}

func TestDevOps_Status_Good_NoContainer(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	// Create fake image to mark as installed
	imagePath := filepath.Join(tempDir, ImageName())
	err := os.WriteFile(imagePath, []byte("fake"), 0644)
	require.NoError(t, err)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	status, err := d.Status(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.Installed)
	assert.False(t, status.Running)
	assert.Empty(t, status.ContainerID)
}

func TestDevOps_IsRunning_Good(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	c := &container.Container{
		ID:        "test-id",
		Name:      "core-dev",
		Status:    container.StatusRunning,
		PID:       os.Getpid(),
		StartedAt: time.Now(),
	}
	err = state.Add(c)
	require.NoError(t, err)

	running, err := d.IsRunning(context.Background())
	assert.NoError(t, err)
	assert.True(t, running)
}

func TestDevOps_IsRunning_Bad_NotRunning(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	running, err := d.IsRunning(context.Background())
	assert.NoError(t, err)
	assert.False(t, running)
}

func TestDevOps_IsRunning_Bad_ContainerStopped(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	c := &container.Container{
		ID:        "test-id",
		Name:      "core-dev",
		Status:    container.StatusStopped,
		PID:       12345,
		StartedAt: time.Now(),
	}
	err = state.Add(c)
	require.NoError(t, err)

	running, err := d.IsRunning(context.Background())
	assert.NoError(t, err)
	assert.False(t, running)
}

func TestDevOps_findContainer_Good(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	c := &container.Container{
		ID:        "test-id",
		Name:      "my-container",
		Status:    container.StatusRunning,
		PID:       os.Getpid(),
		StartedAt: time.Now(),
	}
	err = state.Add(c)
	require.NoError(t, err)

	found, err := d.findContainer(context.Background(), "my-container")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "test-id", found.ID)
	assert.Equal(t, "my-container", found.Name)
}

func TestDevOps_findContainer_Bad_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	found, err := d.findContainer(context.Background(), "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestDevOps_Stop_Bad_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	err = d.Stop(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBootOptions_Custom(t *testing.T) {
	opts := BootOptions{
		Memory: 8192,
		CPUs:   4,
		Name:   "custom-dev",
		Fresh:  true,
	}
	assert.Equal(t, 8192, opts.Memory)
	assert.Equal(t, 4, opts.CPUs)
	assert.Equal(t, "custom-dev", opts.Name)
	assert.True(t, opts.Fresh)
}

func TestDevStatus_Struct(t *testing.T) {
	status := DevStatus{
		Installed:    true,
		Running:      true,
		ImageVersion: "v1.2.3",
		ContainerID:  "abc123",
		Memory:       4096,
		CPUs:         2,
		SSHPort:      2222,
		Uptime:       time.Hour,
	}
	assert.True(t, status.Installed)
	assert.True(t, status.Running)
	assert.Equal(t, "v1.2.3", status.ImageVersion)
	assert.Equal(t, "abc123", status.ContainerID)
	assert.Equal(t, 4096, status.Memory)
	assert.Equal(t, 2, status.CPUs)
	assert.Equal(t, 2222, status.SSHPort)
	assert.Equal(t, time.Hour, status.Uptime)
}

func TestDevOps_Boot_Bad_NotInstalled(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	err = d.Boot(context.Background(), DefaultBootOptions())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}

func TestDevOps_Boot_Bad_AlreadyRunning(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	// Create fake image
	imagePath := filepath.Join(tempDir, ImageName())
	err := os.WriteFile(imagePath, []byte("fake"), 0644)
	require.NoError(t, err)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	// Add a running container
	c := &container.Container{
		ID:        "test-id",
		Name:      "core-dev",
		Status:    container.StatusRunning,
		PID:       os.Getpid(),
		StartedAt: time.Now(),
	}
	err = state.Add(c)
	require.NoError(t, err)

	err = d.Boot(context.Background(), DefaultBootOptions())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestDevOps_Status_Good_WithImageVersion(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	// Create fake image
	imagePath := filepath.Join(tempDir, ImageName())
	err := os.WriteFile(imagePath, []byte("fake"), 0644)
	require.NoError(t, err)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	// Manually set manifest with version info
	mgr.manifest.Images[ImageName()] = ImageInfo{
		Version: "v1.2.3",
		Source:  "test",
	}

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		config:    cfg,
		images:    mgr,
		container: cm,
	}

	status, err := d.Status(context.Background())
	assert.NoError(t, err)
	assert.True(t, status.Installed)
	assert.Equal(t, "v1.2.3", status.ImageVersion)
}

func TestDevOps_findContainer_Good_MultipleContainers(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	// Add multiple containers
	c1 := &container.Container{
		ID:        "id-1",
		Name:      "container-1",
		Status:    container.StatusRunning,
		PID:       os.Getpid(),
		StartedAt: time.Now(),
	}
	c2 := &container.Container{
		ID:        "id-2",
		Name:      "container-2",
		Status:    container.StatusRunning,
		PID:       os.Getpid(),
		StartedAt: time.Now(),
	}
	err = state.Add(c1)
	require.NoError(t, err)
	err = state.Add(c2)
	require.NoError(t, err)

	// Find specific container
	found, err := d.findContainer(context.Background(), "container-2")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "id-2", found.ID)
}

func TestDevOps_Status_Good_ContainerWithUptime(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	startTime := time.Now().Add(-2 * time.Hour)
	c := &container.Container{
		ID:        "test-id",
		Name:      "core-dev",
		Status:    container.StatusRunning,
		PID:       os.Getpid(),
		StartedAt: startTime,
		Memory:    4096,
		CPUs:      2,
	}
	err = state.Add(c)
	require.NoError(t, err)

	status, err := d.Status(context.Background())
	assert.NoError(t, err)
	assert.True(t, status.Running)
	assert.GreaterOrEqual(t, status.Uptime.Hours(), float64(1))
}

func TestDevOps_IsRunning_Bad_DifferentContainerName(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	// Add a container with different name
	c := &container.Container{
		ID:        "test-id",
		Name:      "other-container",
		Status:    container.StatusRunning,
		PID:       os.Getpid(),
		StartedAt: time.Now(),
	}
	err = state.Add(c)
	require.NoError(t, err)

	// IsRunning looks for "core-dev", not "other-container"
	running, err := d.IsRunning(context.Background())
	assert.NoError(t, err)
	assert.False(t, running)
}

func TestDevOps_Boot_Good_FreshFlag(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "devops-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tempDir) })
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	// Create fake image
	imagePath := filepath.Join(tempDir, ImageName())
	err = os.WriteFile(imagePath, []byte("fake"), 0644)
	require.NoError(t, err)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	// Add an existing container with non-existent PID (will be seen as stopped)
	c := &container.Container{
		ID:        "old-id",
		Name:      "core-dev",
		Status:    container.StatusRunning,
		PID:       99999999, // Non-existent PID - List() will mark it as stopped
		StartedAt: time.Now(),
	}
	err = state.Add(c)
	require.NoError(t, err)

	// Boot with Fresh=true should try to stop the existing container
	// then run a new one. The mock hypervisor "succeeds" so this won't error
	opts := BootOptions{
		Memory: 4096,
		CPUs:   2,
		Name:   "core-dev",
		Fresh:  true,
	}
	err = d.Boot(context.Background(), opts)
	// The mock hypervisor's Run succeeds
	assert.NoError(t, err)
}

func TestDevOps_Stop_Bad_ContainerNotRunning(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	// Add a container that's already stopped
	c := &container.Container{
		ID:        "test-id",
		Name:      "core-dev",
		Status:    container.StatusStopped,
		PID:       99999999,
		StartedAt: time.Now(),
	}
	err = state.Add(c)
	require.NoError(t, err)

	// Stop should fail because container is not running
	err = d.Stop(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestDevOps_Boot_Good_FreshWithNoExisting(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	// Create fake image
	imagePath := filepath.Join(tempDir, ImageName())
	err := os.WriteFile(imagePath, []byte("fake"), 0644)
	require.NoError(t, err)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	// Boot with Fresh=true but no existing container
	opts := BootOptions{
		Memory: 4096,
		CPUs:   2,
		Name:   "core-dev",
		Fresh:  true,
	}
	err = d.Boot(context.Background(), opts)
	// The mock hypervisor succeeds
	assert.NoError(t, err)
}

func TestImageName_Format(t *testing.T) {
	name := ImageName()
	// Check format: core-devops-{os}-{arch}.qcow2
	assert.Contains(t, name, "core-devops-")
	assert.Contains(t, name, runtime.GOOS)
	assert.Contains(t, name, runtime.GOARCH)
	assert.True(t, filepath.Ext(name) == ".qcow2")
}

func TestDevOps_Install_Delegates(t *testing.T) {
	// This test verifies the Install method delegates to ImageManager
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	d := &DevOps{
		images: mgr,
	}

	// This will fail because no source is available, but it tests delegation
	err = d.Install(context.Background(), nil)
	assert.Error(t, err)
}

func TestDevOps_CheckUpdate_Delegates(t *testing.T) {
	// This test verifies the CheckUpdate method delegates to ImageManager
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	d := &DevOps{
		images: mgr,
	}

	// This will fail because image not installed, but it tests delegation
	_, _, _, err = d.CheckUpdate(context.Background())
	assert.Error(t, err)
}

func TestDevOps_Boot_Good_Success(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	// Create fake image
	imagePath := filepath.Join(tempDir, ImageName())
	err := os.WriteFile(imagePath, []byte("fake"), 0644)
	require.NoError(t, err)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	statePath := filepath.Join(tempDir, "containers.json")
	state := container.NewState(statePath)
	h := &mockHypervisor{}
	cm := container.NewLinuxKitManagerWithHypervisor(state, h)

	d := &DevOps{
		images:    mgr,
		container: cm,
	}

	// Boot without Fresh flag and no existing container
	opts := DefaultBootOptions()
	err = d.Boot(context.Background(), opts)
	assert.NoError(t, err) // Mock hypervisor succeeds
}

func TestDevOps_Config(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CORE_IMAGES_DIR", tempDir)

	cfg := DefaultConfig()
	mgr, err := NewImageManager(cfg)
	require.NoError(t, err)

	d := &DevOps{
		config: cfg,
		images: mgr,
	}

	assert.NotNil(t, d.config)
	assert.Equal(t, "auto", d.config.Images.Source)
}
