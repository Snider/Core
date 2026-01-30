// Package setup provides workspace setup and bootstrap commands.
package setup

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/i18n"
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
	Short: i18n.T("cmd.setup.short"),
	Long:  i18n.T("cmd.setup.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetupOrchestrator(registryPath, only, dryRun, all, name, build)
	},
}

func init() {
	setupCmd.Flags().StringVar(&registryPath, "registry", "", i18n.T("cmd.setup.flag.registry"))
	setupCmd.Flags().StringVar(&only, "only", "", i18n.T("cmd.setup.flag.only"))
	setupCmd.Flags().BoolVar(&dryRun, "dry-run", false, i18n.T("cmd.setup.flag.dry_run"))
	setupCmd.Flags().BoolVar(&all, "all", false, i18n.T("cmd.setup.flag.all"))
	setupCmd.Flags().StringVar(&name, "name", "", i18n.T("cmd.setup.flag.name"))
	setupCmd.Flags().BoolVar(&build, "build", false, i18n.T("cmd.setup.flag.build"))
}

// AddSetupCommand adds the 'setup' command to the given parent command.
func AddSetupCommand(root *cobra.Command) {
	root.AddCommand(setupCmd)
}
