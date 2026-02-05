package cli

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/host-uk/core/pkg/crypt/openpgp"
	"github.com/host-uk/core/pkg/framework"
	"github.com/host-uk/core/pkg/log"
	"github.com/host-uk/core/pkg/workspace"
	"github.com/spf13/cobra"
)

const (
	// AppName is the CLI application name.
	AppName = "core"
)

// AppVersion is set at build time via ldflags:
//
//	go build -ldflags="-X github.com/host-uk/core/pkg/cli.AppVersion=v1.0.0"
var AppVersion = "dev"

// Main initialises and runs the CLI application.
// This is the main entry point for the CLI.
// Exits with code 1 on error or panic.
func Main() {
	// Recovery from panics
	defer func() {
		if r := recover(); r != nil {
			log.Error("recovered from panic", "error", r, "stack", string(debug.Stack()))
			Shutdown()
			Fatal(fmt.Errorf("panic: %v", r))
		}
	}()

	// Initialise CLI runtime with services
	if err := Init(Options{
		AppName: AppName,
		Version: AppVersion,
		Services: []framework.Option{
			framework.WithName("i18n", NewI18nService(I18nOptions{})),
			framework.WithName("log", NewLogService(log.Options{
				Level: log.LevelInfo,
			})),
			framework.WithName("crypt", openpgp.New),
			framework.WithName("workspace", workspace.New),
		},
	}); err != nil {
		Error(err.Error())
		os.Exit(1)
	}
	defer Shutdown()

	// Add completion command to the CLI's root
	RootCmd().AddCommand(completionCmd)

	if err := Execute(); err != nil {
		code := 1
		var exitErr *ExitError
		if As(err, &exitErr) {
			code = exitErr.Code
		}
		Error(err.Error())
		os.Exit(code)
	}
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
