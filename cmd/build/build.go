// Package build provides project build commands with auto-detection.
package build

import (
	"embed"

	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/i18n"
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
	Short: i18n.T("cmd.build.short"),
	Long:  i18n.T("cmd.build.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectBuild(buildType, ciMode, targets, outputDir, doArchive, doChecksum, configPath, format, push, imageName, noSign, notarize)
	},
}

var fromPathCmd = &cobra.Command{
	Use:   "from-path",
	Short: i18n.T("cmd.build.from_path.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		if fromPath == "" {
			return errPathRequired
		}
		return runBuild(fromPath)
	},
}

var pwaCmd = &cobra.Command{
	Use:   "pwa",
	Short: i18n.T("cmd.build.pwa.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		if pwaURL == "" {
			return errURLRequired
		}
		return runPwaBuild(pwaURL)
	},
}

var sdkBuildCmd = &cobra.Command{
	Use:   "sdk",
	Short: i18n.T("cmd.build.sdk.short"),
	Long:  i18n.T("cmd.build.sdk.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuildSDK(sdkSpec, sdkLang, sdkVersion, sdkDryRun)
	},
}

func init() {
	// Main build command flags
	buildCmd.Flags().StringVar(&buildType, "type", "", i18n.T("cmd.build.flag.type"))
	buildCmd.Flags().BoolVar(&ciMode, "ci", false, i18n.T("cmd.build.flag.ci"))
	buildCmd.Flags().StringVar(&targets, "targets", "", i18n.T("cmd.build.flag.targets"))
	buildCmd.Flags().StringVar(&outputDir, "output", "", i18n.T("cmd.build.flag.output"))
	buildCmd.Flags().BoolVar(&doArchive, "archive", true, i18n.T("cmd.build.flag.archive"))
	buildCmd.Flags().BoolVar(&doChecksum, "checksum", true, i18n.T("cmd.build.flag.checksum"))

	// Docker/LinuxKit specific
	buildCmd.Flags().StringVar(&configPath, "config", "", i18n.T("cmd.build.flag.config"))
	buildCmd.Flags().StringVar(&format, "format", "", i18n.T("cmd.build.flag.format"))
	buildCmd.Flags().BoolVar(&push, "push", false, i18n.T("cmd.build.flag.push"))
	buildCmd.Flags().StringVar(&imageName, "image", "", i18n.T("cmd.build.flag.image"))

	// Signing flags
	buildCmd.Flags().BoolVar(&noSign, "no-sign", false, i18n.T("cmd.build.flag.no_sign"))
	buildCmd.Flags().BoolVar(&notarize, "notarize", false, i18n.T("cmd.build.flag.notarize"))

	// from-path subcommand flags
	fromPathCmd.Flags().StringVar(&fromPath, "path", "", i18n.T("cmd.build.from_path.flag.path"))

	// pwa subcommand flags
	pwaCmd.Flags().StringVar(&pwaURL, "url", "", i18n.T("cmd.build.pwa.flag.url"))

	// sdk subcommand flags
	sdkBuildCmd.Flags().StringVar(&sdkSpec, "spec", "", i18n.T("cmd.build.sdk.flag.spec"))
	sdkBuildCmd.Flags().StringVar(&sdkLang, "lang", "", i18n.T("cmd.build.sdk.flag.lang"))
	sdkBuildCmd.Flags().StringVar(&sdkVersion, "version", "", i18n.T("cmd.build.sdk.flag.version"))
	sdkBuildCmd.Flags().BoolVar(&sdkDryRun, "dry-run", false, i18n.T("cmd.build.sdk.flag.dry_run"))

	// Add subcommands
	buildCmd.AddCommand(fromPathCmd)
	buildCmd.AddCommand(pwaCmd)
	buildCmd.AddCommand(sdkBuildCmd)
}

// AddBuildCommand adds the new build command and its subcommands to the cobra app.
func AddBuildCommand(root *cobra.Command) {
	root.AddCommand(buildCmd)
}
