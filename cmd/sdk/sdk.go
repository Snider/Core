// Package sdk provides SDK validation and API compatibility commands.
package sdk

import (
	"fmt"
	"os"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	sdkpkg "github.com/host-uk/core/pkg/sdk"
	"github.com/spf13/cobra"
)

// SDK styles (aliases to shared)
var (
	sdkHeaderStyle  = cli.TitleStyle
	sdkSuccessStyle = cli.SuccessStyle
	sdkErrorStyle   = cli.ErrorStyle
	sdkDimStyle     = cli.DimStyle
)

var sdkCmd = &cobra.Command{
	Use:   "sdk",
	Short: i18n.T("cmd.sdk.short"),
	Long:  i18n.T("cmd.sdk.long"),
}

var diffBasePath string
var diffSpecPath string

var sdkDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: i18n.T("cmd.sdk.diff.short"),
	Long:  i18n.T("cmd.sdk.diff.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSDKDiff(diffBasePath, diffSpecPath)
	},
}

var validateSpecPath string

var sdkValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: i18n.T("cmd.sdk.validate.short"),
	Long:  i18n.T("cmd.sdk.validate.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSDKValidate(validateSpecPath)
	},
}

func init() {
	// sdk diff flags
	sdkDiffCmd.Flags().StringVar(&diffBasePath, "base", "", i18n.T("cmd.sdk.diff.flag.base"))
	sdkDiffCmd.Flags().StringVar(&diffSpecPath, "spec", "", i18n.T("cmd.sdk.diff.flag.spec"))

	// sdk validate flags
	sdkValidateCmd.Flags().StringVar(&validateSpecPath, "spec", "", i18n.T("common.flag.spec"))

	// Add subcommands
	sdkCmd.AddCommand(sdkDiffCmd)
	sdkCmd.AddCommand(sdkValidateCmd)
}

func runSDKDiff(basePath, specPath string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
	}

	// Detect current spec if not provided
	if specPath == "" {
		s := sdkpkg.New(projectDir, nil)
		specPath, err = s.DetectSpec()
		if err != nil {
			return err
		}
	}

	if basePath == "" {
		return fmt.Errorf(i18n.T("cmd.sdk.diff.error.base_required"))
	}

	fmt.Printf("%s %s\n", sdkHeaderStyle.Render(i18n.T("cmd.sdk.diff.label")), i18n.T("common.progress.checking", map[string]any{"Item": "breaking changes"}))
	fmt.Printf("  %s %s\n", i18n.T("cmd.sdk.diff.base_label"), sdkDimStyle.Render(basePath))
	fmt.Printf("  %s %s\n", i18n.T("common.label.current"), sdkDimStyle.Render(specPath))
	fmt.Println()

	result, err := sdkpkg.Diff(basePath, specPath)
	if err != nil {
		fmt.Printf("%s %v\n", sdkErrorStyle.Render(i18n.T("common.label.error")), err)
		os.Exit(2)
	}

	if result.Breaking {
		fmt.Printf("%s %s\n", sdkErrorStyle.Render(i18n.T("cmd.sdk.diff.breaking")), result.Summary)
		for _, change := range result.Changes {
			fmt.Printf("  - %s\n", change)
		}
		os.Exit(1)
	}

	fmt.Printf("%s %s\n", sdkSuccessStyle.Render(i18n.T("cmd.sdk.label.ok")), result.Summary)
	return nil
}

func runSDKValidate(specPath string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
	}

	s := sdkpkg.New(projectDir, &sdkpkg.Config{Spec: specPath})

	fmt.Printf("%s %s\n", sdkHeaderStyle.Render(i18n.T("cmd.sdk.label.sdk")), i18n.T("cmd.sdk.validate.validating"))

	detectedPath, err := s.DetectSpec()
	if err != nil {
		fmt.Printf("%s %v\n", sdkErrorStyle.Render(i18n.T("common.label.error")), err)
		return err
	}

	fmt.Printf("  %s %s\n", i18n.T("common.label.spec"), sdkDimStyle.Render(detectedPath))
	fmt.Printf("%s %s\n", sdkSuccessStyle.Render(i18n.T("cmd.sdk.label.ok")), i18n.T("cmd.sdk.validate.valid"))
	return nil
}
