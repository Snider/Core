package cmd

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/pkg/build"
	"github.com/host-uk/core/pkg/build/builders"
	"github.com/leaanthony/clir"
	"github.com/leaanthony/debme"
	"github.com/leaanthony/gosod"
	"golang.org/x/net/html"
)

// Build command styles
var (
	buildHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#3b82f6")) // blue-500

	buildTargetStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#e2e8f0")) // gray-200

	buildSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#22c55e")) // green-500

	buildErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ef4444")) // red-500

	buildDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")) // gray-500
)

//go:embed all:tmpl/gui
var guiTemplate embed.FS

// AddBuildCommand adds the new build command and its subcommands to the clir app.
func AddBuildCommand(app *clir.Cli) {
	buildCmd := app.NewSubCommand("build", "Build projects with auto-detection and cross-compilation")
	buildCmd.LongDescription("Builds the current project with automatic type detection.\n" +
		"Supports Go, Wails, Node.js, and PHP projects.\n" +
		"Configuration can be provided via .core/build.yaml or command-line flags.")

	// Flags for the main build command
	var buildType string
	var ciMode bool
	var targets string
	var outputDir string
	var doArchive bool
	var doChecksum bool

	buildCmd.StringFlag("type", "Builder type (go, wails, node, php) - auto-detected if not specified", &buildType)
	buildCmd.BoolFlag("ci", "CI mode - minimal output with JSON artifact list at the end", &ciMode)
	buildCmd.StringFlag("targets", "Comma-separated OS/arch pairs (e.g., linux/amd64,darwin/arm64)", &targets)
	buildCmd.StringFlag("output", "Output directory for artifacts (default: dist)", &outputDir)
	buildCmd.BoolFlag("archive", "Create archives (tar.gz for linux/darwin, zip for windows) - default: true", &doArchive)
	buildCmd.BoolFlag("checksum", "Generate SHA256 checksums and CHECKSUMS.txt - default: true", &doChecksum)

	// Set defaults for archive and checksum (true by default)
	doArchive = true
	doChecksum = true

	// Default action for `core build` (no subcommand)
	buildCmd.Action(func() error {
		return runProjectBuild(buildType, ciMode, targets, outputDir, doArchive, doChecksum)
	})

	// --- `build from-path` command (legacy PWA/GUI build) ---
	fromPathCmd := buildCmd.NewSubCommand("from-path", "Build from a local directory.")
	var fromPath string
	fromPathCmd.StringFlag("path", "The path to the static web application files.", &fromPath)
	fromPathCmd.Action(func() error {
		if fromPath == "" {
			return fmt.Errorf("the --path flag is required")
		}
		return runBuild(fromPath)
	})

	// --- `build pwa` command (legacy PWA build) ---
	pwaCmd := buildCmd.NewSubCommand("pwa", "Build from a live PWA URL.")
	var pwaURL string
	pwaCmd.StringFlag("url", "The URL of the PWA to build.", &pwaURL)
	pwaCmd.Action(func() error {
		if pwaURL == "" {
			return fmt.Errorf("a URL argument is required")
		}
		return runPwaBuild(pwaURL)
	})
}

