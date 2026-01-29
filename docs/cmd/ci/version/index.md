# core ci version

Show the determined release version.

## Usage

```bash
core ci version
```

## Output

```
v1.2.0
```

Version is determined from:
1. `--version` flag (if provided)
2. Git tag on HEAD
3. Latest git tag + increment
4. `v0.0.1` (if no tags exist)
