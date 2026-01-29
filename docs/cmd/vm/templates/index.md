# core vm templates

Manage LinuxKit templates for container images.

## Usage

```bash
core vm templates [command]
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List available templates |
| `show` | Show template details |
| `vars` | Show template variables |

## templates list

List all available LinuxKit templates.

```bash
core vm templates list
```

### Output

```
Available Templates:

  core-dev
    Full development environment with 100+ tools
    Platforms: linux/amd64, linux/arm64

  server-php
    FrankenPHP production server
    Platforms: linux/amd64, linux/arm64

  edge-node
    Minimal edge deployment
    Platforms: linux/amd64, linux/arm64
```

## templates show

Show details of a specific template.

```bash
core vm templates show <name>
```

### Example

```bash
core vm templates show core-dev
```

Output:
```
Template: core-dev

Description: Full development environment with 100+ tools

Platforms:
  - linux/amd64
  - linux/arm64

Formats:
  - iso
  - qcow2

Services:
  - sshd
  - docker
  - frankenphp

Size: ~1.8GB
```

## templates vars

Show variables defined by a template.

```bash
core vm templates vars <name>
```

### Example

```bash
core vm templates vars core-dev
```

Output:
```
Variables for core-dev:
  SSH_KEY      (required)  SSH public key
  MEMORY       (optional)  Memory in MB (default: 4096)
  CPUS         (optional)  CPU count (default: 4)
```

## Template Locations

Templates are searched in order:

1. `.core/linuxkit/` - Project-specific
2. `~/.core/templates/` - User templates
3. Built-in templates

## Creating Templates

Create a LinuxKit YAML in `.core/linuxkit/`. See [Template Format](example.md#template-format) for examples.

Run with:

```bash
core vm run --template myserver
```

## See Also

- [vm command](../) - Run LinuxKit images
- [build command](../../build/) - Build LinuxKit images
