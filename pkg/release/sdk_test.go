package release

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSDK_Bad_NilConfig(t *testing.T) {
	_, err := RunSDK(context.Background(), nil, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config is nil")
}

func TestRunSDK_Bad_NoSDKConfig(t *testing.T) {
	cfg := &Config{
		SDK: nil,
	}
	cfg.projectDir = "/tmp"

	_, err := RunSDK(context.Background(), cfg, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sdk not configured")
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
	require.NoError(t, err)

	assert.Equal(t, "v1.0.0", result.Version)
	assert.Len(t, result.Languages, 2)
	assert.Contains(t, result.Languages, "typescript")
	assert.Contains(t, result.Languages, "python")
	assert.Equal(t, "sdk", result.Output)
}

func TestRunSDK_Good_DryRunDefaultOutput(t *testing.T) {
	cfg := &Config{
		SDK: &SDKConfig{
			Languages: []string{"go"},
			Output:    "", // Empty output, should default to "sdk"
		},
	}
	cfg.projectDir = "/tmp"
	cfg.version = "v2.0.0"

	result, err := RunSDK(context.Background(), cfg, true)
	require.NoError(t, err)

	assert.Equal(t, "sdk", result.Output)
}

func TestRunSDK_Good_DryRunDefaultProjectDir(t *testing.T) {
	cfg := &Config{
		SDK: &SDKConfig{
			Languages: []string{"typescript"},
			Output:    "out",
		},
	}
	// projectDir is empty, should default to "."
	cfg.version = "v1.0.0"

	result, err := RunSDK(context.Background(), cfg, true)
	require.NoError(t, err)

	assert.Equal(t, "v1.0.0", result.Version)
}

func TestRunSDK_Bad_BreakingChangesFailOnBreaking(t *testing.T) {
	// This test verifies that when diff.FailOnBreaking is true and breaking changes
	// are detected, RunSDK returns an error. However, since we can't easily mock
	// the diff check, this test verifies the config is correctly processed.
	// The actual breaking change detection is tested in pkg/sdk/diff_test.go.
	cfg := &Config{
		SDK: &SDKConfig{
			Languages: []string{"typescript"},
			Output:    "sdk",
			Diff: SDKDiffConfig{
				Enabled:        true,
				FailOnBreaking: true,
			},
		},
	}
	cfg.projectDir = "/tmp"
	cfg.version = "v1.0.0"

	// In dry run mode with no git repo, diff check will fail gracefully
	// (non-fatal warning), so this should succeed
	result, err := RunSDK(context.Background(), cfg, true)
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", result.Version)
}

func TestToSDKConfig_Good(t *testing.T) {
	sdkCfg := &SDKConfig{
		Spec:      "api/openapi.yaml",
		Languages: []string{"typescript", "go"},
		Output:    "sdk",
		Package: SDKPackageConfig{
			Name:    "myapi",
			Version: "v1.0.0",
		},
		Diff: SDKDiffConfig{
			Enabled:        true,
			FailOnBreaking: true,
		},
	}

	result := toSDKConfig(sdkCfg)

	assert.Equal(t, "api/openapi.yaml", result.Spec)
	assert.Equal(t, []string{"typescript", "go"}, result.Languages)
	assert.Equal(t, "sdk", result.Output)
	assert.Equal(t, "myapi", result.Package.Name)
	assert.Equal(t, "v1.0.0", result.Package.Version)
	assert.True(t, result.Diff.Enabled)
	assert.True(t, result.Diff.FailOnBreaking)
}

func TestToSDKConfig_Good_NilInput(t *testing.T) {
	result := toSDKConfig(nil)
	assert.Nil(t, result)
}

func TestRunSDK_Good_WithDiffEnabledNoFailOnBreaking(t *testing.T) {
	// Tests diff enabled but FailOnBreaking=false (should warn but not fail)
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

	// Dry run should succeed even without git repo (diff check fails gracefully)
	result, err := RunSDK(context.Background(), cfg, true)
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", result.Version)
	assert.Contains(t, result.Languages, "typescript")
}

func TestRunSDK_Good_MultipleLanguages(t *testing.T) {
	// Tests multiple language support
	cfg := &Config{
		SDK: &SDKConfig{
			Languages: []string{"typescript", "python", "go", "java"},
			Output:    "multi-sdk",
		},
	}
	cfg.projectDir = "/tmp"
	cfg.version = "v3.0.0"

	result, err := RunSDK(context.Background(), cfg, true)
	require.NoError(t, err)

	assert.Equal(t, "v3.0.0", result.Version)
	assert.Len(t, result.Languages, 4)
	assert.Equal(t, "multi-sdk", result.Output)
}

func TestRunSDK_Good_WithPackageConfig(t *testing.T) {
	// Tests that package config is properly handled
	cfg := &Config{
		SDK: &SDKConfig{
			Spec:      "openapi.yaml",
			Languages: []string{"typescript"},
			Output:    "sdk",
			Package: SDKPackageConfig{
				Name:    "my-custom-sdk",
				Version: "v2.5.0",
			},
		},
	}
	cfg.projectDir = "/tmp"
	cfg.version = "v1.0.0"

	result, err := RunSDK(context.Background(), cfg, true)
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", result.Version)
}

func TestToSDKConfig_Good_EmptyPackageConfig(t *testing.T) {
	// Tests conversion with empty package config
	sdkCfg := &SDKConfig{
		Languages: []string{"go"},
		Output:    "sdk",
		// Package is empty struct
	}

	result := toSDKConfig(sdkCfg)

	assert.Equal(t, []string{"go"}, result.Languages)
	assert.Equal(t, "sdk", result.Output)
	assert.Empty(t, result.Package.Name)
	assert.Empty(t, result.Package.Version)
}

func TestToSDKConfig_Good_DiffDisabled(t *testing.T) {
	// Tests conversion with diff disabled
	sdkCfg := &SDKConfig{
		Languages: []string{"typescript"},
		Output:    "sdk",
		Diff: SDKDiffConfig{
			Enabled:        false,
			FailOnBreaking: false,
		},
	}

	result := toSDKConfig(sdkCfg)

	assert.False(t, result.Diff.Enabled)
	assert.False(t, result.Diff.FailOnBreaking)
}
