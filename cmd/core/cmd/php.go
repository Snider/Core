package cmd

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
	"github.com/host-uk/core/pkg/php"
	"github.com/leaanthony/clir"
)

// Service colors for log output
var (
	phpFrankenPHPStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6366f1")) // indigo-500

	phpViteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eab308")) // yellow-500

	phpHorizonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f97316")) // orange-500

	phpReverbStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8b5cf6")) // violet-500

	phpRedisStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ef4444")) // red-500

	phpStatusRunning = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22c55e")). // green-500
				Bold(true)

	phpStatusStopped = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6b7280")) // gray-500

	phpStatusError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ef4444")). // red-500
			Bold(true)
)

// AddPHPCommands adds PHP/Laravel development commands.
func AddPHPCommands(parent *clir.Cli) {
	phpCmd := parent.NewSubCommand("php", "Laravel/PHP development tools")
	phpCmd.LongDescription("Manage Laravel development environment with FrankenPHP.\n\n" +
		"Services orchestrated:\n" +
		"  - FrankenPHP/Octane (port 8000, HTTPS on 443)\n" +
		"  - Vite dev server (port 5173)\n" +
		"  - Laravel Horizon (queue workers)\n" +
		"  - Laravel Reverb (WebSocket, port 8080)\n" +
		"  - Redis (port 6379)")

	addPHPDevCommand(phpCmd)
	addPHPLogsCommand(phpCmd)
	addPHPStopCommand(phpCmd)
	addPHPStatusCommand(phpCmd)
	addPHPSSLCommand(phpCmd)
	addPHPBuildCommand(phpCmd)
	addPHPServeCommand(phpCmd)
	addPHPShellCommand(phpCmd)
	addPHPTestCommand(phpCmd)
	addPHPFmtCommand(phpCmd)
	addPHPAnalyseCommand(phpCmd)
}

