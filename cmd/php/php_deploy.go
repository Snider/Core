package php

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	phppkg "github.com/host-uk/core/pkg/php"
	"github.com/spf13/cobra"
)

// Deploy command styles (aliases to shared)
var (
	phpDeployStyle        = cli.DeploySuccessStyle
	phpDeployPendingStyle = cli.StatusWarningStyle
	phpDeployFailedStyle  = cli.StatusErrorStyle
)

func addPHPDeployCommands(parent *cobra.Command) {
	// Main deploy command
	addPHPDeployCommand(parent)

	// Deploy status subcommand (using colon notation: deploy:status)
	addPHPDeployStatusCommand(parent)

	// Deploy rollback subcommand
	addPHPDeployRollbackCommand(parent)

	// Deploy list subcommand
	addPHPDeployListCommand(parent)
}

var (
	deployStaging bool
	deployForce   bool
	deployWait    bool
)

func addPHPDeployCommand(parent *cobra.Command) {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: i18n.T("cmd.php.deploy.short"),
		Long:  i18n.T("cmd.php.deploy.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
			}

			env := phppkg.EnvProduction
			if deployStaging {
				env = phppkg.EnvStaging
			}

			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.deploy")), i18n.T("cmd.php.deploy.deploying", map[string]interface{}{"Environment": env}))

			ctx := context.Background()

			opts := phppkg.DeployOptions{
				Dir:         cwd,
				Environment: env,
				Force:       deployForce,
				Wait:        deployWait,
			}

			status, err := phppkg.Deploy(ctx, opts)
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.deploy_failed"), err)
			}

			printDeploymentStatus(status)

			if deployWait {
				if phppkg.IsDeploymentSuccessful(status.Status) {
					fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("common.label.done")), i18n.T("cmd.php.deploy.success"))
				} else {
					fmt.Printf("\n%s %s\n", errorStyle.Render(i18n.T("common.label.warning")), i18n.T("cmd.php.deploy.warning_status", map[string]interface{}{"Status": status.Status}))
				}
			} else {
				fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("common.label.done")), i18n.T("cmd.php.deploy.triggered"))
			}

			return nil
		},
	}

	deployCmd.Flags().BoolVar(&deployStaging, "staging", false, i18n.T("cmd.php.deploy.flag.staging"))
	deployCmd.Flags().BoolVar(&deployForce, "force", false, i18n.T("cmd.php.deploy.flag.force"))
	deployCmd.Flags().BoolVar(&deployWait, "wait", false, i18n.T("cmd.php.deploy.flag.wait"))

	parent.AddCommand(deployCmd)
}

var (
	deployStatusStaging      bool
	deployStatusDeploymentID string
)

func addPHPDeployStatusCommand(parent *cobra.Command) {
	statusCmd := &cobra.Command{
		Use:   "deploy:status",
		Short: i18n.T("cmd.php.deploy_status.short"),
		Long:  i18n.T("cmd.php.deploy_status.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
			}

			env := phppkg.EnvProduction
			if deployStatusStaging {
				env = phppkg.EnvStaging
			}

			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.deploy")), i18n.T("cmd.php.deploy_status.checking", map[string]interface{}{"Environment": env}))

			ctx := context.Background()

			opts := phppkg.StatusOptions{
				Dir:          cwd,
				Environment:  env,
				DeploymentID: deployStatusDeploymentID,
			}

			status, err := phppkg.DeployStatus(ctx, opts)
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.status_failed"), err)
			}

			printDeploymentStatus(status)

			return nil
		},
	}

	statusCmd.Flags().BoolVar(&deployStatusStaging, "staging", false, i18n.T("cmd.php.deploy_status.flag.staging"))
	statusCmd.Flags().StringVar(&deployStatusDeploymentID, "id", "", i18n.T("cmd.php.deploy_status.flag.id"))

	parent.AddCommand(statusCmd)
}

var (
	rollbackStaging      bool
	rollbackDeploymentID string
	rollbackWait         bool
)

