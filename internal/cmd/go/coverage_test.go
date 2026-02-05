package gocmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateBlockCoverage(t *testing.T) {
	// Create a dummy coverage profile
	content := `mode: set
github.com/host-uk/core/pkg/foo.go:1.2,3.4 5 1
github.com/host-uk/core/pkg/foo.go:5.6,7.8 2 0
github.com/host-uk/core/pkg/bar.go:10.1,12.20 10 5
`
	tmpfile, err := os.CreateTemp("", "test-coverage-*.out")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(content))
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	// Test calculation
	// 3 blocks total, 2 covered (count > 0)
	// Expect (2/3) * 100 = 66.666...
	pct, err := calculateBlockCoverage(tmpfile.Name())
	assert.NoError(t, err)
	assert.InDelta(t, 66.67, pct, 0.01)

	// Test empty file (only header)
	contentEmpty := "mode: atomic\n"
	tmpfileEmpty, _ := os.CreateTemp("", "test-coverage-empty-*.out")
	defer os.Remove(tmpfileEmpty.Name())
	tmpfileEmpty.Write([]byte(contentEmpty))
	tmpfileEmpty.Close()

	pct, err = calculateBlockCoverage(tmpfileEmpty.Name())
	assert.NoError(t, err)
	assert.Equal(t, 0.0, pct)

	// Test non-existent file
	pct, err = calculateBlockCoverage("non-existent-file")
	assert.Error(t, err)
	assert.Equal(t, 0.0, pct)

	// Test malformed file
	contentMalformed := `mode: set
github.com/host-uk/core/pkg/foo.go:1.2,3.4 5
github.com/host-uk/core/pkg/foo.go:1.2,3.4 5 notanumber
`
	tmpfileMalformed, _ := os.CreateTemp("", "test-coverage-malformed-*.out")
	defer os.Remove(tmpfileMalformed.Name())
	tmpfileMalformed.Write([]byte(contentMalformed))
	tmpfileMalformed.Close()

	pct, err = calculateBlockCoverage(tmpfileMalformed.Name())
	assert.NoError(t, err)
	assert.Equal(t, 0.0, pct)
}

func TestParseOverallCoverage(t *testing.T) {
	output := `ok  	github.com/host-uk/core/pkg/foo	0.100s	coverage: 50.0% of statements
ok  	github.com/host-uk/core/pkg/bar	0.200s	coverage: 100.0% of statements
`
	pct := parseOverallCoverage(output)
	assert.Equal(t, 75.0, pct)

	outputNoCov := "ok  	github.com/host-uk/core/pkg/foo	0.100s"
	pct = parseOverallCoverage(outputNoCov)
	assert.Equal(t, 0.0, pct)
}

func TestFormatCoverage(t *testing.T) {
	// Since formatCoverage uses cli.SuccessStyle etc which might rely on a global state
	// or terminal, we just test if it returns a non-empty string for now.
	// Actually, we can check if it contains the percentage.
	assert.Contains(t, formatCoverage(85.0), "85.0%")
	assert.Contains(t, formatCoverage(65.0), "65.0%")
	assert.Contains(t, formatCoverage(25.0), "25.0%")
}
