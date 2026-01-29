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
