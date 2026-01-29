// Package ci provides release publishing commands for CI/CD pipelines.
//
// Publishes pre-built artifacts from dist/ to configured targets:
//   - GitHub Releases
//   - S3-compatible storage
//   - Custom endpoints
//
// Safe by default: runs in dry-run mode unless --were-go-for-launch is specified.
// Configuration via .core/release.yaml.
package ci

import "github.com/leaanthony/clir"

// AddCommands registers the 'ci' command and all subcommands.
func AddCommands(app *clir.Cli) {
	AddCIReleaseCommand(app)
}