func addPHPDevCommand(parent *clir.Command) {
	var (
		noVite    bool
		noHorizon bool
		noReverb  bool
		noRedis   bool
		https     bool
		domain    string
		port      int
	)

	devCmd := parent.NewSubCommand("dev", "Start Laravel development environment")
	devCmd.LongDescription("Starts all detected Laravel services.\n\n" +
		"Auto-detects:\n" +
		"  - Vite (vite.config.js/ts)\n" +
		"  - Horizon (config/horizon.php)\n" +
		"  - Reverb (config/reverb.php)\n" +
		"  - Redis (from .env)")

	devCmd.BoolFlag("no-vite", "Skip Vite dev server", &noVite)
	devCmd.BoolFlag("no-horizon", "Skip Laravel Horizon", &noHorizon)
	devCmd.BoolFlag("no-reverb", "Skip Laravel Reverb", &noReverb)
	devCmd.BoolFlag("no-redis", "Skip Redis server", &noRedis)
	devCmd.BoolFlag("https", "Enable HTTPS with mkcert", &https)
	devCmd.StringFlag("domain", "Domain for SSL certificate (default: from APP_URL or localhost)", &domain)
	devCmd.IntFlag("port", "FrankenPHP port (default: 8000)", &port)

	devCmd.Action(func() error {
		return runPHPDev(phpDevOptions{
			NoVite:    noVite,
			NoHorizon: noHorizon,
			NoReverb:  noReverb,
			NoRedis:   noRedis,
			HTTPS:     https,
			Domain:    domain,
			Port:      port,
		})
	})
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
	if !php.IsLaravelProject(cwd) {
		return fmt.Errorf("not a Laravel project (missing artisan or laravel/framework)")
	}

	// Get app name for display
	appName := php.GetLaravelAppName(cwd)
	if appName == "" {
		appName = "Laravel"
	}

	fmt.Printf("%s Starting %s development environment\n\n", dimStyle.Render("PHP:"), appName)

	// Detect services
	services := php.DetectServices(cwd)
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

	devOpts := php.Options{
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
	server := php.NewDevServer(devOpts)

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
	appURL := php.GetLaravelAppURL(cwd)
	if appURL == "" {
		if opts.HTTPS {
			appURL = fmt.Sprintf("https://localhost:%d", port)
		} else {
			appURL = fmt.Sprintf("http://localhost:%d", port)
		}
	}
	fmt.Printf("%s %s\n", dimStyle.Render("App URL:"), linkStyle.Render(appURL))

	// Check for Vite
	if !opts.NoVite && containsService(services, php.ServiceVite) {
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

func addPHPLogsCommand(parent *clir.Command) {
	var follow bool
	var service string

	logsCmd := parent.NewSubCommand("logs", "View service logs")
	logsCmd.LongDescription("Stream logs from Laravel services.\n\n" +
		"Services: frankenphp, vite, horizon, reverb, redis")

	logsCmd.BoolFlag("follow", "Follow log output", &follow)
	logsCmd.StringFlag("service", "Specific service (default: all)", &service)

	logsCmd.Action(func() error {
		return runPHPLogs(service, follow)
	})
}

func runPHPLogs(service string, follow bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !php.IsLaravelProject(cwd) {
		return fmt.Errorf("not a Laravel project")
	}

	// Create a minimal server just to access logs
	server := php.NewDevServer(php.Options{Dir: cwd})

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

func addPHPStopCommand(parent *clir.Command) {
	stopCmd := parent.NewSubCommand("stop", "Stop all Laravel services")

	stopCmd.Action(func() error {
		return runPHPStop()
	})
}

func runPHPStop() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	fmt.Printf("%s Stopping services...\n", dimStyle.Render("PHP:"))

	// We need to find running processes
	// This is a simplified version - in practice you'd want to track PIDs
	server := php.NewDevServer(php.Options{Dir: cwd})
	if err := server.Stop(); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	fmt.Printf("%s All services stopped\n", successStyle.Render("Done:"))
	return nil
}

func addPHPStatusCommand(parent *clir.Command) {
	statusCmd := parent.NewSubCommand("status", "Show service status")

	statusCmd.Action(func() error {
		return runPHPStatus()
	})
}

func runPHPStatus() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !php.IsLaravelProject(cwd) {
		return fmt.Errorf("not a Laravel project")
	}

	appName := php.GetLaravelAppName(cwd)
	if appName == "" {
		appName = "Laravel"
	}

	fmt.Printf("%s %s\n\n", dimStyle.Render("Project:"), appName)

	// Detect available services
	services := php.DetectServices(cwd)
	fmt.Printf("%s\n", dimStyle.Render("Detected services:"))
	for _, svc := range services {
		style := getServiceStyle(string(svc))
		fmt.Printf("  %s %s\n", style.Render("*"), svc)
	}
	fmt.Println()

	// Package manager
	pm := php.DetectPackageManager(cwd)
	fmt.Printf("%s %s\n", dimStyle.Render("Package manager:"), pm)

	// FrankenPHP status
	if php.IsFrankenPHPProject(cwd) {
		fmt.Printf("%s %s\n", dimStyle.Render("Octane server:"), "FrankenPHP")
	}

	// SSL status
	appURL := php.GetLaravelAppURL(cwd)
	if appURL != "" {
		domain := php.ExtractDomainFromURL(appURL)
		if php.CertsExist(domain, php.SSLOptions{}) {
			fmt.Printf("%s %s\n", dimStyle.Render("SSL certificates:"), successStyle.Render("installed"))
		} else {
			fmt.Printf("%s %s\n", dimStyle.Render("SSL certificates:"), dimStyle.Render("not setup"))
		}
	}

	return nil
}

func addPHPSSLCommand(parent *clir.Command) {
	var domain string

	sslCmd := parent.NewSubCommand("ssl", "Setup SSL certificates with mkcert")

	sslCmd.StringFlag("domain", "Domain for certificate (default: from APP_URL)", &domain)

	sslCmd.Action(func() error {
		return runPHPSSL(domain)
	})
}

func runPHPSSL(domain string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Get domain from APP_URL if not specified
	if domain == "" {
		appURL := php.GetLaravelAppURL(cwd)
		if appURL != "" {
			domain = php.ExtractDomainFromURL(appURL)
		}
	}
	if domain == "" {
		domain = "localhost"
	}

	// Check if mkcert is installed
	if !php.IsMkcertInstalled() {
		fmt.Printf("%s mkcert is not installed\n", errorStyle.Render("Error:"))
		fmt.Println("\nInstall with:")
		fmt.Println("  macOS:  brew install mkcert")
		fmt.Println("  Linux:  see https://github.com/FiloSottile/mkcert")
		return fmt.Errorf("mkcert not installed")
	}

	fmt.Printf("%s Setting up SSL for %s\n", dimStyle.Render("SSL:"), domain)

	// Check if certs already exist
	if php.CertsExist(domain, php.SSLOptions{}) {
		fmt.Printf("%s Certificates already exist\n", dimStyle.Render("Skip:"))

		certFile, keyFile, _ := php.CertPaths(domain, php.SSLOptions{})
		fmt.Printf("%s %s\n", dimStyle.Render("Cert:"), certFile)
		fmt.Printf("%s %s\n", dimStyle.Render("Key:"), keyFile)
		return nil
	}

	// Setup SSL
	if err := php.SetupSSL(domain, php.SSLOptions{}); err != nil {
		return fmt.Errorf("failed to setup SSL: %w", err)
	}

	certFile, keyFile, _ := php.CertPaths(domain, php.SSLOptions{})

	fmt.Printf("%s SSL certificates created\n", successStyle.Render("Done:"))
	fmt.Printf("%s %s\n", dimStyle.Render("Cert:"), certFile)
	fmt.Printf("%s %s\n", dimStyle.Render("Key:"), keyFile)

	return nil
}

// Helper functions

func printServiceStatuses(statuses []php.ServiceStatus) {
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

func containsService(services []php.DetectedService, target php.DetectedService) bool {
	for _, s := range services {
		if s == target {
			return true
		}
	}
	return false
}

func addPHPBuildCommand(parent *clir.Command) {
	var (
		buildType    string
		imageName    string
		tag          string
		platform     string
		dockerfile   string
		outputPath   string
		format       string
		template     string
		noCache      bool
	)

	buildCmd := parent.NewSubCommand("build", "Build Docker or LinuxKit image")
	buildCmd.LongDescription("Build a production-ready container image for the PHP project.\n\n" +
		"By default, builds a Docker image using FrankenPHP.\n" +
		"Use --type linuxkit to build a LinuxKit VM image instead.\n\n" +
		"Examples:\n" +
		"  core php build                           # Build Docker image\n" +
		"  core php build --name myapp --tag v1.0   # Build with custom name/tag\n" +
		"  core php build --type linuxkit           # Build LinuxKit image\n" +
		"  core php build --type linuxkit --format iso  # Build ISO image")

	buildCmd.StringFlag("type", "Build type: docker (default) or linuxkit", &buildType)
	buildCmd.StringFlag("name", "Image name (default: project directory name)", &imageName)
	buildCmd.StringFlag("tag", "Image tag (default: latest)", &tag)
	buildCmd.StringFlag("platform", "Target platform (e.g., linux/amd64, linux/arm64)", &platform)
	buildCmd.StringFlag("dockerfile", "Path to custom Dockerfile", &dockerfile)
	buildCmd.StringFlag("output", "Output path for LinuxKit image", &outputPath)
	buildCmd.StringFlag("format", "LinuxKit output format: qcow2 (default), iso, raw, vmdk", &format)
	buildCmd.StringFlag("template", "LinuxKit template name (default: server-php)", &template)
	buildCmd.BoolFlag("no-cache", "Build without cache", &noCache)

	buildCmd.Action(func() error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		ctx := context.Background()

		switch strings.ToLower(buildType) {
		case "linuxkit":
			return runPHPBuildLinuxKit(ctx, cwd, linuxKitBuildOptions{
				OutputPath: outputPath,
				Format:     format,
				Template:   template,
			})
		default:
			return runPHPBuildDocker(ctx, cwd, dockerBuildOptions{
				ImageName:  imageName,
				Tag:        tag,
				Platform:   platform,
				Dockerfile: dockerfile,
				NoCache:    noCache,
			})
		}
	})
}

type dockerBuildOptions struct {
	ImageName  string
	Tag        string
	Platform   string
	Dockerfile string
	NoCache    bool
}

type linuxKitBuildOptions struct {
	OutputPath string
	Format     string
	Template   string
}

func runPHPBuildDocker(ctx context.Context, projectDir string, opts dockerBuildOptions) error {
	if !php.IsPHPProject(projectDir) {
		return fmt.Errorf("not a PHP project (missing composer.json)")
	}

	fmt.Printf("%s Building Docker image...\n\n", dimStyle.Render("PHP:"))

	// Show detected configuration
	config, err := php.DetectDockerfileConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to detect project configuration: %w", err)
	}

	fmt.Printf("%s %s\n", dimStyle.Render("PHP Version:"), config.PHPVersion)
	fmt.Printf("%s %v\n", dimStyle.Render("Laravel:"), config.IsLaravel)
	fmt.Printf("%s %v\n", dimStyle.Render("Octane:"), config.HasOctane)
	fmt.Printf("%s %v\n", dimStyle.Render("Frontend:"), config.HasAssets)
	if len(config.PHPExtensions) > 0 {
		fmt.Printf("%s %s\n", dimStyle.Render("Extensions:"), strings.Join(config.PHPExtensions, ", "))
	}
	fmt.Println()

	// Build options
	buildOpts := php.DockerBuildOptions{
		ProjectDir:   projectDir,
		ImageName:    opts.ImageName,
		Tag:          opts.Tag,
		Platform:     opts.Platform,
		Dockerfile:   opts.Dockerfile,
		NoBuildCache: opts.NoCache,
		Output:       os.Stdout,
	}

	if buildOpts.ImageName == "" {
		buildOpts.ImageName = php.GetLaravelAppName(projectDir)
		if buildOpts.ImageName == "" {
			buildOpts.ImageName = "php-app"
		}
		// Sanitize for Docker
		buildOpts.ImageName = strings.ToLower(strings.ReplaceAll(buildOpts.ImageName, " ", "-"))
	}

	if buildOpts.Tag == "" {
		buildOpts.Tag = "latest"
	}

	fmt.Printf("%s %s:%s\n", dimStyle.Render("Image:"), buildOpts.ImageName, buildOpts.Tag)
	if opts.Platform != "" {
		fmt.Printf("%s %s\n", dimStyle.Render("Platform:"), opts.Platform)
	}
	fmt.Println()

	if err := php.BuildDocker(ctx, buildOpts); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("\n%s Docker image built successfully\n", successStyle.Render("Done:"))
	fmt.Printf("%s docker run -p 80:80 -p 443:443 %s:%s\n",
		dimStyle.Render("Run with:"),
		buildOpts.ImageName, buildOpts.Tag)

	return nil
}

func runPHPBuildLinuxKit(ctx context.Context, projectDir string, opts linuxKitBuildOptions) error {
	if !php.IsPHPProject(projectDir) {
		return fmt.Errorf("not a PHP project (missing composer.json)")
	}

	fmt.Printf("%s Building LinuxKit image...\n\n", dimStyle.Render("PHP:"))

	buildOpts := php.LinuxKitBuildOptions{
		ProjectDir: projectDir,
		OutputPath: opts.OutputPath,
		Format:     opts.Format,
		Template:   opts.Template,
		Output:     os.Stdout,
	}

	if buildOpts.Format == "" {
		buildOpts.Format = "qcow2"
	}
	if buildOpts.Template == "" {
		buildOpts.Template = "server-php"
	}

	fmt.Printf("%s %s\n", dimStyle.Render("Template:"), buildOpts.Template)
	fmt.Printf("%s %s\n", dimStyle.Render("Format:"), buildOpts.Format)
	fmt.Println()

	if err := php.BuildLinuxKit(ctx, buildOpts); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("\n%s LinuxKit image built successfully\n", successStyle.Render("Done:"))
	return nil
}

func addPHPServeCommand(parent *clir.Command) {
	var (
		imageName     string
		tag           string
		containerName string
		port          int
		httpsPort     int
		detach        bool
		envFile       string
	)

	serveCmd := parent.NewSubCommand("serve", "Run production container")
	serveCmd.LongDescription("Run a production PHP container.\n\n" +
		"This starts the built Docker image in production mode.\n\n" +
		"Examples:\n" +
		"  core php serve --name myapp              # Run container\n" +
		"  core php serve --name myapp -d           # Run detached\n" +
		"  core php serve --name myapp --port 8080  # Custom port")

	serveCmd.StringFlag("name", "Docker image name (required)", &imageName)
	serveCmd.StringFlag("tag", "Image tag (default: latest)", &tag)
	serveCmd.StringFlag("container", "Container name", &containerName)
	serveCmd.IntFlag("port", "HTTP port (default: 80)", &port)
	serveCmd.IntFlag("https-port", "HTTPS port (default: 443)", &httpsPort)
	serveCmd.BoolFlag("d", "Run in detached mode", &detach)
	serveCmd.StringFlag("env-file", "Path to environment file", &envFile)

	serveCmd.Action(func() error {
		if imageName == "" {
			// Try to detect from current directory
			cwd, err := os.Getwd()
			if err == nil {
				imageName = php.GetLaravelAppName(cwd)
				if imageName != "" {
					imageName = strings.ToLower(strings.ReplaceAll(imageName, " ", "-"))
				}
			}
			if imageName == "" {
				return fmt.Errorf("--name is required: specify the Docker image name")
			}
		}

		ctx := context.Background()

		opts := php.ServeOptions{
			ImageName:     imageName,
			Tag:           tag,
			ContainerName: containerName,
			Port:          port,
			HTTPSPort:     httpsPort,
			Detach:        detach,
			EnvFile:       envFile,
			Output:        os.Stdout,
		}

		fmt.Printf("%s Running production container...\n\n", dimStyle.Render("PHP:"))
		fmt.Printf("%s %s:%s\n", dimStyle.Render("Image:"), imageName, func() string {
			if tag == "" {
				return "latest"
			}
			return tag
		}())

		effectivePort := port
		if effectivePort == 0 {
			effectivePort = 80
		}
		effectiveHTTPSPort := httpsPort
		if effectiveHTTPSPort == 0 {
			effectiveHTTPSPort = 443
		}

		fmt.Printf("%s http://localhost:%d, https://localhost:%d\n",
			dimStyle.Render("Ports:"), effectivePort, effectiveHTTPSPort)
		fmt.Println()

		if err := php.ServeProduction(ctx, opts); err != nil {
			return fmt.Errorf("failed to start container: %w", err)
		}

		if !detach {
			fmt.Printf("\n%s Container stopped\n", dimStyle.Render("PHP:"))
		}

		return nil
	})
}

func addPHPShellCommand(parent *clir.Command) {
	shellCmd := parent.NewSubCommand("shell", "Open shell in running container")
	shellCmd.LongDescription("Open an interactive shell in a running PHP container.\n\n" +
		"Examples:\n" +
		"  core php shell abc123   # Shell into container by ID\n" +
		"  core php shell myapp    # Shell into container by name")

	shellCmd.Action(func() error {
		args := shellCmd.OtherArgs()
		if len(args) == 0 {
			return fmt.Errorf("container ID or name is required")
		}

		ctx := context.Background()

		fmt.Printf("%s Opening shell in container %s...\n", dimStyle.Render("PHP:"), args[0])

		if err := php.Shell(ctx, args[0]); err != nil {
			return fmt.Errorf("failed to open shell: %w", err)
		}

		return nil
	})
}

func addPHPTestCommand(parent *clir.Command) {
	var (
		parallel bool
		coverage bool
		filter   string
		group    string
	)

	testCmd := parent.NewSubCommand("test", "Run PHP tests (PHPUnit/Pest)")
	testCmd.LongDescription("Run PHP tests using PHPUnit or Pest.\n\n" +
		"Auto-detects Pest if tests/Pest.php exists, otherwise uses PHPUnit.\n\n" +
		"Examples:\n" +
		"  core php test                    # Run all tests\n" +
		"  core php test --parallel         # Run tests in parallel\n" +
		"  core php test --coverage         # Run with coverage\n" +
		"  core php test --filter UserTest  # Filter by test name")

	testCmd.BoolFlag("parallel", "Run tests in parallel", &parallel)
	testCmd.BoolFlag("coverage", "Generate code coverage", &coverage)
	testCmd.StringFlag("filter", "Filter tests by name pattern", &filter)
	testCmd.StringFlag("group", "Run only tests in specified group", &group)

	testCmd.Action(func() error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		if !php.IsPHPProject(cwd) {
			return fmt.Errorf("not a PHP project (missing composer.json)")
		}

		// Detect test runner
		runner := php.DetectTestRunner(cwd)
		fmt.Printf("%s Running tests with %s\n\n", dimStyle.Render("PHP:"), runner)

		ctx := context.Background()

		opts := php.TestOptions{
			Dir:      cwd,
			Filter:   filter,
			Parallel: parallel,
			Coverage: coverage,
			Output:   os.Stdout,
		}

		if group != "" {
			opts.Groups = []string{group}
		}

		if err := php.RunTests(ctx, opts); err != nil {
			return fmt.Errorf("tests failed: %w", err)
		}

		return nil
	})
}

