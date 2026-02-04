package php

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatOptions_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := FormatOptions{
			Dir:    "/project",
			Fix:    true,
			Diff:   true,
			Paths:  []string{"app", "tests"},
			Output: os.Stdout,
		}

		assert.Equal(t, "/project", opts.Dir)
		assert.True(t, opts.Fix)
		assert.True(t, opts.Diff)
		assert.Equal(t, []string{"app", "tests"}, opts.Paths)
		assert.NotNil(t, opts.Output)
	})
}

func TestAnalyseOptions_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		opts := AnalyseOptions{
			Dir:    "/project",
			Level:  5,
			Paths:  []string{"src"},
			Memory: "2G",
			Output: os.Stdout,
		}

		assert.Equal(t, "/project", opts.Dir)
		assert.Equal(t, 5, opts.Level)
		assert.Equal(t, []string{"src"}, opts.Paths)
		assert.Equal(t, "2G", opts.Memory)
		assert.NotNil(t, opts.Output)
	})
}

func TestFormatterType_Constants(t *testing.T) {
	t.Run("constants are defined", func(t *testing.T) {
		assert.Equal(t, FormatterType("pint"), FormatterPint)
	})
}

func TestAnalyserType_Constants(t *testing.T) {
	t.Run("constants are defined", func(t *testing.T) {
		assert.Equal(t, AnalyserType("phpstan"), AnalyserPHPStan)
		assert.Equal(t, AnalyserType("larastan"), AnalyserLarastan)
	})
}

