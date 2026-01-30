// Package sdk provides SDK validation and API compatibility commands.
//
// Commands:
//   - diff: Check for breaking API changes between spec versions
//   - validate: Validate OpenAPI spec syntax
//
// Configuration via .core/sdk.yaml. For SDK generation, use: core build sdk
package sdk

import "github.com/spf13/cobra"

// AddCommands registers the 'sdk' command and all subcommands.
func AddCommands(root *cobra.Command) {
	root.AddCommand(sdkCmd)
}
