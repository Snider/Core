// cmd_agent.go manages persistent agent context within task workspaces.
//
// Each agent gets a directory at:
//   .core/workspace/p{epic}/i{issue}/agents/{provider}/{agent-name}/
//
// This directory persists across invocations, allowing agents to build
// understanding over time — QA agents accumulate findings, reviewers
// track patterns, implementors record decisions.
//
// Layout:
//
//	agents/
//	├── claude-opus/implementor/
//	│   ├── memory.md        # Persistent notes, decisions, context
//	│   └── artifacts/       # Generated artifacts (reports, diffs, etc.)
//	├── claude-opus/qa/
//	│   ├── memory.md
//	│   └── artifacts/
//	└── gemini/reviewer/
//	    └── memory.md
package workspace

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/cli"
	coreio "github.com/host-uk/core/pkg/io"
	"github.com/spf13/cobra"
)

var (
	agentProvider string
	agentName    string
)

func addAgentCommands(parent *cobra.Command) {
	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage persistent agent context within task workspaces",
	}

	initCmd := &cobra.Command{
		Use:   "init <provider/agent-name>",
		Short: "Initialize an agent's context directory in the task workspace",
		Long: `Creates agents/{provider}/{agent-name}/ with memory.md and artifacts/
directory. The agent can read/write memory.md across invocations to
build understanding over time.`,
		Args: cobra.ExactArgs(1),
		RunE: runAgentInit,
	}
	initCmd.Flags().IntVar(&taskEpic, "epic", 0, "Epic/project number")
	initCmd.Flags().IntVar(&taskIssue, "issue", 0, "Issue number")
	_ = initCmd.MarkFlagRequired("epic")
	_ = initCmd.MarkFlagRequired("issue")

	agentListCmd := &cobra.Command{
		Use:   "list",
		Short: "List agents in a task workspace",
		RunE:  runAgentList,
	}
	agentListCmd.Flags().IntVar(&taskEpic, "epic", 0, "Epic/project number")
	agentListCmd.Flags().IntVar(&taskIssue, "issue", 0, "Issue number")
	_ = agentListCmd.MarkFlagRequired("epic")
	_ = agentListCmd.MarkFlagRequired("issue")

	pathCmd := &cobra.Command{
		Use:   "path <provider/agent-name>",
		Short: "Print the agent's context directory path",
		Args:  cobra.ExactArgs(1),
		RunE:  runAgentPath,
	}
	pathCmd.Flags().IntVar(&taskEpic, "epic", 0, "Epic/project number")
	pathCmd.Flags().IntVar(&taskIssue, "issue", 0, "Issue number")
	_ = pathCmd.MarkFlagRequired("epic")
	_ = pathCmd.MarkFlagRequired("issue")

	agentCmd.AddCommand(initCmd, agentListCmd, pathCmd)
	parent.AddCommand(agentCmd)
}

// agentContextPath returns the path for an agent's context directory.
func agentContextPath(wsPath, provider, name string) string {
	return filepath.Join(wsPath, "agents", provider, name)
}

// parseAgentID splits "provider/agent-name" into parts.
func parseAgentID(id string) (provider, name string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("agent ID must be provider/agent-name (e.g. claude-opus/qa)")
	}
	return parts[0], parts[1], nil
}

