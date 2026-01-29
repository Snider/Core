# core run

Run LinuxKit images with qemu or hyperkit.

## Usage

```bash
core run <image> [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `-d, --detach` | Run in background |
| `--cpus` | Number of CPUs (default: 1) |
| `--mem` | Memory in MB (default: 1024) |
| `--disk` | Disk size (default: none) |
| `--template` | Use built-in template |

## Examples

### Run ISO Image

```bash
# Run LinuxKit ISO
core run server.iso

# With more resources
core run server.iso --cpus 2 --mem 2048

# Detached mode
core run server.iso -d
```

### Run from Template

```bash
# List available templates
core templates

# Run template
core run --template core-dev
```

### Formats

Supported image formats:
- `.iso` - Bootable ISO
- `.qcow2` - QEMU disk image
- `.raw` - Raw disk image
- `.vmdk` - VMware disk

## Container Management

```bash
# List running containers
core ps

# View logs
core logs <id>

# Follow logs
core logs -f <id>

# Execute command
core exec <id> <command>

# Stop container
core stop <id>
```

## Templates

Built-in templates in `.core/linuxkit/`:

| Template | Description |
|----------|-------------|
| `core-dev` | Development environment |
| `server-php` | FrankenPHP server |

### Custom Templates

Create `.core/linuxkit/mytemplate.yml`:

```yaml
kernel:
  image: linuxkit/kernel:6.6
  cmdline: "console=tty0"

init:
  - linuxkit/init:latest
  - linuxkit/runc:latest
  - linuxkit/containerd:latest

services:
  - name: myservice
    image: myorg/myservice:latest

files:
  - path: /etc/myconfig
    contents: |
      key: value
```

Then run:

```bash
core run --template mytemplate
```

## Networking

LinuxKit VMs get their own network namespace. Port forwarding:

```bash
# Forward port 8080
core run server.iso -p 8080:80

# Multiple ports
core run server.iso -p 8080:80 -p 8443:443
```

## Disk Persistence

```bash
# Create persistent disk
core run server.iso --disk 10G

# Attach existing disk
core run server.iso --disk /path/to/disk.qcow2
```
