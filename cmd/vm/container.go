package vm

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/host-uk/core/pkg/container"
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
		Short: "Run a LinuxKit image or template",
		Long: "Runs a LinuxKit image as a VM using the available hypervisor.\n\n" +
			"Supported image formats: .iso, .qcow2, .vmdk, .raw\n\n" +
			"You can also run from a template using --template, which will build and run\n" +
			"the image automatically. Use --var to set template variables.\n\n" +
			"Examples:\n" +
			"  core vm run image.iso\n" +
			"  core vm run -d image.qcow2\n" +
			"  core vm run --name myvm --memory 2048 --cpus 4 image.iso\n" +
			"  core vm run --template core-dev --var SSH_KEY=\"ssh-rsa AAAA...\"\n" +
			"  core vm run --template server-php --var SSH_KEY=\"...\" --var DOMAIN=example.com",
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
				return fmt.Errorf("image path is required (or use --template)")
			}
			image := args[0]

			return runContainer(image, runName, runDetach, runMemory, runCPUs, runSSHPort)
		},
	}

	runCmd.Flags().StringVar(&runName, "name", "", "Name for the container")
	runCmd.Flags().BoolVarP(&runDetach, "detach", "d", false, "Run in detached mode (background)")
	runCmd.Flags().IntVar(&runMemory, "memory", 0, "Memory in MB (default: 1024)")
	runCmd.Flags().IntVar(&runCPUs, "cpus", 0, "Number of CPUs (default: 1)")
	runCmd.Flags().IntVar(&runSSHPort, "ssh-port", 0, "SSH port for exec commands (default: 2222)")
	runCmd.Flags().StringVar(&runTemplateName, "template", "", "Run from a LinuxKit template (build + run)")
	runCmd.Flags().StringArrayVar(&runVarFlags, "var", nil, "Template variable in KEY=VALUE format (can be repeated)")

	parent.AddCommand(runCmd)
}

func runContainer(image, name string, detach bool, memory, cpus, sshPort int) error {
	manager, err := container.NewLinuxKitManager()
	if err != nil {
		return fmt.Errorf("failed to initialize container manager: %w", err)
	}

	opts := container.RunOptions{
		Name:    name,
		Detach:  detach,
		Memory:  memory,
		CPUs:    cpus,
		SSHPort: sshPort,
	}

	fmt.Printf("%s %s\n", dimStyle.Render("Image:"), image)
	if name != "" {
		fmt.Printf("%s %s\n", dimStyle.Render("Name:"), name)
	}
	fmt.Printf("%s %s\n", dimStyle.Render("Hypervisor:"), manager.Hypervisor().Name())
	fmt.Println()

	ctx := context.Background()
	c, err := manager.Run(ctx, image, opts)
	if err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}

	if detach {
		fmt.Printf("%s %s\n", successStyle.Render("Started:"), c.ID)
		fmt.Printf("%s %d\n", dimStyle.Render("PID:"), c.PID)
		fmt.Println()
		fmt.Printf("Use 'core vm logs %s' to view output\n", c.ID[:8])
		fmt.Printf("Use 'core vm stop %s' to stop\n", c.ID[:8])
	} else {
		fmt.Printf("\n%s %s\n", dimStyle.Render("Container stopped:"), c.ID)
	}

	return nil
}

var psAll bool

// addVMPsCommand adds the 'ps' command under vm.
func addVMPsCommand(parent *cobra.Command) {
	psCmd := &cobra.Command{
		Use:   "ps",
		Short: "List running VMs",
		Long: "Lists all VMs. By default, only shows running VMs.\n\n" +
			"Examples:\n" +
			"  core vm ps\n" +
			"  core vm ps -a",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listContainers(psAll)
		},
	}

	psCmd.Flags().BoolVarP(&psAll, "all", "a", false, "Show all containers (including stopped)")

	parent.AddCommand(psCmd)
}

func listContainers(all bool) error {
	manager, err := container.NewLinuxKitManager()
	if err != nil {
		return fmt.Errorf("failed to initialize container manager: %w", err)
	}

	ctx := context.Background()
	containers, err := manager.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
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
			fmt.Println("No containers")
		} else {
			fmt.Println("No running containers")
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tIMAGE\tSTATUS\tSTARTED\tPID")
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
		Short: "Stop a running VM",
		Long: "Stops a running VM by ID.\n\n" +
			"Examples:\n" +
			"  core vm stop abc12345\n" +
			"  core vm stop abc1",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("container ID is required")
			}
			return stopContainer(args[0])
		},
	}

	parent.AddCommand(stopCmd)
}

func stopContainer(id string) error {
	manager, err := container.NewLinuxKitManager()
	if err != nil {
		return fmt.Errorf("failed to initialize container manager: %w", err)
	}

	// Support partial ID matching
	fullID, err := resolveContainerID(manager, id)
	if err != nil {
		return err
	}

	fmt.Printf("%s %s\n", dimStyle.Render("Stopping:"), fullID[:8])

	ctx := context.Background()
	if err := manager.Stop(ctx, fullID); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	fmt.Printf("%s\n", successStyle.Render("Stopped"))
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
		return "", fmt.Errorf("no container found matching: %s", partialID)
	case 1:
		return matches[0].ID, nil
	default:
		return "", fmt.Errorf("multiple containers match '%s', be more specific", partialID)
	}
}

var logsFollow bool

// addVMLogsCommand adds the 'logs' command under vm.
func addVMLogsCommand(parent *cobra.Command) {
	logsCmd := &cobra.Command{
		Use:   "logs <container-id>",
		Short: "View VM logs",
		Long: "View logs from a VM.\n\n" +
			"Examples:\n" +
			"  core vm logs abc12345\n" +
			"  core vm logs -f abc1",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("container ID is required")
			}
			return viewLogs(args[0], logsFollow)
		},
	}

	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")

	parent.AddCommand(logsCmd)
}

func viewLogs(id string, follow bool) error {
	manager, err := container.NewLinuxKitManager()
	if err != nil {
		return fmt.Errorf("failed to initialize container manager: %w", err)
	}

	fullID, err := resolveContainerID(manager, id)
	if err != nil {
		return err
	}

	ctx := context.Background()
	reader, err := manager.Logs(ctx, fullID, follow)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}
	defer reader.Close()

	_, err = io.Copy(os.Stdout, reader)
	return err
}

// addVMExecCommand adds the 'exec' command under vm.
func addVMExecCommand(parent *cobra.Command) {
	execCmd := &cobra.Command{
		Use:   "exec <container-id> <command> [args...]",
		Short: "Execute a command in a VM",
		Long: "Execute a command inside a running VM via SSH.\n\n" +
			"Examples:\n" +
			"  core vm exec abc12345 ls -la\n" +
			"  core vm exec abc1 /bin/sh",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("container ID and command are required")
			}
			return execInContainer(args[0], args[1:])
		},
	}

	parent.AddCommand(execCmd)
}

func execInContainer(id string, cmd []string) error {
	manager, err := container.NewLinuxKitManager()
	if err != nil {
		return fmt.Errorf("failed to initialize container manager: %w", err)
	}

	fullID, err := resolveContainerID(manager, id)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return manager.Exec(ctx, fullID, cmd)
}
