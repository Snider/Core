// Package sdk provides SDK validation and API compatibility commands.
package sdk

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	sdkpkg "github.com/host-uk/core/pkg/sdk"
	"github.com/spf13/cobra"
)

var (
	sdkHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#3b82f6"))

	sdkSuccessStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#22c55e"))

	sdkErrorStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ef4444"))

	sdkDimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280"))
)

var sdkCmd = &cobra.Command{
	Use:   "sdk",
	Short: "SDK validation and API compatibility tools",
	Long: `Tools for validating OpenAPI specs and checking API compatibility.
To generate SDKs, use: core build sdk

Commands:
  diff      Check for breaking API changes
  validate  Validate OpenAPI spec syntax`,
}

var diffBasePath string
var diffSpecPath string

var sdkDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Check for breaking API changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSDKDiff(diffBasePath, diffSpecPath)
	},
}

var validateSpecPath string

var sdkValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate OpenAPI spec",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSDKValidate(validateSpecPath)
	},
}

func init() {
	// sdk diff flags
	sdkDiffCmd.Flags().StringVar(&diffBasePath, "base", "", "Base spec (version tag or file)")
	sdkDiffCmd.Flags().StringVar(&diffSpecPath, "spec", "", "Current spec file")

	// sdk validate flags
	sdkValidateCmd.Flags().StringVar(&validateSpecPath, "spec", "", "Path to OpenAPI spec file")

	// Add subcommands
	sdkCmd.AddCommand(sdkDiffCmd)
	sdkCmd.AddCommand(sdkValidateCmd)
}

func runSDKDiff(basePath, specPath string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
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
		return fmt.Errorf("--base is required (version tag or file path)")
	}

	fmt.Printf("%s Checking for breaking changes\n", sdkHeaderStyle.Render("SDK Diff:"))
	fmt.Printf("  Base:     %s\n", sdkDimStyle.Render(basePath))
	fmt.Printf("  Current:  %s\n", sdkDimStyle.Render(specPath))
	fmt.Println()

	result, err := sdkpkg.Diff(basePath, specPath)
	if err != nil {
		fmt.Printf("%s %v\n", sdkErrorStyle.Render("Error:"), err)
		os.Exit(2)
	}

	if result.Breaking {
		fmt.Printf("%s %s\n", sdkErrorStyle.Render("Breaking:"), result.Summary)
		for _, change := range result.Changes {
			fmt.Printf("  - %s\n", change)
		}
		os.Exit(1)
	}

	fmt.Printf("%s %s\n", sdkSuccessStyle.Render("OK:"), result.Summary)
	return nil
}

func runSDKValidate(specPath string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	s := sdkpkg.New(projectDir, &sdkpkg.Config{Spec: specPath})

	fmt.Printf("%s Validating OpenAPI spec\n", sdkHeaderStyle.Render("SDK:"))

	detectedPath, err := s.DetectSpec()
	if err != nil {
		fmt.Printf("%s %v\n", sdkErrorStyle.Render("Error:"), err)
		return err
	}

	fmt.Printf("  Spec: %s\n", sdkDimStyle.Render(detectedPath))
	fmt.Printf("%s Spec is valid\n", sdkSuccessStyle.Render("OK:"))
	return nil
}
