# SDK Release Implementation Plan (S3.4)

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `core release --target sdk` to generate SDKs with version and diff checking

**Architecture:** Separate release target that runs diff check then SDK generation, outputs locally

**Tech Stack:** Go, existing pkg/sdk generators, oasdiff for diff

---

## Task 1: Add SetVersion to SDK struct

**Files:**
- Modify: `pkg/sdk/sdk.go`
- Test: `pkg/sdk/sdk_test.go` (create if needed)

**Step 1: Write the failing test**

```go
// pkg/sdk/sdk_test.go
package sdk

import (
    "testing"
)

func TestSDK_Good_SetVersion(t *testing.T) {
    s := New("/tmp", nil)
    s.SetVersion("v1.2.3")

    if s.version != "v1.2.3" {
        t.Errorf("expected version v1.2.3, got %s", s.version)
    }
}

func TestSDK_Good_VersionPassedToGenerator(t *testing.T) {
    config := &Config{
        Languages: []string{"typescript"},
        Output:    "sdk",
        Package: PackageConfig{
            Name: "test-sdk",
        },
    }
    s := New("/tmp", config)
    s.SetVersion("v2.0.0")

    // Version should override config
    if s.config.Package.Version != "v2.0.0" {
        t.Errorf("expected config version v2.0.0, got %s", s.config.Package.Version)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/sdk/... -run TestSDK_Good_SetVersion -v`
Expected: FAIL with "s.version undefined" or similar

**Step 3: Write minimal implementation**

Add to `pkg/sdk/sdk.go`:

```go
// SDK struct - add version field
type SDK struct {
    config     *Config
    projectDir string
    version    string  // ADD THIS
}

// SetVersion sets the SDK version, overriding config.
func (s *SDK) SetVersion(version string) {
    s.version = version
    if s.config != nil {
        s.config.Package.Version = version
    }
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/sdk/... -run TestSDK_Good -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/sdk/sdk.go pkg/sdk/sdk_test.go
git commit -m "feat(sdk): add SetVersion method for release integration"
```

---

## Task 2: Create pkg/release/sdk.go structure

**Files:**
- Create: `pkg/release/sdk.go`

**Step 1: Create file with types and helper**

```go
// pkg/release/sdk.go
package release

import (
    "context"
    "fmt"

    "github.com/host-uk/core/pkg/sdk"
)

// SDKRelease holds the result of an SDK release.
type SDKRelease struct {
    // Version is the SDK version.
    Version string
    // Languages that were generated.
    Languages []string
    // Output directory.
    Output string
}

// toSDKConfig converts release.SDKConfig to sdk.Config.
func toSDKConfig(cfg *SDKConfig) *sdk.Config {
    if cfg == nil {
        return nil
    }
    return &sdk.Config{
        Spec:      cfg.Spec,
        Languages: cfg.Languages,
        Output:    cfg.Output,
        Package: sdk.PackageConfig{
            Name:    cfg.Package.Name,
            Version: cfg.Package.Version,
        },
        Diff: sdk.DiffConfig{
            Enabled:        cfg.Diff.Enabled,
            FailOnBreaking: cfg.Diff.FailOnBreaking,
        },
    }
}
```

**Step 2: Verify it compiles**

Run: `go build ./pkg/release/...`
Expected: Success

**Step 3: Commit**

```bash
git add pkg/release/sdk.go
git commit -m "feat(release): add SDK release types and config converter"
```

---

## Task 3: Implement RunSDK function

**Files:**
- Modify: `pkg/release/sdk.go`
- Test: `pkg/release/sdk_test.go`

**Step 1: Write the failing test**

```go
// pkg/release/sdk_test.go
package release

import (
    "context"
    "testing"
)

func TestRunSDK_Bad_NoConfig(t *testing.T) {
    cfg := &Config{
        SDK: nil,
    }
    cfg.projectDir = "/tmp"

    _, err := RunSDK(context.Background(), cfg, true)
    if err == nil {
        t.Error("expected error when SDK config is nil")
    }
}

func TestRunSDK_Good_DryRun(t *testing.T) {
    cfg := &Config{
        SDK: &SDKConfig{
            Languages: []string{"typescript", "python"},
            Output:    "sdk",
        },
    }
    cfg.projectDir = "/tmp"
    cfg.version = "v1.0.0"

    result, err := RunSDK(context.Background(), cfg, true)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if result.Version != "v1.0.0" {
        t.Errorf("expected version v1.0.0, got %s", result.Version)
    }
    if len(result.Languages) != 2 {
        t.Errorf("expected 2 languages, got %d", len(result.Languages))
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/release/... -run TestRunSDK -v`
Expected: FAIL with "RunSDK undefined"

