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
	"github.com/host-uk/core/pkg/i18n"
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
		Short: i18n.T("cmd.php.dev.short"),
		Long:  i18n.T("cmd.php.dev.long"),
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

	devCmd.Flags().BoolVar(&devNoVite, "no-vite", false, i18n.T("cmd.php.dev.flag.no_vite"))
	devCmd.Flags().BoolVar(&devNoHorizon, "no-horizon", false, i18n.T("cmd.php.dev.flag.no_horizon"))
	devCmd.Flags().BoolVar(&devNoReverb, "no-reverb", false, i18n.T("cmd.php.dev.flag.no_reverb"))
	devCmd.Flags().BoolVar(&devNoRedis, "no-redis", false, i18n.T("cmd.php.dev.flag.no_redis"))
	devCmd.Flags().BoolVar(&devHTTPS, "https", false, i18n.T("cmd.php.dev.flag.https"))
	devCmd.Flags().StringVar(&devDomain, "domain", "", i18n.T("cmd.php.dev.flag.domain"))
	devCmd.Flags().IntVar(&devPort, "port", 0, i18n.T("cmd.php.dev.flag.port"))

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
		return fmt.Errorf(i18n.T("cmd.php.error.not_laravel"))
	}

	// Get app name for display
	appName := phppkg.GetLaravelAppName(cwd)
	if appName == "" {
		appName = "Laravel"
	}

	fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.dev.starting", map[string]interface{}{"AppName": appName}))

	// Detect services
	services := phppkg.DetectServices(cwd)
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.services")), i18n.T("cmd.php.dev.detected_services"))
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
		fmt.Printf("\n%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.dev.shutting_down"))
		cancel()
	}()

	if err := server.Start(ctx, devOpts); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.start_services"), err)
	}

	// Print status
	fmt.Printf("%s %s\n", successStyle.Render(i18n.T("cmd.php.label.running")), i18n.T("cmd.php.dev.services_started"))
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
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.app_url")), linkStyle.Render(appURL))

	// Check for Vite
	if !opts.NoVite && containsService(services, phppkg.ServiceVite) {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.vite")), linkStyle.Render("http://localhost:5173"))
	}

	fmt.Printf("\n%s\n\n", dimStyle.Render(i18n.T("cmd.php.dev.press_ctrl_c")))

	// Stream unified logs
	logsReader, err := server.Logs("", true)
	if err != nil {
		fmt.Printf("%s %s\n", errorStyle.Render(i18n.T("cmd.php.label.warning")), i18n.T("cmd.php.dev.logs_failed", map[string]interface{}{"Error": err}))
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
		fmt.Printf("%s %s\n", errorStyle.Render(i18n.T("cmd.php.label.error")), i18n.T("cmd.php.dev.stop_error", map[string]interface{}{"Error": err}))
	}

	fmt.Printf("%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.dev.all_stopped"))
	return nil
}

var (
	logsFollow  bool
	logsService string
)

func addPHPLogsCommand(parent *cobra.Command) {
	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: i18n.T("cmd.php.logs.short"),
		Long:  i18n.T("cmd.php.logs.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPHPLogs(logsService, logsFollow)
		},
	}

	logsCmd.Flags().BoolVar(&logsFollow, "follow", false, i18n.T("cmd.php.logs.flag.follow"))
	logsCmd.Flags().StringVar(&logsService, "service", "", i18n.T("cmd.php.logs.flag.service"))

	parent.AddCommand(logsCmd)
}

func runPHPLogs(service string, follow bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !phppkg.IsLaravelProject(cwd) {
		return fmt.Errorf(i18n.T("cmd.php.error.not_laravel_short"))
	}

	// Create a minimal server just to access logs
	server := phppkg.NewDevServer(phppkg.Options{Dir: cwd})

	logsReader, err := server.Logs(service, follow)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.get_logs"), err)
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
		Short: i18n.T("cmd.php.stop.short"),
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

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.stop.stopping"))

	// We need to find running processes
	// This is a simplified version - in practice you'd want to track PIDs
	server := phppkg.NewDevServer(phppkg.Options{Dir: cwd})
	if err := server.Stop(); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.stop_services"), err)
	}

	fmt.Printf("%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.dev.all_stopped"))
	return nil
}

