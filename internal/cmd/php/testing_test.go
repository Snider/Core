package php

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectTestRunner_Good(t *testing.T) {
	t.Run("detects Pest when tests/Pest.php exists", func(t *testing.T) {
		dir := t.TempDir()
		testsDir := filepath.Join(dir, "tests")
		err := os.MkdirAll(testsDir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(testsDir, "Pest.php"), []byte("<?php\n"), 0644)
		require.NoError(t, err)

		runner := DetectTestRunner(dir)
		assert.Equal(t, TestRunnerPest, runner)
	})

	t.Run("returns PHPUnit when no Pest.php", func(t *testing.T) {
		dir := t.TempDir()

		runner := DetectTestRunner(dir)
		assert.Equal(t, TestRunnerPHPUnit, runner)
	})

	t.Run("returns PHPUnit when tests directory exists but no Pest.php", func(t *testing.T) {
		dir := t.TempDir()
		testsDir := filepath.Join(dir, "tests")
		err := os.MkdirAll(testsDir, 0755)
		require.NoError(t, err)

		runner := DetectTestRunner(dir)
		assert.Equal(t, TestRunnerPHPUnit, runner)
	})
}

func TestBuildPestCommand_Good(t *testing.T) {
	t.Run("basic command", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir}

		cmd, args := buildPestCommand(opts)
		assert.Equal(t, "pest", cmd)
		assert.Empty(t, args)
	})

	t.Run("with filter", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir, Filter: "UserTest"}

		_, args := buildPestCommand(opts)
		assert.Contains(t, args, "--filter")
		assert.Contains(t, args, "UserTest")
	})

	t.Run("with parallel", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir, Parallel: true}

		_, args := buildPestCommand(opts)
		assert.Contains(t, args, "--parallel")
	})

	t.Run("with coverage", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir, Coverage: true}

		_, args := buildPestCommand(opts)
		assert.Contains(t, args, "--coverage")
	})

	t.Run("with coverage HTML format", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir, Coverage: true, CoverageFormat: "html"}

		_, args := buildPestCommand(opts)
		assert.Contains(t, args, "--coverage-html")
		assert.Contains(t, args, "coverage")
	})

	t.Run("with coverage clover format", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir, Coverage: true, CoverageFormat: "clover"}

		_, args := buildPestCommand(opts)
		assert.Contains(t, args, "--coverage-clover")
		assert.Contains(t, args, "coverage.xml")
	})

	t.Run("with groups", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir, Groups: []string{"unit", "integration"}}

		_, args := buildPestCommand(opts)
		assert.Contains(t, args, "--group")
		assert.Contains(t, args, "unit")
		assert.Contains(t, args, "integration")
	})

	t.Run("uses vendor binary when exists", func(t *testing.T) {
		dir := t.TempDir()
		binDir := filepath.Join(dir, "vendor", "bin")
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)

		pestPath := filepath.Join(binDir, "pest")
		err = os.WriteFile(pestPath, []byte("#!/bin/bash"), 0755)
		require.NoError(t, err)

		opts := TestOptions{Dir: dir}
		cmd, _ := buildPestCommand(opts)
		assert.Equal(t, pestPath, cmd)
	})

	t.Run("all options combined", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{
			Dir:            dir,
			Filter:         "Test",
			Parallel:       true,
			Coverage:       true,
			CoverageFormat: "html",
			Groups:         []string{"unit"},
		}

		_, args := buildPestCommand(opts)
		assert.Contains(t, args, "--filter")
		assert.Contains(t, args, "--parallel")
		assert.Contains(t, args, "--coverage-html")
		assert.Contains(t, args, "--group")
	})
}

func TestBuildPHPUnitCommand_Good(t *testing.T) {
	t.Run("basic command", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir}

		cmd, args := buildPHPUnitCommand(opts)
		assert.Equal(t, "phpunit", cmd)
		assert.Empty(t, args)
	})

	t.Run("with filter", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir, Filter: "UserTest"}

		_, args := buildPHPUnitCommand(opts)
		assert.Contains(t, args, "--filter")
		assert.Contains(t, args, "UserTest")
	})

	t.Run("with parallel uses paratest", func(t *testing.T) {
		dir := t.TempDir()
		binDir := filepath.Join(dir, "vendor", "bin")
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)

		paratestPath := filepath.Join(binDir, "paratest")
		err = os.WriteFile(paratestPath, []byte("#!/bin/bash"), 0755)
		require.NoError(t, err)

		opts := TestOptions{Dir: dir, Parallel: true}
		cmd, _ := buildPHPUnitCommand(opts)
		assert.Equal(t, paratestPath, cmd)
	})

	t.Run("parallel without paratest stays phpunit", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir, Parallel: true}

		cmd, _ := buildPHPUnitCommand(opts)
		assert.Equal(t, "phpunit", cmd)
	})

	t.Run("with coverage", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir, Coverage: true}

		_, args := buildPHPUnitCommand(opts)
		assert.Contains(t, args, "--coverage-text")
	})

	t.Run("with coverage HTML format", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir, Coverage: true, CoverageFormat: "html"}

		_, args := buildPHPUnitCommand(opts)
		assert.Contains(t, args, "--coverage-html")
		assert.Contains(t, args, "coverage")
	})

	t.Run("with coverage clover format", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir, Coverage: true, CoverageFormat: "clover"}

		_, args := buildPHPUnitCommand(opts)
		assert.Contains(t, args, "--coverage-clover")
		assert.Contains(t, args, "coverage.xml")
	})

	t.Run("with groups", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{Dir: dir, Groups: []string{"unit", "integration"}}

		_, args := buildPHPUnitCommand(opts)
		assert.Contains(t, args, "--group")
		assert.Contains(t, args, "unit")
		assert.Contains(t, args, "integration")
	})

	t.Run("uses vendor binary when exists", func(t *testing.T) {
		dir := t.TempDir()
		binDir := filepath.Join(dir, "vendor", "bin")
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)

		phpunitPath := filepath.Join(binDir, "phpunit")
		err = os.WriteFile(phpunitPath, []byte("#!/bin/bash"), 0755)
		require.NoError(t, err)

		opts := TestOptions{Dir: dir}
		cmd, _ := buildPHPUnitCommand(opts)
		assert.Equal(t, phpunitPath, cmd)
	})
}

