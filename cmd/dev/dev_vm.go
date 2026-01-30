package dev

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/host-uk/core/pkg/devops"
	"github.com/spf13/cobra"
)

// addVMCommands adds the dev environment VM commands to the dev parent command.
// These are added as direct subcommands: core dev install, core dev boot, etc.
func addVMCommands(parent *cobra.Command) {
	addVMInstallCommand(parent)
	addVMBootCommand(parent)
	addVMStopCommand(parent)
	addVMStatusCommand(parent)
	addVMShellCommand(parent)
	addVMServeCommand(parent)
	addVMTestCommand(parent)
	addVMClaudeCommand(parent)
	addVMUpdateCommand(parent)
}

// addVMInstallCommand adds the 'dev install' command.
func addVMInstallCommand(parent *cobra.Command) {
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Download and install the dev environment image",
		Long: `Downloads the platform-specific dev environment image.

The image includes Go, PHP, Node.js, Python, Docker, and Claude CLI.
Downloads are cached at ~/.core/images/

Examples:
  core dev install`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMInstall()
		},
	}

	parent.AddCommand(installCmd)
}

func runVMInstall() error {
	d, err := devops.New()
	if err != nil {
		return err
	}

	if d.IsInstalled() {
		fmt.Println(successStyle.Render("Dev environment already installed"))
		fmt.Println()
		fmt.Printf("Use %s to check for updates\n", dimStyle.Render("core dev update"))
		return nil
	}

	fmt.Printf("%s %s\n", dimStyle.Render("Image:"), devops.ImageName())
	fmt.Println()
	fmt.Println("Downloading dev environment...")
	fmt.Println()

	ctx := context.Background()
	start := time.Now()
	var lastProgress int64

	err = d.Install(ctx, func(downloaded, total int64) {
		if total > 0 {
			pct := int(float64(downloaded) / float64(total) * 100)
			if pct != int(float64(lastProgress)/float64(total)*100) {
				fmt.Printf("\r%s %d%%", dimStyle.Render("Progress:"), pct)
				lastProgress = downloaded
			}
		}
	})

	fmt.Println() // Clear progress line

	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	elapsed := time.Since(start).Round(time.Second)
	fmt.Println()
	fmt.Printf("%s in %s\n", successStyle.Render("Installed"), elapsed)
	fmt.Println()
	fmt.Printf("Start with: %s\n", dimStyle.Render("core dev boot"))

	return nil
}

// VM boot command flags
var (
	vmBootMemory int
	vmBootCPUs   int
	vmBootFresh  bool
)

// addVMBootCommand adds the 'devops boot' command.
func addVMBootCommand(parent *cobra.Command) {
	bootCmd := &cobra.Command{
		Use:   "boot",
		Short: "Start the dev environment",
		Long: `Boots the dev environment VM.

Examples:
  core dev boot
  core dev boot --memory 8192 --cpus 4
  core dev boot --fresh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMBoot(vmBootMemory, vmBootCPUs, vmBootFresh)
		},
	}

	bootCmd.Flags().IntVar(&vmBootMemory, "memory", 0, "Memory in MB (default: 4096)")
	bootCmd.Flags().IntVar(&vmBootCPUs, "cpus", 0, "Number of CPUs (default: 2)")
	bootCmd.Flags().BoolVar(&vmBootFresh, "fresh", false, "Stop existing and start fresh")

	parent.AddCommand(bootCmd)
}

func runVMBoot(memory, cpus int, fresh bool) error {
	d, err := devops.New()
	if err != nil {
		return err
	}

	if !d.IsInstalled() {
		return fmt.Errorf("dev environment not installed (run 'core dev install' first)")
	}

	opts := devops.DefaultBootOptions()
	if memory > 0 {
		opts.Memory = memory
	}
	if cpus > 0 {
		opts.CPUs = cpus
	}
	opts.Fresh = fresh

	fmt.Printf("%s %dMB, %d CPUs\n", dimStyle.Render("Config:"), opts.Memory, opts.CPUs)
	fmt.Println()
	fmt.Println("Booting dev environment...")

	ctx := context.Background()
	if err := d.Boot(ctx, opts); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(successStyle.Render("Dev environment running"))
	fmt.Println()
	fmt.Printf("Connect with: %s\n", dimStyle.Render("core dev shell"))
	fmt.Printf("SSH port:     %s\n", dimStyle.Render("2222"))

	return nil
}

// addVMStopCommand adds the 'devops stop' command.
func addVMStopCommand(parent *cobra.Command) {
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the dev environment",
		Long: `Stops the running dev environment VM.

Examples:
  core dev stop`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMStop()
		},
	}

	parent.AddCommand(stopCmd)
}

func runVMStop() error {
	d, err := devops.New()
	if err != nil {
		return err
	}

	ctx := context.Background()
	running, err := d.IsRunning(ctx)
	if err != nil {
		return err
	}

	if !running {
		fmt.Println(dimStyle.Render("Dev environment is not running"))
		return nil
	}

	fmt.Println("Stopping dev environment...")

	if err := d.Stop(ctx); err != nil {
		return err
	}

	fmt.Println(successStyle.Render("Stopped"))
	return nil
}

// addVMStatusCommand adds the 'devops status' command.
func addVMStatusCommand(parent *cobra.Command) {
	statusCmd := &cobra.Command{
		Use:   "vm-status",
		Short: "Show dev environment status",
		Long: `Shows the current status of the dev environment.

Examples:
  core dev vm-status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMStatus()
		},
	}

	parent.AddCommand(statusCmd)
}

