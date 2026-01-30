// build_sdk.go implements SDK generation from OpenAPI specifications.
//
// Generates typed API clients for TypeScript, Python, Go, and PHP
// from OpenAPI/Swagger specifications.

package build

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/host-uk/core/pkg/sdk"
)

// runBuildSDK handles the `core build sdk` command.
func runBuildSDK(specPath, lang, version string, dryRun bool) error {
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
	if version != "" {
		s.SetVersion(version)
	}

	fmt.Printf("%s Generating SDKs\n", buildHeaderStyle.Render("Build SDK:"))
	if dryRun {
		fmt.Printf("  %s\n", buildDimStyle.Render("(dry-run mode)"))
	}
	fmt.Println()

	// Detect spec
	detectedSpec, err := s.DetectSpec()
	if err != nil {
		fmt.Printf("%s %v\n", buildErrorStyle.Render("Error:"), err)
		return err
	}
	fmt.Printf("  Spec:      %s\n", buildTargetStyle.Render(detectedSpec))

	if dryRun {
		if lang != "" {
			fmt.Printf("  Language:  %s\n", buildTargetStyle.Render(lang))
		} else {
			fmt.Printf("  Languages: %s\n", buildTargetStyle.Render(strings.Join(config.Languages, ", ")))
		}
		fmt.Println()
		fmt.Printf("%s Would generate SDKs (dry-run)\n", buildSuccessStyle.Render("OK:"))
		return nil
	}

	if lang != "" {
		// Generate single language
		if err := s.GenerateLanguage(ctx, lang); err != nil {
			fmt.Printf("%s %v\n", buildErrorStyle.Render("Error:"), err)
			return err
		}
		fmt.Printf("  Generated: %s\n", buildTargetStyle.Render(lang))
	} else {
		// Generate all
		if err := s.Generate(ctx); err != nil {
			fmt.Printf("%s %v\n", buildErrorStyle.Render("Error:"), err)
			return err
		}
		fmt.Printf("  Generated: %s\n", buildTargetStyle.Render(strings.Join(config.Languages, ", ")))
	}

	fmt.Println()
	fmt.Printf("%s SDK generation complete\n", buildSuccessStyle.Render("Success:"))
	return nil
}
