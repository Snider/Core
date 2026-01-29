package php

import (
	"context"
	"fmt"
	"os"
	"strings"

	phppkg "github.com/host-uk/core/pkg/php"
	"github.com/leaanthony/clir"
)

func addPHPBuildCommand(parent *clir.Command) {
	var (
		buildType  string
		imageName  string
		tag        string
		platform   string
		dockerfile string
		outputPath string
		format     string
		template   string
		noCache    bool
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
	if !phppkg.IsPHPProject(projectDir) {
		return fmt.Errorf("not a PHP project (missing composer.json)")
	}

	fmt.Printf("%s Building Docker image...\n\n", dimStyle.Render("PHP:"))

	// Show detected configuration
	config, err := phppkg.DetectDockerfileConfig(projectDir)
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
	buildOpts := phppkg.DockerBuildOptions{
		ProjectDir:   projectDir,
		ImageName:    opts.ImageName,
		Tag:          opts.Tag,
		Platform:     opts.Platform,
		Dockerfile:   opts.Dockerfile,
		NoBuildCache: opts.NoCache,
		Output:       os.Stdout,
	}

	if buildOpts.ImageName == "" {
		buildOpts.ImageName = phppkg.GetLaravelAppName(projectDir)
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

	if err := phppkg.BuildDocker(ctx, buildOpts); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("\n%s Docker image built successfully\n", successStyle.Render("Done:"))
	fmt.Printf("%s docker run -p 80:80 -p 443:443 %s:%s\n",
		dimStyle.Render("Run with:"),
		buildOpts.ImageName, buildOpts.Tag)

	return nil
}

func runPHPBuildLinuxKit(ctx context.Context, projectDir string, opts linuxKitBuildOptions) error {
	if !phppkg.IsPHPProject(projectDir) {
		return fmt.Errorf("not a PHP project (missing composer.json)")
	}

	fmt.Printf("%s Building LinuxKit image...\n\n", dimStyle.Render("PHP:"))

	buildOpts := phppkg.LinuxKitBuildOptions{
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

	if err := phppkg.BuildLinuxKit(ctx, buildOpts); err != nil {
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
				imageName = phppkg.GetLaravelAppName(cwd)
				if imageName != "" {
					imageName = strings.ToLower(strings.ReplaceAll(imageName, " ", "-"))
				}
			}
			if imageName == "" {
				return fmt.Errorf("--name is required: specify the Docker image name")
			}
		}

		ctx := context.Background()

		opts := phppkg.ServeOptions{
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

		if err := phppkg.ServeProduction(ctx, opts); err != nil {
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

		if err := phppkg.Shell(ctx, args[0]); err != nil {
			return fmt.Errorf("failed to open shell: %w", err)
		}

		return nil
	})
}