func addPHPFmtCommand(parent *clir.Command) {
	var (
		fix  bool
		diff bool
	)

	fmtCmd := parent.NewSubCommand("fmt", "Format PHP code with Laravel Pint")
	fmtCmd.LongDescription("Format PHP code using Laravel Pint.\n\n" +
		"Examples:\n" +
		"  core php fmt           # Check formatting (dry-run)\n" +
		"  core php fmt --fix     # Auto-fix formatting issues\n" +
		"  core php fmt --diff    # Show diff of changes")

	fmtCmd.BoolFlag("fix", "Auto-fix formatting issues", &fix)
	fmtCmd.BoolFlag("diff", "Show diff of changes", &diff)

	fmtCmd.Action(func() error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		if !php.IsPHPProject(cwd) {
			return fmt.Errorf("not a PHP project (missing composer.json)")
		}

		// Detect formatter
		formatter, found := php.DetectFormatter(cwd)
		if !found {
			return fmt.Errorf("no formatter found (install Laravel Pint: composer require laravel/pint --dev)")
		}

		action := "Checking"
		if fix {
			action = "Formatting"
		}
		fmt.Printf("%s %s code with %s\n\n", dimStyle.Render("PHP:"), action, formatter)

		ctx := context.Background()

		opts := php.FormatOptions{
			Dir:    cwd,
			Fix:    fix,
			Diff:   diff,
			Output: os.Stdout,
		}

		// Get any additional paths from args
		if args := fmtCmd.OtherArgs(); len(args) > 0 {
			opts.Paths = args
		}

		if err := php.Format(ctx, opts); err != nil {
			if fix {
				return fmt.Errorf("formatting failed: %w", err)
			}
			return fmt.Errorf("formatting issues found: %w", err)
		}

		if fix {
			fmt.Printf("\n%s Code formatted successfully\n", successStyle.Render("Done:"))
		} else {
			fmt.Printf("\n%s No formatting issues found\n", successStyle.Render("Done:"))
		}

		return nil
	})
}

