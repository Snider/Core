// Package cmd implements the core CLI application.
//
// The CLI provides commands for:
//   - Multi-repo development workflows (dev)
//   - AI agent task management (ai)
//   - Go and PHP development tools (go, php)
//   - Build and release automation (build, ci)
//   - SDK validation and API compatibility (sdk)
//   - Package and environment management (pkg, vm)
//   - Documentation and testing (docs, test)
//   - Environment health checks (doctor)
//   - Repository setup and cloning (setup)
//
// Two build variants exist:
//   - Default build: Full development toolset
//   - CI build (-tags ci): Minimal release toolset
package cmd

import (
	"os"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/framework"
	"github.com/spf13/cobra"
)

const (
	appName    = "core"
	appVersion = "0.1.0"
)

// Terminal styles using Tailwind colour palette (from shared package).
var (
	// coreStyle is used for primary headings and the CLI name.
	coreStyle = cli.RepoNameStyle

	// linkStyle is used for URLs and clickable references.
	linkStyle = cli.LinkStyle
)

// rootCmd is the base command for the CLI.
var rootCmd = &cobra.Command{
	Use:     appName,
	Short:   "CLI tool for development and production",
	Version: appVersion,
}

// Execute initialises and runs the CLI application.
// Commands are registered based on build tags (see core_ci.go and core_dev.go).
func Execute() error {
	// Initialise CLI runtime with services
	if err := cli.Init(cli.Options{
		AppName: appName,
		Version: appVersion,
		Services: []framework.Option{
			framework.WithName("i18n", cli.NewI18nService(cli.I18nOptions{})),
			framework.WithName("log", cli.NewLogService(cli.LogOptions{
				Level: cli.LogLevelInfo,
			})),
		},
	}); err != nil {
		return err
	}
	defer cli.Shutdown()

	return rootCmd.Execute()
}

func init() {
	// Add shell completion command
	rootCmd.AddCommand(completionCmd)
}

// completionCmd generates shell completion scripts.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for the specified shell.

To load completions:

Bash:
  $ source <(core completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ core completion bash > /etc/bash_completion.d/core
  # macOS:
  $ core completion bash > $(brew --prefix)/etc/bash_completion.d/core

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ core completion zsh > "${fpath[1]}/_core"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ core completion fish | source

  # To load completions for each session, execute once:
  $ core completion fish > ~/.config/fish/completions/core.fish

PowerShell:
  PS> core completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> core completion powershell > core.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			_ = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			_ = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			_ = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			_ = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}