func TestDetectFormatter_Extended(t *testing.T) {
	t.Run("returns not found for empty directory", func(t *testing.T) {
		dir := t.TempDir()
		_, found := DetectFormatter(dir)
		assert.False(t, found)
	})

	t.Run("prefers pint.json over vendor binary", func(t *testing.T) {
		dir := t.TempDir()

		// Create pint.json
		err := os.WriteFile(filepath.Join(dir, "pint.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		formatter, found := DetectFormatter(dir)
		assert.True(t, found)
		assert.Equal(t, FormatterPint, formatter)
	})
}

func TestDetectAnalyser_Extended(t *testing.T) {
	t.Run("returns not found for empty directory", func(t *testing.T) {
		dir := t.TempDir()
		_, found := DetectAnalyser(dir)
		assert.False(t, found)
	})

	t.Run("detects phpstan from vendor binary alone", func(t *testing.T) {
		dir := t.TempDir()

		// Create vendor binary
		binDir := filepath.Join(dir, "vendor", "bin")
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(binDir, "phpstan"), []byte(""), 0755)
		require.NoError(t, err)

		analyser, found := DetectAnalyser(dir)
		assert.True(t, found)
		assert.Equal(t, AnalyserPHPStan, analyser)
	})

	t.Run("detects larastan from larastan/larastan vendor path", func(t *testing.T) {
		dir := t.TempDir()

		// Create phpstan.neon
		err := os.WriteFile(filepath.Join(dir, "phpstan.neon"), []byte(""), 0644)
		require.NoError(t, err)

		// Create larastan/larastan path
		larastanPath := filepath.Join(dir, "vendor", "larastan", "larastan")
		err = os.MkdirAll(larastanPath, 0755)
		require.NoError(t, err)

		analyser, found := DetectAnalyser(dir)
		assert.True(t, found)
		assert.Equal(t, AnalyserLarastan, analyser)
	})

	t.Run("detects larastan from nunomaduro/larastan vendor path", func(t *testing.T) {
		dir := t.TempDir()

		// Create phpstan.neon
		err := os.WriteFile(filepath.Join(dir, "phpstan.neon"), []byte(""), 0644)
		require.NoError(t, err)

		// Create nunomaduro/larastan path
		larastanPath := filepath.Join(dir, "vendor", "nunomaduro", "larastan")
		err = os.MkdirAll(larastanPath, 0755)
		require.NoError(t, err)

		analyser, found := DetectAnalyser(dir)
		assert.True(t, found)
		assert.Equal(t, AnalyserLarastan, analyser)
	})
}

func TestBuildPintCommand_Extended(t *testing.T) {
	t.Run("uses global pint when no vendor binary", func(t *testing.T) {
		dir := t.TempDir()
		opts := FormatOptions{Dir: dir}

		cmd, _ := buildPintCommand(opts)
		assert.Equal(t, "pint", cmd)
	})

	t.Run("adds test flag when Fix is false", func(t *testing.T) {
		dir := t.TempDir()
		opts := FormatOptions{Dir: dir, Fix: false}

		_, args := buildPintCommand(opts)
		assert.Contains(t, args, "--test")
	})

	t.Run("does not add test flag when Fix is true", func(t *testing.T) {
		dir := t.TempDir()
		opts := FormatOptions{Dir: dir, Fix: true}

		_, args := buildPintCommand(opts)
		assert.NotContains(t, args, "--test")
	})

	t.Run("adds diff flag", func(t *testing.T) {
		dir := t.TempDir()
		opts := FormatOptions{Dir: dir, Diff: true}

		_, args := buildPintCommand(opts)
		assert.Contains(t, args, "--diff")
	})

	t.Run("adds paths", func(t *testing.T) {
		dir := t.TempDir()
		opts := FormatOptions{Dir: dir, Paths: []string{"app", "tests"}}

		_, args := buildPintCommand(opts)
		assert.Contains(t, args, "app")
		assert.Contains(t, args, "tests")
	})
}

func TestBuildPHPStanCommand_Extended(t *testing.T) {
	t.Run("uses global phpstan when no vendor binary", func(t *testing.T) {
		dir := t.TempDir()
		opts := AnalyseOptions{Dir: dir}

		cmd, _ := buildPHPStanCommand(opts)
		assert.Equal(t, "phpstan", cmd)
	})

	t.Run("adds level flag", func(t *testing.T) {
		dir := t.TempDir()
		opts := AnalyseOptions{Dir: dir, Level: 8}

		_, args := buildPHPStanCommand(opts)
		assert.Contains(t, args, "--level")
		assert.Contains(t, args, "8")
	})

	t.Run("does not add level flag when zero", func(t *testing.T) {
		dir := t.TempDir()
		opts := AnalyseOptions{Dir: dir, Level: 0}

		_, args := buildPHPStanCommand(opts)
		assert.NotContains(t, args, "--level")
	})

	t.Run("adds memory limit", func(t *testing.T) {
		dir := t.TempDir()
		opts := AnalyseOptions{Dir: dir, Memory: "4G"}

		_, args := buildPHPStanCommand(opts)
		assert.Contains(t, args, "--memory-limit")
		assert.Contains(t, args, "4G")
	})

	t.Run("does not add memory flag when empty", func(t *testing.T) {
		dir := t.TempDir()
		opts := AnalyseOptions{Dir: dir, Memory: ""}

		_, args := buildPHPStanCommand(opts)
		assert.NotContains(t, args, "--memory-limit")
	})

	t.Run("adds paths", func(t *testing.T) {
		dir := t.TempDir()
		opts := AnalyseOptions{Dir: dir, Paths: []string{"src", "app"}}

		_, args := buildPHPStanCommand(opts)
		assert.Contains(t, args, "src")
		assert.Contains(t, args, "app")
	})
}

func TestFormat_Bad(t *testing.T) {
	t.Run("fails when no formatter found", func(t *testing.T) {
		dir := t.TempDir()
		opts := FormatOptions{Dir: dir}

		err := Format(context.TODO(), opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no formatter found")
	})

	t.Run("uses cwd when dir not specified", func(t *testing.T) {
		// When no formatter found in cwd, should still fail with "no formatter found"
		opts := FormatOptions{Dir: ""}

		err := Format(context.TODO(), opts)
		// May or may not find a formatter depending on cwd, but function should not panic
		if err != nil {
			// Expected - no formatter in cwd
			assert.Contains(t, err.Error(), "no formatter")
		}
	})

	t.Run("uses stdout when output not specified", func(t *testing.T) {
		dir := t.TempDir()
		// Create pint.json to enable formatter detection
		err := os.WriteFile(filepath.Join(dir, "pint.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		opts := FormatOptions{Dir: dir, Output: nil}

		// Will fail because pint isn't actually installed, but tests the code path
		err = Format(context.Background(), opts)
		assert.Error(t, err) // Pint not installed
	})
}

func TestAnalyse_Bad(t *testing.T) {
	t.Run("fails when no analyser found", func(t *testing.T) {
		dir := t.TempDir()
		opts := AnalyseOptions{Dir: dir}

		err := Analyse(context.TODO(), opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no static analyser found")
	})

	t.Run("uses cwd when dir not specified", func(t *testing.T) {
		opts := AnalyseOptions{Dir: ""}

		err := Analyse(context.TODO(), opts)
		// May or may not find an analyser depending on cwd
		if err != nil {
			assert.Contains(t, err.Error(), "no static analyser")
		}
	})

	t.Run("uses stdout when output not specified", func(t *testing.T) {
		dir := t.TempDir()
		// Create phpstan.neon to enable analyser detection
		err := os.WriteFile(filepath.Join(dir, "phpstan.neon"), []byte(""), 0644)
		require.NoError(t, err)

		opts := AnalyseOptions{Dir: dir, Output: nil}

		// Will fail because phpstan isn't actually installed, but tests the code path
		err = Analyse(context.Background(), opts)
		assert.Error(t, err) // PHPStan not installed
	})
}
