package vm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/host-uk/core/pkg/container"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

var (
	runName         string
	runDetach       bool
	runMemory       int
	runCPUs         int
	runSSHPort      int
	runTemplateName string
	runVarFlags     []string
)

// addVMRunCommand adds the 'run' command under vm.
func addVMRunCommand(parent *cobra.Command) {
	runCmd := &cobra.Command{
		Use:   "run [image]",
		Short: i18n.T("cmd.vm.run.short"),
		Long:  i18n.T("cmd.vm.run.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := container.RunOptions{
				Name:    runName,
				Detach:  runDetach,
				Memory:  runMemory,
				CPUs:    runCPUs,
				SSHPort: runSSHPort,
			}

			// If template is specified, build and run from template
			if runTemplateName != "" {
				vars := ParseVarFlags(runVarFlags)
				return RunFromTemplate(runTemplateName, vars, opts)
			}

			// Otherwise, require an image path
			if len(args) == 0 {
				return errors.New(i18n.T("cmd.vm.run.error.image_required"))
			}
			image := args[0]

			return runContainer(image, runName, runDetach, runMemory, runCPUs, runSSHPort)
		},
	}

	runCmd.Flags().StringVar(&runName, "name", "", i18n.T("cmd.vm.run.flag.name"))
	runCmd.Flags().BoolVarP(&runDetach, "detach", "d", false, i18n.T("cmd.vm.run.flag.detach"))
	runCmd.Flags().IntVar(&runMemory, "memory", 0, i18n.T("cmd.vm.run.flag.memory"))
	runCmd.Flags().IntVar(&runCPUs, "cpus", 0, i18n.T("cmd.vm.run.flag.cpus"))
	runCmd.Flags().IntVar(&runSSHPort, "ssh-port", 0, i18n.T("cmd.vm.run.flag.ssh_port"))
	runCmd.Flags().StringVar(&runTemplateName, "template", "", i18n.T("cmd.vm.run.flag.template"))
	runCmd.Flags().StringArrayVar(&runVarFlags, "var", nil, i18n.T("cmd.vm.run.flag.var"))

	parent.AddCommand(runCmd)
}

func runContainer(image, name string, detach bool, memory, cpus, sshPort int) error {
	manager, err := container.NewLinuxKitManager()
	if err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "initialize container manager"})+": %w", err)
	}

	opts := container.RunOptions{
		Name:    name,
		Detach:  detach,
		Memory:  memory,
		CPUs:    cpus,
		SSHPort: sshPort,
	}

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.image")), image)
	if name != "" {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.vm.label.name")), name)
	}
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.vm.label.hypervisor")), manager.Hypervisor().Name())
	fmt.Println()

	ctx := context.Background()
	c, err := manager.Run(ctx, image, opts)
	if err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "run container"})+": %w", err)
	}

	if detach {
		fmt.Printf("%s %s\n", successStyle.Render(i18n.T("common.label.started")), c.ID)
		fmt.Printf("%s %d\n", dimStyle.Render(i18n.T("cmd.vm.label.pid")), c.PID)
		fmt.Println()
		fmt.Println(i18n.T("cmd.vm.hint.view_logs", map[string]interface{}{"ID": c.ID[:8]}))
		fmt.Println(i18n.T("cmd.vm.hint.stop", map[string]interface{}{"ID": c.ID[:8]}))
	} else {
		fmt.Printf("\n%s %s\n", dimStyle.Render(i18n.T("cmd.vm.label.container_stopped")), c.ID)
	}

	return nil
}

var psAll bool

// addVMPsCommand adds the 'ps' command under vm.
func addVMPsCommand(parent *cobra.Command) {
	psCmd := &cobra.Command{
		Use:   "ps",
		Short: i18n.T("cmd.vm.ps.short"),
		Long:  i18n.T("cmd.vm.ps.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listContainers(psAll)
		},
	}

	psCmd.Flags().BoolVarP(&psAll, "all", "a", false, i18n.T("cmd.vm.ps.flag.all"))

	parent.AddCommand(psCmd)
}

