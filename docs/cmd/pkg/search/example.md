# Package Search Examples

```bash
# Find all core-* packages
core pkg search core-

# Search term
core pkg search api

# Different org
core pkg search --org myorg query
```

## Output

```
┌──────────────┬─────────────────────────────┐
│ Package      │ Description                 │
├──────────────┼─────────────────────────────┤
│ core-api     │ REST API framework          │
│ core-auth    │ Authentication utilities    │
└──────────────┴─────────────────────────────┘
```