func addPHPDeployRollbackCommand(parent *cobra.Command) {
	rollbackCmd := &cobra.Command{
		Use:   "deploy:rollback",
		Short: i18n.T("cmd.php.deploy_rollback.short"),
		Long:  i18n.T("cmd.php.deploy_rollback.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
			}

			env := phppkg.EnvProduction
			if rollbackStaging {
				env = phppkg.EnvStaging
			}

			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.deploy")), i18n.T("cmd.php.deploy_rollback.rolling_back", map[string]interface{}{"Environment": env}))

			ctx := context.Background()

			opts := phppkg.RollbackOptions{
				Dir:          cwd,
				Environment:  env,
				DeploymentID: rollbackDeploymentID,
				Wait:         rollbackWait,
			}

			status, err := phppkg.Rollback(ctx, opts)
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.rollback_failed"), err)
			}

			printDeploymentStatus(status)

			if rollbackWait {
				if phppkg.IsDeploymentSuccessful(status.Status) {
					fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("common.label.done")), i18n.T("cmd.php.deploy_rollback.success"))
				} else {
					fmt.Printf("\n%s %s\n", errorStyle.Render(i18n.T("common.label.warning")), i18n.T("cmd.php.deploy_rollback.warning_status", map[string]interface{}{"Status": status.Status}))
				}
			} else {
				fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("common.label.done")), i18n.T("cmd.php.deploy_rollback.triggered"))
			}

			return nil
		},
	}

	rollbackCmd.Flags().BoolVar(&rollbackStaging, "staging", false, i18n.T("cmd.php.deploy_rollback.flag.staging"))
	rollbackCmd.Flags().StringVar(&rollbackDeploymentID, "id", "", i18n.T("cmd.php.deploy_rollback.flag.id"))
	rollbackCmd.Flags().BoolVar(&rollbackWait, "wait", false, i18n.T("cmd.php.deploy_rollback.flag.wait"))

	parent.AddCommand(rollbackCmd)
}

var (
	deployListStaging bool
	deployListLimit   int
)

func addPHPDeployListCommand(parent *cobra.Command) {
	listCmd := &cobra.Command{
		Use:   "deploy:list",
		Short: i18n.T("cmd.php.deploy_list.short"),
		Long:  i18n.T("cmd.php.deploy_list.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
			}

			env := phppkg.EnvProduction
			if deployListStaging {
				env = phppkg.EnvStaging
			}

			limit := deployListLimit
			if limit == 0 {
				limit = 10
			}

			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.deploy")), i18n.T("cmd.php.deploy_list.recent", map[string]interface{}{"Environment": env}))

			ctx := context.Background()

			deployments, err := phppkg.ListDeployments(ctx, cwd, env, limit)
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.list_deployments"), err)
			}

			if len(deployments) == 0 {
				fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.info")), i18n.T("cmd.php.deploy_list.none_found"))
				return nil
			}

			for i, d := range deployments {
				printDeploymentSummary(i+1, &d)
			}

			return nil
		},
	}

	listCmd.Flags().BoolVar(&deployListStaging, "staging", false, i18n.T("cmd.php.deploy_list.flag.staging"))
	listCmd.Flags().IntVar(&deployListLimit, "limit", 0, i18n.T("cmd.php.deploy_list.flag.limit"))

	parent.AddCommand(listCmd)
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

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.status")), statusStyle.Render(status.Status))

	if status.ID != "" {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.id")), status.ID)
	}

	if status.URL != "" {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.url")), linkStyle.Render(status.URL))
	}

	if status.Branch != "" {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.branch")), status.Branch)
	}

	if status.Commit != "" {
		commit := status.Commit
		if len(commit) > 7 {
			commit = commit[:7]
		}
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.commit")), commit)
		if status.CommitMessage != "" {
			// Truncate long messages
			msg := status.CommitMessage
			if len(msg) > 60 {
				msg = msg[:57] + "..."
			}
			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.message")), msg)
		}
	}

	if !status.StartedAt.IsZero() {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.started")), status.StartedAt.Format(time.RFC3339))
	}

	if !status.CompletedAt.IsZero() {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.completed")), status.CompletedAt.Format(time.RFC3339))
		if !status.StartedAt.IsZero() {
			duration := status.CompletedAt.Sub(status.StartedAt)
			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.duration")), duration.Round(time.Second))
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
		return i18n.T("cli.time.just_now")
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return i18n.T("cli.time.minute_ago")
		}
		return i18n.T("cli.time.minutes_ago", map[string]interface{}{"Count": mins})
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return i18n.T("cli.time.hour_ago")
		}
		return i18n.T("cli.time.hours_ago", map[string]interface{}{"Count": hours})
	default:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return i18n.T("cli.time.day_ago")
		}
		return i18n.T("cli.time.days_ago", map[string]interface{}{"Count": days})
	}
}
