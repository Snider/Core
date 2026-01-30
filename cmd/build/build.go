// Package build provides project build commands with auto-detection.
package build

import (
	"embed"

	"github.com/host-uk/core/cmd/shared"
	"github.com/spf13/cobra"
)

// Style aliases from shared package
var (
	buildHeaderStyle  = shared.TitleStyle
	buildTargetStyle  = shared.ValueStyle
	buildSuccessStyle = shared.SuccessStyle
	buildErrorStyle   = shared.ErrorStyle
	buildDimStyle     = shared.DimStyle
)

//go:embed all:tmpl/gui
var guiTemplate embed.FS

// Flags for the main build command
var (
	buildType  string
	ciMode     bool
	targets    string
	outputDir  string
	doArchive  bool
	doChecksum bool

	// Docker/LinuxKit specific flags
	configPath string
	format     string
	push       bool
	imageName  string

	// Signing flags
	noSign   bool
	notarize bool

	// from-path subcommand
	fromPath string

	// pwa subcommand
	pwaURL string

	// sdk subcommand
	sdkSpec    string
	sdkLang    string
	sdkVersion string
	sdkDryRun  bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build projects with auto-detection and cross-compilation",
	Long: `Builds the current project with automatic type detection.
Supports Go, Wails, Docker, LinuxKit, and Taskfile projects.
Configuration can be provided via .core/build.yaml or command-line flags.

Examples:
  core build                              # Auto-detect and build
  core build --type docker                # Build Docker image
  core build --type linuxkit              # Build LinuxKit image
  core build --type linuxkit --config linuxkit.yml --format qcow2-bios`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectBuild(buildType, ciMode, targets, outputDir, doArchive, doChecksum, configPath, format, push, imageName, noSign, notarize)
	},
}

var fromPathCmd = &cobra.Command{
	Use:   "from-path",
	Short: "Build from a local directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if fromPath == "" {
			return errPathRequired
		}
		return runBuild(fromPath)
	},
}

var pwaCmd = &cobra.Command{
	Use:   "pwa",
	Short: "Build from a live PWA URL.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if pwaURL == "" {
			return errURLRequired
		}
		return runPwaBuild(pwaURL)
	},
}

var sdkBuildCmd = &cobra.Command{
	Use:   "sdk",
	Short: "Generate API SDKs from OpenAPI spec",
	Long: `Generates typed API clients from OpenAPI specifications.
Supports TypeScript, Python, Go, and PHP.

Examples:
  core build sdk                    # Generate all configured SDKs
  core build sdk --lang typescript  # Generate only TypeScript SDK
  core build sdk --spec api.yaml    # Use specific OpenAPI spec`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuildSDK(sdkSpec, sdkLang, sdkVersion, sdkDryRun)
	},
}

func init() {
	// Main build command flags
	buildCmd.Flags().StringVar(&buildType, "type", "", "Builder type (go, wails, docker, linuxkit, taskfile) - auto-detected if not specified")
	buildCmd.Flags().BoolVar(&ciMode, "ci", false, "CI mode - minimal output with JSON artifact list at the end")
	buildCmd.Flags().StringVar(&targets, "targets", "", "Comma-separated OS/arch pairs (e.g., linux/amd64,darwin/arm64)")
	buildCmd.Flags().StringVar(&outputDir, "output", "", "Output directory for artifacts (default: dist)")
	buildCmd.Flags().BoolVar(&doArchive, "archive", true, "Create archives (tar.gz for linux/darwin, zip for windows)")
	buildCmd.Flags().BoolVar(&doChecksum, "checksum", true, "Generate SHA256 checksums and CHECKSUMS.txt")

	// Docker/LinuxKit specific
	buildCmd.Flags().StringVar(&configPath, "config", "", "Config file path (for linuxkit: YAML config, for docker: Dockerfile)")
	buildCmd.Flags().StringVar(&format, "format", "", "Output format for linuxkit (iso-bios, qcow2-bios, raw, vmdk)")
	buildCmd.Flags().BoolVar(&push, "push", false, "Push Docker image after build")
	buildCmd.Flags().StringVar(&imageName, "image", "", "Docker image name (e.g., host-uk/core-devops)")

	// Signing flags
	buildCmd.Flags().BoolVar(&noSign, "no-sign", false, "Skip all code signing")
	buildCmd.Flags().BoolVar(&notarize, "notarize", false, "Enable macOS notarization (requires Apple credentials)")

	// from-path subcommand flags
	fromPathCmd.Flags().StringVar(&fromPath, "path", "", "The path to the static web application files.")

	// pwa subcommand flags
	pwaCmd.Flags().StringVar(&pwaURL, "url", "", "The URL of the PWA to build.")

	// sdk subcommand flags
	sdkBuildCmd.Flags().StringVar(&sdkSpec, "spec", "", "Path to OpenAPI spec file")
	sdkBuildCmd.Flags().StringVar(&sdkLang, "lang", "", "Generate only this language (typescript, python, go, php)")
	sdkBuildCmd.Flags().StringVar(&sdkVersion, "version", "", "Version to embed in generated SDKs")
	sdkBuildCmd.Flags().BoolVar(&sdkDryRun, "dry-run", false, "Show what would be generated without writing files")

	// Add subcommands
	buildCmd.AddCommand(fromPathCmd)
	buildCmd.AddCommand(pwaCmd)
	buildCmd.AddCommand(sdkBuildCmd)
}

// AddBuildCommand adds the new build command and its subcommands to the cobra app.
func AddBuildCommand(root *cobra.Command) {
	root.AddCommand(buildCmd)
}
