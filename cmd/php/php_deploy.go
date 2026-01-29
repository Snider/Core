package php

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	phppkg "github.com/host-uk/core/pkg/php"
	"github.com/leaanthony/clir"
)

// Deploy command styles
var (
	phpDeployStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10b981")) // emerald-500

	phpDeployPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f59e0b")) // amber-500

	phpDeployFailedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ef4444")) // red-500
)

func addPHPDeployCommands(parent *clir.Command) {
	// Main deploy command
	addPHPDeployCommand(parent)

	// Deploy status subcommand (using colon notation: deploy:status)
	addPHPDeployStatusCommand(parent)

	// Deploy rollback subcommand
	addPHPDeployRollbackCommand(parent)

	// Deploy list subcommand
	addPHPDeployListCommand(parent)
}

func addPHPDeployCommand(parent *clir.Command) {
	var (
		staging bool
		force   bool
		wait    bool
	)

	deployCmd := parent.NewSubCommand("deploy", "Deploy to Coolify")
	deployCmd.LongDescription("Deploy the PHP application to Coolify.\n\n" +
		"Requires configuration in .env:\n" +
		"  COOLIFY_URL=https://coolify.example.com\n" +
		"  COOLIFY_TOKEN=your-api-token\n" +
		"  COOLIFY_APP_ID=production-app-id\n" +
		"  COOLIFY_STAGING_APP_ID=staging-app-id (optional)\n\n" +
		"Examples:\n" +
		"  core php deploy              # Deploy to production\n" +
		"  core php deploy --staging    # Deploy to staging\n" +
		"  core php deploy --force      # Force deployment\n" +
		"  core php deploy --wait       # Wait for deployment to complete")

	deployCmd.BoolFlag("staging", "Deploy to staging environment", &staging)
	deployCmd.BoolFlag("force", "Force deployment even if no changes detected", &force)
	deployCmd.BoolFlag("wait", "Wait for deployment to complete", &wait)

	deployCmd.Action(func() error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		env := phppkg.EnvProduction
		if staging {
			env = phppkg.EnvStaging
		}

		fmt.Printf("%s Deploying to %s...\n\n", dimStyle.Render("Deploy:"), env)

		ctx := context.Background()

		opts := phppkg.DeployOptions{
			Dir:         cwd,
			Environment: env,
			Force:       force,
			Wait:        wait,
		}

		status, err := phppkg.Deploy(ctx, opts)
		if err != nil {
			return fmt.Errorf("deployment failed: %w", err)
		}

		printDeploymentStatus(status)

		if wait {
			if phppkg.IsDeploymentSuccessful(status.Status) {
				fmt.Printf("\n%s Deployment completed successfully\n", successStyle.Render("Done:"))
			} else {
				fmt.Printf("\n%s Deployment ended with status: %s\n", errorStyle.Render("Warning:"), status.Status)
			}
		} else {
			fmt.Printf("\n%s Deployment triggered. Use 'core php deploy:status' to check progress.\n", successStyle.Render("Done:"))
		}

		return nil
	})
}

func addPHPDeployStatusCommand(parent *clir.Command) {
	var (
		staging      bool
		deploymentID string
	)

	statusCmd := parent.NewSubCommand("deploy:status", "Show deployment status")
	statusCmd.LongDescription("Show the status of a deployment.\n\n" +
		"Examples:\n" +
		"  core php deploy:status                    # Latest production deployment\n" +
		"  core php deploy:status --staging          # Latest staging deployment\n" +
		"  core php deploy:status --id abc123        # Specific deployment")

	statusCmd.BoolFlag("staging", "Check staging environment", &staging)
	statusCmd.StringFlag("id", "Specific deployment ID", &deploymentID)

	statusCmd.Action(func() error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		env := phppkg.EnvProduction
		if staging {
			env = phppkg.EnvStaging
		}

		fmt.Printf("%s Checking %s deployment status...\n\n", dimStyle.Render("Deploy:"), env)

		ctx := context.Background()

		opts := phppkg.StatusOptions{
			Dir:          cwd,
			Environment:  env,
			DeploymentID: deploymentID,
		}

		status, err := phppkg.DeployStatus(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		printDeploymentStatus(status)

		return nil
	})
}

func addPHPDeployRollbackCommand(parent *clir.Command) {
	var (
		staging      bool
		deploymentID string
		wait         bool
	)

	rollbackCmd := parent.NewSubCommand("deploy:rollback", "Rollback to previous deployment")
	rollbackCmd.LongDescription("Rollback to a previous deployment.\n\n" +
		"If no deployment ID is specified, rolls back to the most recent\n" +
		"successful deployment.\n\n" +
		"Examples:\n" +
		"  core php deploy:rollback                  # Rollback to previous\n" +
		"  core php deploy:rollback --staging        # Rollback staging\n" +
		"  core php deploy:rollback --id abc123      # Rollback to specific deployment")

	rollbackCmd.BoolFlag("staging", "Rollback staging environment", &staging)
	rollbackCmd.StringFlag("id", "Specific deployment ID to rollback to", &deploymentID)
	rollbackCmd.BoolFlag("wait", "Wait for rollback to complete", &wait)

	rollbackCmd.Action(func() error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		env := phppkg.EnvProduction
		if staging {
			env = phppkg.EnvStaging
		}

		fmt.Printf("%s Rolling back %s...\n\n", dimStyle.Render("Deploy:"), env)

		ctx := context.Background()

		opts := phppkg.RollbackOptions{
			Dir:          cwd,
			Environment:  env,
			DeploymentID: deploymentID,
			Wait:         wait,
		}

		status, err := phppkg.Rollback(ctx, opts)
		if err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		printDeploymentStatus(status)

		if wait {
			if phppkg.IsDeploymentSuccessful(status.Status) {
				fmt.Printf("\n%s Rollback completed successfully\n", successStyle.Render("Done:"))
			} else {
				fmt.Printf("\n%s Rollback ended with status: %s\n", errorStyle.Render("Warning:"), status.Status)
			}
		} else {
			fmt.Printf("\n%s Rollback triggered. Use 'core php deploy:status' to check progress.\n", successStyle.Render("Done:"))
		}

		return nil
	})
}

