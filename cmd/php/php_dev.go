package php

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	phppkg "github.com/host-uk/core/pkg/php"
	"github.com/spf13/cobra"
)

var (
	devNoVite    bool
	devNoHorizon bool
	devNoReverb  bool
	devNoRedis   bool
	devHTTPS     bool
	devDomain    string
	devPort      int
)

func addPHPDevCommand(parent *cobra.Command) {
	devCmd := &cobra.Command{
		Use:   "dev",
		Short: "Start Laravel development environment",
		Long: "Starts all detected Laravel services.\n\n" +
			"Auto-detects:\n" +
			"  - Vite (vite.config.js/ts)\n" +
			"  - Horizon (config/horizon.php)\n" +
			"  - Reverb (config/reverb.php)\n" +
			"  - Redis (from .env)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPHPDev(phpDevOptions{
				NoVite:    devNoVite,
				NoHorizon: devNoHorizon,
				NoReverb:  devNoReverb,
				NoRedis:   devNoRedis,
				HTTPS:     devHTTPS,
				Domain:    devDomain,
				Port:      devPort,
			})
		},
	}

	devCmd.Flags().BoolVar(&devNoVite, "no-vite", false, "Skip Vite dev server")
	devCmd.Flags().BoolVar(&devNoHorizon, "no-horizon", false, "Skip Laravel Horizon")
	devCmd.Flags().BoolVar(&devNoReverb, "no-reverb", false, "Skip Laravel Reverb")
	devCmd.Flags().BoolVar(&devNoRedis, "no-redis", false, "Skip Redis server")
	devCmd.Flags().BoolVar(&devHTTPS, "https", false, "Enable HTTPS with mkcert")
	devCmd.Flags().StringVar(&devDomain, "domain", "", "Domain for SSL certificate (default: from APP_URL or localhost)")
	devCmd.Flags().IntVar(&devPort, "port", 0, "FrankenPHP port (default: 8000)")

	parent.AddCommand(devCmd)
}

type phpDevOptions struct {
	NoVite    bool
	NoHorizon bool
	NoReverb  bool
	NoRedis   bool
	HTTPS     bool
	Domain    string
	Port      int
}

func runPHPDev(opts phpDevOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check if this is a Laravel project
	if !phppkg.IsLaravelProject(cwd) {
		return fmt.Errorf("not a Laravel project (missing artisan or laravel/framework)")
	}

	// Get app name for display
	appName := phppkg.GetLaravelAppName(cwd)
	if appName == "" {
		appName = "Laravel"
	}

	fmt.Printf("%s Starting %s development environment\n\n", dimStyle.Render("PHP:"), appName)

	// Detect services
	services := phppkg.DetectServices(cwd)
	fmt.Printf("%s Detected services:\n", dimStyle.Render("Services:"))
	for _, svc := range services {
		fmt.Printf("  %s %s\n", successStyle.Render("*"), svc)
	}
	fmt.Println()

	// Setup options
	port := opts.Port
	if port == 0 {
		port = 8000
	}

	devOpts := phppkg.Options{
		Dir:            cwd,
		NoVite:         opts.NoVite,
		NoHorizon:      opts.NoHorizon,
		NoReverb:       opts.NoReverb,
		NoRedis:        opts.NoRedis,
		HTTPS:          opts.HTTPS,
		Domain:         opts.Domain,
		FrankenPHPPort: port,
	}

	// Create and start dev server
	server := phppkg.NewDevServer(devOpts)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Printf("\n%s Shutting down...\n", dimStyle.Render("PHP:"))
		cancel()
	}()

	if err := server.Start(ctx, devOpts); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Print status
	fmt.Printf("%s Services started:\n", successStyle.Render("Running:"))
	printServiceStatuses(server.Status())
	fmt.Println()

	// Print URLs
	appURL := phppkg.GetLaravelAppURL(cwd)
	if appURL == "" {
		if opts.HTTPS {
			appURL = fmt.Sprintf("https://localhost:%d", port)
		} else {
			appURL = fmt.Sprintf("http://localhost:%d", port)
		}
	}
	fmt.Printf("%s %s\n", dimStyle.Render("App URL:"), linkStyle.Render(appURL))

	// Check for Vite
	if !opts.NoVite && containsService(services, phppkg.ServiceVite) {
		fmt.Printf("%s %s\n", dimStyle.Render("Vite:"), linkStyle.Render("http://localhost:5173"))
	}

	fmt.Printf("\n%s\n\n", dimStyle.Render("Press Ctrl+C to stop all services"))

	// Stream unified logs
	logsReader, err := server.Logs("", true)
	if err != nil {
		fmt.Printf("%s Failed to get logs: %v\n", errorStyle.Render("Warning:"), err)
	} else {
		defer logsReader.Close()

		scanner := bufio.NewScanner(logsReader)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				goto shutdown
			default:
				line := scanner.Text()
				printColoredLog(line)
			}
		}
	}

