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

	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/sdk"
)

// runBuildSDK handles the `core build sdk` command.
func runBuildSDK(specPath, lang, version string, dryRun bool) error {
	ctx := context.Background()

	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.build.error.working_dir"), err)
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

	fmt.Printf("%s %s\n", buildHeaderStyle.Render(i18n.T("cmd.build.sdk.label")), i18n.T("cmd.build.sdk.generating"))
	if dryRun {
		fmt.Printf("  %s\n", buildDimStyle.Render(i18n.T("cmd.build.sdk.dry_run_mode")))
	}
	fmt.Println()

	// Detect spec
	detectedSpec, err := s.DetectSpec()
	if err != nil {
		fmt.Printf("%s %v\n", buildErrorStyle.Render(i18n.T("cmd.build.label.error")), err)
		return err
	}
	fmt.Printf("  %s %s\n", i18n.T("cmd.build.sdk.spec_label"), buildTargetStyle.Render(detectedSpec))

	if dryRun {
		if lang != "" {
			fmt.Printf("  %s %s\n", i18n.T("cmd.build.sdk.language_label"), buildTargetStyle.Render(lang))
		} else {
			fmt.Printf("  %s %s\n", i18n.T("cmd.build.sdk.languages_label"), buildTargetStyle.Render(strings.Join(config.Languages, ", ")))
		}
		fmt.Println()
		fmt.Printf("%s %s\n", buildSuccessStyle.Render(i18n.T("cmd.build.label.ok")), i18n.T("cmd.build.sdk.would_generate"))
		return nil
	}

	if lang != "" {
		// Generate single language
		if err := s.GenerateLanguage(ctx, lang); err != nil {
			fmt.Printf("%s %v\n", buildErrorStyle.Render(i18n.T("cmd.build.label.error")), err)
			return err
		}
		fmt.Printf("  %s %s\n", i18n.T("cmd.build.sdk.generated_label"), buildTargetStyle.Render(lang))
	} else {
		// Generate all
		if err := s.Generate(ctx); err != nil {
			fmt.Printf("%s %v\n", buildErrorStyle.Render(i18n.T("cmd.build.label.error")), err)
			return err
		}
		fmt.Printf("  %s %s\n", i18n.T("cmd.build.sdk.generated_label"), buildTargetStyle.Render(strings.Join(config.Languages, ", ")))
	}

	fmt.Println()
	fmt.Printf("%s %s\n", buildSuccessStyle.Render(i18n.T("cmd.build.label.success")), i18n.T("cmd.build.sdk.complete"))
	return nil
}
