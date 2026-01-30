package php

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/host-uk/core/pkg/i18n"
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
		Short: i18n.T("cmd.php.build.short"),
		Long:  i18n.T("cmd.php.build.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.working_dir"), err)
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

	buildCmd.Flags().StringVar(&buildType, "type", "", i18n.T("cmd.php.build.flag.type"))
	buildCmd.Flags().StringVar(&buildImageName, "name", "", i18n.T("cmd.php.build.flag.name"))
	buildCmd.Flags().StringVar(&buildTag, "tag", "", i18n.T("cmd.php.build.flag.tag"))
	buildCmd.Flags().StringVar(&buildPlatform, "platform", "", i18n.T("cmd.php.build.flag.platform"))
	buildCmd.Flags().StringVar(&buildDockerfile, "dockerfile", "", i18n.T("cmd.php.build.flag.dockerfile"))
	buildCmd.Flags().StringVar(&buildOutputPath, "output", "", i18n.T("cmd.php.build.flag.output"))
	buildCmd.Flags().StringVar(&buildFormat, "format", "", i18n.T("cmd.php.build.flag.format"))
	buildCmd.Flags().StringVar(&buildTemplate, "template", "", i18n.T("cmd.php.build.flag.template"))
	buildCmd.Flags().BoolVar(&buildNoCache, "no-cache", false, i18n.T("cmd.php.build.flag.no_cache"))

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
		return fmt.Errorf(i18n.T("cmd.php.error.not_php"))
	}

	fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.build.building_docker"))

	// Show detected configuration
	config, err := phppkg.DetectDockerfileConfig(projectDir)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.detect_config"), err)
	}

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.build.php_version")), config.PHPVersion)
	fmt.Printf("%s %v\n", dimStyle.Render(i18n.T("cmd.php.build.laravel")), config.IsLaravel)
	fmt.Printf("%s %v\n", dimStyle.Render(i18n.T("cmd.php.build.octane")), config.HasOctane)
	fmt.Printf("%s %v\n", dimStyle.Render(i18n.T("cmd.php.build.frontend")), config.HasAssets)
	if len(config.PHPExtensions) > 0 {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.build.extensions")), strings.Join(config.PHPExtensions, ", "))
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

	fmt.Printf("%s %s:%s\n", dimStyle.Render(i18n.T("cmd.php.build.image")), buildOpts.ImageName, buildOpts.Tag)
	if opts.Platform != "" {
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.build.platform")), opts.Platform)
	}
	fmt.Println()

	if err := phppkg.BuildDocker(ctx, buildOpts); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.build_failed"), err)
	}

	fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.build.docker_success"))
	fmt.Printf("%s docker run -p 80:80 -p 443:443 %s:%s\n",
		dimStyle.Render(i18n.T("cmd.php.build.docker_run_with")),
		buildOpts.ImageName, buildOpts.Tag)

	return nil
}

func runPHPBuildLinuxKit(ctx context.Context, projectDir string, opts linuxKitBuildOptions) error {
	if !phppkg.IsPHPProject(projectDir) {
		return fmt.Errorf(i18n.T("cmd.php.error.not_php"))
	}

	fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.build.building_linuxkit"))

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

	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.build.template")), buildOpts.Template)
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.build.format")), buildOpts.Format)
	fmt.Println()

	if err := phppkg.BuildLinuxKit(ctx, buildOpts); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.build_failed"), err)
	}

	fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.build.linuxkit_success"))
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
		Short: i18n.T("cmd.php.serve.short"),
		Long:  i18n.T("cmd.php.serve.long"),
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
					return fmt.Errorf(i18n.T("cmd.php.serve.name_required"))
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

			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.serve.running"))
			fmt.Printf("%s %s:%s\n", dimStyle.Render(i18n.T("cmd.php.build.image")), imageName, func() string {
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
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.start_container"), err)
			}

			if !serveDetach {
				fmt.Printf("\n%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.serve.stopped"))
			}

			return nil
		},
	}

	serveCmd.Flags().StringVar(&serveImageName, "name", "", i18n.T("cmd.php.serve.flag.name"))
	serveCmd.Flags().StringVar(&serveTag, "tag", "", i18n.T("cmd.php.serve.flag.tag"))
	serveCmd.Flags().StringVar(&serveContainerName, "container", "", i18n.T("cmd.php.serve.flag.container"))
	serveCmd.Flags().IntVar(&servePort, "port", 0, i18n.T("cmd.php.serve.flag.port"))
	serveCmd.Flags().IntVar(&serveHTTPSPort, "https-port", 0, i18n.T("cmd.php.serve.flag.https_port"))
	serveCmd.Flags().BoolVarP(&serveDetach, "detach", "d", false, i18n.T("cmd.php.serve.flag.detach"))
	serveCmd.Flags().StringVar(&serveEnvFile, "env-file", "", i18n.T("cmd.php.serve.flag.env_file"))

	parent.AddCommand(serveCmd)
}

func addPHPShellCommand(parent *cobra.Command) {
	shellCmd := &cobra.Command{
		Use:   "shell [container]",
		Short: i18n.T("cmd.php.shell.short"),
		Long:  i18n.T("cmd.php.shell.long"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.shell.opening", map[string]interface{}{"Container": args[0]}))

			if err := phppkg.Shell(ctx, args[0]); err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.open_shell"), err)
			}

			return nil
		},
	}

	parent.AddCommand(shellCmd)
}
