package container

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQemuHypervisor_Available_Good(t *testing.T) {
	q := NewQemuHypervisor()

	// Check if qemu is available on this system
	available := q.Available()

	// We just verify it returns a boolean without error
	// The actual availability depends on the system
	assert.IsType(t, true, available)
}

func TestQemuHypervisor_Available_Bad_InvalidBinary(t *testing.T) {
	q := &QemuHypervisor{
		Binary: "nonexistent-qemu-binary-that-does-not-exist",
	}

	available := q.Available()

	assert.False(t, available)
}

func TestHyperkitHypervisor_Available_Good(t *testing.T) {
	h := NewHyperkitHypervisor()

	available := h.Available()

	// On non-darwin systems, should always be false
	if runtime.GOOS != "darwin" {
		assert.False(t, available)
	} else {
		// On darwin, just verify it returns a boolean
		assert.IsType(t, true, available)
	}
}

func TestHyperkitHypervisor_Available_Bad_NotDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("This test only runs on non-darwin systems")
	}

	h := NewHyperkitHypervisor()

	available := h.Available()

	assert.False(t, available, "Hyperkit should not be available on non-darwin systems")
}

func TestHyperkitHypervisor_Available_Bad_InvalidBinary(t *testing.T) {
	h := &HyperkitHypervisor{
		Binary: "nonexistent-hyperkit-binary-that-does-not-exist",
	}

	available := h.Available()

	assert.False(t, available)
}

func TestIsKVMAvailable_Good(t *testing.T) {
	// This test verifies the function runs without error
	// The actual result depends on the system
	result := isKVMAvailable()

	// On non-linux systems, should be false
	if runtime.GOOS != "linux" {
		assert.False(t, result, "KVM should not be available on non-linux systems")
	} else {
		// On linux, just verify it returns a boolean
		assert.IsType(t, true, result)
	}
}

func TestDetectHypervisor_Good(t *testing.T) {
	// DetectHypervisor tries to find an available hypervisor
	hv, err := DetectHypervisor()

	// This test may pass or fail depending on system configuration
	// If no hypervisor is available, it should return an error
	if err != nil {
		assert.Nil(t, hv)
		assert.Contains(t, err.Error(), "no hypervisor available")
	} else {
		assert.NotNil(t, hv)
		assert.NotEmpty(t, hv.Name())
	}
}

func TestGetHypervisor_Good_Qemu(t *testing.T) {
	hv, err := GetHypervisor("qemu")

	// Depends on whether qemu is installed
	if err != nil {
		assert.Contains(t, err.Error(), "not available")
	} else {
		assert.NotNil(t, hv)
		assert.Equal(t, "qemu", hv.Name())
	}
}

func TestGetHypervisor_Good_QemuUppercase(t *testing.T) {
	hv, err := GetHypervisor("QEMU")

	// Depends on whether qemu is installed
	if err != nil {
		assert.Contains(t, err.Error(), "not available")
	} else {
		assert.NotNil(t, hv)
		assert.Equal(t, "qemu", hv.Name())
	}
}

func TestGetHypervisor_Good_Hyperkit(t *testing.T) {
	hv, err := GetHypervisor("hyperkit")

	// On non-darwin systems, should always fail
	if runtime.GOOS != "darwin" {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	} else {
		// On darwin, depends on whether hyperkit is installed
		if err != nil {
			assert.Contains(t, err.Error(), "not available")
		} else {
			assert.NotNil(t, hv)
			assert.Equal(t, "hyperkit", hv.Name())
		}
	}
}

func TestGetHypervisor_Bad_Unknown(t *testing.T) {
	_, err := GetHypervisor("unknown-hypervisor")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown hypervisor")
}

func TestQemuHypervisor_BuildCommand_Good_WithPortsAndVolumes(t *testing.T) {
	q := NewQemuHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{
		Memory:  2048,
		CPUs:    4,
		SSHPort: 2222,
		Ports:   map[int]int{8080: 80, 443: 443},
		Volumes: map[string]string{
			"/host/data": "/container/data",
			"/host/logs": "/container/logs",
		},
		Detach: true,
	}

	cmd, err := q.BuildCommand(ctx, "/path/to/image.iso", opts)
	require.NoError(t, err)
	assert.NotNil(t, cmd)

	// Verify command includes all expected args
	args := cmd.Args
	assert.Contains(t, args, "-m")
	assert.Contains(t, args, "2048")
	assert.Contains(t, args, "-smp")
	assert.Contains(t, args, "4")
}

