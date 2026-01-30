// Package build provides project build commands with auto-detection.
package build

import (
	"embed"

	"github.com/charmbracelet/lipgloss"
	"github.com/leaanthony/clir"
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
		"Supports Go, Wails, Docker, LinuxKit, and Taskfile projects.\n" +
		"Configuration can be provided via .core/build.yaml or command-line flags.\n\n" +
		"Examples:\n" +
		"  core build                              # Auto-detect and build\n" +
		"  core build --type docker                # Build Docker image\n" +
		"  core build --type linuxkit              # Build LinuxKit image\n" +
		"  core build --type linuxkit --config linuxkit.yml --format qcow2-bios")

	// Flags for the main build command
	var buildType string
	var ciMode bool
	var targets string
	var outputDir string
	var doArchive bool
	var doChecksum bool

	// Docker/LinuxKit specific flags
	var configPath string
	var format string
	var push bool
	var imageName string

	// Signing flags
	var noSign bool
	var notarize bool

	buildCmd.StringFlag("type", "Builder type (go, wails, docker, linuxkit, taskfile) - auto-detected if not specified", &buildType)
	buildCmd.BoolFlag("ci", "CI mode - minimal output with JSON artifact list at the end", &ciMode)
	buildCmd.StringFlag("targets", "Comma-separated OS/arch pairs (e.g., linux/amd64,darwin/arm64)", &targets)
	buildCmd.StringFlag("output", "Output directory for artifacts (default: dist)", &outputDir)
	buildCmd.BoolFlag("archive", "Create archives (tar.gz for linux/darwin, zip for windows) - default: true", &doArchive)
	buildCmd.BoolFlag("checksum", "Generate SHA256 checksums and CHECKSUMS.txt - default: true", &doChecksum)

	// Docker/LinuxKit specific
	buildCmd.StringFlag("config", "Config file path (for linuxkit: YAML config, for docker: Dockerfile)", &configPath)
	buildCmd.StringFlag("format", "Output format for linuxkit (iso-bios, qcow2-bios, raw, vmdk)", &format)
	buildCmd.BoolFlag("push", "Push Docker image after build (default: false)", &push)
	buildCmd.StringFlag("image", "Docker image name (e.g., host-uk/core-devops)", &imageName)

	// Signing flags
	buildCmd.BoolFlag("no-sign", "Skip all code signing", &noSign)
	buildCmd.BoolFlag("notarize", "Enable macOS notarization (requires Apple credentials)", &notarize)

	// Set defaults for archive and checksum (true by default)
	doArchive = true
	doChecksum = true

	// Default action for `core build` (no subcommand)
	buildCmd.Action(func() error {
		return runProjectBuild(buildType, ciMode, targets, outputDir, doArchive, doChecksum, configPath, format, push, imageName, noSign, notarize)
	})

	// --- `build from-path` command (legacy PWA/GUI build) ---
	fromPathCmd := buildCmd.NewSubCommand("from-path", "Build from a local directory.")
	var fromPath string
	fromPathCmd.StringFlag("path", "The path to the static web application files.", &fromPath)
	fromPathCmd.Action(func() error {
		if fromPath == "" {
			return errPathRequired
		}
		return runBuild(fromPath)
	})

	// --- `build pwa` command (legacy PWA build) ---
	pwaCmd := buildCmd.NewSubCommand("pwa", "Build from a live PWA URL.")
	var pwaURL string
	pwaCmd.StringFlag("url", "The URL of the PWA to build.", &pwaURL)
	pwaCmd.Action(func() error {
		if pwaURL == "" {
			return errURLRequired
		}
		return runPwaBuild(pwaURL)
	})

	// --- `build sdk` command ---
	sdkBuildCmd := buildCmd.NewSubCommand("sdk", "Generate API SDKs from OpenAPI spec")
	sdkBuildCmd.LongDescription("Generates typed API clients from OpenAPI specifications.\n" +
		"Supports TypeScript, Python, Go, and PHP.\n\n" +
		"Examples:\n" +
		"  core build sdk                    # Generate all configured SDKs\n" +
		"  core build sdk --lang typescript  # Generate only TypeScript SDK\n" +
		"  core build sdk --spec api.yaml    # Use specific OpenAPI spec")

	var sdkSpec, sdkLang, sdkVersion string
	var sdkDryRun bool
	sdkBuildCmd.StringFlag("spec", "Path to OpenAPI spec file", &sdkSpec)
	sdkBuildCmd.StringFlag("lang", "Generate only this language (typescript, python, go, php)", &sdkLang)
	sdkBuildCmd.StringFlag("version", "Version to embed in generated SDKs", &sdkVersion)
	sdkBuildCmd.BoolFlag("dry-run", "Show what would be generated without writing files", &sdkDryRun)
	sdkBuildCmd.Action(func() error {
		return runBuildSDK(sdkSpec, sdkLang, sdkVersion, sdkDryRun)
	})
}
