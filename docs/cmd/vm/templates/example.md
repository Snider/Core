# VM Templates Examples

## List

```bash
core vm templates
```

## Show

```bash
core vm templates show core-dev
```

## Variables

```bash
core vm templates vars core-dev
```

## Output

```
Variables for core-dev:
  SSH_KEY      (required)  SSH public key
  MEMORY       (optional)  Memory in MB (default: 4096)
  CPUS         (optional)  CPU count (default: 4)
```

## Using Templates

```bash
core vm run --template core-dev --var SSH_KEY="ssh-rsa AAAA..."
```

## Template Format

`.core/linuxkit/myserver.yml`:

```yaml
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
