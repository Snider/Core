package container

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Hypervisor defines the interface for VM hypervisors.
type Hypervisor interface {
	// Name returns the name of the hypervisor.
	Name() string
	// Available checks if the hypervisor is available on the system.
	Available() bool
	// BuildCommand builds the command to run a VM with the given options.
	BuildCommand(ctx context.Context, image string, opts *HypervisorOptions) (*exec.Cmd, error)
}

// HypervisorOptions contains options for running a VM.
type HypervisorOptions struct {
	// Memory in MB.
	Memory int
	// CPUs count.
	CPUs int
	// LogFile path for output.
	LogFile string
	// SSHPort for SSH access.
	SSHPort int
	// Ports maps host ports to guest ports.
	Ports map[int]int
	// Volumes maps host paths to guest paths (9p shares).
	Volumes map[string]string
	// Detach runs in background (nographic mode).
	Detach bool
}

// QemuHypervisor implements Hypervisor for QEMU.
type QemuHypervisor struct {
	// Binary is the path to the qemu binary (defaults to qemu-system-x86_64).
	Binary string
}

// NewQemuHypervisor creates a new QEMU hypervisor instance.
func NewQemuHypervisor() *QemuHypervisor {
	return &QemuHypervisor{
		Binary: "qemu-system-x86_64",
	}
}

// Name returns the hypervisor name.
func (q *QemuHypervisor) Name() string {
	return "qemu"
}

// Available checks if QEMU is installed and accessible.
func (q *QemuHypervisor) Available() bool {
	_, err := exec.LookPath(q.Binary)
	return err == nil
}

// BuildCommand creates the QEMU command for running a VM.
func (q *QemuHypervisor) BuildCommand(ctx context.Context, image string, opts *HypervisorOptions) (*exec.Cmd, error) {
	format := DetectImageFormat(image)
	if format == FormatUnknown {
		return nil, fmt.Errorf("unknown image format: %s", image)
	}

	args := []string{
		"-m", fmt.Sprintf("%d", opts.Memory),
		"-smp", fmt.Sprintf("%d", opts.CPUs),
		"-enable-kvm",
	}

	// Add the image based on format
	switch format {
	case FormatISO:
		args = append(args, "-cdrom", image)
		args = append(args, "-boot", "d")
	case FormatQCOW2:
		args = append(args, "-drive", fmt.Sprintf("file=%s,format=qcow2", image))
	case FormatVMDK:
		args = append(args, "-drive", fmt.Sprintf("file=%s,format=vmdk", image))
	case FormatRaw:
		args = append(args, "-drive", fmt.Sprintf("file=%s,format=raw", image))
	}

	// Always run in nographic mode for container-like behavior
	args = append(args, "-nographic")

	// Add serial console for log output
	args = append(args, "-serial", "stdio")

	// Network with port forwarding
	netdev := "user,id=net0"
	if opts.SSHPort > 0 {
		netdev += fmt.Sprintf(",hostfwd=tcp::%d-:22", opts.SSHPort)
	}
	for hostPort, guestPort := range opts.Ports {
		netdev += fmt.Sprintf(",hostfwd=tcp::%d-:%d", hostPort, guestPort)
	}
	args = append(args, "-netdev", netdev)
	args = append(args, "-device", "virtio-net-pci,netdev=net0")

	// Add 9p shares for volumes
	shareID := 0
	for hostPath, guestPath := range opts.Volumes {
		tag := fmt.Sprintf("share%d", shareID)
		args = append(args,
			"-fsdev", fmt.Sprintf("local,id=%s,path=%s,security_model=none", tag, hostPath),
			"-device", fmt.Sprintf("virtio-9p-pci,fsdev=%s,mount_tag=%s", tag, filepath.Base(guestPath)),
		)
		shareID++
	}

	// Check if KVM is available on Linux, remove -enable-kvm if not
	if runtime.GOOS != "linux" || !isKVMAvailable() {
		// Remove -enable-kvm from args
		newArgs := make([]string, 0, len(args))
		for _, arg := range args {
			if arg != "-enable-kvm" {
				newArgs = append(newArgs, arg)
			}
		}
		args = newArgs

		// On macOS, use HVF acceleration if available
		if runtime.GOOS == "darwin" {
			args = append(args, "-accel", "hvf")
		}
	}

	cmd := exec.CommandContext(ctx, q.Binary, args...)
	return cmd, nil
}

