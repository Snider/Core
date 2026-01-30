package dev

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/host-uk/core/pkg/devops"
	"github.com/host-uk/core/pkg/i18n"
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
		Short: i18n.T("cmd.dev.vm.install.short"),
		Long:  i18n.T("cmd.dev.vm.install.long"),
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
		fmt.Println(successStyle.Render(i18n.T("cmd.dev.vm.already_installed")))
		fmt.Println()
		fmt.Println(i18n.T("cmd.dev.vm.check_updates", map[string]interface{}{"Command": dimStyle.Render("core dev update")}))
		return nil
	}

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.image")), devops.ImageName())
	fmt.Println()
	fmt.Println(i18n.T("cmd.dev.vm.downloading"))
	fmt.Println()

	ctx := context.Background()
	start := time.Now()
	var lastProgress int64

	err = d.Install(ctx, func(downloaded, total int64) {
		if total > 0 {
			pct := int(float64(downloaded) / float64(total) * 100)
			if pct != int(float64(lastProgress)/float64(total)*100) {
				fmt.Printf("\r%s %d%%", dimStyle.Render(i18n.T("cmd.dev.vm.progress_label")), pct)
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
	fmt.Println(i18n.T("cmd.dev.vm.installed_in", map[string]interface{}{"Duration": elapsed}))
	fmt.Println()
	fmt.Println(i18n.T("cmd.dev.vm.start_with", map[string]interface{}{"Command": dimStyle.Render("core dev boot")}))

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
		Short: i18n.T("cmd.dev.vm.boot.short"),
		Long:  i18n.T("cmd.dev.vm.boot.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMBoot(vmBootMemory, vmBootCPUs, vmBootFresh)
		},
	}

	bootCmd.Flags().IntVar(&vmBootMemory, "memory", 0, i18n.T("cmd.dev.vm.boot.flag.memory"))
	bootCmd.Flags().IntVar(&vmBootCPUs, "cpus", 0, i18n.T("cmd.dev.vm.boot.flag.cpus"))
	bootCmd.Flags().BoolVar(&vmBootFresh, "fresh", false, i18n.T("cmd.dev.vm.boot.flag.fresh"))

	parent.AddCommand(bootCmd)
}

func runVMBoot(memory, cpus int, fresh bool) error {
	d, err := devops.New()
	if err != nil {
		return err
	}

	if !d.IsInstalled() {
		return fmt.Errorf(i18n.T("cmd.dev.vm.not_installed"))
	}

	opts := devops.DefaultBootOptions()
	if memory > 0 {
		opts.Memory = memory
	}
	if cpus > 0 {
		opts.CPUs = cpus
	}
	opts.Fresh = fresh

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.config_label")), i18n.T("cmd.dev.vm.config_value", map[string]interface{}{"Memory": opts.Memory, "CPUs": opts.CPUs}))
	fmt.Println()
	fmt.Println(i18n.T("cmd.dev.vm.booting"))

	ctx := context.Background()
	if err := d.Boot(ctx, opts); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(successStyle.Render(i18n.T("cmd.dev.vm.running")))
	fmt.Println()
	fmt.Println(i18n.T("cmd.dev.vm.connect_with", map[string]interface{}{"Command": dimStyle.Render("core dev shell")}))
	fmt.Printf("%s %s\n", i18n.T("cmd.dev.vm.ssh_port"), dimStyle.Render("2222"))

	return nil
}

// addVMStopCommand adds the 'devops stop' command.
func addVMStopCommand(parent *cobra.Command) {
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: i18n.T("cmd.dev.vm.stop.short"),
		Long:  i18n.T("cmd.dev.vm.stop.long"),
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
		fmt.Println(dimStyle.Render(i18n.T("cmd.dev.vm.not_running")))
		return nil
	}

	fmt.Println(i18n.T("cmd.dev.vm.stopping"))

	if err := d.Stop(ctx); err != nil {
		return err
	}

	fmt.Println(successStyle.Render(i18n.T("common.status.stopped")))
	return nil
}

// addVMStatusCommand adds the 'devops status' command.
func addVMStatusCommand(parent *cobra.Command) {
	statusCmd := &cobra.Command{
		Use:   "vm-status",
		Short: i18n.T("cmd.dev.vm.status.short"),
		Long:  i18n.T("cmd.dev.vm.status.long"),
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

	fmt.Println(headerStyle.Render(i18n.T("cmd.dev.vm.status_title")))
	fmt.Println()

	// Installation status
	if status.Installed {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.installed_label")), successStyle.Render(i18n.T("cmd.dev.vm.installed_yes")))
		if status.ImageVersion != "" {
			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.version")), status.ImageVersion)
		}
	} else {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.installed_label")), errorStyle.Render(i18n.T("cmd.dev.vm.installed_no")))
		fmt.Println()
		fmt.Println(i18n.T("cmd.dev.vm.install_with", map[string]interface{}{"Command": dimStyle.Render("core dev install")}))
		return nil
	}

	fmt.Println()

	// Running status
	if status.Running {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.status")), successStyle.Render(i18n.T("common.status.running")))
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.container_label")), status.ContainerID[:8])
		fmt.Printf("%s %dMB\n", dimStyle.Render(i18n.T("cmd.dev.vm.memory_label")), status.Memory)
		fmt.Printf("%s %d\n", dimStyle.Render(i18n.T("cmd.dev.vm.cpus_label")), status.CPUs)
		fmt.Printf("%s %d\n", dimStyle.Render(i18n.T("cmd.dev.vm.ssh_port")), status.SSHPort)
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.uptime_label")), formatVMUptime(status.Uptime))
	} else {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.status")), dimStyle.Render(i18n.T("common.status.stopped")))
		fmt.Println()
		fmt.Println(i18n.T("cmd.dev.vm.start_with", map[string]interface{}{"Command": dimStyle.Render("core dev boot")}))
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
		Short: i18n.T("cmd.dev.vm.shell.short"),
		Long:  i18n.T("cmd.dev.vm.shell.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMShell(vmShellConsole, args)
		},
	}

	shellCmd.Flags().BoolVar(&vmShellConsole, "console", false, i18n.T("cmd.dev.vm.shell.flag.console"))

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
		Short: i18n.T("cmd.dev.vm.serve.short"),
		Long:  i18n.T("cmd.dev.vm.serve.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMServe(vmServePort, vmServePath)
		},
	}

	serveCmd.Flags().IntVarP(&vmServePort, "port", "p", 0, i18n.T("cmd.dev.vm.serve.flag.port"))
	serveCmd.Flags().StringVar(&vmServePath, "path", "", i18n.T("cmd.dev.vm.serve.flag.path"))

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
		Short: i18n.T("cmd.dev.vm.test.short"),
		Long:  i18n.T("cmd.dev.vm.test.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMTest(vmTestName, args)
		},
	}

	testCmd.Flags().StringVarP(&vmTestName, "name", "n", "", i18n.T("cmd.dev.vm.test.flag.name"))

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
		Short: i18n.T("cmd.dev.vm.claude.short"),
		Long:  i18n.T("cmd.dev.vm.claude.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMClaude(vmClaudeNoAuth, vmClaudeModel, vmClaudeAuthFlags)
		},
	}

	claudeCmd.Flags().BoolVar(&vmClaudeNoAuth, "no-auth", false, i18n.T("cmd.dev.vm.claude.flag.no_auth"))
	claudeCmd.Flags().StringVarP(&vmClaudeModel, "model", "m", "", i18n.T("cmd.dev.vm.claude.flag.model"))
	claudeCmd.Flags().StringSliceVar(&vmClaudeAuthFlags, "auth", nil, i18n.T("cmd.dev.vm.claude.flag.auth"))

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
		Short: i18n.T("cmd.dev.vm.update.short"),
		Long:  i18n.T("cmd.dev.vm.update.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVMUpdate(vmUpdateApply)
		},
	}

	updateCmd.Flags().BoolVar(&vmUpdateApply, "apply", false, i18n.T("cmd.dev.vm.update.flag.apply"))

	parent.AddCommand(updateCmd)
}

