# VM Examples

## Running VMs

```bash
# Run image
core vm run server.iso

# Detached with resources
core vm run -d --memory 4096 --cpus 4 server.iso

# From template
core vm run --template core-dev --var SSH_KEY="ssh-rsa AAAA..."
```

## Management

```bash
# List running
core vm ps

# Include stopped
core vm ps -a

# Stop
core vm stop abc123

# View logs
core vm logs abc123

# Follow logs
core vm logs -f abc123

# Execute command
core vm exec abc123 ls -la

# Shell
core vm exec abc123 /bin/sh
```

## Templates

```bash
# List
core vm templates

# Show content
core vm templates show core-dev

# Show variables
core vm templates vars core-dev
```
