package deploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/ansible"
	"github.com/host-uk/core/pkg/cli"
	"github.com/spf13/cobra"
)

var (
	ansibleInventory string
	ansibleLimit     string
	ansibleTags      string
	ansibleSkipTags  string
	ansibleVars      []string
	ansibleVerbose   int
	ansibleCheck     bool
)

var ansibleCmd = &cobra.Command{
	Use:   "ansible <playbook>",
	Short: "Run Ansible playbooks natively (no Python required)",
	Long: `Execute Ansible playbooks using a pure Go implementation.

This command parses Ansible YAML playbooks and executes them natively,
without requiring Python or ansible-playbook to be installed.

Supported modules:
  - shell, command, raw, script
  - copy, template, file, lineinfile, stat, slurp, fetch, get_url
  - apt, apt_key, apt_repository, package, pip
  - service, systemd
  - user, group
  - uri, wait_for, git, unarchive
  - debug, fail, assert, set_fact, pause

Examples:
  core deploy ansible playbooks/coolify/create.yml -i inventory/
  core deploy ansible site.yml -l production
  core deploy ansible deploy.yml -e "version=1.2.3" -e "env=prod"`,
	Args: cobra.ExactArgs(1),
	RunE: runAnsible,
}

var ansibleTestCmd = &cobra.Command{
	Use:   "test <host>",
	Short: "Test SSH connectivity to a host",
	Long: `Test SSH connection and gather facts from a host.

Examples:
  core deploy ansible test linux.snider.dev -u claude -p claude
  core deploy ansible test server.example.com -i ~/.ssh/id_rsa`,
	Args: cobra.ExactArgs(1),
	RunE: runAnsibleTest,
}

var (
	testUser     string
	testPassword string
	testKeyFile  string
	testPort     int
)

func init() {
	// ansible command flags
	ansibleCmd.Flags().StringVarP(&ansibleInventory, "inventory", "i", "", "Inventory file or directory")
	ansibleCmd.Flags().StringVarP(&ansibleLimit, "limit", "l", "", "Limit to specific hosts")
	ansibleCmd.Flags().StringVarP(&ansibleTags, "tags", "t", "", "Only run plays and tasks tagged with these values")
	ansibleCmd.Flags().StringVar(&ansibleSkipTags, "skip-tags", "", "Skip plays and tasks tagged with these values")
	ansibleCmd.Flags().StringArrayVarP(&ansibleVars, "extra-vars", "e", nil, "Set additional variables (key=value)")
	ansibleCmd.Flags().CountVarP(&ansibleVerbose, "verbose", "v", "Increase verbosity")
	ansibleCmd.Flags().BoolVar(&ansibleCheck, "check", false, "Don't make any changes (dry run)")

	// test command flags
	ansibleTestCmd.Flags().StringVarP(&testUser, "user", "u", "root", "SSH user")
	ansibleTestCmd.Flags().StringVarP(&testPassword, "password", "p", "", "SSH password")
	ansibleTestCmd.Flags().StringVarP(&testKeyFile, "key", "i", "", "SSH private key file")
	ansibleTestCmd.Flags().IntVar(&testPort, "port", 22, "SSH port")

	// Add subcommands
	ansibleCmd.AddCommand(ansibleTestCmd)
	Cmd.AddCommand(ansibleCmd)
}

