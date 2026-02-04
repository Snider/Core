# Issue 258: Smart Test Detection

## Original Issue
<https://github.com/host-uk/core/issues/258>

## Summary
Make `core test` smart — detect changed Go files and run only relevant tests.

> **Scope:** Go-only. The existing `core test` command (`internal/cmd/test/`) targets Go projects (requires `go.mod`). Future language support (PHP, etc.) would be added as separate detection strategies, but this issue covers Go only.

## Commands
```bash
core test                    # Run tests for changed files only
core test --all              # Run all tests (skip detection)
core test --filter UserTest  # Run specific test pattern
core test --coverage         # With coverage report
core test --base origin/dev  # Compare against specific base branch (CI)
```

## Acceptance Criteria
- [ ] Detect changed `.go` files via `git diff` (local: `HEAD`, CI: `origin/dev...HEAD`)
- [ ] Handle renames, deletes, and new files via `git diff --name-status`
- [ ] Map source files to test files using N:M discovery (`foo.go` → `foo_test.go`, `foo_integration_test.go`, etc.)
- [ ] Warn when changed files have no corresponding tests
- [ ] Execute tests through existing `runTest()` infrastructure (not raw `go test`)
- [ ] Support `--all` flag to skip detection and run all tests
- [ ] Support `--filter` flag for test name pattern matching
- [ ] Support `--coverage` flag for coverage reports
- [ ] Support `--base` flag for CI/PR diff context

## Technical Context
- Existing `core test` command: `internal/cmd/test/cmd_main.go`
- Existing test runner: `internal/cmd/test/cmd_runner.go` (`runTest()`)
- Output parsing: `internal/cmd/test/cmd_output.go`
- Command registration: `internal/cmd/test/cmd_commands.go` via `cli.RegisterCommands()`
- Follow existing patterns in `internal/cmd/test/`
