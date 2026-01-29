# CLI Documentation Status

Documentation coverage for core CLI commands.

## Summary

| Package | Coverage | Status |
|---------|----------|--------|
| ai | 100% | ✓ Complete |
| build | 50% | Partial |
| ci | 100% | ✓ Complete |
| dev | 100% | ✓ Complete |
| docs | 100% | ✓ Complete |
| doctor | 100% | ✓ Complete |
| go | 100% | ✓ Complete |
| php | 100% | ✓ Complete |
| pkg | 100% | ✓ Complete |
| sdk | 100% | ✓ Complete |
| setup | 100% | ✓ Complete |
| test | 100% | ✓ Complete |
| vm | 100% | ✓ Complete |

## Remaining Gaps

### build

| Command | Status |
|---------|--------|
| build from-path | Missing (legacy command) |
| build pwa | Missing (legacy command) |

These are legacy commands for PWA/GUI builds that may be deprecated.

## Documentation Structure

All documented packages follow this structure:
```
/docs/cmd/{package}/
├── index.md          (main command + all subcommands)
├── example.md        (optional usage examples)
└── {subcommand}/
    └── index.md      (detailed subcommand docs, if needed)
```

Most packages document all commands in their index.md file with anchor links.

## Recent Updates

- 2026-01-29: Setup wizard implemented with three modes (registry, bootstrap, repo setup)
- 2026-01-29: CI flag renamed to --we-are-go-for-launch
- 2026-01-29: AI command examples fixed (core dev -> core ai)
- 2026-01-29: Added ai package documentation
- 2026-01-29: Updated php, pkg, vm documentation with full command coverage
- 2026-01-29: Task commands moved from dev to ai

Last verified: 2026-01-29
