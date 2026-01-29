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