func addPHPStatusCommand(parent *cobra.Command) {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: i18n.T("cmd.php.status.short"),
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
		return fmt.Errorf(i18n.T("cmd.php.error.not_laravel_short"))
	}

	appName := phppkg.GetLaravelAppName(cwd)
	if appName == "" {
		appName = "Laravel"
	}

	fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.status.project")), appName)

	// Detect available services
	services := phppkg.DetectServices(cwd)
	fmt.Printf("%s\n", dimStyle.Render(i18n.T("cmd.php.status.detected_services")))
	for _, svc := range services {
		style := getServiceStyle(string(svc))
		fmt.Printf("  %s %s\n", style.Render("*"), svc)
	}
	fmt.Println()

	// Package manager
	pm := phppkg.DetectPackageManager(cwd)
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.status.package_manager")), pm)

	// FrankenPHP status
	if phppkg.IsFrankenPHPProject(cwd) {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.status.octane_server")), "FrankenPHP")
	}

	// SSL status
	appURL := phppkg.GetLaravelAppURL(cwd)
	if appURL != "" {
		domain := phppkg.ExtractDomainFromURL(appURL)
		if phppkg.CertsExist(domain, phppkg.SSLOptions{}) {
			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.status.ssl_certs")), successStyle.Render(i18n.T("cmd.php.status.ssl_installed")))
		} else {
			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.status.ssl_certs")), dimStyle.Render(i18n.T("cmd.php.status.ssl_not_setup")))
		}
	}

	return nil
}

var sslDomain string

func addPHPSSLCommand(parent *cobra.Command) {
	sslCmd := &cobra.Command{
		Use:   "ssl",
		Short: i18n.T("cmd.php.ssl.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPHPSSL(sslDomain)
		},
	}

	sslCmd.Flags().StringVar(&sslDomain, "domain", "", i18n.T("cmd.php.ssl.flag.domain"))

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
		fmt.Printf("%s %s\n", errorStyle.Render(i18n.T("cmd.php.label.error")), i18n.T("cmd.php.ssl.mkcert_not_installed"))
		fmt.Printf("\n%s\n", i18n.T("cmd.php.ssl.install_with"))
		fmt.Printf("  %s\n", i18n.T("cmd.php.ssl.install_macos"))
		fmt.Printf("  %s\n", i18n.T("cmd.php.ssl.install_linux"))
		return fmt.Errorf(i18n.T("cmd.php.error.mkcert_not_installed"))
	}

	fmt.Printf("%s %s\n", dimStyle.Render("SSL:"), i18n.T("cmd.php.ssl.setting_up", map[string]interface{}{"Domain": domain}))

	// Check if certs already exist
	if phppkg.CertsExist(domain, phppkg.SSLOptions{}) {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.skip")), i18n.T("cmd.php.ssl.certs_exist"))

		certFile, keyFile, _ := phppkg.CertPaths(domain, phppkg.SSLOptions{})
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.ssl.cert_label")), certFile)
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.ssl.key_label")), keyFile)
		return nil
	}

	// Setup SSL
	if err := phppkg.SetupSSL(domain, phppkg.SSLOptions{}); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.ssl_setup"), err)
	}

	certFile, keyFile, _ := phppkg.CertPaths(domain, phppkg.SSLOptions{})

	fmt.Printf("%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.ssl.certs_created"))
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.ssl.cert_label")), certFile)
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.ssl.key_label")), keyFile)

	return nil
}

// Helper functions for dev commands

func printServiceStatuses(statuses []phppkg.ServiceStatus) {
	for _, s := range statuses {
		style := getServiceStyle(s.Name)
		var statusText string

		if s.Error != nil {
			statusText = phpStatusError.Render(i18n.T("cmd.php.status.error", map[string]interface{}{"Error": s.Error}))
		} else if s.Running {
			statusText = phpStatusRunning.Render(i18n.T("cmd.php.status.running"))
			if s.Port > 0 {
				statusText += dimStyle.Render(fmt.Sprintf(" (%s)", i18n.T("cmd.php.status.port", map[string]interface{}{"Port": s.Port})))
			}
			if s.PID > 0 {
				statusText += dimStyle.Render(fmt.Sprintf(" [%s]", i18n.T("cmd.php.status.pid", map[string]interface{}{"PID": s.PID})))
			}
		} else {
			statusText = phpStatusStopped.Render(i18n.T("cmd.php.status.stopped"))
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
