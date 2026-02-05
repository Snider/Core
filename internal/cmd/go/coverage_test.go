package gocmd

import (
	"os"
	"testing"

	"github.com/host-uk/core/pkg/cli"
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

	// Test malformed file - missing fields
	contentMalformed2 := `mode: set
github.com/host-uk/core/pkg/foo.go:1.2,3.4 5
`
	tmpfileMalformed2, _ := os.CreateTemp("", "test-coverage-malformed2-*.out")
	defer os.Remove(tmpfileMalformed2.Name())
	tmpfileMalformed2.Write([]byte(contentMalformed2))
	tmpfileMalformed2.Close()

	pct, err = calculateBlockCoverage(tmpfileMalformed2.Name())
	assert.NoError(t, err)
	assert.Equal(t, 0.0, pct)

	// Test completely empty file
	tmpfileEmpty2, _ := os.CreateTemp("", "test-coverage-empty2-*.out")
	defer os.Remove(tmpfileEmpty2.Name())
	tmpfileEmpty2.Close()
	pct, err = calculateBlockCoverage(tmpfileEmpty2.Name())
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
	assert.Contains(t, formatCoverage(85.0), "85.0%")
	assert.Contains(t, formatCoverage(65.0), "65.0%")
	assert.Contains(t, formatCoverage(25.0), "25.0%")
}

func TestAddGoCovCommand(t *testing.T) {
	cmd := &cli.Command{Use: "test"}
	addGoCovCommand(cmd)
	assert.True(t, cmd.HasSubCommands())
	sub := cmd.Commands()[0]
	assert.Equal(t, "cov", sub.Name())
}

func TestAddGoQACommand(t *testing.T) {
	cmd := &cli.Command{Use: "test"}
	addGoQACommand(cmd)
	assert.True(t, cmd.HasSubCommands())
	sub := cmd.Commands()[0]
	assert.Equal(t, "qa", sub.Name())
}

func TestDetermineChecks(t *testing.T) {
	// Default checks
	qaOnly = ""
	qaSkip = ""
	qaRace = false
	qaBench = false
	checks := determineChecks()
	assert.Contains(t, checks, "fmt")
	assert.Contains(t, checks, "test")

	// Only
	qaOnly = "fmt,lint"
	checks = determineChecks()
	assert.Equal(t, []string{"fmt", "lint"}, checks)

	// Skip
	qaOnly = ""
	qaSkip = "fmt,lint"
	checks = determineChecks()
	assert.NotContains(t, checks, "fmt")
	assert.NotContains(t, checks, "lint")
	assert.Contains(t, checks, "test")

	// Race
	qaSkip = ""
	qaRace = true
	checks = determineChecks()
	assert.Contains(t, checks, "race")
	assert.NotContains(t, checks, "test")

	// Reset
	qaRace = false
}

func TestBuildCheck(t *testing.T) {
	qaFix = false
	c := buildCheck("fmt")
	assert.Equal(t, "format", c.Name)
	assert.Equal(t, []string{"-l", "."}, c.Args)

	qaFix = true
	c = buildCheck("fmt")
	assert.Equal(t, []string{"-w", "."}, c.Args)

	c = buildCheck("vet")
	assert.Equal(t, "vet", c.Name)

	c = buildCheck("lint")
	assert.Equal(t, "lint", c.Name)

	c = buildCheck("test")
	assert.Equal(t, "test", c.Name)

	c = buildCheck("race")
	assert.Equal(t, "race", c.Name)

	c = buildCheck("bench")
	assert.Equal(t, "bench", c.Name)

	c = buildCheck("vuln")
	assert.Equal(t, "vuln", c.Name)

	c = buildCheck("sec")
	assert.Equal(t, "sec", c.Name)

	c = buildCheck("fuzz")
	assert.Equal(t, "fuzz", c.Name)

	c = buildCheck("docblock")
	assert.Equal(t, "docblock", c.Name)

	c = buildCheck("unknown")
	assert.Equal(t, "", c.Name)
}

func TestBuildChecks(t *testing.T) {
	checks := buildChecks([]string{"fmt", "vet", "unknown"})
	assert.Equal(t, 2, len(checks))
	assert.Equal(t, "format", checks[0].Name)
	assert.Equal(t, "vet", checks[1].Name)
}

func TestFixHintFor(t *testing.T) {
	assert.Contains(t, fixHintFor("format", ""), "core go qa fmt --fix")
	assert.Contains(t, fixHintFor("vet", ""), "go vet")
	assert.Contains(t, fixHintFor("lint", ""), "core go qa lint --fix")
	assert.Contains(t, fixHintFor("test", "--- FAIL: TestFoo"), "TestFoo")
	assert.Contains(t, fixHintFor("race", ""), "Data race")
	assert.Contains(t, fixHintFor("bench", ""), "Benchmark regression")
	assert.Contains(t, fixHintFor("vuln", ""), "govulncheck")
	assert.Contains(t, fixHintFor("sec", ""), "gosec")
	assert.Contains(t, fixHintFor("fuzz", ""), "crashing input")
	assert.Contains(t, fixHintFor("docblock", ""), "doc comments")
	assert.Equal(t, "", fixHintFor("unknown", ""))
}

func TestRunGoQA_NoGoMod(t *testing.T) {
	// runGoQA should fail if go.mod is not present in CWD
	// We run it in a temp dir without go.mod
	tmpDir, _ := os.MkdirTemp("", "test-qa-*")
	defer os.RemoveAll(tmpDir)
	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	cmd := &cli.Command{Use: "qa"}
	err := runGoQA(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no go.mod found")
}
