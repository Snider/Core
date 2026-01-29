# core templates

Manage LinuxKit templates for container images.

## Usage

```bash
core templates [command]
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List available templates |
| `show` | Show template details |

## templates list

List all available LinuxKit templates.

```bash
core templates list
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
core templates show <name>
```

### Example

```bash
core templates show core-dev
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

## Template Locations

Templates are searched in order:

1. `.core/linuxkit/` - Project-specific
2. `~/.core/templates/` - User templates
3. Built-in templates

## Creating Templates

Create a LinuxKit YAML in `.core/linuxkit/`:

```yaml
# .core/linuxkit/myserver.yml
kernel:
  image: linuxkit/kernel:5.15
  cmdline: "console=tty0"

init:
  - linuxkit/init:v1.0.0

services:
  - name: sshd
    image: linuxkit/sshd:v1.0.0
  - name: myapp
    image: ghcr.io/myorg/myapp:latest
```

Run with:

```bash
core run --template myserver
```

## See Also

- [run command](run.md) - Run LinuxKit images
- [build command](build.md) - Build LinuxKit images
