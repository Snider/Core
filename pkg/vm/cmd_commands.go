// Package vm provides LinuxKit virtual machine management commands.
//
// Commands:
//   - run: Run a VM from image (.iso, .qcow2, .vmdk, .raw) or template
//   - ps: List running VMs
//   - stop: Stop a running VM
//   - logs: View VM logs
//   - exec: Execute command in VM via SSH
//   - templates: Manage LinuxKit templates (list, build)
//
// Uses qemu or hyperkit depending on system availability.
// Templates are built from YAML definitions and can include variables.
package vm
