# core dev commit

Claude-assisted commits across repositories.

Uses Claude to create commits for dirty repos. Shows uncommitted changes and invokes Claude to generate commit messages.

## Usage

```bash
core dev commit [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to repos.yaml (auto-detected if not specified) |
| `--all` | Commit all dirty repos without prompting |

## Examples

```bash
# Interactive commit (prompts for each repo)
core dev commit

# Commit all dirty repos automatically
core dev commit --all

# Use specific registry
core dev commit --registry ~/projects/repos.yaml
```

## How It Works

1. Scans all repositories for uncommitted changes
2. For each dirty repo:
   - Shows the diff
   - Invokes Claude to generate a commit message
   - Creates the commit with `Co-Authored-By: Claude`
3. Reports success/failure for each repo

## See Also

- [health command](../health/) - Check repo status
- [push command](../push/) - Push commits after committing
- [work command](../work/) - Full workflow (status + commit + push)
