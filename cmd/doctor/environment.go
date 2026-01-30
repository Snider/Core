package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/repos"
)

// checkGitHubSSH checks if SSH keys exist for GitHub access
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

// checkGitHubCLI checks if the GitHub CLI is authenticated
func checkGitHubCLI() bool {
	cmd := exec.Command("gh", "auth", "status")
	output, _ := cmd.CombinedOutput()
	// Check for any successful login (even if there's also a failing token)
	return strings.Contains(string(output), "Logged in to")
}

// checkWorkspace checks for repos.yaml and counts cloned repos
func checkWorkspace() {
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
}
