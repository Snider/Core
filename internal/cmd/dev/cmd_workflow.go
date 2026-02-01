package dev

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

// Workflow command flags
var (
	workflowRegistryPath string
	workflowDryRun       bool
)

// addWorkflowCommands adds the 'workflow' subcommand and its subcommands.
func addWorkflowCommands(parent *cli.Command) {
	workflowCmd := &cli.Command{
		Use:   "workflow",
		Short: i18n.T("cmd.dev.workflow.short"),
		Long:  i18n.T("cmd.dev.workflow.long"),
	}

	// Shared flags
	workflowCmd.PersistentFlags().StringVar(&workflowRegistryPath, "registry", "", i18n.T("common.flag.registry"))

	// Subcommands
	addWorkflowListCommand(workflowCmd)
	addWorkflowSyncCommand(workflowCmd)

	parent.AddCommand(workflowCmd)
}

// addWorkflowListCommand adds the 'workflow list' subcommand.
func addWorkflowListCommand(parent *cli.Command) {
	listCmd := &cli.Command{
		Use:   "list",
		Short: i18n.T("cmd.dev.workflow.list.short"),
		Long:  i18n.T("cmd.dev.workflow.list.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runWorkflowList(workflowRegistryPath)
		},
	}

	parent.AddCommand(listCmd)
}

// addWorkflowSyncCommand adds the 'workflow sync' subcommand.
func addWorkflowSyncCommand(parent *cli.Command) {
	syncCmd := &cli.Command{
		Use:   "sync <workflow>",
		Short: i18n.T("cmd.dev.workflow.sync.short"),
		Long:  i18n.T("cmd.dev.workflow.sync.long"),
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			return runWorkflowSync(workflowRegistryPath, args[0], workflowDryRun)
		},
	}

	syncCmd.Flags().BoolVar(&workflowDryRun, "dry-run", false, i18n.T("cmd.dev.workflow.sync.flag.dry_run"))

	parent.AddCommand(syncCmd)
}

// runWorkflowList shows a table of repos vs workflows.
func runWorkflowList(registryPath string) error {
	reg, registryDir, err := loadRegistryWithConfig(registryPath)
	if err != nil {
		return err
	}

	repoList := reg.List()
	if len(repoList) == 0 {
		cli.Text(i18n.T("cmd.dev.no_git_repos"))
		return nil
	}

	// Sort repos by name for consistent output
	sort.Slice(repoList, func(i, j int) bool {
		return repoList[i].Name < repoList[j].Name
	})

	// Collect all unique workflow files across all repos
	workflowSet := make(map[string]bool)
	repoWorkflows := make(map[string]map[string]bool)

	for _, repo := range repoList {
		workflows := findWorkflows(repo.Path)
		repoWorkflows[repo.Name] = make(map[string]bool)
		for _, wf := range workflows {
			workflowSet[wf] = true
			repoWorkflows[repo.Name][wf] = true
		}
	}

	// Sort workflow names
	var workflowNames []string
	for wf := range workflowSet {
		workflowNames = append(workflowNames, wf)
	}
	sort.Strings(workflowNames)

	if len(workflowNames) == 0 {
		cli.Text(i18n.T("cmd.dev.workflow.no_workflows"))
		return nil
	}

	// Check for template workflows in the registry directory
	templateWorkflows := findWorkflows(filepath.Join(registryDir, ".github", "workflow-templates"))
	if len(templateWorkflows) == 0 {
		// Also check .github/workflows in the devops repo itself
		templateWorkflows = findWorkflows(filepath.Join(registryDir, ".github", "workflows"))
	}
	templateSet := make(map[string]bool)
	for _, wf := range templateWorkflows {
		templateSet[wf] = true
	}

	// Build table
	headers := []string{i18n.T("cmd.dev.workflow.header.repo")}
	headers = append(headers, workflowNames...)
	table := cli.NewTable(headers...)

	for _, repo := range repoList {
		row := []string{repo.Name}
		for _, wf := range workflowNames {
			if repoWorkflows[repo.Name][wf] {
				row = append(row, successStyle.Render(cli.Glyph(":check:")))
			} else {
				row = append(row, errorStyle.Render(cli.Glyph(":cross:")))
			}
		}
		table.AddRow(row...)
	}

	cli.Blank()
	table.Render()

	return nil
}