shutdown:
	// Stop services
	if err := server.Stop(); err != nil {
		fmt.Printf("%s Error stopping services: %v\n", errorStyle.Render("Error:"), err)
	}

	fmt.Printf("%s All services stopped\n", successStyle.Render("Done:"))
	return nil
}

var (
	logsFollow  bool
	logsService string
)

func addPHPLogsCommand(parent *cobra.Command) {
	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "View service logs",
		Long: "Stream logs from Laravel services.\n\n" +
			"Services: frankenphp, vite, horizon, reverb, redis",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPHPLogs(logsService, logsFollow)
		},
	}

	logsCmd.Flags().BoolVar(&logsFollow, "follow", false, "Follow log output")
	logsCmd.Flags().StringVar(&logsService, "service", "", "Specific service (default: all)")

	parent.AddCommand(logsCmd)
}

func runPHPLogs(service string, follow bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !phppkg.IsLaravelProject(cwd) {
		return fmt.Errorf("not a Laravel project")
	}

	// Create a minimal server just to access logs
	server := phppkg.NewDevServer(phppkg.Options{Dir: cwd})

	logsReader, err := server.Logs(service, follow)
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}
	defer logsReader.Close()

	// Handle interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	scanner := bufio.NewScanner(logsReader)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
			printColoredLog(scanner.Text())
		}
	}

	return scanner.Err()
}

func addPHPStopCommand(parent *cobra.Command) {
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop all Laravel services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPHPStop()
		},
	}

	parent.AddCommand(stopCmd)
}

func runPHPStop() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	fmt.Printf("%s Stopping services...\n", dimStyle.Render("PHP:"))

	// We need to find running processes
	// This is a simplified version - in practice you'd want to track PIDs
	server := phppkg.NewDevServer(phppkg.Options{Dir: cwd})
	if err := server.Stop(); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	fmt.Printf("%s All services stopped\n", successStyle.Render("Done:"))
	return nil
}

func addPHPStatusCommand(parent *cobra.Command) {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show service status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPHPStatus()
		},
	}

	parent.AddCommand(statusCmd)
}

func runPHPStatus() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !phppkg.IsLaravelProject(cwd) {
		return fmt.Errorf("not a Laravel project")
	}

	appName := phppkg.GetLaravelAppName(cwd)
	if appName == "" {
		appName = "Laravel"
	}

	fmt.Printf("%s %s\n\n", dimStyle.Render("Project:"), appName)

	// Detect available services
	services := phppkg.DetectServices(cwd)
	fmt.Printf("%s\n", dimStyle.Render("Detected services:"))
	for _, svc := range services {
		style := getServiceStyle(string(svc))
		fmt.Printf("  %s %s\n", style.Render("*"), svc)
	}
	fmt.Println()

	// Package manager
	pm := phppkg.DetectPackageManager(cwd)
	fmt.Printf("%s %s\n", dimStyle.Render("Package manager:"), pm)

	// FrankenPHP status
	if phppkg.IsFrankenPHPProject(cwd) {
		fmt.Printf("%s %s\n", dimStyle.Render("Octane server:"), "FrankenPHP")
	}

	// SSL status
	appURL := phppkg.GetLaravelAppURL(cwd)
	if appURL != "" {
		domain := phppkg.ExtractDomainFromURL(appURL)
		if phppkg.CertsExist(domain, phppkg.SSLOptions{}) {
			fmt.Printf("%s %s\n", dimStyle.Render("SSL certificates:"), successStyle.Render("installed"))
		} else {
			fmt.Printf("%s %s\n", dimStyle.Render("SSL certificates:"), dimStyle.Render("not setup"))
		}
	}

	return nil
}

var sslDomain string

func addPHPSSLCommand(parent *cobra.Command) {
	sslCmd := &cobra.Command{
		Use:   "ssl",
		Short: "Setup SSL certificates with mkcert",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPHPSSL(sslDomain)
		},
	}

	sslCmd.Flags().StringVar(&sslDomain, "domain", "", "Domain for certificate (default: from APP_URL)")

	parent.AddCommand(sslCmd)
}

