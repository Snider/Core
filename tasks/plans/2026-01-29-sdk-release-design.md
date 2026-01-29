# SDK Release Integration Design (S3.4)

## Summary

Add `core release --target sdk` to generate SDKs as a separate release target. Runs breaking change detection before generating, uses release version for SDK versioning, outputs locally for manual publishing.

## Design Decisions

- **Separate target**: `--target sdk` runs ONLY SDK generation (no binary builds)
- **Local output**: Generates to `sdk/` directory, user handles publishing
- **Diff first**: Run breaking change detection before generating
- **Match version**: SDK version matches release version from git tags

## CLI

```bash
core release --target sdk                    # Generate SDKs only
core release --target sdk --version v1.2.3   # Explicit version
core release --target sdk --dry-run          # Preview what would generate
core release                                 # Normal release (unchanged)
```

## Config Schema

In `.core/release.yaml`:

```yaml
sdk:
  spec: openapi.yaml           # or auto-detect
  languages: [typescript, python, go, php]
  output: sdk                  # output directory
  package:
    name: myapi-sdk
  diff:
    enabled: true
    fail_on_breaking: false    # warn but continue
```

## Flow

```
core release --target sdk
    ↓
1. Load release config (.core/release.yaml)
    ↓
2. Check sdk config exists (error if not configured)
    ↓
3. Determine version (git tag or --version flag)
    ↓
4. If diff.enabled:
   - Get previous tag
   - Run oasdiff against current spec
   - If breaking && fail_on_breaking: abort
   - If breaking && !fail_on_breaking: warn, continue
    ↓
5. Generate SDKs for each language
   - Pass version to generators
   - Output to sdk/{language}/
    ↓
6. Print summary (languages generated, output paths)
```

## Package Structure

```
pkg/release/
├── sdk.go          # RunSDK() orchestration + diff helper  ← NEW
├── release.go      # Existing Run() unchanged
└── config.go       # Existing SDKConfig unchanged

pkg/sdk/
└── sdk.go          # Add SetVersion() method  ← MODIFY

cmd/core/cmd/
└── release.go      # Add --target flag  ← MODIFY
```

## RunSDK Implementation

```go
// pkg/release/sdk.go

// RunSDK executes SDK-only release: diff check + generate.
func RunSDK(ctx context.Context, cfg *Config, dryRun bool) (*SDKRelease, error) {
    if cfg.SDK == nil {
        return nil, fmt.Errorf("sdk not configured in .core/release.yaml")
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
            return nil, fmt.Errorf("failed to determine version: %w", err)
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
                return nil, fmt.Errorf("breaking API changes detected")
            }
            fmt.Printf("Warning: breaking API changes detected\n")
        }
    }

    if dryRun {
        return &SDKRelease{
            Version:   version,
            Languages: cfg.SDK.Languages,
            Output:    cfg.SDK.Output,
        }, nil
    }

    // Generate SDKs
    sdkCfg := toSDKConfig(cfg.SDK)
    s := sdk.New(projectDir, sdkCfg)
    s.SetVersion(version)

    if err := s.Generate(ctx); err != nil {
        return nil, fmt.Errorf("sdk generation failed: %w", err)
    }

    return &SDKRelease{
        Version:   version,
        Languages: cfg.SDK.Languages,
        Output:    cfg.SDK.Output,
    }, nil
}

// SDKRelease holds the result of an SDK release.
type SDKRelease struct {
    Version   string
    Languages []string
    Output    string
}
```

## CLI Integration

```go
// cmd/core/cmd/release.go

var target string
releaseCmd.StringFlag("target", "Release target (sdk)", &target)

releaseCmd.Action(func() error {
    if target == "sdk" {
        return runReleaseSDK(dryRun, version)
    }
    return runRelease(dryRun, version, draft, prerelease)
})

func runReleaseSDK(dryRun bool, version string) error {
    ctx := context.Background()
    projectDir, _ := os.Getwd()

    cfg, err := release.LoadConfig(projectDir)
    if err != nil {
        return err
    }

    if version != "" {
        cfg.SetVersion(version)
    }

    fmt.Printf("%s Generating SDKs\n", releaseHeaderStyle.Render("SDK Release:"))
    if dryRun {
        fmt.Printf("  %s\n", releaseDimStyle.Render("(dry-run mode)"))
    }

    result, err := release.RunSDK(ctx, cfg, dryRun)
    if err != nil {
        fmt.Printf("%s %v\n", releaseErrorStyle.Render("Error:"), err)
        return err
    }

    fmt.Printf("%s SDK generation complete\n", releaseSuccessStyle.Render("Success:"))
    fmt.Printf("  Version:   %s\n", result.Version)
    fmt.Printf("  Languages: %v\n", result.Languages)
    fmt.Printf("  Output:    %s/\n", result.Output)

    return nil
}
```

## Implementation Steps

1. Add `SetVersion()` method to `pkg/sdk/sdk.go`
2. Create `pkg/release/sdk.go` with `RunSDK()` and helpers
3. Add `--target` flag to `cmd/core/cmd/release.go`
4. Add `runReleaseSDK()` function to CLI
5. Add tests for `pkg/release/sdk_test.go`
6. Final verification and TODO update

## Dependencies

- `oasdiff` CLI (for breaking change detection)
- Existing SDK generators (openapi-generator, etc.)
