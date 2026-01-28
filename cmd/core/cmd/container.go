package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/host-uk/core/pkg/container"
	"github.com/leaanthony/clir"
)

// AddContainerCommands adds container-related commands to the CLI.
func AddContainerCommands(parent *clir.Cli) {
	AddRunCommand(parent)
	AddPsCommand(parent)
	AddStopCommand(parent)
	AddLogsCommand(parent)
	AddExecCommand(parent)
}

// AddRunCommand adds the 'run' command.
func AddRunCommand(parent *clir.Cli) {
	var (
		name         string
		detach       bool
		memory       int
		cpus         int
		sshPort      int
		templateName string
		varFlags     []string
	)

	runCmd := parent.NewSubCommand("run", "Run a LinuxKit image or template")
	runCmd.LongDescription("Runs a LinuxKit image as a VM using the available hypervisor.\n\n" +
		"Supported image formats: .iso, .qcow2, .vmdk, .raw\n\n" +
		"You can also run from a template using --template, which will build and run\n" +
		"the image automatically. Use --var to set template variables.\n\n" +
		"Examples:\n" +
		"  core run image.iso\n" +
		"  core run -d image.qcow2\n" +
		"  core run --name myvm --memory 2048 --cpus 4 image.iso\n" +
		"  core run --template core-dev --var SSH_KEY=\"ssh-rsa AAAA...\"\n" +
		"  core run --template server-php --var SSH_KEY=\"...\" --var DOMAIN=example.com")

	runCmd.StringFlag("name", "Name for the container", &name)
	runCmd.BoolFlag("d", "Run in detached mode (background)", &detach)
	runCmd.IntFlag("memory", "Memory in MB (default: 1024)", &memory)
	runCmd.IntFlag("cpus", "Number of CPUs (default: 1)", &cpus)
	runCmd.IntFlag("ssh-port", "SSH port for exec commands (default: 2222)", &sshPort)
	runCmd.StringFlag("template", "Run from a LinuxKit template (build + run)", &templateName)
	runCmd.StringsFlag("var", "Template variable in KEY=VALUE format (can be repeated)", &varFlags)

	runCmd.Action(func() error {
		opts := container.RunOptions{
			Name:    name,
			Detach:  detach,
			Memory:  memory,
			CPUs:    cpus,
			SSHPort: sshPort,
		}

		// If template is specified, build and run from template
		if templateName != "" {
			vars := ParseVarFlags(varFlags)
			return RunFromTemplate(templateName, vars, opts)
		}

		// Otherwise, require an image path
		args := runCmd.OtherArgs()
		if len(args) == 0 {
			return fmt.Errorf("image path is required (or use --template)")
		}
		image := args[0]

		return runContainer(image, name, detach, memory, cpus, sshPort)
	})
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
		fmt.Printf("Use 'core logs %s' to view output\n", c.ID[:8])
		fmt.Printf("Use 'core stop %s' to stop\n", c.ID[:8])
	} else {
		fmt.Printf("\n%s %s\n", dimStyle.Render("Container stopped:"), c.ID)
	}

	return nil
}

// AddPsCommand adds the 'ps' command.
func AddPsCommand(parent *clir.Cli) {
	var all bool

	psCmd := parent.NewSubCommand("ps", "List running containers")
	psCmd.LongDescription("Lists all containers. By default, only shows running containers.\n\n" +
		"Examples:\n" +
		"  core ps\n" +
		"  core ps -a")

	psCmd.BoolFlag("a", "Show all containers (including stopped)", &all)

	psCmd.Action(func() error {
		return listContainers(all)
	})
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

// AddStopCommand adds the 'stop' command.
func AddStopCommand(parent *clir.Cli) {
	stopCmd := parent.NewSubCommand("stop", "Stop a running container")
	stopCmd.LongDescription("Stops a running container by ID.\n\n" +
		"Examples:\n" +
		"  core stop abc12345\n" +
		"  core stop abc1")

	stopCmd.Action(func() error {
		args := stopCmd.OtherArgs()
		if len(args) == 0 {
			return fmt.Errorf("container ID is required")
		}
		return stopContainer(args[0])
	})
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

// AddLogsCommand adds the 'logs' command.
func AddLogsCommand(parent *clir.Cli) {
	var follow bool

	logsCmd := parent.NewSubCommand("logs", "View container logs")
	logsCmd.LongDescription("View logs from a container.\n\n" +
		"Examples:\n" +
		"  core logs abc12345\n" +
		"  core logs -f abc1")

	logsCmd.BoolFlag("f", "Follow log output", &follow)

	logsCmd.Action(func() error {
		args := logsCmd.OtherArgs()
		if len(args) == 0 {
			return fmt.Errorf("container ID is required")
		}
		return viewLogs(args[0], follow)
	})
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

// AddExecCommand adds the 'exec' command.
func AddExecCommand(parent *clir.Cli) {
	execCmd := parent.NewSubCommand("exec", "Execute a command in a container")
	execCmd.LongDescription("Execute a command inside a running container via SSH.\n\n" +
		"Examples:\n" +
		"  core exec abc12345 ls -la\n" +
		"  core exec abc1 /bin/sh")

	execCmd.Action(func() error {
		args := execCmd.OtherArgs()
		if len(args) < 2 {
			return fmt.Errorf("container ID and command are required")
		}
		return execInContainer(args[0], args[1:])
	})
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
