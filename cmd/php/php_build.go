package php

import (
	"context"
	"fmt"
	"os"
	"strings"

	phppkg "github.com/host-uk/core/pkg/php"
	"github.com/spf13/cobra"
)

var (
	buildType       string
	buildImageName  string
	buildTag        string
	buildPlatform   string
	buildDockerfile string
	buildOutputPath string
	buildFormat     string
	buildTemplate   string
	buildNoCache    bool
)

func addPHPBuildCommand(parent *cobra.Command) {
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build Docker or LinuxKit image",
		Long: "Build a production-ready container image for the PHP project.\n\n" +
			"By default, builds a Docker image using FrankenPHP.\n" +
			"Use --type linuxkit to build a LinuxKit VM image instead.\n\n" +
			"Examples:\n" +
			"  core php build                           # Build Docker image\n" +
			"  core php build --name myapp --tag v1.0   # Build with custom name/tag\n" +
			"  core php build --type linuxkit           # Build LinuxKit image\n" +
			"  core php build --type linuxkit --format iso  # Build ISO image",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			ctx := context.Background()

			switch strings.ToLower(buildType) {
			case "linuxkit":
				return runPHPBuildLinuxKit(ctx, cwd, linuxKitBuildOptions{
					OutputPath: buildOutputPath,
					Format:     buildFormat,
					Template:   buildTemplate,
				})
			default:
				return runPHPBuildDocker(ctx, cwd, dockerBuildOptions{
					ImageName:  buildImageName,
					Tag:        buildTag,
					Platform:   buildPlatform,
					Dockerfile: buildDockerfile,
					NoCache:    buildNoCache,
				})
			}
		},
	}

	buildCmd.Flags().StringVar(&buildType, "type", "", "Build type: docker (default) or linuxkit")
	buildCmd.Flags().StringVar(&buildImageName, "name", "", "Image name (default: project directory name)")
	buildCmd.Flags().StringVar(&buildTag, "tag", "", "Image tag (default: latest)")
	buildCmd.Flags().StringVar(&buildPlatform, "platform", "", "Target platform (e.g., linux/amd64, linux/arm64)")
	buildCmd.Flags().StringVar(&buildDockerfile, "dockerfile", "", "Path to custom Dockerfile")
	buildCmd.Flags().StringVar(&buildOutputPath, "output", "", "Output path for LinuxKit image")
	buildCmd.Flags().StringVar(&buildFormat, "format", "", "LinuxKit output format: qcow2 (default), iso, raw, vmdk")
	buildCmd.Flags().StringVar(&buildTemplate, "template", "", "LinuxKit template name (default: server-php)")
	buildCmd.Flags().BoolVar(&buildNoCache, "no-cache", false, "Build without cache")

	parent.AddCommand(buildCmd)
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

var (
	serveImageName     string
	serveTag           string
	serveContainerName string
	servePort          int
	serveHTTPSPort     int
	serveDetach        bool
	serveEnvFile       string
)

func addPHPServeCommand(parent *cobra.Command) {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Run production container",
		Long: "Run a production PHP container.\n\n" +
			"This starts the built Docker image in production mode.\n\n" +
			"Examples:\n" +
			"  core php serve --name myapp              # Run container\n" +
			"  core php serve --name myapp -d           # Run detached\n" +
			"  core php serve --name myapp --port 8080  # Custom port",
		RunE: func(cmd *cobra.Command, args []string) error {
			imageName := serveImageName
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
				Tag:           serveTag,
				ContainerName: serveContainerName,
				Port:          servePort,
				HTTPSPort:     serveHTTPSPort,
				Detach:        serveDetach,
				EnvFile:       serveEnvFile,
				Output:        os.Stdout,
			}

			fmt.Printf("%s Running production container...\n\n", dimStyle.Render("PHP:"))
			fmt.Printf("%s %s:%s\n", dimStyle.Render("Image:"), imageName, func() string {
				if serveTag == "" {
					return "latest"
				}
				return serveTag
			}())

			effectivePort := servePort
			if effectivePort == 0 {
				effectivePort = 80
			}
			effectiveHTTPSPort := serveHTTPSPort
			if effectiveHTTPSPort == 0 {
				effectiveHTTPSPort = 443
			}

			fmt.Printf("%s http://localhost:%d, https://localhost:%d\n",
				dimStyle.Render("Ports:"), effectivePort, effectiveHTTPSPort)
			fmt.Println()

			if err := phppkg.ServeProduction(ctx, opts); err != nil {
				return fmt.Errorf("failed to start container: %w", err)
			}

			if !serveDetach {
				fmt.Printf("\n%s Container stopped\n", dimStyle.Render("PHP:"))
			}

			return nil
		},
	}

	serveCmd.Flags().StringVar(&serveImageName, "name", "", "Docker image name (required)")
	serveCmd.Flags().StringVar(&serveTag, "tag", "", "Image tag (default: latest)")
	serveCmd.Flags().StringVar(&serveContainerName, "container", "", "Container name")
	serveCmd.Flags().IntVar(&servePort, "port", 0, "HTTP port (default: 80)")
	serveCmd.Flags().IntVar(&serveHTTPSPort, "https-port", 0, "HTTPS port (default: 443)")
	serveCmd.Flags().BoolVarP(&serveDetach, "detach", "d", false, "Run in detached mode")
	serveCmd.Flags().StringVar(&serveEnvFile, "env-file", "", "Path to environment file")

	parent.AddCommand(serveCmd)
}

func addPHPShellCommand(parent *cobra.Command) {
	shellCmd := &cobra.Command{
		Use:   "shell [container]",
		Short: "Open shell in running container",
		Long: "Open an interactive shell in a running PHP container.\n\n" +
			"Examples:\n" +
			"  core php shell abc123   # Shell into container by ID\n" +
			"  core php shell myapp    # Shell into container by name",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			fmt.Printf("%s Opening shell in container %s...\n", dimStyle.Render("PHP:"), args[0])

			if err := phppkg.Shell(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to open shell: %w", err)
			}

			return nil
		},
	}

	parent.AddCommand(shellCmd)
}
