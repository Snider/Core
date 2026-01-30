# Commit Instructions

## Context

When asked to commit changes in these repositories:

1. **All changes are pre-approved** - Both tracked (modified) and untracked files have been reviewed and approved for commit
2. **Include everything** - Commit ALL modified files AND all untracked files (including new directories)
3. **Don't skip files** - Do not leave out untracked files assuming they need separate review
4. **Single commit preferred** - Combine all changes into one cohesive commit unless explicitly told otherwise

The user has already validated these changes. Proceed with confidence.

## Handling Ignored Files

If you see untracked directories that should typically be ignored (like `node_modules/`, `vendor/`, `.cache/`, `dist/`, `build/`):

1. **Fix the .gitignore** - Create or update `.gitignore` to exclude these directories
2. **Commit the .gitignore** - Include this fix in your commit
3. **Don't ask** - Just fix it and commit

Common patterns to add to .gitignore:
```
node_modules/
vendor/
.cache/
dist/
build/
*.log
.env
.DS_Store
```

## Commit Message Style

- Use conventional commit format: `type(scope): description`
- Common types: `refactor`, `feat`, `fix`, `docs`, `chore`
- Keep the first line under 72 characters
- Add body for complex changes explaining the "why"
- Include `Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>`

## Task

Review the uncommitted changes and create an appropriate commit. Be concise.
