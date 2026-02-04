// Package publishers provides release publishing implementations.
package publishers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LinuxKitConfig holds configuration for the LinuxKit publisher.
type LinuxKitConfig struct {
	// Config is the path to the LinuxKit YAML configuration file.
	Config string `yaml:"config"`
	// Formats are the output formats to build.
	// Supported: iso, iso-bios, iso-efi, raw, raw-bios, raw-efi,
	//            qcow2, qcow2-bios, qcow2-efi, vmdk, vhd, gcp, aws,
	//            docker (tarball for `docker load`), tar, kernel+initrd
	Formats []string `yaml:"formats"`
	// Platforms are the target platforms (linux/amd64, linux/arm64).
	Platforms []string `yaml:"platforms"`
}

// LinuxKitPublisher builds and publishes LinuxKit images.
type LinuxKitPublisher struct{}

// NewLinuxKitPublisher creates a new LinuxKit publisher.
func NewLinuxKitPublisher() *LinuxKitPublisher {
	return &LinuxKitPublisher{}
}

// Name returns the publisher's identifier.
func (p *LinuxKitPublisher) Name() string {
	return "linuxkit"
}

// Publish builds LinuxKit images and uploads them to the GitHub release.
func (p *LinuxKitPublisher) Publish(ctx context.Context, release *Release, pubCfg PublisherConfig, relCfg ReleaseConfig, dryRun bool) error {
	// Validate linuxkit CLI is available
	if err := validateLinuxKitCli(); err != nil {
		return err
	}

	// Parse LinuxKit-specific config from publisher config
	lkCfg := p.parseConfig(pubCfg, release.ProjectDir)

	// Validate config file exists
	if !release.FS.Exists(lkCfg.Config) {
		return fmt.Errorf("linuxkit.Publish: config file not found: %s", lkCfg.Config)
	}

	// Determine repository for artifact upload
	repo := ""
	if relCfg != nil {
		repo = relCfg.GetRepository()
	}
	if repo == "" {
		detectedRepo, err := detectRepository(release.ProjectDir)
		if err != nil {
			return fmt.Errorf("linuxkit.Publish: could not determine repository: %w", err)
		}
		repo = detectedRepo
	}

	if dryRun {
		return p.dryRunPublish(release, lkCfg, repo)
	}

	return p.executePublish(ctx, release, lkCfg, repo)
}

// parseConfig extracts LinuxKit-specific configuration.
func (p *LinuxKitPublisher) parseConfig(pubCfg PublisherConfig, projectDir string) LinuxKitConfig {
	cfg := LinuxKitConfig{
		Config:    filepath.Join(projectDir, ".core", "linuxkit", "server.yml"),
		Formats:   []string{"iso"},
		Platforms: []string{"linux/amd64"},
	}

	// Override from extended config if present
	if ext, ok := pubCfg.Extended.(map[string]any); ok {
		if configPath, ok := ext["config"].(string); ok && configPath != "" {
			if filepath.IsAbs(configPath) {
				cfg.Config = configPath
			} else {
				cfg.Config = filepath.Join(projectDir, configPath)
			}
		}
		if formats, ok := ext["formats"].([]any); ok && len(formats) > 0 {
			cfg.Formats = make([]string, 0, len(formats))
			for _, f := range formats {
				if s, ok := f.(string); ok {
					cfg.Formats = append(cfg.Formats, s)
				}
			}
		}
		if platforms, ok := ext["platforms"].([]any); ok && len(platforms) > 0 {
			cfg.Platforms = make([]string, 0, len(platforms))
			for _, p := range platforms {
				if s, ok := p.(string); ok {
					cfg.Platforms = append(cfg.Platforms, s)
				}
			}
		}
	}

	return cfg
}

// dryRunPublish shows what would be done without actually building.
func (p *LinuxKitPublisher) dryRunPublish(release *Release, cfg LinuxKitConfig, repo string) error {
	fmt.Println()
	fmt.Println("=== DRY RUN: LinuxKit Build & Publish ===")
	fmt.Println()
	fmt.Printf("Repository:    %s\n", repo)
	fmt.Printf("Version:       %s\n", release.Version)
	fmt.Printf("Config:        %s\n", cfg.Config)
	fmt.Printf("Formats:       %s\n", strings.Join(cfg.Formats, ", "))
	fmt.Printf("Platforms:     %s\n", strings.Join(cfg.Platforms, ", "))
	fmt.Println()

	outputDir := filepath.Join(release.ProjectDir, "dist", "linuxkit")
	baseName := p.buildBaseName(release.Version)

	fmt.Println("Would execute commands:")
	for _, platform := range cfg.Platforms {
		parts := strings.Split(platform, "/")
		arch := "amd64"
		if len(parts) == 2 {
			arch = parts[1]
		}

		for _, format := range cfg.Formats {
			outputName := fmt.Sprintf("%s-%s", baseName, arch)
			args := p.buildLinuxKitArgs(cfg.Config, format, outputName, outputDir, arch)
			fmt.Printf("  linuxkit %s\n", strings.Join(args, " "))
		}
	}
	fmt.Println()

	fmt.Println("Would upload artifacts to release:")
	for _, platform := range cfg.Platforms {
		parts := strings.Split(platform, "/")
		arch := "amd64"
		if len(parts) == 2 {
			arch = parts[1]
		}

		for _, format := range cfg.Formats {
			outputName := fmt.Sprintf("%s-%s", baseName, arch)
			artifactPath := p.getArtifactPath(outputDir, outputName, format)
			fmt.Printf("  - %s\n", filepath.Base(artifactPath))
			if format == "docker" {
				fmt.Printf("    Usage: docker load < %s\n", filepath.Base(artifactPath))
			}
		}
	}

	fmt.Println()
	fmt.Println("=== END DRY RUN ===")

	return nil
}

