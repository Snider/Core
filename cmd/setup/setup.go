// Package setup provides workspace setup and bootstrap commands.
package setup

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/leaanthony/clir"
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

// AddSetupCommand adds the 'setup' command to the given parent command.
func AddSetupCommand(parent *clir.Cli) {
	var registryPath string
	var only string
	var dryRun bool
	var all bool
	var name string
	var build bool

	setupCmd := parent.NewSubCommand("setup", "Bootstrap workspace or clone packages from registry")
	setupCmd.LongDescription("Sets up a development workspace.\n\n" +
		"REGISTRY MODE (repos.yaml exists):\n" +
		"  Clones all repositories defined in repos.yaml into packages/.\n" +
		"  Skips repos that already exist. Use --only to filter by type.\n\n" +
		"BOOTSTRAP MODE (no repos.yaml):\n" +
		"  1. Clones core-devops to set up the workspace\n" +
		"  2. Presents an interactive wizard to select packages\n" +
		"  3. Clones selected packages\n\n" +
		"Use --all to skip the wizard and clone everything.")

	setupCmd.StringFlag("registry", "Path to repos.yaml (auto-detected if not specified)", &registryPath)
	setupCmd.StringFlag("only", "Only clone repos of these types (comma-separated: foundation,module,product)", &only)
	setupCmd.BoolFlag("dry-run", "Show what would be cloned without cloning", &dryRun)
	setupCmd.BoolFlag("all", "Skip wizard, clone all packages (non-interactive)", &all)
	setupCmd.StringFlag("name", "Project directory name for bootstrap mode", &name)
	setupCmd.BoolFlag("build", "Run build after cloning", &build)

	setupCmd.Action(func() error {
		return runSetupOrchestrator(registryPath, only, dryRun, all, name, build)
	})
}
