// Package release provides release automation with changelog generation and publishing.
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
	// Get previous tag for comparison (uses getPreviousTag from changelog.go)
	prevTag, err := getPreviousTag(projectDir, "HEAD")
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
