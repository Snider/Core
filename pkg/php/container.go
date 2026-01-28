package php

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DockerBuildOptions configures Docker image building for PHP projects.
type DockerBuildOptions struct {
	// ProjectDir is the path to the PHP/Laravel project.
	ProjectDir string

	// ImageName is the name for the Docker image.
	ImageName string

	// Tag is the image tag (default: "latest").
	Tag string

	// Platform specifies the target platform (e.g., "linux/amd64", "linux/arm64").
	Platform string

	// Dockerfile is the path to a custom Dockerfile.
	// If empty, one will be auto-generated for FrankenPHP.
	Dockerfile string

	// NoBuildCache disables Docker build cache.
	NoBuildCache bool

	// BuildArgs are additional build arguments.
	BuildArgs map[string]string

	// Output is the writer for build output (default: os.Stdout).
	Output io.Writer
}

// LinuxKitBuildOptions configures LinuxKit image building for PHP projects.
type LinuxKitBuildOptions struct {
	// ProjectDir is the path to the PHP/Laravel project.
	ProjectDir string

	// OutputPath is the path for the output image.
	OutputPath string

	// Format is the output format: "iso", "qcow2", "raw", "vmdk".
	Format string

	// Template is the LinuxKit template name (default: "server-php").
	Template string

	// Variables are template variables to apply.
	Variables map[string]string

	// Output is the writer for build output (default: os.Stdout).
	Output io.Writer
}

// ServeOptions configures running a production PHP container.
type ServeOptions struct {
	// ImageName is the Docker image to run.
	ImageName string

	// Tag is the image tag (default: "latest").
	Tag string

	// ContainerName is the name for the container.
	ContainerName string

	// Port is the host port to bind (default: 80).
	Port int

	// HTTPSPort is the host HTTPS port to bind (default: 443).
	HTTPSPort int

	// Detach runs the container in detached mode.
	Detach bool

	// EnvFile is the path to an environment file.
	EnvFile string

	// Volumes maps host paths to container paths.
	Volumes map[string]string

	// Output is the writer for output (default: os.Stdout).
	Output io.Writer
}

// BuildDocker builds a Docker image for the PHP project.
func BuildDocker(ctx context.Context, opts DockerBuildOptions) error {
	if opts.ProjectDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		opts.ProjectDir = cwd
	}

	// Validate project directory
	if !IsPHPProject(opts.ProjectDir) {
		return fmt.Errorf("not a PHP project: %s (missing composer.json)", opts.ProjectDir)
	}

	// Set defaults
	if opts.ImageName == "" {
		opts.ImageName = filepath.Base(opts.ProjectDir)
	}
	if opts.Tag == "" {
		opts.Tag = "latest"
	}
	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	// Determine Dockerfile path
	dockerfilePath := opts.Dockerfile
	var tempDockerfile string

	if dockerfilePath == "" {
		// Generate Dockerfile
		content, err := GenerateDockerfile(opts.ProjectDir)
		if err != nil {
			return fmt.Errorf("failed to generate Dockerfile: %w", err)
		}

		// Write to temporary file
		tempDockerfile = filepath.Join(opts.ProjectDir, "Dockerfile.core-generated")
		if err := os.WriteFile(tempDockerfile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write Dockerfile: %w", err)
		}
		defer os.Remove(tempDockerfile)

		dockerfilePath = tempDockerfile
	}

	// Build Docker image
	imageRef := fmt.Sprintf("%s:%s", opts.ImageName, opts.Tag)

	args := []string{"build", "-t", imageRef, "-f", dockerfilePath}

	if opts.Platform != "" {
		args = append(args, "--platform", opts.Platform)
	}

	if opts.NoBuildCache {
		args = append(args, "--no-cache")
	}

	for key, value := range opts.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	args = append(args, opts.ProjectDir)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = opts.ProjectDir
	cmd.Stdout = opts.Output
	cmd.Stderr = opts.Output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	return nil
}

// BuildLinuxKit builds a LinuxKit image for the PHP project.
func BuildLinuxKit(ctx context.Context, opts LinuxKitBuildOptions) error {
	if opts.ProjectDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		opts.ProjectDir = cwd
	}

	// Validate project directory
	if !IsPHPProject(opts.ProjectDir) {
		return fmt.Errorf("not a PHP project: %s (missing composer.json)", opts.ProjectDir)
	}

	// Set defaults
	if opts.Template == "" {
		opts.Template = "server-php"
	}
	if opts.Format == "" {
		opts.Format = "qcow2"
	}
	if opts.OutputPath == "" {
		opts.OutputPath = filepath.Join(opts.ProjectDir, "dist", filepath.Base(opts.ProjectDir))
	}
	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(opts.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Find linuxkit binary
	linuxkitPath, err := lookupLinuxKit()
	if err != nil {
		return err
	}

	// Get template content
	templateContent, err := getLinuxKitTemplate(opts.Template)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Apply variables
	if opts.Variables == nil {
		opts.Variables = make(map[string]string)
	}
	// Add project-specific variables
	opts.Variables["PROJECT_DIR"] = opts.ProjectDir
	opts.Variables["PROJECT_NAME"] = filepath.Base(opts.ProjectDir)

	content, err := applyTemplateVariables(templateContent, opts.Variables)
	if err != nil {
		return fmt.Errorf("failed to apply template variables: %w", err)
	}

	// Write template to temp file
	tempYAML := filepath.Join(opts.ProjectDir, ".core-linuxkit.yml")
	if err := os.WriteFile(tempYAML, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}
	defer os.Remove(tempYAML)

	// Build LinuxKit image
	args := []string{
		"build",
		"-format", opts.Format,
		"-name", opts.OutputPath,
		tempYAML,
	}

	cmd := exec.CommandContext(ctx, linuxkitPath, args...)
	cmd.Dir = opts.ProjectDir
	cmd.Stdout = opts.Output
	cmd.Stderr = opts.Output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("linuxkit build failed: %w", err)
	}

	return nil
}