// runProjectBuild handles the main `core build` command with auto-detection.
func runProjectBuild(buildType string, ciMode bool, targetsFlag string, outputDir string, doArchive bool, doChecksum bool) error {
	// Get current working directory as project root
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load configuration from .core/build.yaml (or defaults)
	buildCfg, err := build.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Detect project type if not specified
	var projectType build.ProjectType
	if buildType != "" {
		projectType = build.ProjectType(buildType)
	} else {
		projectType, err = build.PrimaryType(projectDir)
		if err != nil {
			return fmt.Errorf("failed to detect project type: %w", err)
		}
		if projectType == "" {
			return fmt.Errorf("no supported project type detected in %s\n"+
				"Supported types: go (go.mod), wails (wails.json), node (package.json), php (composer.json)", projectDir)
		}
	}

	// Determine targets
	var buildTargets []build.Target
	if targetsFlag != "" {
		// Parse from command line
		buildTargets, err = parseTargets(targetsFlag)
		if err != nil {
			return err
		}
	} else if len(buildCfg.Targets) > 0 {
		// Use config targets
		buildTargets = buildCfg.ToTargets()
	} else {
		// Fall back to current OS/arch
		buildTargets = []build.Target{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		}
	}

	// Determine output directory
	if outputDir == "" {
		outputDir = "dist"
	}

	// Determine binary name
	binaryName := buildCfg.Project.Binary
	if binaryName == "" {
		binaryName = buildCfg.Project.Name
	}
	if binaryName == "" {
		binaryName = filepath.Base(projectDir)
	}

	// Print build info (unless CI mode)
	if !ciMode {
		fmt.Printf("%s Building project\n", buildHeaderStyle.Render("Build:"))
		fmt.Printf("  Type:    %s\n", buildTargetStyle.Render(string(projectType)))
		fmt.Printf("  Output:  %s\n", buildTargetStyle.Render(outputDir))
		fmt.Printf("  Binary:  %s\n", buildTargetStyle.Render(binaryName))
		fmt.Printf("  Targets: %s\n", buildTargetStyle.Render(formatTargets(buildTargets)))
		fmt.Println()
	}

	// Get the appropriate builder
	builder, err := getBuilder(projectType)
	if err != nil {
		return err
	}

	// Create build config for the builder
	cfg := &build.Config{
		ProjectDir: projectDir,
		OutputDir:  outputDir,
		Name:       binaryName,
		Version:    buildCfg.Project.Name, // Could be enhanced with git describe
		LDFlags:    buildCfg.Build.LDFlags,
	}

	// Execute build
	ctx := context.Background()
	artifacts, err := builder.Build(ctx, cfg, buildTargets)
	if err != nil {
		if !ciMode {
			fmt.Printf("%s Build failed: %v\n", buildErrorStyle.Render("Error:"), err)
		}
		return err
	}

	if !ciMode {
		fmt.Printf("%s Built %d artifact(s)\n", buildSuccessStyle.Render("Success:"), len(artifacts))
		fmt.Println()
		for _, artifact := range artifacts {
			relPath, err := filepath.Rel(projectDir, artifact.Path)
			if err != nil {
				relPath = artifact.Path
			}
			fmt.Printf("  %s %s %s\n",
				buildSuccessStyle.Render("✓"),
				buildTargetStyle.Render(relPath),
				buildDimStyle.Render(fmt.Sprintf("(%s/%s)", artifact.OS, artifact.Arch)),
			)
		}
	}

	// Archive artifacts if enabled
	var archivedArtifacts []build.Artifact
	if doArchive && len(artifacts) > 0 {
		if !ciMode {
			fmt.Println()
			fmt.Printf("%s Creating archives...\n", buildHeaderStyle.Render("Archive:"))
		}

		archivedArtifacts, err = build.ArchiveAll(artifacts)
		if err != nil {
			if !ciMode {
				fmt.Printf("%s Archive failed: %v\n", buildErrorStyle.Render("Error:"), err)
			}
			return err
		}

		if !ciMode {
			for _, artifact := range archivedArtifacts {
				relPath, err := filepath.Rel(projectDir, artifact.Path)
				if err != nil {
					relPath = artifact.Path
				}
				fmt.Printf("  %s %s %s\n",
					buildSuccessStyle.Render("✓"),
					buildTargetStyle.Render(relPath),
					buildDimStyle.Render(fmt.Sprintf("(%s/%s)", artifact.OS, artifact.Arch)),
				)
			}
		}
	}

	// Compute checksums if enabled
	var checksummedArtifacts []build.Artifact
	if doChecksum && len(archivedArtifacts) > 0 {
		if !ciMode {
			fmt.Println()
			fmt.Printf("%s Computing checksums...\n", buildHeaderStyle.Render("Checksum:"))
		}

		checksummedArtifacts, err = build.ChecksumAll(archivedArtifacts)
		if err != nil {
			if !ciMode {
				fmt.Printf("%s Checksum failed: %v\n", buildErrorStyle.Render("Error:"), err)
			}
			return err
		}

		// Write CHECKSUMS.txt
		checksumPath := filepath.Join(outputDir, "CHECKSUMS.txt")
		if err := build.WriteChecksumFile(checksummedArtifacts, checksumPath); err != nil {
			if !ciMode {
				fmt.Printf("%s Failed to write CHECKSUMS.txt: %v\n", buildErrorStyle.Render("Error:"), err)
			}
			return err
		}

		if !ciMode {
			for _, artifact := range checksummedArtifacts {
				relPath, err := filepath.Rel(projectDir, artifact.Path)
				if err != nil {
					relPath = artifact.Path
				}
				fmt.Printf("  %s %s\n",
					buildSuccessStyle.Render("✓"),
					buildTargetStyle.Render(relPath),
				)
				fmt.Printf("    %s\n", buildDimStyle.Render(artifact.Checksum))
			}

			relChecksumPath, err := filepath.Rel(projectDir, checksumPath)
			if err != nil {
				relChecksumPath = checksumPath
			}
			fmt.Printf("  %s %s\n",
				buildSuccessStyle.Render("✓"),
				buildTargetStyle.Render(relChecksumPath),
			)
		}
	} else if doChecksum && len(artifacts) > 0 && !doArchive {
		// Checksum raw binaries if archiving is disabled
		if !ciMode {
			fmt.Println()
			fmt.Printf("%s Computing checksums...\n", buildHeaderStyle.Render("Checksum:"))
		}

		checksummedArtifacts, err = build.ChecksumAll(artifacts)
		if err != nil {
			if !ciMode {
				fmt.Printf("%s Checksum failed: %v\n", buildErrorStyle.Render("Error:"), err)
			}
			return err
		}

		// Write CHECKSUMS.txt
		checksumPath := filepath.Join(outputDir, "CHECKSUMS.txt")
		if err := build.WriteChecksumFile(checksummedArtifacts, checksumPath); err != nil {
			if !ciMode {
				fmt.Printf("%s Failed to write CHECKSUMS.txt: %v\n", buildErrorStyle.Render("Error:"), err)
			}
			return err
		}

		if !ciMode {
			for _, artifact := range checksummedArtifacts {
				relPath, err := filepath.Rel(projectDir, artifact.Path)
				if err != nil {
					relPath = artifact.Path
				}
				fmt.Printf("  %s %s\n",
					buildSuccessStyle.Render("✓"),
					buildTargetStyle.Render(relPath),
				)
				fmt.Printf("    %s\n", buildDimStyle.Render(artifact.Checksum))
			}

			relChecksumPath, err := filepath.Rel(projectDir, checksumPath)
			if err != nil {
				relChecksumPath = checksumPath
			}
			fmt.Printf("  %s %s\n",
				buildSuccessStyle.Render("✓"),
				buildTargetStyle.Render(relChecksumPath),
			)
		}
	}

	// Output results for CI mode
	if ciMode {
		// Determine which artifacts to output (prefer checksummed > archived > raw)
		var outputArtifacts []build.Artifact
		if len(checksummedArtifacts) > 0 {
			outputArtifacts = checksummedArtifacts
		} else if len(archivedArtifacts) > 0 {
			outputArtifacts = archivedArtifacts
		} else {
			outputArtifacts = artifacts
		}

		// JSON output for CI
		output, err := json.MarshalIndent(outputArtifacts, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal artifacts: %w", err)
		}
		fmt.Println(string(output))
	}

	return nil
}

