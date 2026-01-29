# Setup Examples

```bash
# Clone all repos
core setup

# Preview what would be cloned
core setup --dry-run

# Only foundation packages
core setup --only foundation

# Multiple types
core setup --only foundation,module
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