// isKVMAvailable checks if KVM is available on the system.
func isKVMAvailable() bool {
	_, err := os.Stat("/dev/kvm")
	return err == nil
}

// HyperkitHypervisor implements Hypervisor for macOS Hyperkit.
type HyperkitHypervisor struct {
	// Binary is the path to the hyperkit binary.
	Binary string
}

// NewHyperkitHypervisor creates a new Hyperkit hypervisor instance.
func NewHyperkitHypervisor() *HyperkitHypervisor {
	return &HyperkitHypervisor{
		Binary: "hyperkit",
	}
}

// Name returns the hypervisor name.
func (h *HyperkitHypervisor) Name() string {
	return "hyperkit"
}

// Available checks if Hyperkit is installed and accessible.
func (h *HyperkitHypervisor) Available() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	_, err := exec.LookPath(h.Binary)
	return err == nil
}

// BuildCommand creates the Hyperkit command for running a VM.
func (h *HyperkitHypervisor) BuildCommand(ctx context.Context, image string, opts *HypervisorOptions) (*exec.Cmd, error) {
	format := DetectImageFormat(image)
	if format == FormatUnknown {
		return nil, fmt.Errorf("unknown image format: %s", image)
	}

	args := []string{
		"-m", fmt.Sprintf("%dM", opts.Memory),
		"-c", fmt.Sprintf("%d", opts.CPUs),
		"-A", // ACPI
		"-u", // Unlimited console output
		"-s", "0:0,hostbridge",
		"-s", "31,lpc",
		"-l", "com1,stdio", // Serial console
	}

	// Add PCI slot for disk (slot 2)
	switch format {
	case FormatISO:
		args = append(args, "-s", fmt.Sprintf("2:0,ahci-cd,%s", image))
	case FormatQCOW2, FormatVMDK, FormatRaw:
		args = append(args, "-s", fmt.Sprintf("2:0,virtio-blk,%s", image))
	}

	// Network with port forwarding (slot 3)
	netArgs := "virtio-net"
	if opts.SSHPort > 0 || len(opts.Ports) > 0 {
		// Hyperkit uses slirp for user networking with port forwarding
		portForwards := make([]string, 0)
		if opts.SSHPort > 0 {
			portForwards = append(portForwards, fmt.Sprintf("tcp:%d:22", opts.SSHPort))
		}
		for hostPort, guestPort := range opts.Ports {
			portForwards = append(portForwards, fmt.Sprintf("tcp:%d:%d", hostPort, guestPort))
		}
		if len(portForwards) > 0 {
			netArgs += "," + strings.Join(portForwards, ",")
		}
	}
	args = append(args, "-s", "3:0,"+netArgs)

	cmd := exec.CommandContext(ctx, h.Binary, args...)
	return cmd, nil
}

// DetectImageFormat determines the image format from its file extension.
func DetectImageFormat(path string) ImageFormat {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".iso":
		return FormatISO
	case ".qcow2":
		return FormatQCOW2
	case ".vmdk":
		return FormatVMDK
	case ".raw", ".img":
		return FormatRaw
	default:
		return FormatUnknown
	}
}

// DetectHypervisor returns the best available hypervisor for the current platform.
func DetectHypervisor() (Hypervisor, error) {
	// On macOS, prefer Hyperkit if available, fall back to QEMU
	if runtime.GOOS == "darwin" {
		hk := NewHyperkitHypervisor()
		if hk.Available() {
			return hk, nil
		}
	}

	// Try QEMU on all platforms
	qemu := NewQemuHypervisor()
	if qemu.Available() {
		return qemu, nil
	}

	return nil, fmt.Errorf("no hypervisor available: install qemu or hyperkit (macOS)")
}

// GetHypervisor returns a specific hypervisor by name.
func GetHypervisor(name string) (Hypervisor, error) {
	switch strings.ToLower(name) {
	case "qemu":
		h := NewQemuHypervisor()
		if !h.Available() {
			return nil, fmt.Errorf("qemu is not available")
		}
		return h, nil
	case "hyperkit":
		h := NewHyperkitHypervisor()
		if !h.Available() {
			return nil, fmt.Errorf("hyperkit is not available (requires macOS)")
		}
		return h, nil
	default:
		return nil, fmt.Errorf("unknown hypervisor: %s", name)
	}
}
