package devops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ensureHostKey ensures that the host key for the dev environment is in the known hosts file.
// This is used after boot to allow StrictHostKeyChecking=yes to work.
func ensureHostKey(ctx context.Context, port int) error {
	// Skip if requested (used in tests)
	if os.Getenv("CORE_SKIP_SSH_SCAN") == "true" {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	knownHostsPath := filepath.Join(home, ".core", "known_hosts")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(knownHostsPath), 0755); err != nil {
		return fmt.Errorf("create known_hosts dir: %w", err)
	}

	// Get host key using ssh-keyscan
	cmd := exec.CommandContext(ctx, "ssh-keyscan", "-p", fmt.Sprintf("%d", port), "localhost")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("ssh-keyscan failed: %w", err)
	}

	if len(out) == 0 {
		return fmt.Errorf("ssh-keyscan returned no keys")
	}

	// Read existing known_hosts to avoid duplicates
	existing, _ := os.ReadFile(knownHostsPath)
	existingStr := string(existing)

	// Append new keys that aren't already there
	f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open known_hosts: %w", err)
	}
	defer f.Close()

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.Contains(existingStr, line) {
			if _, err := f.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("write known_hosts: %w", err)
			}
		}
	}

	return nil
}
