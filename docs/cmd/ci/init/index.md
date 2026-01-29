# core ci init

Initialize release configuration.

## Usage

```bash
core ci init
```

Creates `.core/release.yaml` with default configuration.

## Example Output

```yaml
version: 1

project:
  name: myapp

publishers:
  - type: github
```

See [config.md](../config.md) for full configuration options.
