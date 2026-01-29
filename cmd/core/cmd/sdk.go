package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/pkg/sdk"
	"github.com/leaanthony/clir"
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

// AddSDKCommand adds the sdk command and its subcommands.
func AddSDKCommand(app *clir.Cli) {
	sdkCmd := app.NewSubCommand("sdk", "Generate and manage API SDKs")
	sdkCmd.LongDescription("Generate typed API clients from OpenAPI specs.\n" +
		"Supports TypeScript, Python, Go, and PHP.")

	// sdk generate
	genCmd := sdkCmd.NewSubCommand("generate", "Generate SDKs from OpenAPI spec")
	var specPath, lang string
	genCmd.StringFlag("spec", "Path to OpenAPI spec file", &specPath)
	genCmd.StringFlag("lang", "Generate only this language", &lang)
	genCmd.Action(func() error {
		return runSDKGenerate(specPath, lang)
	})

	// sdk diff
	diffCmd := sdkCmd.NewSubCommand("diff", "Check for breaking API changes")
	var basePath string
	diffCmd.StringFlag("base", "Base spec (version tag or file)", &basePath)
	diffCmd.StringFlag("spec", "Current spec file", &specPath)
	diffCmd.Action(func() error {
		return runSDKDiff(basePath, specPath)
	})

	// sdk validate
	validateCmd := sdkCmd.NewSubCommand("validate", "Validate OpenAPI spec")
	validateCmd.StringFlag("spec", "Path to OpenAPI spec file", &specPath)
	validateCmd.Action(func() error {
		return runSDKValidate(specPath)
	})
}

func runSDKGenerate(specPath, lang string) error {
	ctx := context.Background()

	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load config
	config := sdk.DefaultConfig()
	if specPath != "" {
		config.Spec = specPath
	}

	s := sdk.New(projectDir, config)

	fmt.Printf("%s Generating SDKs\n", sdkHeaderStyle.Render("SDK:"))

	if lang != "" {
		// Generate single language
		if err := s.GenerateLanguage(ctx, lang); err != nil {
			fmt.Printf("%s %v\n", sdkErrorStyle.Render("Error:"), err)
			return err
		}
	} else {
		// Generate all
		if err := s.Generate(ctx); err != nil {
			fmt.Printf("%s %v\n", sdkErrorStyle.Render("Error:"), err)
			return err
		}
	}

	fmt.Printf("%s SDK generation complete\n", sdkSuccessStyle.Render("Success:"))
	return nil
}

func runSDKDiff(basePath, specPath string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Detect current spec if not provided
	if specPath == "" {
		s := sdk.New(projectDir, nil)
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

	result, err := sdk.Diff(basePath, specPath)
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

	s := sdk.New(projectDir, &sdk.Config{Spec: specPath})

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