func TestTestOptions_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := TestOptions{
			Dir:            "/test",
			Filter:         "TestName",
			Parallel:       true,
			Coverage:       true,
			CoverageFormat: "html",
			Groups:         []string{"unit"},
			Output:         os.Stdout,
		}

		assert.Equal(t, "/test", opts.Dir)
		assert.Equal(t, "TestName", opts.Filter)
		assert.True(t, opts.Parallel)
		assert.True(t, opts.Coverage)
		assert.Equal(t, "html", opts.CoverageFormat)
		assert.Equal(t, []string{"unit"}, opts.Groups)
		assert.NotNil(t, opts.Output)
	})
}

func TestTestRunner_Constants(t *testing.T) {
	t.Run("constants are defined", func(t *testing.T) {
		assert.Equal(t, TestRunner("pest"), TestRunnerPest)
		assert.Equal(t, TestRunner("phpunit"), TestRunnerPHPUnit)
	})
}

func TestRunTests_Bad(t *testing.T) {
	t.Skip("requires PHP test runner installed")
}

func TestRunParallel_Bad(t *testing.T) {
	t.Skip("requires PHP test runner installed")
}

func TestRunTests_Integration(t *testing.T) {
	t.Skip("requires PHP/Pest/PHPUnit installed")
}

func TestBuildPestCommand_CoverageOptions(t *testing.T) {
	tests := []struct {
		name           string
		coverageFormat string
		expectedArg    string
	}{
		{"default coverage", "", "--coverage"},
		{"html coverage", "html", "--coverage-html"},
		{"clover coverage", "clover", "--coverage-clover"},
		{"unknown format uses default", "unknown", "--coverage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			opts := TestOptions{
				Dir:            dir,
				Coverage:       true,
				CoverageFormat: tt.coverageFormat,
			}

			_, args := buildPestCommand(opts)

			// For unknown format, should fall through to default
			if tt.coverageFormat == "unknown" {
				assert.Contains(t, args, "--coverage")
			} else {
				assert.Contains(t, args, tt.expectedArg)
			}
		})
	}
}

func TestBuildPHPUnitCommand_CoverageOptions(t *testing.T) {
	tests := []struct {
		name           string
		coverageFormat string
		expectedArg    string
	}{
		{"default coverage", "", "--coverage-text"},
		{"html coverage", "html", "--coverage-html"},
		{"clover coverage", "clover", "--coverage-clover"},
		{"unknown format uses default", "unknown", "--coverage-text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			opts := TestOptions{
				Dir:            dir,
				Coverage:       true,
				CoverageFormat: tt.coverageFormat,
			}

			_, args := buildPHPUnitCommand(opts)

			if tt.coverageFormat == "unknown" {
				assert.Contains(t, args, "--coverage-text")
			} else {
				assert.Contains(t, args, tt.expectedArg)
			}
		})
	}
}

func TestBuildPestCommand_MultipleGroups(t *testing.T) {
	t.Run("adds multiple group flags", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{
			Dir:    dir,
			Groups: []string{"unit", "integration", "feature"},
		}

		_, args := buildPestCommand(opts)

		// Should have --group for each group
		groupCount := 0
		for _, arg := range args {
			if arg == "--group" {
				groupCount++
			}
		}
		assert.Equal(t, 3, groupCount)
	})
}

func TestBuildPHPUnitCommand_MultipleGroups(t *testing.T) {
	t.Run("adds multiple group flags", func(t *testing.T) {
		dir := t.TempDir()
		opts := TestOptions{
			Dir:    dir,
			Groups: []string{"unit", "integration"},
		}

		_, args := buildPHPUnitCommand(opts)

		groupCount := 0
		for _, arg := range args {
			if arg == "--group" {
				groupCount++
			}
		}
		assert.Equal(t, 2, groupCount)
	})
}