func runPHPSSL(domain string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Get domain from APP_URL if not specified
	if domain == "" {
		appURL := phppkg.GetLaravelAppURL(cwd)
		if appURL != "" {
			domain = phppkg.ExtractDomainFromURL(appURL)
		}
	}
	if domain == "" {
		domain = "localhost"
	}

	// Check if mkcert is installed
	if !phppkg.IsMkcertInstalled() {
		fmt.Printf("%s mkcert is not installed\n", errorStyle.Render("Error:"))
		fmt.Println("\nInstall with:")
		fmt.Println("  macOS:  brew install mkcert")
		fmt.Println("  Linux:  see https://github.com/FiloSottile/mkcert")
		return fmt.Errorf("mkcert not installed")
	}

	fmt.Printf("%s Setting up SSL for %s\n", dimStyle.Render("SSL:"), domain)

	// Check if certs already exist
	if phppkg.CertsExist(domain, phppkg.SSLOptions{}) {
		fmt.Printf("%s Certificates already exist\n", dimStyle.Render("Skip:"))

		certFile, keyFile, _ := phppkg.CertPaths(domain, phppkg.SSLOptions{})
		fmt.Printf("%s %s\n", dimStyle.Render("Cert:"), certFile)
		fmt.Printf("%s %s\n", dimStyle.Render("Key:"), keyFile)
		return nil
	}

	// Setup SSL
	if err := phppkg.SetupSSL(domain, phppkg.SSLOptions{}); err != nil {
		return fmt.Errorf("failed to setup SSL: %w", err)
	}

	certFile, keyFile, _ := phppkg.CertPaths(domain, phppkg.SSLOptions{})

	fmt.Printf("%s SSL certificates created\n", successStyle.Render("Done:"))
	fmt.Printf("%s %s\n", dimStyle.Render("Cert:"), certFile)
	fmt.Printf("%s %s\n", dimStyle.Render("Key:"), keyFile)

	return nil
}

// Helper functions for dev commands

func printServiceStatuses(statuses []phppkg.ServiceStatus) {
	for _, s := range statuses {
		style := getServiceStyle(s.Name)
		var statusText string

		if s.Error != nil {
			statusText = phpStatusError.Render(fmt.Sprintf("error: %v", s.Error))
		} else if s.Running {
			statusText = phpStatusRunning.Render("running")
			if s.Port > 0 {
				statusText += dimStyle.Render(fmt.Sprintf(" (port %d)", s.Port))
			}
			if s.PID > 0 {
				statusText += dimStyle.Render(fmt.Sprintf(" [pid %d]", s.PID))
			}
		} else {
			statusText = phpStatusStopped.Render("stopped")
		}

		fmt.Printf("  %s %s\n", style.Render(s.Name+":"), statusText)
	}
}

func printColoredLog(line string) {
	// Parse service prefix from log line
	timestamp := time.Now().Format("15:04:05")

	var style lipgloss.Style
	serviceName := ""

	if strings.HasPrefix(line, "[FrankenPHP]") {
		style = phpFrankenPHPStyle
		serviceName = "FrankenPHP"
		line = strings.TrimPrefix(line, "[FrankenPHP] ")
	} else if strings.HasPrefix(line, "[Vite]") {
		style = phpViteStyle
		serviceName = "Vite"
		line = strings.TrimPrefix(line, "[Vite] ")
	} else if strings.HasPrefix(line, "[Horizon]") {
		style = phpHorizonStyle
		serviceName = "Horizon"
		line = strings.TrimPrefix(line, "[Horizon] ")
	} else if strings.HasPrefix(line, "[Reverb]") {
		style = phpReverbStyle
		serviceName = "Reverb"
		line = strings.TrimPrefix(line, "[Reverb] ")
	} else if strings.HasPrefix(line, "[Redis]") {
		style = phpRedisStyle
		serviceName = "Redis"
		line = strings.TrimPrefix(line, "[Redis] ")
	} else {
		// Unknown service, print as-is
		fmt.Printf("%s %s\n", dimStyle.Render(timestamp), line)
		return
	}

	fmt.Printf("%s %s %s\n",
		dimStyle.Render(timestamp),
		style.Render(fmt.Sprintf("[%s]", serviceName)),
		line,
	)
}

func getServiceStyle(name string) lipgloss.Style {
	switch strings.ToLower(name) {
	case "frankenphp":
		return phpFrankenPHPStyle
	case "vite":
		return phpViteStyle
	case "horizon":
		return phpHorizonStyle
	case "reverb":
		return phpReverbStyle
	case "redis":
		return phpRedisStyle
	default:
		return dimStyle
	}
}

func containsService(services []phppkg.DetectedService, target phppkg.DetectedService) bool {
	for _, s := range services {
		if s == target {
			return true
		}
	}
	return false
}
