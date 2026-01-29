// Package release provides release automation with changelog generation and publishing.
package release

import (
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
