package php

import (
	"context"
	"time"

	"github.com/host-uk/core/pkg/cli"
)

// Environment represents a deployment environment.
type Environment string

const (
	// EnvProduction is the production environment.
	EnvProduction Environment = "production"
	// EnvStaging is the staging environment.
	EnvStaging Environment = "staging"
)

// DeployOptions configures a deployment.
type DeployOptions struct {
	// Dir is the project directory containing .env config.
	Dir string

	// Environment is the target environment (production or staging).
	Environment Environment

	// Force triggers a deployment even if no changes are detected.
	Force bool

	// Wait blocks until deployment completes.
	Wait bool

	// WaitTimeout is the maximum time to wait for deployment.
	// Defaults to 10 minutes.
	WaitTimeout time.Duration

	// PollInterval is how often to check deployment status when waiting.
	// Defaults to 5 seconds.
	PollInterval time.Duration
}

// StatusOptions configures a status check.
type StatusOptions struct {
	// Dir is the project directory containing .env config.
	Dir string

	// Environment is the target environment (production or staging).
	Environment Environment

	// DeploymentID is a specific deployment to check.
	// If empty, returns the latest deployment.
	DeploymentID string
}

// RollbackOptions configures a rollback.
type RollbackOptions struct {
	// Dir is the project directory containing .env config.
	Dir string

	// Environment is the target environment (production or staging).
	Environment Environment

	// DeploymentID is the deployment to rollback to.
	// If empty, rolls back to the previous successful deployment.
	DeploymentID string

	// Wait blocks until rollback completes.
	Wait bool

	// WaitTimeout is the maximum time to wait for rollback.
	WaitTimeout time.Duration
}

// DeploymentStatus represents the status of a deployment.
type DeploymentStatus struct {
	// ID is the deployment identifier.
	ID string

	// Status is the current deployment status.
	// Values: queued, building, deploying, finished, failed, cancelled
	Status string

	// URL is the deployed application URL.
	URL string

	// Commit is the git commit SHA.
	Commit string

	// CommitMessage is the git commit message.
	CommitMessage string

	// Branch is the git branch.
	Branch string

	// StartedAt is when the deployment started.
	StartedAt time.Time

	// CompletedAt is when the deployment completed.
	CompletedAt time.Time

	// Log contains deployment logs.
	Log string
}

// Deploy triggers a deployment to Coolify.
func Deploy(ctx context.Context, opts DeployOptions) (*DeploymentStatus, error) {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if opts.Environment == "" {
		opts.Environment = EnvProduction
	}
	if opts.WaitTimeout == 0 {
		opts.WaitTimeout = 10 * time.Minute
	}
	if opts.PollInterval == 0 {
		opts.PollInterval = 5 * time.Second
	}

	// Load config
	config, err := LoadCoolifyConfig(opts.Dir)
	if err != nil {
		return nil, cli.WrapVerb(err, "load", "Coolify config")
	}

	// Get app ID for environment
	appID := getAppIDForEnvironment(config, opts.Environment)
	if appID == "" {
		return nil, cli.Err("no app ID configured for %s environment", opts.Environment)
	}

	// Create client
	client := NewCoolifyClient(config.URL, config.Token)

	// Trigger deployment
	deployment, err := client.TriggerDeploy(ctx, appID, opts.Force)
	if err != nil {
		return nil, cli.WrapVerb(err, "trigger", "deployment")
	}

	status := convertDeployment(deployment)

	// Wait for completion if requested
	if opts.Wait && deployment.ID != "" {
		status, err = waitForDeployment(ctx, client, appID, deployment.ID, opts.WaitTimeout, opts.PollInterval)
		if err != nil {
			return status, err
		}
	}

	// Get app info for URL
	app, err := client.GetApp(ctx, appID)
	if err == nil && app.FQDN != "" {
		status.URL = app.FQDN
	}

	return status, nil
}

// DeployStatus retrieves the status of a deployment.
func DeployStatus(ctx context.Context, opts StatusOptions) (*DeploymentStatus, error) {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if opts.Environment == "" {
		opts.Environment = EnvProduction
	}

	// Load config
	config, err := LoadCoolifyConfig(opts.Dir)
	if err != nil {
		return nil, cli.WrapVerb(err, "load", "Coolify config")
	}

	// Get app ID for environment
	appID := getAppIDForEnvironment(config, opts.Environment)
	if appID == "" {
		return nil, cli.Err("no app ID configured for %s environment", opts.Environment)
	}

	// Create client
	client := NewCoolifyClient(config.URL, config.Token)

	var deployment *CoolifyDeployment

	if opts.DeploymentID != "" {
		// Get specific deployment
		deployment, err = client.GetDeployment(ctx, appID, opts.DeploymentID)
		if err != nil {
			return nil, cli.WrapVerb(err, "get", "deployment")
		}
	} else {
		// Get latest deployment
		deployments, err := client.ListDeployments(ctx, appID, 1)
		if err != nil {
			return nil, cli.WrapVerb(err, "list", "deployments")
		}
		if len(deployments) == 0 {
			return nil, cli.Err("no deployments found")
		}
		deployment = &deployments[0]
	}

	status := convertDeployment(deployment)

	// Get app info for URL
	app, err := client.GetApp(ctx, appID)
	if err == nil && app.FQDN != "" {
		status.URL = app.FQDN
	}

	return status, nil
}