func addPHPAnalyseCommand(parent *clir.Command) {
	var (
		level  int
		memory string
	)

	analyseCmd := parent.NewSubCommand("analyse", "Run PHPStan static analysis")
	analyseCmd.LongDescription("Run PHPStan or Larastan static analysis.\n\n" +
		"Auto-detects Larastan if installed, otherwise uses PHPStan.\n\n" +
		"Examples:\n" +
		"  core php analyse              # Run analysis\n" +
		"  core php analyse --level 9    # Run at max strictness\n" +
		"  core php analyse --memory 2G  # Increase memory limit")

	analyseCmd.IntFlag("level", "PHPStan analysis level (0-9)", &level)
	analyseCmd.StringFlag("memory", "Memory limit (e.g., 2G)", &memory)

	analyseCmd.Action(func() error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		if !php.IsPHPProject(cwd) {
			return fmt.Errorf("not a PHP project (missing composer.json)")
		}

		// Detect analyser
		analyser, found := php.DetectAnalyser(cwd)
		if !found {
			return fmt.Errorf("no static analyser found (install PHPStan: composer require phpstan/phpstan --dev)")
		}

		fmt.Printf("%s Running static analysis with %s\n\n", dimStyle.Render("PHP:"), analyser)

		ctx := context.Background()

		opts := php.AnalyseOptions{
			Dir:    cwd,
			Level:  level,
			Memory: memory,
			Output: os.Stdout,
		}

		// Get any additional paths from args
		if args := analyseCmd.OtherArgs(); len(args) > 0 {
			opts.Paths = args
		}

		if err := php.Analyse(ctx, opts); err != nil {
			return fmt.Errorf("analysis found issues: %w", err)
		}

		fmt.Printf("\n%s No issues found\n", successStyle.Render("Done:"))
		return nil
	})
}
