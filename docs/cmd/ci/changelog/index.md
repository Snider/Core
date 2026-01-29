# core ci changelog

Generate changelog from conventional commits.

## Usage

```bash
core ci changelog
```

## Output

Generates markdown changelog from git commits since last tag:

```markdown
## v1.2.0

### Features
- Add user authentication (#123)
- Support dark mode (#124)

### Bug Fixes
- Fix memory leak in worker (#125)
```

## Configuration

See [configuration.md](../../../configuration.md) for changelog configuration options.