// parseTargets parses a comma-separated list of OS/arch pairs.
func parseTargets(targetsFlag string) ([]build.Target, error) {
	parts := strings.Split(targetsFlag, ",")
	var targets []build.Target

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		osArch := strings.Split(part, "/")
		if len(osArch) != 2 {
			return nil, fmt.Errorf("invalid target format %q, expected OS/arch (e.g., linux/amd64)", part)
		}

		targets = append(targets, build.Target{
			OS:   strings.TrimSpace(osArch[0]),
			Arch: strings.TrimSpace(osArch[1]),
		})
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no valid targets specified")
	}

	return targets, nil
}

// formatTargets returns a human-readable string of targets.
func formatTargets(targets []build.Target) string {
	var parts []string
	for _, t := range targets {
		parts = append(parts, t.String())
	}
	return strings.Join(parts, ", ")
}

// getBuilder returns the appropriate builder for the project type.
func getBuilder(projectType build.ProjectType) (build.Builder, error) {
	switch projectType {
	case build.ProjectTypeGo, build.ProjectTypeWails:
		return builders.NewGoBuilder(), nil
	case build.ProjectTypeNode:
		return nil, fmt.Errorf("Node.js builder not yet implemented")
	case build.ProjectTypePHP:
		return nil, fmt.Errorf("PHP builder not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported project type: %s", projectType)
	}
}

// --- PWA Build Logic ---