// Rollback triggers a rollback to a previous deployment.
func Rollback(ctx context.Context, opts RollbackOptions) (*DeploymentStatus, error) {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if opts.Environment == "" {
		opts.Environment = EnvProduction
	}
	if opts.WaitTimeout == 0 {
		opts.WaitTimeout = 10 * time.Minute
	}

	// Load config
	config, err := LoadCoolifyConfig(opts.Dir)
	if err != nil {
		return nil, cli.WrapVerb(err, "load", "Coolify config")
	}

	// Get app ID for environment
	appID := getAppIDForEnvironment(config, opts.Environment)
	if appID == "" {
		return nil, cli.Err("no app ID configured for %s environment", opts.Environment)
	}

	// Create client
	client := NewCoolifyClient(config.URL, config.Token)

	// Find deployment to rollback to
	deploymentID := opts.DeploymentID
	if deploymentID == "" {
		// Find previous successful deployment
		deployments, err := client.ListDeployments(ctx, appID, 10)
		if err != nil {
			return nil, cli.WrapVerb(err, "list", "deployments")
		}

		// Skip the first (current) deployment, find the last successful one
		for i, d := range deployments {
			if i == 0 {
				continue // Skip current deployment
			}
			if d.Status == "finished" || d.Status == "success" {
				deploymentID = d.ID
				break
			}
		}

		if deploymentID == "" {
			return nil, cli.Err("no previous successful deployment found to rollback to")
		}
	}

	// Trigger rollback
	deployment, err := client.Rollback(ctx, appID, deploymentID)
	if err != nil {
		return nil, cli.WrapVerb(err, "trigger", "rollback")
	}

	status := convertDeployment(deployment)

	// Wait for completion if requested
	if opts.Wait && deployment.ID != "" {
		status, err = waitForDeployment(ctx, client, appID, deployment.ID, opts.WaitTimeout, 5*time.Second)
		if err != nil {
			return status, err
		}
	}

	return status, nil
}

// ListDeployments retrieves recent deployments.
func ListDeployments(ctx context.Context, dir string, env Environment, limit int) ([]DeploymentStatus, error) {
	if dir == "" {
		dir = "."
	}
	if env == "" {
		env = EnvProduction
	}
	if limit == 0 {
		limit = 10
	}

	// Load config
	config, err := LoadCoolifyConfig(dir)
	if err != nil {
		return nil, cli.WrapVerb(err, "load", "Coolify config")
	}

	// Get app ID for environment
	appID := getAppIDForEnvironment(config, env)
	if appID == "" {
		return nil, cli.Err("no app ID configured for %s environment", env)
	}

	// Create client
	client := NewCoolifyClient(config.URL, config.Token)

	deployments, err := client.ListDeployments(ctx, appID, limit)
	if err != nil {
		return nil, cli.WrapVerb(err, "list", "deployments")
	}

	result := make([]DeploymentStatus, len(deployments))
	for i, d := range deployments {
		result[i] = *convertDeployment(&d)
	}

	return result, nil
}

// getAppIDForEnvironment returns the app ID for the given environment.
func getAppIDForEnvironment(config *CoolifyConfig, env Environment) string {
	switch env {
	case EnvStaging:
		if config.StagingAppID != "" {
			return config.StagingAppID
		}
		return config.AppID // Fallback to production
	default:
		return config.AppID
	}
}

// convertDeployment converts a CoolifyDeployment to DeploymentStatus.
func convertDeployment(d *CoolifyDeployment) *DeploymentStatus {
	return &DeploymentStatus{
		ID:            d.ID,
		Status:        d.Status,
		URL:           d.DeployedURL,
		Commit:        d.CommitSHA,
		CommitMessage: d.CommitMsg,
		Branch:        d.Branch,
		StartedAt:     d.CreatedAt,
		CompletedAt:   d.FinishedAt,
		Log:           d.Log,
	}
}

// waitForDeployment polls for deployment completion.
func waitForDeployment(ctx context.Context, client *CoolifyClient, appID, deploymentID string, timeout, interval time.Duration) (*DeploymentStatus, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		deployment, err := client.GetDeployment(ctx, appID, deploymentID)
		if err != nil {
			return nil, cli.WrapVerb(err, "get", "deployment status")
		}

		status := convertDeployment(deployment)

		// Check if deployment is complete
		switch deployment.Status {
		case "finished", "success":
			return status, nil
		case "failed", "error":
			return status, cli.Err("deployment failed: %s", deployment.Status)
		case "cancelled":
			return status, cli.Err("deployment was cancelled")
		}

		// Still in progress, wait and retry
		select {
		case <-ctx.Done():
			return status, ctx.Err()
		case <-time.After(interval):
		}
	}

	return nil, cli.Err("deployment timed out after %v", timeout)
}

// IsDeploymentComplete returns true if the status indicates completion.
func IsDeploymentComplete(status string) bool {
	switch status {
	case "finished", "success", "failed", "error", "cancelled":
		return true
	default:
		return false
	}
}

// IsDeploymentSuccessful returns true if the status indicates success.
func IsDeploymentSuccessful(status string) bool {
	return status == "finished" || status == "success"
}