**Step 3: Write implementation**

Add to `pkg/release/sdk.go`:

```go
// RunSDK executes SDK-only release: diff check + generate.
// If dryRun is true, it shows what would be done without generating.
func RunSDK(ctx context.Context, cfg *Config, dryRun bool) (*SDKRelease, error) {
    if cfg == nil {
        return nil, fmt.Errorf("release.RunSDK: config is nil")
    }
    if cfg.SDK == nil {
        return nil, fmt.Errorf("release.RunSDK: sdk not configured in .core/release.yaml")
    }

    projectDir := cfg.projectDir
    if projectDir == "" {
        projectDir = "."
    }

    // Determine version
    version := cfg.version
    if version == "" {
        var err error
        version, err = DetermineVersion(projectDir)
        if err != nil {
            return nil, fmt.Errorf("release.RunSDK: failed to determine version: %w", err)
        }
    }

    // Run diff check if enabled
    if cfg.SDK.Diff.Enabled {
        breaking, err := checkBreakingChanges(projectDir, cfg.SDK)
        if err != nil {
            // Non-fatal: warn and continue
            fmt.Printf("Warning: diff check failed: %v\n", err)
        } else if breaking {
            if cfg.SDK.Diff.FailOnBreaking {
                return nil, fmt.Errorf("release.RunSDK: breaking API changes detected")
            }
            fmt.Printf("Warning: breaking API changes detected\n")
        }
    }

    // Prepare result
    output := cfg.SDK.Output
    if output == "" {
        output = "sdk"
    }

    result := &SDKRelease{
        Version:   version,
        Languages: cfg.SDK.Languages,
        Output:    output,
    }

    if dryRun {
        return result, nil
    }

    // Generate SDKs
    sdkCfg := toSDKConfig(cfg.SDK)
    s := sdk.New(projectDir, sdkCfg)
    s.SetVersion(version)

    if err := s.Generate(ctx); err != nil {
        return nil, fmt.Errorf("release.RunSDK: generation failed: %w", err)
    }

    return result, nil
}

// checkBreakingChanges runs oasdiff to detect breaking changes.
func checkBreakingChanges(projectDir string, cfg *SDKConfig) (bool, error) {
    // Get previous tag for comparison
    prevTag, err := getPreviousTag(projectDir)
    if err != nil {
        return false, fmt.Errorf("no previous tag found: %w", err)
    }

    // Detect spec path
    specPath := cfg.Spec
    if specPath == "" {
        s := sdk.New(projectDir, nil)
        specPath, err = s.DetectSpec()
        if err != nil {
            return false, err
        }
    }

    // Run diff
    result, err := sdk.Diff(prevTag, specPath)
    if err != nil {
        return false, err
    }

    return result.Breaking, nil
}

// getPreviousTag gets the most recent tag before HEAD.
func getPreviousTag(projectDir string) (string, error) {
    // Use git describe to get previous tag
    // This is a simplified version - may need refinement
    cmd := exec.Command("git", "describe", "--tags", "--abbrev=0", "HEAD^")
    cmd.Dir = projectDir
    out, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(out)), nil
}
```

Add import for `os/exec` and `strings`.

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/release/... -run TestRunSDK -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/release/sdk.go pkg/release/sdk_test.go
git commit -m "feat(release): implement RunSDK for SDK-only releases"
```

---

## Task 4: Add --target flag to CLI

**Files:**
- Modify: `cmd/core/cmd/release.go`

**Step 1: Add target flag and routing**

In `AddReleaseCommand`, add:

```go
var target string
releaseCmd.StringFlag("target", "Release target (sdk)", &target)

