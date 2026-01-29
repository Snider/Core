# core vm

LinuxKit VM management.

LinuxKit VMs are lightweight, immutable VMs built from YAML templates.
They run using qemu or hyperkit depending on your system.

## Usage

```bash
core vm <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| [`run`](#vm-run) | Run a LinuxKit image or template |
| [`ps`](#vm-ps) | List running VMs |
| [`stop`](#vm-stop) | Stop a VM |
| [`logs`](#vm-logs) | View VM logs |
| [`exec`](#vm-exec) | Execute command in VM |
| [templates](templates/) | Manage LinuxKit templates |

---

## vm run

Run a LinuxKit image or build from a template.

```bash
core vm run <image> [flags]
core vm run --template <name> [flags]
```

Supported image formats: `.iso`, `.qcow2`, `.vmdk`, `.raw`

### Flags

| Flag | Description |
|------|-------------|
| `--template` | Run from a LinuxKit template (build + run) |
| `--var` | Template variable in KEY=VALUE format (repeatable) |
| `--name` | Name for the container |
| `--memory` | Memory in MB (default: 1024) |
| `--cpus` | CPU count (default: 1) |
| `--ssh-port` | SSH port for exec commands (default: 2222) |
| `-d` | Run in detached mode (background) |

### Examples

```bash
# Run from image file
core vm run image.iso

# Run detached with more resources
core vm run -d image.qcow2 --memory 2048 --cpus 4

# Run from template
core vm run --template core-dev --var SSH_KEY="ssh-rsa AAAA..."

# Multiple template variables
core vm run --template server-php --var SSH_KEY="..." --var DOMAIN=example.com
```

---

## vm ps

List running VMs.

```bash
core vm ps [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `-a` | Show all (including stopped) |

### Output

```
ID        NAME      IMAGE                STATUS    STARTED   PID
abc12345  myvm      ...core-dev.qcow2    running   5m        12345
```

---

## vm stop

Stop a running VM by ID or name.

```bash
core vm stop <id>
```

Supports partial ID matching.

### Examples

```bash
# Full ID
core vm stop abc12345678

# Partial ID
core vm stop abc1
```

---

## vm logs

View VM logs.

```bash
core vm logs <id> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `-f` | Follow log output |

### Examples

```bash
# View logs
core vm logs abc12345

# Follow logs
core vm logs -f abc1
```

---

## vm exec

Execute a command in a running VM via SSH.

```bash
core vm exec <id> <command...>
```

### Examples

```bash
# List files
core vm exec abc12345 ls -la

# Open shell
core vm exec abc1 /bin/sh
```

---

## See Also

- [templates](templates/) - Manage LinuxKit templates
- [build](../build/) - Build LinuxKit images
- [dev](../dev/) - Dev environment management
