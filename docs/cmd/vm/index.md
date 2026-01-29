# core vm

LinuxKit VM management.

## Usage

```bash
core vm <command> [flags]
```

## Commands

| Command | Description |
|---------|-------------|
| `run` | Run a LinuxKit image |
| `ps` | List running VMs |
| `stop` | Stop a VM |
| `logs` | View VM logs |
| `exec` | Execute command in VM |
| [templates](templates/) | Manage LinuxKit templates |

## vm run

Run a LinuxKit image.

```bash
core vm run <image> [flags]
core vm run --template <name> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--template` | Use a template instead of image file |
| `--var` | Set template variable (KEY=value) |
| `--name` | Name for the container |
| `--memory` | Memory in MB (default: 1024) |
| `--cpus` | CPU count (default: 1) |
| `--ssh-port` | SSH port for exec commands (default: 2222) |
| `-d` | Run in detached mode (background) |

## vm ps

List running VMs.

```bash
core vm ps [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `-a` | Show all (including stopped) |

## vm stop

Stop a running VM.

```bash
core vm stop <id>
```

## vm logs

View VM logs.

```bash
core vm logs <id> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `-f` | Follow log output |

## vm exec

Execute a command in a running VM.

```bash
core vm exec <id> <command>
```

## See Also

- [build command](../build/) - Build LinuxKit images