func addPHPDeployListCommand(parent *clir.Command) {
	var (
		staging bool
		limit   int
	)

	listCmd := parent.NewSubCommand("deploy:list", "List recent deployments")
	listCmd.LongDescription("List recent deployments.\n\n" +
		"Examples:\n" +
		"  core php deploy:list                      # List production deployments\n" +
		"  core php deploy:list --staging            # List staging deployments\n" +
		"  core php deploy:list --limit 20           # List more deployments")

	listCmd.BoolFlag("staging", "List staging deployments", &staging)
	listCmd.IntFlag("limit", "Number of deployments to list (default: 10)", &limit)

	listCmd.Action(func() error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		env := phppkg.EnvProduction
		if staging {
			env = phppkg.EnvStaging
		}

		if limit == 0 {
			limit = 10
		}

		fmt.Printf("%s Recent %s deployments:\n\n", dimStyle.Render("Deploy:"), env)

		ctx := context.Background()

		deployments, err := phppkg.ListDeployments(ctx, cwd, env, limit)
		if err != nil {
			return fmt.Errorf("failed to list deployments: %w", err)
		}

		if len(deployments) == 0 {
			fmt.Printf("%s No deployments found\n", dimStyle.Render("Info:"))
			return nil
		}

		for i, d := range deployments {
			printDeploymentSummary(i+1, &d)
		}

		return nil
	})
}

func printDeploymentStatus(status *phppkg.DeploymentStatus) {
	// Status with color
	statusStyle := phpDeployStyle
	switch status.Status {
	case "queued", "building", "deploying", "pending", "rolling_back":
		statusStyle = phpDeployPendingStyle
	case "failed", "error", "cancelled":
		statusStyle = phpDeployFailedStyle
	}

	fmt.Printf("%s %s\n", dimStyle.Render("Status:"), statusStyle.Render(status.Status))

	if status.ID != "" {
		fmt.Printf("%s %s\n", dimStyle.Render("ID:"), status.ID)
	}

	if status.URL != "" {
		fmt.Printf("%s %s\n", dimStyle.Render("URL:"), linkStyle.Render(status.URL))
	}

	if status.Branch != "" {
		fmt.Printf("%s %s\n", dimStyle.Render("Branch:"), status.Branch)
	}

	if status.Commit != "" {
		commit := status.Commit
		if len(commit) > 7 {
			commit = commit[:7]
		}
		fmt.Printf("%s %s\n", dimStyle.Render("Commit:"), commit)
		if status.CommitMessage != "" {
			// Truncate long messages
			msg := status.CommitMessage
			if len(msg) > 60 {
				msg = msg[:57] + "..."
			}
			fmt.Printf("%s %s\n", dimStyle.Render("Message:"), msg)
		}
	}

	if !status.StartedAt.IsZero() {
		fmt.Printf("%s %s\n", dimStyle.Render("Started:"), status.StartedAt.Format(time.RFC3339))
	}

	if !status.CompletedAt.IsZero() {
		fmt.Printf("%s %s\n", dimStyle.Render("Completed:"), status.CompletedAt.Format(time.RFC3339))
		if !status.StartedAt.IsZero() {
			duration := status.CompletedAt.Sub(status.StartedAt)
			fmt.Printf("%s %s\n", dimStyle.Render("Duration:"), duration.Round(time.Second))
		}
	}
}

func printDeploymentSummary(index int, status *phppkg.DeploymentStatus) {
	// Status with color
	statusStyle := phpDeployStyle
	switch status.Status {
	case "queued", "building", "deploying", "pending", "rolling_back":
		statusStyle = phpDeployPendingStyle
	case "failed", "error", "cancelled":
		statusStyle = phpDeployFailedStyle
	}

	// Format: #1 [finished] abc1234 - commit message (2 hours ago)
	id := status.ID
	if len(id) > 8 {
		id = id[:8]
	}

	commit := status.Commit
	if len(commit) > 7 {
		commit = commit[:7]
	}

	msg := status.CommitMessage
	if len(msg) > 40 {
		msg = msg[:37] + "..."
	}

	age := ""
	if !status.StartedAt.IsZero() {
		age = formatTimeAgo(status.StartedAt)
	}

	fmt.Printf("  %s %s %s",
		dimStyle.Render(fmt.Sprintf("#%d", index)),
		statusStyle.Render(fmt.Sprintf("[%s]", status.Status)),
		id,
	)

	if commit != "" {
		fmt.Printf(" %s", commit)
	}

	if msg != "" {
		fmt.Printf(" - %s", msg)
	}

	if age != "" {
		fmt.Printf(" %s", dimStyle.Render(fmt.Sprintf("(%s)", age)))
	}

	fmt.Println()
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