func runVMStatus() error {
	d, err := devops.New()
	if err != nil {
		return err
	}

	ctx := context.Background()
	status, err := d.Status(ctx)
	if err != nil {
		return err
	}

	fmt.Println(headerStyle.Render("Dev Environment Status"))
	fmt.Println()

	// Installation status
	if status.Installed {
		fmt.Printf("%s %s\n", dimStyle.Render("Installed:"), successStyle.Render("Yes"))
		if status.ImageVersion != "" {
			fmt.Printf("%s %s\n", dimStyle.Render("Version:"), status.ImageVersion)
		}
	} else {
		fmt.Printf("%s %s\n", dimStyle.Render("Installed:"), errorStyle.Render("No"))
		fmt.Println()
		fmt.Printf("Install with: %s\n", dimStyle.Render("core dev install"))
		return nil
	}

	fmt.Println()

	// Running status
	if status.Running {
		fmt.Printf("%s %s\n", dimStyle.Render("Status:"), successStyle.Render("Running"))
		fmt.Printf("%s %s\n", dimStyle.Render("Container:"), status.ContainerID[:8])
		fmt.Printf("%s %dMB\n", dimStyle.Render("Memory:"), status.Memory)
		fmt.Printf("%s %d\n", dimStyle.Render("CPUs:"), status.CPUs)
		fmt.Printf("%s %d\n", dimStyle.Render("SSH Port:"), status.SSHPort)
		fmt.Printf("%s %s\n", dimStyle.Render("Uptime:"), formatVMUptime(status.Uptime))
	} else {
		fmt.Printf("%s %s\n", dimStyle.Render("Status:"), dimStyle.Render("Stopped"))
		fmt.Println()
		fmt.Printf("Start with: %s\n", dimStyle.Render("core dev boot"))
	}

	return nil
}

func formatVMUptime(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	return fmt.Sprintf("%dd %dh", int(d.Hours()/24), int(d.Hours())%24)
}

// VM shell command flags
var vmShellConsole bool

// addVMShellCommand adds the 'devops shell' command.
func addVMShellCommand(parent *cobra.Command) {
	shellCmd := &cobra.Command{
		Use:   "shell [-- command...]",
		Short: "Connect to the dev environment",
		Long: `Opens an interactive shell in the dev environment.

Uses SSH by default, or serial console with --console.

Examples:
  core dev shell
  core dev shell --console
  core dev shell -- ls -la`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMShell(vmShellConsole, args)
		},
	}

	shellCmd.Flags().BoolVar(&vmShellConsole, "console", false, "Use serial console instead of SSH")

	parent.AddCommand(shellCmd)
}

func runVMShell(console bool, command []string) error {
	d, err := devops.New()
	if err != nil {
		return err
	}

	opts := devops.ShellOptions{
		Console: console,
		Command: command,
	}

	ctx := context.Background()
	return d.Shell(ctx, opts)
}

// VM serve command flags
var (
	vmServePort int
	vmServePath string
)

// addVMServeCommand adds the 'devops serve' command.
func addVMServeCommand(parent *cobra.Command) {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Mount project and start dev server",
		Long: `Mounts the current project into the dev environment and starts a dev server.

Auto-detects the appropriate serve command based on project files.

Examples:
  core dev serve
  core dev serve --port 3000
  core dev serve --path public`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMServe(vmServePort, vmServePath)
		},
	}

	serveCmd.Flags().IntVarP(&vmServePort, "port", "p", 0, "Port to serve on (default: 8000)")
	serveCmd.Flags().StringVar(&vmServePath, "path", "", "Subdirectory to serve")

	parent.AddCommand(serveCmd)
}

func runVMServe(port int, path string) error {
	d, err := devops.New()
	if err != nil {
		return err
	}

	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	opts := devops.ServeOptions{
		Port: port,
		Path: path,
	}

	ctx := context.Background()
	return d.Serve(ctx, projectDir, opts)
}

// VM test command flags
var vmTestName string

