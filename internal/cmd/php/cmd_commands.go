// Package php provides Laravel/PHP development and deployment commands.
//
// Development Commands:
//   - dev: Start Laravel environment (FrankenPHP, Vite, Horizon, Reverb, Redis)
//   - logs: Stream unified service logs
//   - stop: Stop all running services
//   - status: Show service status
//   - ssl: Setup SSL certificates with mkcert
//
// Build Commands:
//   - build: Build Docker or LinuxKit image
//   - serve: Run production container
//   - shell: Open shell in running container
//
// Code Quality:
//   - test: Run PHPUnit/Pest tests
//   - fmt: Format code with Laravel Pint
//   - stan: Run PHPStan/Larastan static analysis
//   - psalm: Run Psalm static analysis
//   - audit: Security audit for dependencies
//   - security: Security vulnerability scanning
//   - qa: Run full QA pipeline
//   - rector: Automated code refactoring
//   - infection: Mutation testing for test quality
//
// Package Management:
//   - packages link/unlink/update/list: Manage local Composer packages
//
// Deployment (Coolify):
//   - deploy: Deploy to Coolify
//   - deploy:status: Check deployment status
//   - deploy:rollback: Rollback deployment
//   - deploy:list: List recent deployments
package php

import "github.com/spf13/cobra"

// AddCommands registers the 'php' command and all subcommands.
func AddCommands(root *cobra.Command) {
	AddPHPCommands(root)
}