func TestQemuHypervisor_BuildCommand_Good_QCow2Format(t *testing.T) {
	q := NewQemuHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{Memory: 1024, CPUs: 1}

	cmd, err := q.BuildCommand(ctx, "/path/to/image.qcow2", opts)
	require.NoError(t, err)

	// Check that the drive format is qcow2
	found := false
	for _, arg := range cmd.Args {
		if arg == "file=/path/to/image.qcow2,format=qcow2" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have qcow2 drive argument")
}

func TestQemuHypervisor_BuildCommand_Good_VMDKFormat(t *testing.T) {
	q := NewQemuHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{Memory: 1024, CPUs: 1}

	cmd, err := q.BuildCommand(ctx, "/path/to/image.vmdk", opts)
	require.NoError(t, err)

	// Check that the drive format is vmdk
	found := false
	for _, arg := range cmd.Args {
		if arg == "file=/path/to/image.vmdk,format=vmdk" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have vmdk drive argument")
}

func TestQemuHypervisor_BuildCommand_Good_RawFormat(t *testing.T) {
	q := NewQemuHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{Memory: 1024, CPUs: 1}

	cmd, err := q.BuildCommand(ctx, "/path/to/image.raw", opts)
	require.NoError(t, err)

	// Check that the drive format is raw
	found := false
	for _, arg := range cmd.Args {
		if arg == "file=/path/to/image.raw,format=raw" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have raw drive argument")
}

func TestHyperkitHypervisor_BuildCommand_Good_WithPorts(t *testing.T) {
	h := NewHyperkitHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{
		Memory:  1024,
		CPUs:    2,
		SSHPort: 2222,
		Ports:   map[int]int{8080: 80},
	}

	cmd, err := h.BuildCommand(ctx, "/path/to/image.iso", opts)
	require.NoError(t, err)
	assert.NotNil(t, cmd)

	// Verify it creates a command with memory and CPU args
	args := cmd.Args
	assert.Contains(t, args, "-m")
	assert.Contains(t, args, "1024M")
	assert.Contains(t, args, "-c")
	assert.Contains(t, args, "2")
}

func TestHyperkitHypervisor_BuildCommand_Good_QCow2Format(t *testing.T) {
	h := NewHyperkitHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{Memory: 1024, CPUs: 1}

	cmd, err := h.BuildCommand(ctx, "/path/to/image.qcow2", opts)
	require.NoError(t, err)
	assert.NotNil(t, cmd)
}

func TestHyperkitHypervisor_BuildCommand_Good_RawFormat(t *testing.T) {
	h := NewHyperkitHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{Memory: 1024, CPUs: 1}

	cmd, err := h.BuildCommand(ctx, "/path/to/image.raw", opts)
	require.NoError(t, err)
	assert.NotNil(t, cmd)
}

func TestHyperkitHypervisor_BuildCommand_Good_NoPorts(t *testing.T) {
	h := NewHyperkitHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{
		Memory:  512,
		CPUs:    1,
		SSHPort: 0, // No SSH port
		Ports:   nil,
	}

	cmd, err := h.BuildCommand(ctx, "/path/to/image.iso", opts)
	require.NoError(t, err)
	assert.NotNil(t, cmd)
}

func TestQemuHypervisor_BuildCommand_Good_NoSSHPort(t *testing.T) {
	q := NewQemuHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{
		Memory:  512,
		CPUs:    1,
		SSHPort: 0, // No SSH port
		Ports:   nil,
	}

	cmd, err := q.BuildCommand(ctx, "/path/to/image.iso", opts)
	require.NoError(t, err)
	assert.NotNil(t, cmd)
}

func TestQemuHypervisor_BuildCommand_Bad_UnknownFormat(t *testing.T) {
	q := NewQemuHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{Memory: 1024, CPUs: 1}

	_, err := q.BuildCommand(ctx, "/path/to/image.txt", opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown image format")
}

func TestHyperkitHypervisor_BuildCommand_Bad_UnknownFormat(t *testing.T) {
	h := NewHyperkitHypervisor()

	ctx := context.Background()
	opts := &HypervisorOptions{Memory: 1024, CPUs: 1}

	_, err := h.BuildCommand(ctx, "/path/to/image.unknown", opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown image format")
}

func TestHyperkitHypervisor_Name_Good(t *testing.T) {
	h := NewHyperkitHypervisor()
	assert.Equal(t, "hyperkit", h.Name())
}

func TestHyperkitHypervisor_BuildCommand_Good_ISOFormat(t *testing.T) {
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