func runAnsible(cmd *cobra.Command, args []string) error {
	playbookPath := args[0]

	// Resolve playbook path
	if !filepath.IsAbs(playbookPath) {
		cwd, _ := os.Getwd()
		playbookPath = filepath.Join(cwd, playbookPath)
	}

	if _, err := os.Stat(playbookPath); os.IsNotExist(err) {
		return fmt.Errorf("playbook not found: %s", playbookPath)
	}

	// Create executor
	basePath := filepath.Dir(playbookPath)
	executor := ansible.NewExecutor(basePath)
	defer executor.Close()

	// Set options
	executor.Limit = ansibleLimit
	executor.CheckMode = ansibleCheck
	executor.Verbose = ansibleVerbose

	if ansibleTags != "" {
		executor.Tags = strings.Split(ansibleTags, ",")
	}
	if ansibleSkipTags != "" {
		executor.SkipTags = strings.Split(ansibleSkipTags, ",")
	}

	// Parse extra vars
	for _, v := range ansibleVars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			executor.SetVar(parts[0], parts[1])
		}
	}

	// Load inventory
	if ansibleInventory != "" {
		invPath := ansibleInventory
		if !filepath.IsAbs(invPath) {
			cwd, _ := os.Getwd()
			invPath = filepath.Join(cwd, invPath)
		}

		// Check if it's a directory
		info, err := os.Stat(invPath)
		if err != nil {
			return fmt.Errorf("inventory not found: %s", invPath)
		}

		if info.IsDir() {
			// Look for inventory.yml or hosts.yml
			for _, name := range []string{"inventory.yml", "hosts.yml", "inventory.yaml", "hosts.yaml"} {
				p := filepath.Join(invPath, name)
				if _, err := os.Stat(p); err == nil {
					invPath = p
					break
				}
			}
		}

		if err := executor.SetInventory(invPath); err != nil {
			return fmt.Errorf("load inventory: %w", err)
		}
	}

	// Set up callbacks
	executor.OnPlayStart = func(play *ansible.Play) {
		fmt.Printf("\n%s %s\n", cli.TitleStyle.Render("PLAY"), cli.BoldStyle.Render("["+play.Name+"]"))
		fmt.Println(strings.Repeat("*", 70))
	}

	executor.OnTaskStart = func(host string, task *ansible.Task) {
		taskName := task.Name
		if taskName == "" {
			taskName = task.Module
		}
		fmt.Printf("\n%s %s\n", cli.TitleStyle.Render("TASK"), cli.BoldStyle.Render("["+taskName+"]"))
		if ansibleVerbose > 0 {
			fmt.Printf("%s\n", cli.DimStyle.Render("host: "+host))
		}
	}

	executor.OnTaskEnd = func(host string, task *ansible.Task, result *ansible.TaskResult) {
		status := "ok"
		style := cli.SuccessStyle

		if result.Failed {
			status = "failed"
			style = cli.ErrorStyle
		} else if result.Skipped {
			status = "skipping"
			style = cli.DimStyle
		} else if result.Changed {
			status = "changed"
			style = cli.WarningStyle
		}

		fmt.Printf("%s: [%s]", style.Render(status), host)
		if result.Msg != "" && ansibleVerbose > 0 {
			fmt.Printf(" => %s", result.Msg)
		}
		if result.Duration > 0 && ansibleVerbose > 1 {
			fmt.Printf(" (%s)", result.Duration.Round(time.Millisecond))
		}
		fmt.Println()

		if result.Failed && result.Stderr != "" {
			fmt.Printf("%s\n", cli.ErrorStyle.Render(result.Stderr))
		}

		if ansibleVerbose > 1 {
			if result.Stdout != "" {
				fmt.Printf("stdout: %s\n", strings.TrimSpace(result.Stdout))
			}
		}
	}

	executor.OnPlayEnd = func(play *ansible.Play) {
		fmt.Println()
	}

	// Run playbook
	ctx := context.Background()
	start := time.Now()

	fmt.Printf("%s Running playbook: %s\n", cli.BoldStyle.Render("▶"), playbookPath)

	if err := executor.Run(ctx, playbookPath); err != nil {
		return fmt.Errorf("playbook failed: %w", err)
	}

	fmt.Printf("\n%s Playbook completed in %s\n",
		cli.SuccessStyle.Render("✓"),
		time.Since(start).Round(time.Millisecond))

	return nil
}

func runAnsibleTest(cmd *cobra.Command, args []string) error {
	host := args[0]

	fmt.Printf("Testing SSH connection to %s...\n", cli.BoldStyle.Render(host))

	cfg := ansible.SSHConfig{
		Host:     host,
		Port:     testPort,
		User:     testUser,
		Password: testPassword,
		KeyFile:  testKeyFile,
		Timeout:  30 * time.Second,
	}

	client, err := ansible.NewSSHClient(cfg)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test connection
	start := time.Now()
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	connectTime := time.Since(start)

	fmt.Printf("%s Connected in %s\n", cli.SuccessStyle.Render("✓"), connectTime.Round(time.Millisecond))

	// Gather facts
	fmt.Println("\nGathering facts...")

	// Hostname
	stdout, _, _, _ := client.Run(ctx, "hostname -f 2>/dev/null || hostname")
	fmt.Printf("  Hostname: %s\n", cli.BoldStyle.Render(strings.TrimSpace(stdout)))

	// OS
	stdout, _, _, _ = client.Run(ctx, "cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d'\"' -f2")
	if stdout != "" {
		fmt.Printf("  OS: %s\n", strings.TrimSpace(stdout))
	}

	// Kernel
	stdout, _, _, _ = client.Run(ctx, "uname -r")
	fmt.Printf("  Kernel: %s\n", strings.TrimSpace(stdout))

	// Architecture
	stdout, _, _, _ = client.Run(ctx, "uname -m")
	fmt.Printf("  Architecture: %s\n", strings.TrimSpace(stdout))

	// Memory
	stdout, _, _, _ = client.Run(ctx, "free -h | grep Mem | awk '{print $2}'")
	fmt.Printf("  Memory: %s\n", strings.TrimSpace(stdout))

	// Disk
	stdout, _, _, _ = client.Run(ctx, "df -h / | tail -1 | awk '{print $2 \" total, \" $4 \" available\"}'")
	fmt.Printf("  Disk: %s\n", strings.TrimSpace(stdout))

	// Docker
	stdout, _, _, err = client.Run(ctx, "docker --version 2>/dev/null")
	if err == nil {
		fmt.Printf("  Docker: %s\n", cli.SuccessStyle.Render(strings.TrimSpace(stdout)))
	} else {
		fmt.Printf("  Docker: %s\n", cli.DimStyle.Render("not installed"))
	}

	// Check if Coolify is running
	stdout, _, _, _ = client.Run(ctx, "docker ps 2>/dev/null | grep -q coolify && echo 'running' || echo 'not running'")
	if strings.TrimSpace(stdout) == "running" {
		fmt.Printf("  Coolify: %s\n", cli.SuccessStyle.Render("running"))
	} else {
		fmt.Printf("  Coolify: %s\n", cli.DimStyle.Render("not installed"))
	}

	fmt.Printf("\n%s SSH test passed\n", cli.SuccessStyle.Render("✓"))

	return nil
}