// AgentManifest tracks agent metadata for a task workspace.
type AgentManifest struct {
	Provider  string    `json:"provider"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
}

func runAgentInit(cmd *cobra.Command, args []string) error {
	provider, name, err := parseAgentID(args[0])
	if err != nil {
		return err
	}

	root, err := FindWorkspaceRoot()
	if err != nil {
		return cli.Err("not in a workspace")
	}

	wsPath := taskWorkspacePath(root, taskEpic, taskIssue)
	if !coreio.Local.IsDir(wsPath) {
		return cli.Err("task workspace does not exist: p%d/i%d — create it first with `core workspace task create`", taskEpic, taskIssue)
	}

	agentDir := agentContextPath(wsPath, provider, name)

	if coreio.Local.IsDir(agentDir) {
		// Update last_seen
		updateAgentManifest(agentDir, provider, name)
		cli.Print("Agent %s/%s already initialized at p%d/i%d\n",
			cli.ValueStyle.Render(provider), cli.ValueStyle.Render(name), taskEpic, taskIssue)
		cli.Print("Path: %s\n", cli.DimStyle.Render(agentDir))
		return nil
	}

	// Create directory structure
	if err := coreio.Local.EnsureDir(agentDir); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}
	if err := coreio.Local.EnsureDir(filepath.Join(agentDir, "artifacts")); err != nil {
		return fmt.Errorf("failed to create artifacts directory: %w", err)
	}

	// Create initial memory.md
	memoryContent := fmt.Sprintf(`# %s/%s — Issue #%d (EPIC #%d)

## Context
- **Task workspace:** p%d/i%d
- **Initialized:** %s

## Notes

<!-- Add observations, decisions, and findings below -->
`, provider, name, taskIssue, taskEpic, taskEpic, taskIssue, time.Now().Format(time.RFC3339))

	if err := coreio.Local.Write(filepath.Join(agentDir, "memory.md"), memoryContent); err != nil {
		return fmt.Errorf("failed to create memory.md: %w", err)
	}

	// Write manifest
	updateAgentManifest(agentDir, provider, name)

	cli.Print("%s Agent %s/%s initialized at p%d/i%d\n",
		cli.SuccessStyle.Render("Done:"),
		cli.ValueStyle.Render(provider), cli.ValueStyle.Render(name),
		taskEpic, taskIssue)
	cli.Print("Memory: %s\n", cli.DimStyle.Render(filepath.Join(agentDir, "memory.md")))

	return nil
}

func runAgentList(cmd *cobra.Command, args []string) error {
	root, err := FindWorkspaceRoot()
	if err != nil {
		return cli.Err("not in a workspace")
	}

	wsPath := taskWorkspacePath(root, taskEpic, taskIssue)
	agentsDir := filepath.Join(wsPath, "agents")

	if !coreio.Local.IsDir(agentsDir) {
		cli.Println("No agents in this workspace.")
		return nil
	}

	providers, err := coreio.Local.List(agentsDir)
	if err != nil {
		return fmt.Errorf("failed to list agents: %w", err)
	}

	found := false
	for _, providerEntry := range providers {
		if !providerEntry.IsDir() {
			continue
		}
		providerDir := filepath.Join(agentsDir, providerEntry.Name())
		agents, err := coreio.Local.List(providerDir)
		if err != nil {
			continue
		}

		for _, agentEntry := range agents {
			if !agentEntry.IsDir() {
				continue
			}
			found = true
			agentDir := filepath.Join(providerDir, agentEntry.Name())

			// Read manifest for last_seen
			lastSeen := ""
			manifestPath := filepath.Join(agentDir, "manifest.json")
			if data, err := coreio.Local.Read(manifestPath); err == nil {
				var m AgentManifest
				if json.Unmarshal([]byte(data), &m) == nil {
					lastSeen = m.LastSeen.Format("2006-01-02 15:04")
				}
			}

			// Check if memory has content beyond the template
			memorySize := ""
			if content, err := coreio.Local.Read(filepath.Join(agentDir, "memory.md")); err == nil {
				lines := len(strings.Split(content, "\n"))
				memorySize = fmt.Sprintf("%d lines", lines)
			}

			cli.Print("  %s/%s  %s",
				cli.ValueStyle.Render(providerEntry.Name()),
				cli.ValueStyle.Render(agentEntry.Name()),
				cli.DimStyle.Render(memorySize))
			if lastSeen != "" {
				cli.Print("  last: %s", cli.DimStyle.Render(lastSeen))
			}
			cli.Print("\n")
		}
	}

	if !found {
		cli.Println("No agents in this workspace.")
	}

	return nil
}

func runAgentPath(cmd *cobra.Command, args []string) error {
	provider, name, err := parseAgentID(args[0])
	if err != nil {
		return err
	}

	root, err := FindWorkspaceRoot()
	if err != nil {
		return cli.Err("not in a workspace")
	}

	wsPath := taskWorkspacePath(root, taskEpic, taskIssue)
	agentDir := agentContextPath(wsPath, provider, name)

	if !coreio.Local.IsDir(agentDir) {
		return cli.Err("agent %s/%s not initialized — run `core workspace agent init %s/%s`", provider, name, provider, name)
	}

	// Print just the path (useful for scripting: cd $(core workspace agent path ...))
	cli.Text(agentDir)
	return nil
}

func updateAgentManifest(agentDir, provider, name string) {
	now := time.Now()
	manifest := AgentManifest{
		Provider:  provider,
		Name:      name,
		CreatedAt: now,
		LastSeen:  now,
	}

	// Try to preserve created_at from existing manifest
	manifestPath := filepath.Join(agentDir, "manifest.json")
	if data, err := coreio.Local.Read(manifestPath); err == nil {
		var existing AgentManifest
		if json.Unmarshal([]byte(data), &existing) == nil {
			manifest.CreatedAt = existing.CreatedAt
		}
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return
	}
	_ = coreio.Local.Write(manifestPath, string(data))
}