// addVMTestCommand adds the 'devops test' command.
func addVMTestCommand(parent *cobra.Command) {
	testCmd := &cobra.Command{
		Use:   "test [-- command...]",
		Short: "Run tests in the dev environment",
		Long: `Runs tests in the dev environment.

Auto-detects the test command based on project files, or uses .core/test.yaml.

Examples:
  core dev test
  core dev test --name integration
  core dev test -- go test -v ./...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMTest(vmTestName, args)
		},
	}

	testCmd.Flags().StringVarP(&vmTestName, "name", "n", "", "Run named test command from .core/test.yaml")

	parent.AddCommand(testCmd)
}

func runVMTest(name string, command []string) error {
	d, err := devops.New()
	if err != nil {
		return err
	}

	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	opts := devops.TestOptions{
		Name:    name,
		Command: command,
	}

	ctx := context.Background()
	return d.Test(ctx, projectDir, opts)
}

// VM claude command flags
var (
	vmClaudeNoAuth    bool
	vmClaudeModel     string
	vmClaudeAuthFlags []string
)

// addVMClaudeCommand adds the 'devops claude' command.
func addVMClaudeCommand(parent *cobra.Command) {
	claudeCmd := &cobra.Command{
		Use:   "claude",
		Short: "Start sandboxed Claude session",
		Long: `Starts a Claude Code session inside the dev environment sandbox.

Provides isolation while forwarding selected credentials.
Auto-boots the dev environment if not running.

Auth options (default: all):
  gh        - GitHub CLI auth
  anthropic - Anthropic API key
  ssh       - SSH agent forwarding
  git       - Git config (name, email)

Examples:
  core dev claude
  core dev claude --model opus
  core dev claude --auth gh,anthropic
  core dev claude --no-auth`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMClaude(vmClaudeNoAuth, vmClaudeModel, vmClaudeAuthFlags)
		},
	}

	claudeCmd.Flags().BoolVar(&vmClaudeNoAuth, "no-auth", false, "Don't forward any auth credentials")
	claudeCmd.Flags().StringVarP(&vmClaudeModel, "model", "m", "", "Model to use (opus, sonnet)")
	claudeCmd.Flags().StringSliceVar(&vmClaudeAuthFlags, "auth", nil, "Selective auth forwarding (gh,anthropic,ssh,git)")

	parent.AddCommand(claudeCmd)
}

func runVMClaude(noAuth bool, model string, authFlags []string) error {
	d, err := devops.New()
	if err != nil {
		return err
	}

	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	opts := devops.ClaudeOptions{
		NoAuth: noAuth,
		Model:  model,
		Auth:   authFlags,
	}

	ctx := context.Background()
	return d.Claude(ctx, projectDir, opts)
}

// VM update command flags
var vmUpdateApply bool

// addVMUpdateCommand adds the 'devops update' command.
func addVMUpdateCommand(parent *cobra.Command) {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Check for and apply updates",
		Long: `Checks for dev environment updates and optionally applies them.

Examples:
  core dev update
  core dev update --apply`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMUpdate(vmUpdateApply)
		},
	}

	updateCmd.Flags().BoolVar(&vmUpdateApply, "apply", false, "Download and apply the update")

	parent.AddCommand(updateCmd)
}

func runVMUpdate(apply bool) error {
	d, err := devops.New()
	if err != nil {
		return err
	}

	ctx := context.Background()

	fmt.Println("Checking for updates...")
	fmt.Println()

	current, latest, hasUpdate, err := d.CheckUpdate(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	fmt.Printf("%s %s\n", dimStyle.Render("Current:"), valueStyle.Render(current))
	fmt.Printf("%s %s\n", dimStyle.Render("Latest:"), valueStyle.Render(latest))
	fmt.Println()

	if !hasUpdate {
		fmt.Println(successStyle.Render("Already up to date"))
		return nil
	}

	fmt.Println(warningStyle.Render("Update available"))
	fmt.Println()

	if !apply {
		fmt.Printf("Run %s to update\n", dimStyle.Render("core dev update --apply"))
		return nil
	}

	// Stop if running
	running, _ := d.IsRunning(ctx)
	if running {
		fmt.Println("Stopping current instance...")
		_ = d.Stop(ctx)
	}

	fmt.Println("Downloading update...")
	fmt.Println()

	start := time.Now()
	err = d.Install(ctx, func(downloaded, total int64) {
		if total > 0 {
			pct := int(float64(downloaded) / float64(total) * 100)
			fmt.Printf("\r%s %d%%", dimStyle.Render("Progress:"), pct)
		}
	})

	fmt.Println()

	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	elapsed := time.Since(start).Round(time.Second)
	fmt.Println()
	fmt.Printf("%s in %s\n", successStyle.Render("Updated"), elapsed)

	return nil
}
