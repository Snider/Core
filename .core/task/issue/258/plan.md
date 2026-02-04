# Implementation Plan: Issue 258

## Phase 1: Command Structure
1. Extend existing `internal/cmd/test/cmd_main.go` with smart detection flags
2. Add flags: `--all`, `--filter` (alias for `--run`)
3. Existing flags (`--coverage`, `--verbose`, `--short`, `--race`, `--json`, `--pkg`, `--run`) are already registered

## Phase 2: Change Detection
1. Determine diff strategy based on context:
   - **Local development** (default): `git diff --name-only HEAD` for uncommitted changes, plus `git diff --name-only --cached` for staged changes
   - **CI/PR context**: `git diff --name-only origin/dev...HEAD` to compare against base branch
   - Auto-detect CI via `CI` or `GITHUB_ACTIONS` env vars; allow override via `--base` flag
2. Filter for `.go` files (exclude `_test.go`)
3. Use `git diff --name-status` to detect renames (R), adds (A), and deletes (D):
   - **Renames**: Map tests to the new file path
   - **Deletes**: Skip deleted source files (do not run orphaned tests)
   - **New files without tests**: Log a warning
4. Map each changed file to test file(s) using N:M discovery:
   - Search for `*_test.go` files in the same package directory (not just `<file>_test.go`)
   - Handle shared test files that cover multiple source files
   - `internal/foo/bar.go` → `internal/foo/bar_test.go`, `internal/foo/bar_integration_test.go`, etc.
   - Skip if no matching test files exist (warn user)

## Phase 3: Test Execution
1. Reuse existing `runTest()` from `internal/cmd/test/cmd_runner.go`
   - This preserves environment setup (`MACOSX_DEPLOYMENT_TARGET`), output filtering (linker warnings), coverage parsing, JSON support, and consistent styling
2. Map smart detection flags to existing `runTest()` parameters:
   - `--coverage` → `coverage` param (already exists)
   - `--filter` → `run` param (mapped to `-run`)
   - Detected test packages → `pkg` param (comma-joined or iterated)
3. Do not invoke `go test` directly — all execution goes through `runTest()`

## Phase 4: Edge Cases
- No changed files → inform user, suggest `--all`
- No matching test files → inform user with list of changed files that lack tests
- `--all` flag → skip detection, call `runTest()` with `pkg="./..."` (uses existing infrastructure, not raw `go test`)
- Mixed renames and edits → deduplicate test file list
- Non-Go files changed → skip silently (only `.go` files trigger detection)

## Files to Modify
- `internal/cmd/test/cmd_main.go` (add `--all`, `--filter`, `--base` flags)
- `internal/cmd/test/cmd_runner.go` (add change detection logic before calling existing `runTest()`)
- `internal/cmd/test/cmd_detect.go` (new — git diff parsing and file-to-test mapping)

## Testing
- Add `internal/cmd/test/cmd_detect_test.go` with unit tests for:
  - File-to-test mapping (1:1, 1:N, renames, deletes)
  - Git diff parsing (`--name-only`, `--name-status`)
  - CI vs local context detection
- Manual testing with actual git changes