func runVMUpdate(apply bool) error {
	d, err := devops.New()
	if err != nil {
		return err
	}

	ctx := context.Background()

	fmt.Println(i18n.T("common.progress.checking_updates"))
	fmt.Println()

	current, latest, hasUpdate, err := d.CheckUpdate(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.current")), valueStyle.Render(current))
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.latest_label")), valueStyle.Render(latest))
	fmt.Println()

	if !hasUpdate {
		fmt.Println(successStyle.Render(i18n.T("cmd.dev.vm.up_to_date")))
		return nil
	}

	fmt.Println(warningStyle.Render(i18n.T("cmd.dev.vm.update_available")))
	fmt.Println()

	if !apply {
		fmt.Println(i18n.T("cmd.dev.vm.run_to_update", map[string]interface{}{"Command": dimStyle.Render("core dev update --apply")}))
		return nil
	}

	// Stop if running
	running, _ := d.IsRunning(ctx)
	if running {
		fmt.Println(i18n.T("cmd.dev.vm.stopping_current"))
		_ = d.Stop(ctx)
	}

	fmt.Println(i18n.T("cmd.dev.vm.downloading_update"))
	fmt.Println()

	start := time.Now()
	err = d.Install(ctx, func(downloaded, total int64) {
		if total > 0 {
			pct := int(float64(downloaded) / float64(total) * 100)
			fmt.Printf("\r%s %d%%", dimStyle.Render(i18n.T("cmd.dev.vm.progress_label")), pct)
		}
	})

	fmt.Println()

	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	elapsed := time.Since(start).Round(time.Second)
	fmt.Println()
	fmt.Println(i18n.T("cmd.dev.vm.updated_in", map[string]interface{}{"Duration": elapsed}))

	return nil
}
