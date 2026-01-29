# Setup Examples

```bash
# Clone all repos
core setup

# Specific directory
core setup --dir ~/Code/host-uk

# Use SSH
core setup --ssh
```

## Configuration

`repos.yaml`:

```yaml
org: host-uk
repos:
  core-php:
    type: package
  core-tenant:
    type: package
    depends: [core-php]
  core-admin:
    type: package
    depends: [core-php, core-tenant]
```
