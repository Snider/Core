# core go work

Go workspace management commands.

## Subcommands

| Command | Description |
|---------|-------------|
| `sync` | Sync go.work with modules |
| `init` | Initialize go.work |
| `use` | Add module to workspace |

## Examples

```bash
core go work sync               # Sync workspace
core go work init               # Initialize workspace
core go work use ./pkg/mymodule # Add module to workspace
```