// Update the action
releaseCmd.Action(func() error {
    if target == "sdk" {
        return runReleaseSDK(dryRun, version)
    }
    return runRelease(dryRun, version, draft, prerelease)
})
```

**Step 2: Verify it compiles**

Run: `go build ./cmd/core/...`
Expected: FAIL with "runReleaseSDK undefined"

**Step 3: Commit partial progress**

```bash
git add cmd/core/cmd/release.go
git commit -m "feat(cli): add --target flag to release command"
```

---

## Task 5: Implement runReleaseSDK CLI function

**Files:**
- Modify: `cmd/core/cmd/release.go`

**Step 1: Add the function**

```go
// runReleaseSDK executes SDK-only release.
func runReleaseSDK(dryRun bool, version string) error {
    ctx := context.Background()

    projectDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("failed to get working directory: %w", err)
    }

    // Load configuration
    cfg, err := release.LoadConfig(projectDir)
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    // Apply CLI overrides
    if version != "" {
        cfg.SetVersion(version)
    }

    // Print header
    fmt.Printf("%s Generating SDKs\n", releaseHeaderStyle.Render("SDK Release:"))
    if dryRun {
        fmt.Printf("  %s\n", releaseDimStyle.Render("(dry-run mode)"))
    }
    fmt.Println()

    // Run SDK release
    result, err := release.RunSDK(ctx, cfg, dryRun)
    if err != nil {
        fmt.Printf("%s %v\n", releaseErrorStyle.Render("Error:"), err)
        return err
    }

    // Print summary
    fmt.Println()
    fmt.Printf("%s SDK generation complete!\n", releaseSuccessStyle.Render("Success:"))
    fmt.Printf("  Version:   %s\n", releaseValueStyle.Render(result.Version))
    fmt.Printf("  Languages: %v\n", result.Languages)
    fmt.Printf("  Output:    %s/\n", releaseValueStyle.Render(result.Output))

    return nil
}
```

**Step 2: Verify it compiles and help shows flag**

Run: `go build -o bin/core ./cmd/core && ./bin/core release --help`
Expected: Shows `--target` flag in help output

**Step 3: Commit**

```bash
git add cmd/core/cmd/release.go
git commit -m "feat(cli): implement runReleaseSDK for SDK generation"
```

---

## Task 6: Add integration tests

**Files:**
- Modify: `pkg/release/sdk_test.go`

**Step 1: Add more test cases**

```go
func TestRunSDK_Good_WithDiffEnabled(t *testing.T) {
    cfg := &Config{
        SDK: &SDKConfig{
            Languages: []string{"typescript"},
            Output:    "sdk",
            Diff: SDKDiffConfig{
                Enabled:        true,
                FailOnBreaking: false,
            },
        },
    }
    cfg.projectDir = "/tmp"
    cfg.version = "v1.0.0"

    // Dry run should succeed even without git repo
    result, err := RunSDK(context.Background(), cfg, true)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.Version != "v1.0.0" {
        t.Errorf("expected v1.0.0, got %s", result.Version)
    }
}

func TestRunSDK_Good_DefaultOutput(t *testing.T) {
    cfg := &Config{
        SDK: &SDKConfig{
            Languages: []string{"go"},
            // Output not set - should default to "sdk"
        },
    }
    cfg.projectDir = "/tmp"
    cfg.version = "v1.0.0"

    result, err := RunSDK(context.Background(), cfg, true)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.Output != "sdk" {
        t.Errorf("expected default output 'sdk', got %s", result.Output)
    }
}

func TestToSDKConfig_Good_Conversion(t *testing.T) {
    relCfg := &SDKConfig{
        Spec:      "api.yaml",
        Languages: []string{"typescript", "python"},
        Output:    "generated",
        Package: SDKPackageConfig{
            Name:    "my-sdk",
            Version: "v2.0.0",
        },
        Diff: SDKDiffConfig{
            Enabled:        true,
            FailOnBreaking: true,
        },
    }

    sdkCfg := toSDKConfig(relCfg)

    if sdkCfg.Spec != "api.yaml" {
        t.Errorf("expected spec api.yaml, got %s", sdkCfg.Spec)
    }
    if len(sdkCfg.Languages) != 2 {
        t.Errorf("expected 2 languages, got %d", len(sdkCfg.Languages))
    }
    if sdkCfg.Package.Name != "my-sdk" {
        t.Errorf("expected package name my-sdk, got %s", sdkCfg.Package.Name)
    }
}

func TestToSDKConfig_Good_NilInput(t *testing.T) {
    result := toSDKConfig(nil)
    if result != nil {
        t.Error("expected nil for nil input")
    }
}
```

**Step 2: Run all tests**

Run: `go test ./pkg/release/... -v`
Expected: All tests PASS

**Step 3: Commit**

```bash
git add pkg/release/sdk_test.go
git commit -m "test(release): add SDK release integration tests"
```

---

## Task 7: Final verification and TODO update

**Step 1: Build CLI**

Run: `go build -o bin/core ./cmd/core`
Expected: Success

**Step 2: Test help output**

Run: `./bin/core release --help`
Expected: Shows `--target` flag

**Step 3: Run all tests**

Run: `go test ./pkg/release/... ./pkg/sdk/... -v`
Expected: All PASS

**Step 4: Update TODO.md**

Mark S3.4 `core release --target sdk` as complete in `tasks/TODO.md`.

**Step 5: Commit**

```bash
git add tasks/TODO.md
git commit -m "docs: mark S3.4 SDK release integration as complete"
```