// ServeProduction runs a production PHP container.
func ServeProduction(ctx context.Context, opts ServeOptions) error {
	if opts.ImageName == "" {
		return fmt.Errorf("image name is required")
	}

	// Set defaults
	if opts.Tag == "" {
		opts.Tag = "latest"
	}
	if opts.Port == 0 {
		opts.Port = 80
	}
	if opts.HTTPSPort == 0 {
		opts.HTTPSPort = 443
	}
	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	imageRef := fmt.Sprintf("%s:%s", opts.ImageName, opts.Tag)

	args := []string{"run"}

	if opts.Detach {
		args = append(args, "-d")
	} else {
		args = append(args, "--rm")
	}

	if opts.ContainerName != "" {
		args = append(args, "--name", opts.ContainerName)
	}

	// Port mappings
	args = append(args, "-p", fmt.Sprintf("%d:80", opts.Port))
	args = append(args, "-p", fmt.Sprintf("%d:443", opts.HTTPSPort))

	// Environment file
	if opts.EnvFile != "" {
		args = append(args, "--env-file", opts.EnvFile)
	}

	// Volume mounts
	for hostPath, containerPath := range opts.Volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", hostPath, containerPath))
	}

	args = append(args, imageRef)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = opts.Output
	cmd.Stderr = opts.Output

	if opts.Detach {
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to start container: %w", err)
		}
		containerID := strings.TrimSpace(string(output))
		fmt.Fprintf(opts.Output, "Container started: %s\n", containerID[:12])
		return nil
	}

	return cmd.Run()
}

// Shell opens a shell in a running container.
func Shell(ctx context.Context, containerID string) error {
	if containerID == "" {
		return fmt.Errorf("container ID is required")
	}

	// Resolve partial container ID
	fullID, err := resolveDockerContainerID(ctx, containerID)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "docker", "exec", "-it", fullID, "/bin/sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// IsPHPProject checks if the given directory is a PHP project.
func IsPHPProject(dir string) bool {
	composerPath := filepath.Join(dir, "composer.json")
	_, err := os.Stat(composerPath)
	return err == nil
}

// lookupLinuxKit finds the linuxkit binary.
func lookupLinuxKit() (string, error) {
	// Check PATH first
	if path, err := exec.LookPath("linuxkit"); err == nil {
		return path, nil
	}

	// Check common locations
	paths := []string{
		"/usr/local/bin/linuxkit",
		"/opt/homebrew/bin/linuxkit",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("linuxkit not found. Install with: brew install linuxkit (macOS) or see https://github.com/linuxkit/linuxkit")
}

// getLinuxKitTemplate retrieves a LinuxKit template by name.
func getLinuxKitTemplate(name string) (string, error) {
	// Default server-php template for PHP projects
	if name == "server-php" {
		return defaultServerPHPTemplate, nil
	}

	// Try to load from container package templates
	// This would integrate with github.com/host-uk/core/pkg/container
	return "", fmt.Errorf("template not found: %s", name)
}

// applyTemplateVariables applies variable substitution to template content.
func applyTemplateVariables(content string, vars map[string]string) (string, error) {
	result := content
	for key, value := range vars {
		placeholder := "${" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result, nil
}

// resolveDockerContainerID resolves a partial container ID to a full ID.
func resolveDockerContainerID(ctx context.Context, partialID string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--no-trunc", "--format", "{{.ID}}")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var matches []string

	for _, line := range lines {
		if strings.HasPrefix(line, partialID) {
			matches = append(matches, line)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no container found matching: %s", partialID)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("multiple containers match '%s', be more specific", partialID)
	}
}

// defaultServerPHPTemplate is the default LinuxKit template for PHP servers.
const defaultServerPHPTemplate = `# LinuxKit configuration for PHP/FrankenPHP server
kernel:
  image: linuxkit/kernel:6.6.13
  cmdline: "console=tty0 console=ttyS0"
init:
  - linuxkit/init:v1.0.1
  - linuxkit/runc:v1.0.1
  - linuxkit/containerd:v1.0.1
onboot:
  - name: sysctl
    image: linuxkit/sysctl:v1.0.1
  - name: dhcpcd
    image: linuxkit/dhcpcd:v1.0.1
    command: ["/sbin/dhcpcd", "--nobackground", "-f", "/dhcpcd.conf"]
services:
  - name: getty
    image: linuxkit/getty:v1.0.1
    env:
      - INSECURE=true
  - name: sshd
    image: linuxkit/sshd:v1.0.1
files:
  - path: etc/ssh/authorized_keys
    contents: |
      ${SSH_KEY:-}
`
