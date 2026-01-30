// Package setup provides workspace setup and bootstrap commands.
package setup

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/spf13/cobra"
)

// Style aliases from shared package
var (
	repoNameStyle = shared.RepoNameStyle
	successStyle  = shared.SuccessStyle
	errorStyle    = shared.ErrorStyle
	dimStyle      = shared.DimStyle
)

// Default organization and devops repo for bootstrap
const (
	defaultOrg      = "host-uk"
	devopsRepo      = "core-devops"
	devopsReposYaml = "repos.yaml"
)

// Setup command flags
var (
	registryPath string
	only         string
	dryRun       bool
	all          bool
	name         string
	build        bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Bootstrap workspace or clone packages from registry",
	Long: `Sets up a development workspace.

REGISTRY MODE (repos.yaml exists):
  Clones all repositories defined in repos.yaml into packages/.
  Skips repos that already exist. Use --only to filter by type.

BOOTSTRAP MODE (no repos.yaml):
  1. Clones core-devops to set up the workspace
  2. Presents an interactive wizard to select packages
  3. Clones selected packages

Use --all to skip the wizard and clone everything.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetupOrchestrator(registryPath, only, dryRun, all, name, build)
	},
}

func init() {
	setupCmd.Flags().StringVar(&registryPath, "registry", "", "Path to repos.yaml (auto-detected if not specified)")
	setupCmd.Flags().StringVar(&only, "only", "", "Only clone repos of these types (comma-separated: foundation,module,product)")
	setupCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be cloned without cloning")
	setupCmd.Flags().BoolVar(&all, "all", false, "Skip wizard, clone all packages (non-interactive)")
	setupCmd.Flags().StringVar(&name, "name", "", "Project directory name for bootstrap mode")
	setupCmd.Flags().BoolVar(&build, "build", false, "Run build after cloning")
}

// AddSetupCommand adds the 'setup' command to the given parent command.
func AddSetupCommand(root *cobra.Command) {
	root.AddCommand(setupCmd)
}
