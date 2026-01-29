# Dev Work Examples

```bash
# Full workflow: status → commit → push
core dev work

# Status only
core dev work --status
```

## Output

```
┌─────────────┬────────┬──────────┬─────────┐
│ Repo        │ Branch │ Status   │ Behind  │
├─────────────┼────────┼──────────┼─────────┤
│ core-php    │ main   │ clean    │ 0       │
│ core-tenant │ main   │ 2 files  │ 0       │
│ core-admin  │ dev    │ clean    │ 3       │
└─────────────┴────────┴──────────┴─────────┘
```

## Registry

```yaml
repos:
  - name: core
    path: ./core
    url: https://github.com/host-uk/core
  - name: core-php
    path: ./core-php
    url: https://github.com/host-uk/core-php
```