// executePublish builds LinuxKit images and uploads them.
func (p *LinuxKitPublisher) executePublish(ctx context.Context, release *Release, cfg LinuxKitConfig, repo string) error {
	outputDir := filepath.Join(release.ProjectDir, "dist", "linuxkit")

	// Create output directory
	if err := release.FS.EnsureDir(outputDir); err != nil {
		return fmt.Errorf("linuxkit.Publish: failed to create output directory: %w", err)
	}

	baseName := p.buildBaseName(release.Version)
	var artifacts []string

	// Build for each platform and format
	for _, platform := range cfg.Platforms {
		parts := strings.Split(platform, "/")
		arch := "amd64"
		if len(parts) == 2 {
			arch = parts[1]
		}

		for _, format := range cfg.Formats {
			outputName := fmt.Sprintf("%s-%s", baseName, arch)

			// Build the image
			args := p.buildLinuxKitArgs(cfg.Config, format, outputName, outputDir, arch)
			cmd := exec.CommandContext(ctx, "linuxkit", args...)
			cmd.Dir = release.ProjectDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			fmt.Printf("Building LinuxKit image: %s (%s)\n", outputName, format)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("linuxkit.Publish: build failed for %s/%s: %w", platform, format, err)
			}

			// Track artifact for upload
			artifactPath := p.getArtifactPath(outputDir, outputName, format)
			artifacts = append(artifacts, artifactPath)
		}
	}

	// Upload artifacts to GitHub release
	for _, artifactPath := range artifacts {
		if !release.FS.Exists(artifactPath) {
			return fmt.Errorf("linuxkit.Publish: artifact not found after build: %s", artifactPath)
		}

		if err := UploadArtifact(ctx, repo, release.Version, artifactPath); err != nil {
			return fmt.Errorf("linuxkit.Publish: failed to upload %s: %w", filepath.Base(artifactPath), err)
		}

		// Print helpful usage info for docker format
		if strings.HasSuffix(artifactPath, ".docker.tar") {
			fmt.Printf("  Load with: docker load < %s\n", filepath.Base(artifactPath))
		}
	}

	return nil
}

// buildBaseName creates the base name for output files.
func (p *LinuxKitPublisher) buildBaseName(version string) string {
	// Strip leading 'v' if present for cleaner filenames
	name := strings.TrimPrefix(version, "v")
	return fmt.Sprintf("linuxkit-%s", name)
}

// buildLinuxKitArgs builds the arguments for linuxkit build command.
func (p *LinuxKitPublisher) buildLinuxKitArgs(configPath, format, outputName, outputDir, arch string) []string {
	args := []string{"build"}

	// Output format
	args = append(args, "--format", format)

	// Output name
	args = append(args, "--name", outputName)

	// Output directory
	args = append(args, "--dir", outputDir)

	// Architecture (if not amd64)
	if arch != "amd64" {
		args = append(args, "--arch", arch)
	}

	// Config file
	args = append(args, configPath)

	return args
}

// getArtifactPath returns the expected path of the built artifact.
func (p *LinuxKitPublisher) getArtifactPath(outputDir, outputName, format string) string {
	ext := p.getFormatExtension(format)
	return filepath.Join(outputDir, outputName+ext)
}

// getFormatExtension returns the file extension for a LinuxKit output format.
func (p *LinuxKitPublisher) getFormatExtension(format string) string {
	switch format {
	case "iso", "iso-bios", "iso-efi":
		return ".iso"
	case "raw", "raw-bios", "raw-efi":
		return ".raw"
	case "qcow2", "qcow2-bios", "qcow2-efi":
		return ".qcow2"
	case "vmdk":
		return ".vmdk"
	case "vhd":
		return ".vhd"
	case "gcp":
		return ".img.tar.gz"
	case "aws":
		return ".raw"
	case "docker":
		// Docker format outputs a tarball that can be loaded with `docker load`
		return ".docker.tar"
	case "tar":
		return ".tar"
	case "kernel+initrd":
		return "-initrd.img"
	default:
		return "." + format
	}
}

// validateLinuxKitCli checks if the linuxkit CLI is available.
func validateLinuxKitCli() error {
	cmd := exec.Command("linuxkit", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("linuxkit: linuxkit CLI not found. Install it from https://github.com/linuxkit/linuxkit")
	}
	return nil
}
