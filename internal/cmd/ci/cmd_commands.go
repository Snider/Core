// Package ci provides release publishing commands for CI/CD pipelines.
//
// Publishes pre-built artifacts from dist/ to configured targets:
//   - GitHub Releases
//   - S3-compatible storage
//   - Custom endpoints
//
// Safe by default: runs in dry-run mode unless --we-are-go-for-launch is specified.
// Configuration via .core/release.yaml.
package ci

import (
	"github.com/host-uk/core/pkg/cli"
)

func init() {
	cli.RegisterCommands(AddCICommands)
}

// AddCICommands registers the 'ci' command and all subcommands.
func AddCICommands(root *cli.Command) {
	root.AddCommand(ciCmd)
}
