# CI Changelog Examples

```bash
core ci changelog
```

## Output

```markdown
## v1.2.0

### Features
- Add user authentication (#123)
- Support dark mode (#124)

### Bug Fixes
- Fix memory leak in worker (#125)

### Performance
- Optimize database queries (#126)
```

## Configuration

`.core/release.yaml`:

```yaml
changelog:
  include:
    - feat
    - fix
    - perf
  exclude:
    - chore
    - docs
```