func listContainers(all bool) error {
	manager, err := container.NewLinuxKitManager()
	if err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "initialize container manager"})+": %w", err)
	}

	ctx := context.Background()
	containers, err := manager.List(ctx)
	if err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "list containers"})+": %w", err)
	}

	// Filter if not showing all
	if !all {
		filtered := make([]*container.Container, 0)
		for _, c := range containers {
			if c.Status == container.StatusRunning {
				filtered = append(filtered, c)
			}
		}
		containers = filtered
	}

	if len(containers) == 0 {
		if all {
			fmt.Println(i18n.T("cmd.vm.ps.no_containers"))
		} else {
			fmt.Println(i18n.T("cmd.vm.ps.no_running"))
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, i18n.T("cmd.vm.ps.header"))
	fmt.Fprintln(w, "--\t----\t-----\t------\t-------\t---")

	for _, c := range containers {
		// Shorten image path
		imageName := c.Image
		if len(imageName) > 30 {
			imageName = "..." + imageName[len(imageName)-27:]
		}

		// Format duration
		duration := formatDuration(time.Since(c.StartedAt))

		// Status with color
		status := string(c.Status)
		switch c.Status {
		case container.StatusRunning:
			status = successStyle.Render(status)
		case container.StatusStopped:
			status = dimStyle.Render(status)
		case container.StatusError:
			status = errorStyle.Render(status)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n",
			c.ID[:8], c.Name, imageName, status, duration, c.PID)
	}

	w.Flush()
	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

// addVMStopCommand adds the 'stop' command under vm.
func addVMStopCommand(parent *cobra.Command) {
	stopCmd := &cobra.Command{
		Use:   "stop <container-id>",
		Short: i18n.T("cmd.vm.stop.short"),
		Long:  i18n.T("cmd.vm.stop.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New(i18n.T("cmd.vm.error.id_required"))
			}
			return stopContainer(args[0])
		},
	}

	parent.AddCommand(stopCmd)
}

func stopContainer(id string) error {
	manager, err := container.NewLinuxKitManager()
	if err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "initialize container manager"})+": %w", err)
	}

	// Support partial ID matching
	fullID, err := resolveContainerID(manager, id)
	if err != nil {
		return err
	}

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.vm.stop.stopping")), fullID[:8])

	ctx := context.Background()
	if err := manager.Stop(ctx, fullID); err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "stop container"})+": %w", err)
	}

	fmt.Printf("%s\n", successStyle.Render(i18n.T("common.status.stopped")))
	return nil
}

// resolveContainerID resolves a partial ID to a full ID.
func resolveContainerID(manager *container.LinuxKitManager, partialID string) (string, error) {
	ctx := context.Background()
	containers, err := manager.List(ctx)
	if err != nil {
		return "", err
	}

	var matches []*container.Container
	for _, c := range containers {
		if strings.HasPrefix(c.ID, partialID) || strings.HasPrefix(c.Name, partialID) {
			matches = append(matches, c)
		}
	}

	switch len(matches) {
	case 0:
		return "", errors.New(i18n.T("cmd.vm.error.no_match", map[string]interface{}{"ID": partialID}))
	case 1:
		return matches[0].ID, nil
	default:
		return "", errors.New(i18n.T("cmd.vm.error.multiple_match", map[string]interface{}{"ID": partialID}))
	}
}

var logsFollow bool

// addVMLogsCommand adds the 'logs' command under vm.
func addVMLogsCommand(parent *cobra.Command) {
	logsCmd := &cobra.Command{
		Use:   "logs <container-id>",
		Short: i18n.T("cmd.vm.logs.short"),
		Long:  i18n.T("cmd.vm.logs.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New(i18n.T("cmd.vm.error.id_required"))
			}
			return viewLogs(args[0], logsFollow)
		},
	}

	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, i18n.T("common.flag.follow"))

	parent.AddCommand(logsCmd)
}

func viewLogs(id string, follow bool) error {
	manager, err := container.NewLinuxKitManager()
	if err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "initialize container manager"})+": %w", err)
	}

	fullID, err := resolveContainerID(manager, id)
	if err != nil {
		return err
	}

	ctx := context.Background()
	reader, err := manager.Logs(ctx, fullID, follow)
	if err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "get logs"})+": %w", err)
	}
	defer reader.Close()

	_, err = io.Copy(os.Stdout, reader)
	return err
}

// addVMExecCommand adds the 'exec' command under vm.
func addVMExecCommand(parent *cobra.Command) {
	execCmd := &cobra.Command{
		Use:   "exec <container-id> <command> [args...]",
		Short: i18n.T("cmd.vm.exec.short"),
		Long:  i18n.T("cmd.vm.exec.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New(i18n.T("cmd.vm.error.id_and_cmd_required"))
			}
			return execInContainer(args[0], args[1:])
		},
	}

	parent.AddCommand(execCmd)
}

func execInContainer(id string, cmd []string) error {
	manager, err := container.NewLinuxKitManager()
	if err != nil {
		return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "initialize container manager"})+": %w", err)
	}

	fullID, err := resolveContainerID(manager, id)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return manager.Exec(ctx, fullID, cmd)
}
