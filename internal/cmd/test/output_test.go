package testcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortenPackageName(t *testing.T) {
	assert.Equal(t, "pkg/foo", shortenPackageName("github.com/host-uk/core/pkg/foo"))
	assert.Equal(t, "core-php", shortenPackageName("github.com/host-uk/core-php"))
	assert.Equal(t, "bar", shortenPackageName("github.com/other/bar"))
}

func TestFormatCoverageTest(t *testing.T) {
	assert.Contains(t, formatCoverage(85.0), "85.0%")
	assert.Contains(t, formatCoverage(65.0), "65.0%")
	assert.Contains(t, formatCoverage(25.0), "25.0%")
}

func TestParseTestOutput(t *testing.T) {
	output := `ok  	github.com/host-uk/core/pkg/foo	0.100s	coverage: 50.0% of statements
FAIL	github.com/host-uk/core/pkg/bar
?   	github.com/host-uk/core/pkg/baz	[no test files]
`
	results := parseTestOutput(output)
	assert.Equal(t, 1, results.passed)
	assert.Equal(t, 1, results.failed)
	assert.Equal(t, 1, results.skipped)
	assert.Equal(t, 1, len(results.failedPkgs))
	assert.Equal(t, "github.com/host-uk/core/pkg/bar", results.failedPkgs[0])
	assert.Equal(t, 1, len(results.packages))
	assert.Equal(t, 50.0, results.packages[0].coverage)
}

func TestPrintCoverageSummarySafe(t *testing.T) {
	// This tests the bug fix for long package names causing negative Repeat count
	results := testResults{
		packages: []packageCoverage{
			{name: "github.com/host-uk/core/pkg/short", coverage: 100, hasCov: true},
			{name: "github.com/host-uk/core/pkg/a-very-very-very-very-very-long-package-name-that-might-cause-issues", coverage: 80, hasCov: true},
		},
		passed:   2,
		totalCov: 180,
		covCount: 2,
	}

	// Should not panic
	assert.NotPanics(t, func() {
		printCoverageSummary(results)
	})
}
