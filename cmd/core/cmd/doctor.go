package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Snider/Core/pkg/repos"
	"github.com/leaanthony/clir"
)

// AddDoctorCommand adds the 'doctor' command to the given parent command.
func AddDoctorCommand(parent *clir.Cli) {
	var verbose bool

	doctorCmd := parent.NewSubCommand("doctor", "Check development environment")
	doctorCmd.LongDescription("Checks that all required tools are installed and configured.\n" +
		"Run this before `core setup` to ensure your environment is ready.")

	doctorCmd.BoolFlag("verbose", "Show detailed version information", &verbose)

	doctorCmd.Action(func() error {
		return runDoctor(verbose)
	})
}

type check struct {
	name        string
	description string
	command     string
	args        []string
	required    bool
	versionFlag string
}

func runDoctor(verbose bool) error {
	fmt.Println("Checking development environment...")
	fmt.Println()

	checks := []check{
		// Required tools
		{
			name:        "Git",
			description: "Version control",
			command:     "git",
			args:        []string{"--version"},
			required:    true,
			versionFlag: "--version",
		},
		{
			name:        "GitHub CLI",
			description: "GitHub integration (issues, PRs, CI)",
			command:     "gh",
			args:        []string{"--version"},
			required:    true,
			versionFlag: "--version",
		},
		{
			name:        "PHP",
			description: "Laravel packages",
			command:     "php",
			args:        []string{"-v"},
			required:    true,
			versionFlag: "-v",
		},
		{
			name:        "Composer",
			description: "PHP dependencies",
			command:     "composer",
			args:        []string{"--version"},
			required:    true,
			versionFlag: "--version",
		},
		{
			name:        "Node.js",
			description: "Frontend builds",
			command:     "node",
			args:        []string{"--version"},
			required:    true,
			versionFlag: "--version",
		},
		// Optional tools
		{
			name:        "pnpm",
			description: "Fast package manager",
			command:     "pnpm",
			args:        []string{"--version"},
			required:    false,
			versionFlag: "--version",
		},
		{
			name:        "Claude Code",
			description: "AI-assisted development",
			command:     "claude",
			args:        []string{"--version"},
			required:    false,
			versionFlag: "--version",
		},
		{
			name:        "Docker",
			description: "Container runtime",
			command:     "docker",
			args:        []string{"--version"},
			required:    false,
			versionFlag: "--version",
		},
	}

	var passed, failed, optional int

	fmt.Println("Required:")
	for _, c := range checks {
		if !c.required {
			continue
		}
		ok, version := runCheck(c)
		if ok {
			if verbose && version != "" {
				fmt.Printf("  %s %s %s\n", successStyle.Render("✓"), c.name, dimStyle.Render(version))
			} else {
				fmt.Printf("  %s %s\n", successStyle.Render("✓"), c.name)
			}
			passed++
		} else {
			fmt.Printf("  %s %s - %s\n", errorStyle.Render("✗"), c.name, c.description)
			failed++
		}
	}

	fmt.Println("\nOptional:")
	for _, c := range checks {
		if c.required {
			continue
		}
		ok, version := runCheck(c)
		if ok {
			if verbose && version != "" {
				fmt.Printf("  %s %s %s\n", successStyle.Render("✓"), c.name, dimStyle.Render(version))
			} else {
				fmt.Printf("  %s %s\n", successStyle.Render("✓"), c.name)
			}
			passed++
		} else {
			fmt.Printf("  %s %s - %s\n", dimStyle.Render("○"), c.name, dimStyle.Render(c.description))
			optional++
		}
	}

	// Check SSH
	fmt.Println("\nGitHub Access:")
	if checkGitHubSSH() {
		fmt.Printf("  %s SSH key found\n", successStyle.Render("✓"))
	} else {
		fmt.Printf("  %s SSH key missing - run: ssh-keygen && gh ssh-key add\n", errorStyle.Render("✗"))
		failed++
	}

	if checkGitHubCLI() {
		fmt.Printf("  %s CLI authenticated\n", successStyle.Render("✓"))
	} else {
		fmt.Printf("  %s CLI authentication - run: gh auth login\n", errorStyle.Render("✗"))
		failed++
	}

	// Check workspace
	fmt.Println("\nWorkspace:")
	registryPath, err := repos.FindRegistry()
	if err == nil {
		fmt.Printf("  %s Found repos.yaml at %s\n", successStyle.Render("✓"), registryPath)

		reg, err := repos.LoadRegistry(registryPath)
		if err == nil {
			basePath := reg.BasePath
			if basePath == "" {
				basePath = "./packages"
			}
			if !filepath.IsAbs(basePath) {
				basePath = filepath.Join(filepath.Dir(registryPath), basePath)
			}
			if strings.HasPrefix(basePath, "~/") {
				home, _ := os.UserHomeDir()
				basePath = filepath.Join(home, basePath[2:])
			}

			// Count existing repos
			allRepos := reg.List()
			var cloned int
			for _, repo := range allRepos {
				repoPath := filepath.Join(basePath, repo.Name)
				if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
					cloned++
				}
			}
			fmt.Printf("  %s %d/%d repos cloned\n", successStyle.Render("✓"), cloned, len(allRepos))
		}
	} else {
		fmt.Printf("  %s No repos.yaml found (run from workspace directory)\n", dimStyle.Render("○"))
	}

	// Summary
	fmt.Println()
	if failed > 0 {
		fmt.Printf("%s %d issues found\n", errorStyle.Render("Doctor:"), failed)
		fmt.Println("\nInstall missing tools:")
		printInstallInstructions()
		return fmt.Errorf("%d required tools missing", failed)
	}

	fmt.Printf("%s Environment ready\n", successStyle.Render("Doctor:"))
	return nil
}

func runCheck(c check) (bool, string) {
	cmd := exec.Command(c.command, c.args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, ""
	}

	// Extract first line as version
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 {
		return true, strings.TrimSpace(lines[0])
	}
	return true, ""
}

func checkGitHubSSH() bool {
	// Just check if SSH keys exist - don't try to authenticate
	// (key might be locked/passphrase protected)
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	sshDir := filepath.Join(home, ".ssh")
	keyPatterns := []string{"id_rsa", "id_ed25519", "id_ecdsa", "id_dsa"}

	for _, key := range keyPatterns {
		keyPath := filepath.Join(sshDir, key)
		if _, err := os.Stat(keyPath); err == nil {
			return true
		}
	}

	return false
}

func checkGitHubCLI() bool {
	cmd := exec.Command("gh", "auth", "status")
	output, _ := cmd.CombinedOutput()
	// Check for any successful login (even if there's also a failing token)
	return strings.Contains(string(output), "Logged in to")
}

func printInstallInstructions() {
	switch runtime.GOOS {
	case "darwin":
		fmt.Println("  brew install git gh php composer node pnpm docker")
		fmt.Println("  brew install --cask claude")
	case "linux":
		fmt.Println("  # Install via your package manager or:")
		fmt.Println("  # Git: apt install git")
		fmt.Println("  # GitHub CLI: https://cli.github.com/")
		fmt.Println("  # PHP: apt install php8.3-cli")
		fmt.Println("  # Node: https://nodejs.org/")
		fmt.Println("  # pnpm: npm install -g pnpm")
	default:
		fmt.Println("  See documentation for your OS")
	}
}
