package sdk

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"
)

// DiffResult holds the result of comparing two OpenAPI specs.
type DiffResult struct {
	// Breaking is true if breaking changes were detected.
	Breaking bool
	// Changes is the list of breaking changes.
	Changes []string
	// Summary is a human-readable summary.
	Summary string
}

// Diff compares two OpenAPI specs and detects breaking changes.
func Diff(basePath, revisionPath string) (*DiffResult, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// Load specs
	baseSpec, err := load.NewSpecInfo(loader, load.NewSource(basePath))
	if err != nil {
		return nil, fmt.Errorf("sdk.Diff: failed to load base spec: %w", err)
	}

	revSpec, err := load.NewSpecInfo(loader, load.NewSource(revisionPath))
	if err != nil {
		return nil, fmt.Errorf("sdk.Diff: failed to load revision spec: %w", err)
	}

	// Compute diff with operations sources map for better error reporting
	diffResult, operationsSources, err := diff.GetWithOperationsSourcesMap(diff.NewConfig(), baseSpec, revSpec)
	if err != nil {
		return nil, fmt.Errorf("sdk.Diff: failed to compute diff: %w", err)
	}

	// Check for breaking changes
	config := checker.NewConfig(checker.GetAllChecks())
	breaks := checker.CheckBackwardCompatibilityUntilLevel(
		config,
		diffResult,
		operationsSources,
		checker.ERR, // Only errors (breaking changes)
	)

	// Build result
	result := &DiffResult{
		Breaking: len(breaks) > 0,
		Changes:  make([]string, 0, len(breaks)),
	}

	localizer := checker.NewDefaultLocalizer()
	for _, b := range breaks {
		result.Changes = append(result.Changes, b.GetUncolorizedText(localizer))
	}

	if result.Breaking {
		result.Summary = fmt.Sprintf("%d breaking change(s) detected", len(breaks))
	} else {
		result.Summary = "No breaking changes"
	}

	return result, nil
}

// DiffExitCode returns the exit code for CI integration.
// 0 = no breaking changes, 1 = breaking changes, 2 = error
func DiffExitCode(result *DiffResult, err error) int {
	if err != nil {
		return 2
	}
	if result.Breaking {
		return 1
	}
	return 0
}
