package dev

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/devops"
	"github.com/host-uk/core/pkg/i18n"
)

// addVMCommands adds the dev environment VM commands to the dev parent command.
// These are added as direct subcommands: core dev install, core dev boot, etc.
func addVMCommands(parent *cli.Command) {
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
func addVMInstallCommand(parent *cli.Command) {
	installCmd := &cli.Command{
		Use:   "install",
		Short: i18n.T("cmd.dev.vm.install.short"),
		Long:  i18n.T("cmd.dev.vm.install.long"),
		RunE: func(cmd *cli.Command, args []string) error {
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
		cli.Text(successStyle.Render(i18n.T("cmd.dev.vm.already_installed")))
		cli.Line("")
		cli.Text(i18n.T("cmd.dev.vm.check_updates", map[string]interface{}{"Command": dimStyle.Render("core dev update")}))
		return nil
	}

	cli.Print("%s %s\n", dimStyle.Render(i18n.Label("image")), devops.ImageName())
	cli.Line("")
	cli.Text(i18n.T("cmd.dev.vm.downloading"))
	cli.Line("")

	ctx := context.Background()
	start := time.Now()
	var lastProgress int64

	err = d.Install(ctx, func(downloaded, total int64) {
		if total > 0 {
			pct := int(float64(downloaded) / float64(total) * 100)
			if pct != int(float64(lastProgress)/float64(total)*100) {
				cli.Print("\r%s %d%%", dimStyle.Render(i18n.T("cmd.dev.vm.progress_label")), pct)
				lastProgress = downloaded
			}
		}
	})

	cli.Line("") // Clear progress line

	if err != nil {
		return cli.Wrap(err, "install failed")
	}

	elapsed := time.Since(start).Round(time.Second)
	cli.Line("")
	cli.Text(i18n.T("cmd.dev.vm.installed_in", map[string]interface{}{"Duration": elapsed}))
	cli.Line("")
	cli.Text(i18n.T("cmd.dev.vm.start_with", map[string]interface{}{"Command": dimStyle.Render("core dev boot")}))

	return nil
}

// VM boot command flags
var (
	vmBootMemory int
	vmBootCPUs   int
	vmBootFresh  bool
)

// addVMBootCommand adds the 'devops boot' command.
func addVMBootCommand(parent *cli.Command) {
	bootCmd := &cli.Command{
		Use:   "boot",
		Short: i18n.T("cmd.dev.vm.boot.short"),
		Long:  i18n.T("cmd.dev.vm.boot.long"),
		RunE: func(cmd *cli.Command, args []string) error {
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
		return errors.New(i18n.T("cmd.dev.vm.not_installed"))
	}

	opts := devops.DefaultBootOptions()
	if memory > 0 {
		opts.Memory = memory
	}
	if cpus > 0 {
		opts.CPUs = cpus
	}
	opts.Fresh = fresh

	cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.config_label")), i18n.T("cmd.dev.vm.config_value", map[string]interface{}{"Memory": opts.Memory, "CPUs": opts.CPUs}))
	cli.Line("")
	cli.Text(i18n.T("cmd.dev.vm.booting"))

	ctx := context.Background()
	if err := d.Boot(ctx, opts); err != nil {
		return err
	}

	cli.Line("")
	cli.Text(successStyle.Render(i18n.T("cmd.dev.vm.running")))
	cli.Line("")
	cli.Text(i18n.T("cmd.dev.vm.connect_with", map[string]interface{}{"Command": dimStyle.Render("core dev shell")}))
	cli.Print("%s %s\n", i18n.T("cmd.dev.vm.ssh_port"), dimStyle.Render("2222"))

	return nil
}

// addVMStopCommand adds the 'devops stop' command.
func addVMStopCommand(parent *cli.Command) {
	stopCmd := &cli.Command{
		Use:   "stop",
		Short: i18n.T("cmd.dev.vm.stop.short"),
		Long:  i18n.T("cmd.dev.vm.stop.long"),
		RunE: func(cmd *cli.Command, args []string) error {
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
		cli.Text(dimStyle.Render(i18n.T("cmd.dev.vm.not_running")))
		return nil
	}

	cli.Text(i18n.T("cmd.dev.vm.stopping"))

	if err := d.Stop(ctx); err != nil {
		return err
	}

	cli.Text(successStyle.Render(i18n.T("common.status.stopped")))
	return nil
}

// addVMStatusCommand adds the 'devops status' command.
func addVMStatusCommand(parent *cli.Command) {
	statusCmd := &cli.Command{
		Use:   "vm-status",
		Short: i18n.T("cmd.dev.vm.status.short"),
		Long:  i18n.T("cmd.dev.vm.status.long"),
		RunE: func(cmd *cli.Command, args []string) error {
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

	cli.Text(headerStyle.Render(i18n.T("cmd.dev.vm.status_title")))
	cli.Line("")

	// Installation status
	if status.Installed {
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.installed_label")), successStyle.Render(i18n.T("cmd.dev.vm.installed_yes")))
		if status.ImageVersion != "" {
			cli.Print("%s %s\n", dimStyle.Render(i18n.Label("version")), status.ImageVersion)
		}
	} else {
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.installed_label")), errorStyle.Render(i18n.T("cmd.dev.vm.installed_no")))
		cli.Line("")
		cli.Text(i18n.T("cmd.dev.vm.install_with", map[string]interface{}{"Command": dimStyle.Render("core dev install")}))
		return nil
	}

	cli.Line("")

	// Running status
	if status.Running {
		cli.Print("%s %s\n", dimStyle.Render(i18n.Label("status")), successStyle.Render(i18n.T("common.status.running")))
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.container_label")), status.ContainerID[:8])
		cli.Print("%s %dMB\n", dimStyle.Render(i18n.T("cmd.dev.vm.memory_label")), status.Memory)
		cli.Print("%s %d\n", dimStyle.Render(i18n.T("cmd.dev.vm.cpus_label")), status.CPUs)
		cli.Print("%s %d\n", dimStyle.Render(i18n.T("cmd.dev.vm.ssh_port")), status.SSHPort)
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.uptime_label")), formatVMUptime(status.Uptime))
	} else {
		cli.Print("%s %s\n", dimStyle.Render(i18n.Label("status")), dimStyle.Render(i18n.T("common.status.stopped")))
		cli.Line("")
		cli.Text(i18n.T("cmd.dev.vm.start_with", map[string]interface{}{"Command": dimStyle.Render("core dev boot")}))
	}

	return nil
}

func formatVMUptime(d time.Duration) string {
	if d < time.Minute {
		return cli.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return cli.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return cli.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	return cli.Sprintf("%dd %dh", int(d.Hours()/24), int(d.Hours())%24)
}

// VM shell command flags
var vmShellConsole bool

// addVMShellCommand adds the 'devops shell' command.
func addVMShellCommand(parent *cli.Command) {
	shellCmd := &cli.Command{
		Use:   "shell [-- command...]",
		Short: i18n.T("cmd.dev.vm.shell.short"),
		Long:  i18n.T("cmd.dev.vm.shell.long"),
		RunE: func(cmd *cli.Command, args []string) error {
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
func addVMServeCommand(parent *cli.Command) {
	serveCmd := &cli.Command{
		Use:   "serve",
		Short: i18n.T("cmd.dev.vm.serve.short"),
		Long:  i18n.T("cmd.dev.vm.serve.long"),
		RunE: func(cmd *cli.Command, args []string) error {
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
func addVMTestCommand(parent *cli.Command) {
	testCmd := &cli.Command{
		Use:   "test [-- command...]",
		Short: i18n.T("cmd.dev.vm.test.short"),
		Long:  i18n.T("cmd.dev.vm.test.long"),
		RunE: func(cmd *cli.Command, args []string) error {
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
func addVMClaudeCommand(parent *cli.Command) {
	claudeCmd := &cli.Command{
		Use:   "claude",
		Short: i18n.T("cmd.dev.vm.claude.short"),
		Long:  i18n.T("cmd.dev.vm.claude.long"),
		RunE: func(cmd *cli.Command, args []string) error {
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
func addVMUpdateCommand(parent *cli.Command) {
	updateCmd := &cli.Command{
		Use:   "update",
		Short: i18n.T("cmd.dev.vm.update.short"),
		Long:  i18n.T("cmd.dev.vm.update.long"),
		RunE: func(cmd *cli.Command, args []string) error {
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

	cli.Text(i18n.T("common.progress.checking_updates"))
	cli.Line("")

	current, latest, hasUpdate, err := d.CheckUpdate(ctx)
	if err != nil {
		return cli.Wrap(err, "failed to check for updates")
	}

	cli.Print("%s %s\n", dimStyle.Render(i18n.Label("current")), valueStyle.Render(current))
	cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.vm.latest_label")), valueStyle.Render(latest))
	cli.Line("")

	if !hasUpdate {
		cli.Text(successStyle.Render(i18n.T("cmd.dev.vm.up_to_date")))
		return nil
	}

	cli.Text(warningStyle.Render(i18n.T("cmd.dev.vm.update_available")))
	cli.Line("")

	if !apply {
		cli.Text(i18n.T("cmd.dev.vm.run_to_update", map[string]interface{}{"Command": dimStyle.Render("core dev update --apply")}))
		return nil
	}

	// Stop if running
	running, _ := d.IsRunning(ctx)
	if running {
		cli.Text(i18n.T("cmd.dev.vm.stopping_current"))
		_ = d.Stop(ctx)
	}

	cli.Text(i18n.T("cmd.dev.vm.downloading_update"))
	cli.Line("")

	start := time.Now()
	err = d.Install(ctx, func(downloaded, total int64) {
		if total > 0 {
			pct := int(float64(downloaded) / float64(total) * 100)
			cli.Print("\r%s %d%%", dimStyle.Render(i18n.T("cmd.dev.vm.progress_label")), pct)
		}
	})

	cli.Line("")

	if err != nil {
		return cli.Wrap(err, "update failed")
	}

	elapsed := time.Since(start).Round(time.Second)
	cli.Line("")
	cli.Text(i18n.T("cmd.dev.vm.updated_in", map[string]interface{}{"Duration": elapsed}))

	return nil
}