func runPwaBuild(pwaURL string) error {
	fmt.Printf("Starting PWA build from URL: %s\n", pwaURL)

	tempDir, err := os.MkdirTemp("", "core-pwa-build-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	// defer os.RemoveAll(tempDir) // Keep temp dir for debugging
	fmt.Printf("Downloading PWA to temporary directory: %s\n", tempDir)

	if err := downloadPWA(pwaURL, tempDir); err != nil {
		return fmt.Errorf("failed to download PWA: %w", err)
	}

	return runBuild(tempDir)
}

func downloadPWA(baseURL, destDir string) error {
	// Fetch the main HTML page
	resp, err := http.Get(baseURL)
	if err != nil {
		return fmt.Errorf("failed to fetch URL %s: %w", baseURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Find the manifest URL from the HTML
	manifestURL, err := findManifestURL(string(body), baseURL)
	if err != nil {
		// If no manifest, it's not a PWA, but we can still try to package it as a simple site.
		fmt.Println("Warning: no manifest file found. Proceeding with basic site download.")
		if err := os.WriteFile(filepath.Join(destDir, "index.html"), body, 0644); err != nil {
			return fmt.Errorf("failed to write index.html: %w", err)
		}
		return nil
	}

	fmt.Printf("Found manifest: %s\n", manifestURL)

	// Fetch and parse the manifest
	manifest, err := fetchManifest(manifestURL)
	if err != nil {
		return fmt.Errorf("failed to fetch or parse manifest: %w", err)
	}

	// Download all assets listed in the manifest
	assets := collectAssets(manifest, manifestURL)
	for _, assetURL := range assets {
		if err := downloadAsset(assetURL, destDir); err != nil {
			fmt.Printf("Warning: failed to download asset %s: %v\n", assetURL, err)
		}
	}

	// Also save the root index.html
	if err := os.WriteFile(filepath.Join(destDir, "index.html"), body, 0644); err != nil {
		return fmt.Errorf("failed to write index.html: %w", err)
	}

	fmt.Println("PWA download complete.")
	return nil
}

func findManifestURL(htmlContent, baseURL string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var manifestPath string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "link" {
			var rel, href string
			for _, a := range n.Attr {
				if a.Key == "rel" {
					rel = a.Val
				}
				if a.Key == "href" {
					href = a.Val
				}
			}
			if rel == "manifest" && href != "" {
				manifestPath = href
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if manifestPath == "" {
		return "", fmt.Errorf("no <link rel=\"manifest\"> tag found")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	manifestURL, err := base.Parse(manifestPath)
	if err != nil {
		return "", err
	}

	return manifestURL.String(), nil
}

func fetchManifest(manifestURL string) (map[string]interface{}, error) {
	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var manifest map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func collectAssets(manifest map[string]interface{}, manifestURL string) []string {
	var assets []string
	base, _ := url.Parse(manifestURL)

	// Add start_url
	if startURL, ok := manifest["start_url"].(string); ok {
		if resolved, err := base.Parse(startURL); err == nil {
			assets = append(assets, resolved.String())
		}
	}

	// Add icons
	if icons, ok := manifest["icons"].([]interface{}); ok {
		for _, icon := range icons {
			if iconMap, ok := icon.(map[string]interface{}); ok {
				if src, ok := iconMap["src"].(string); ok {
					if resolved, err := base.Parse(src); err == nil {
						assets = append(assets, resolved.String())
					}
				}
			}
		}
	}

	return assets
}

func downloadAsset(assetURL, destDir string) error {
	resp, err := http.Get(assetURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	u, err := url.Parse(assetURL)
	if err != nil {
		return err
	}

	path := filepath.Join(destDir, filepath.FromSlash(u.Path))
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// --- Standard Build Logic ---

func runBuild(fromPath string) error {
	fmt.Printf("Starting build from path: %s\n", fromPath)

	info, err := os.Stat(fromPath)
	if err != nil {
		return fmt.Errorf("invalid path specified: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path specified must be a directory")
	}

	buildDir := ".core/build/app"
	htmlDir := filepath.Join(buildDir, "html")
	appName := filepath.Base(fromPath)
	if strings.HasPrefix(appName, "core-pwa-build-") {
		appName = "pwa-app"
	}
	outputExe := appName

	if err := os.RemoveAll(buildDir); err != nil {
		return fmt.Errorf("failed to clean build directory: %w", err)
	}

	// 1. Generate the project from the embedded template
	fmt.Println("Generating application from template...")
	templateFS, err := debme.FS(guiTemplate, "tmpl/gui")
	if err != nil {
		return fmt.Errorf("failed to anchor template filesystem: %w", err)
	}
	sod := gosod.New(templateFS)
	if sod != nil {
		return fmt.Errorf("failed to create new sod instance: %w", sod)
	}

	templateData := map[string]string{"AppName": appName}
	if err := sod.Extract(buildDir, templateData); err != nil {
		return fmt.Errorf("failed to extract template: %w", err)
	}

	// 2. Copy the user's web app files
	fmt.Println("Copying application files...")
	if err := copyDir(fromPath, htmlDir); err != nil {
		return fmt.Errorf("failed to copy application files: %w", err)
	}

	// 3. Compile the application
	fmt.Println("Compiling application...")

	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	// Run go build
	cmd = exec.Command("go", "build", "-o", outputExe)
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	fmt.Printf("\nBuild successful! Executable created at: %s/%s\n", buildDir, outputExe)
	return nil
}

// copyDir recursively copies a directory from src to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}