// runWorkflowSync copies a workflow template to all repos.
func runWorkflowSync(registryPath string, workflowFile string, dryRun bool) error {
	reg, registryDir, err := loadRegistryWithConfig(registryPath)
	if err != nil {
		return err
	}

	// Find the template workflow
	templatePath := findTemplateWorkflow(registryDir, workflowFile)
	if templatePath == "" {
		return cli.Err("%s", i18n.T("cmd.dev.workflow.template_not_found", map[string]interface{}{"File": workflowFile}))
	}

	// Read template content
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return cli.Wrap(err, i18n.T("cmd.dev.workflow.read_template_error"))
	}

	repoList := reg.List()
	if len(repoList) == 0 {
		cli.Text(i18n.T("cmd.dev.no_git_repos"))
		return nil
	}

	// Sort repos by name for consistent output
	sort.Slice(repoList, func(i, j int) bool {
		return repoList[i].Name < repoList[j].Name
	})

	if dryRun {
		cli.Text(i18n.T("cmd.dev.workflow.dry_run_mode"))
		cli.Blank()
	}

	var synced, skipped, failed int

	for _, repo := range repoList {
		if !repo.IsGitRepo() {
			skipped++
			continue
		}

		destDir := filepath.Join(repo.Path, ".github", "workflows")
		destPath := filepath.Join(destDir, workflowFile)

		// Check if workflow already exists and is identical
		if existingContent, err := os.ReadFile(destPath); err == nil {
			if string(existingContent) == string(templateContent) {
				cli.Print("  %s %s %s\n",
					dimStyle.Render("-"),
					repoNameStyle.Render(repo.Name),
					dimStyle.Render(i18n.T("cmd.dev.workflow.up_to_date")))
				skipped++
				continue
			}
		}

		if dryRun {
			cli.Print("  %s %s %s\n",
				warningStyle.Render("*"),
				repoNameStyle.Render(repo.Name),
				i18n.T("cmd.dev.workflow.would_sync"))
			synced++
			continue
		}

		// Create .github/workflows directory if needed
		if err := os.MkdirAll(destDir, 0755); err != nil {
			cli.Print("  %s %s %s\n",
				errorStyle.Render(cli.Glyph(":cross:")),
				repoNameStyle.Render(repo.Name),
				err.Error())
			failed++
			continue
		}

		// Write workflow file
		if err := os.WriteFile(destPath, templateContent, 0644); err != nil {
			cli.Print("  %s %s %s\n",
				errorStyle.Render(cli.Glyph(":cross:")),
				repoNameStyle.Render(repo.Name),
				err.Error())
			failed++
			continue
		}

		cli.Print("  %s %s %s\n",
			successStyle.Render(cli.Glyph(":check:")),
			repoNameStyle.Render(repo.Name),
			i18n.T("cmd.dev.workflow.synced"))
		synced++
	}

	cli.Blank()

	// Summary
	if dryRun {
		cli.Print("%s %s\n",
			i18n.T("cmd.dev.workflow.would_sync_count", map[string]interface{}{"Count": synced}),
			dimStyle.Render(i18n.T("cmd.dev.workflow.skipped_count", map[string]interface{}{"Count": skipped})))
		cli.Text(i18n.T("cmd.dev.workflow.run_without_dry_run"))
	} else {
		cli.Print("%s %s\n",
			successStyle.Render(i18n.T("cmd.dev.workflow.synced_count", map[string]interface{}{"Count": synced})),
			dimStyle.Render(i18n.T("cmd.dev.workflow.skipped_count", map[string]interface{}{"Count": skipped})))
		if failed > 0 {
			cli.Print("%s\n", errorStyle.Render(i18n.T("cmd.dev.workflow.failed_count", map[string]interface{}{"Count": failed})))
		}
	}

	return nil
}

// findWorkflows returns a list of workflow file names in a directory.
func findWorkflows(dir string) []string {
	workflowsDir := filepath.Join(dir, ".github", "workflows")
	// If dir already ends with workflows path, use it directly
	if strings.HasSuffix(dir, "workflows") || strings.HasSuffix(dir, "workflow-templates") {
		workflowsDir = dir
	}

	entries, err := os.ReadDir(workflowsDir)
	if err != nil {
		return nil
	}

	var workflows []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			workflows = append(workflows, name)
		}
	}

	return workflows
}

// findTemplateWorkflow finds a workflow template file in common locations.
func findTemplateWorkflow(registryDir, workflowFile string) string {
	// Ensure .yml extension
	if !strings.HasSuffix(workflowFile, ".yml") && !strings.HasSuffix(workflowFile, ".yaml") {
		workflowFile = workflowFile + ".yml"
	}

	// Check common template locations
	candidates := []string{
		filepath.Join(registryDir, ".github", "workflow-templates", workflowFile),
		filepath.Join(registryDir, ".github", "workflows", workflowFile),
		filepath.Join(registryDir, "workflow-templates", workflowFile),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}
