package dev

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/host-uk/core/pkg/devops"
	"github.com/leaanthony/clir"
)

// addVMCommands adds the dev environment VM commands to the dev parent command.
// These are added as direct subcommands: core dev install, core dev boot, etc.
func addVMCommands(parent *clir.Command) {
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
func addVMInstallCommand(parent *clir.Command) {
	installCmd := parent.NewSubCommand("install", "Download and install the dev environment image")
	installCmd.LongDescription("Downloads the platform-specific dev environment image.\n\n" +
		"The image includes Go, PHP, Node.js, Python, Docker, and Claude CLI.\n" +
		"Downloads are cached at ~/.core/images/\n\n" +
		"Examples:\n" +
		"  core dev install")

	installCmd.Action(func() error {
		return runVMInstall()
	})
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

// addVMBootCommand adds the 'devops boot' command.
func addVMBootCommand(parent *clir.Command) {
	var memory int
	var cpus int
	var fresh bool

	bootCmd := parent.NewSubCommand("boot", "Start the dev environment")
	bootCmd.LongDescription("Boots the dev environment VM.\n\n" +
		"Examples:\n" +
		"  core dev boot\n" +
		"  core dev boot --memory 8192 --cpus 4\n" +
		"  core dev boot --fresh")

	bootCmd.IntFlag("memory", "Memory in MB (default: 4096)", &memory)
	bootCmd.IntFlag("cpus", "Number of CPUs (default: 2)", &cpus)
	bootCmd.BoolFlag("fresh", "Stop existing and start fresh", &fresh)

	bootCmd.Action(func() error {
		return runVMBoot(memory, cpus, fresh)
	})
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
func addVMStopCommand(parent *clir.Command) {
	stopCmd := parent.NewSubCommand("stop", "Stop the dev environment")
	stopCmd.LongDescription("Stops the running dev environment VM.\n\n" +
		"Examples:\n" +
		"  core dev stop")

	stopCmd.Action(func() error {
		return runVMStop()
	})
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
func addVMStatusCommand(parent *clir.Command) {
	statusCmd := parent.NewSubCommand("vm-status", "Show dev environment status")
	statusCmd.LongDescription("Shows the current status of the dev environment.\n\n" +
		"Examples:\n" +
		"  core dev vm-status")

	statusCmd.Action(func() error {
		return runVMStatus()
	})
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

// addVMShellCommand adds the 'devops shell' command.
func addVMShellCommand(parent *clir.Command) {
	var console bool

	shellCmd := parent.NewSubCommand("shell", "Connect to the dev environment")
	shellCmd.LongDescription("Opens an interactive shell in the dev environment.\n\n" +
		"Uses SSH by default, or serial console with --console.\n\n" +
		"Examples:\n" +
		"  core dev shell\n" +
		"  core dev shell --console\n" +
		"  core dev shell -- ls -la")

	shellCmd.BoolFlag("console", "Use serial console instead of SSH", &console)

	shellCmd.Action(func() error {
		args := shellCmd.OtherArgs()
		return runVMShell(console, args)
	})
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

// addVMServeCommand adds the 'devops serve' command.
func addVMServeCommand(parent *clir.Command) {
	var port int
	var path string

	serveCmd := parent.NewSubCommand("serve", "Mount project and start dev server")
	serveCmd.LongDescription("Mounts the current project into the dev environment and starts a dev server.\n\n" +
		"Auto-detects the appropriate serve command based on project files.\n\n" +
		"Examples:\n" +
		"  core dev serve\n" +
		"  core dev serve --port 3000\n" +
		"  core dev serve --path public")

	serveCmd.IntFlag("port", "Port to serve on (default: 8000)", &port)
	serveCmd.StringFlag("path", "Subdirectory to serve", &path)

	serveCmd.Action(func() error {
		return runVMServe(port, path)
	})
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

// addVMTestCommand adds the 'devops test' command.
func addVMTestCommand(parent *clir.Command) {
	var name string

	testCmd := parent.NewSubCommand("test", "Run tests in the dev environment")
	testCmd.LongDescription("Runs tests in the dev environment.\n\n" +
		"Auto-detects the test command based on project files, or uses .core/test.yaml.\n\n" +
		"Examples:\n" +
		"  core dev test\n" +
		"  core dev test --name integration\n" +
		"  core dev test -- go test -v ./...")

	testCmd.StringFlag("name", "Run named test command from .core/test.yaml", &name)

	testCmd.Action(func() error {
		args := testCmd.OtherArgs()
		return runVMTest(name, args)
	})
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

// addVMClaudeCommand adds the 'devops claude' command.
func addVMClaudeCommand(parent *clir.Command) {
	var noAuth bool
	var model string
	var authFlags []string

	claudeCmd := parent.NewSubCommand("claude", "Start sandboxed Claude session")
	claudeCmd.LongDescription("Starts a Claude Code session inside the dev environment sandbox.\n\n" +
		"Provides isolation while forwarding selected credentials.\n" +
		"Auto-boots the dev environment if not running.\n\n" +
		"Auth options (default: all):\n" +
		"  gh        - GitHub CLI auth\n" +
		"  anthropic - Anthropic API key\n" +
		"  ssh       - SSH agent forwarding\n" +
		"  git       - Git config (name, email)\n\n" +
		"Examples:\n" +
		"  core dev claude\n" +
		"  core dev claude --model opus\n" +
		"  core dev claude --auth gh,anthropic\n" +
		"  core dev claude --no-auth")

	claudeCmd.BoolFlag("no-auth", "Don't forward any auth credentials", &noAuth)
	claudeCmd.StringFlag("model", "Model to use (opus, sonnet)", &model)
	claudeCmd.StringsFlag("auth", "Selective auth forwarding (gh,anthropic,ssh,git)", &authFlags)

	claudeCmd.Action(func() error {
		return runVMClaude(noAuth, model, authFlags)
	})
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

// addVMUpdateCommand adds the 'devops update' command.
func addVMUpdateCommand(parent *clir.Command) {
	var apply bool

	updateCmd := parent.NewSubCommand("update", "Check for and apply updates")
	updateCmd.LongDescription("Checks for dev environment updates and optionally applies them.\n\n" +
		"Examples:\n" +
		"  core dev update\n" +
		"  core dev update --apply")

	updateCmd.BoolFlag("apply", "Download and apply the update", &apply)

	updateCmd.Action(func() error {
		return runVMUpdate(apply)
	})
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
