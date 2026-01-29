package php

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectFormatter_Good(t *testing.T) {
	t.Run("detects pint.json", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "pint.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		formatter, found := DetectFormatter(dir)
		assert.True(t, found)
		assert.Equal(t, FormatterPint, formatter)
	})

	t.Run("detects vendor binary", func(t *testing.T) {
		dir := t.TempDir()
		binDir := filepath.Join(dir, "vendor", "bin")
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(binDir, "pint"), []byte(""), 0755)
		require.NoError(t, err)

		formatter, found := DetectFormatter(dir)
		assert.True(t, found)
		assert.Equal(t, FormatterPint, formatter)
	})
}

func TestDetectFormatter_Bad(t *testing.T) {
	t.Run("no formatter", func(t *testing.T) {
		dir := t.TempDir()
		_, found := DetectFormatter(dir)
		assert.False(t, found)
	})
}

func TestDetectAnalyser_Good(t *testing.T) {
	t.Run("detects phpstan.neon", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "phpstan.neon"), []byte(""), 0644)
		require.NoError(t, err)

		analyser, found := DetectAnalyser(dir)
		assert.True(t, found)
		assert.Equal(t, AnalyserPHPStan, analyser)
	})

	t.Run("detects phpstan.neon.dist", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "phpstan.neon.dist"), []byte(""), 0644)
		require.NoError(t, err)

		analyser, found := DetectAnalyser(dir)
		assert.True(t, found)
		assert.Equal(t, AnalyserPHPStan, analyser)
	})

	t.Run("detects larastan", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "phpstan.neon"), []byte(""), 0644)
		require.NoError(t, err)

		larastanDir := filepath.Join(dir, "vendor", "larastan", "larastan")
		err = os.MkdirAll(larastanDir, 0755)
		require.NoError(t, err)

		analyser, found := DetectAnalyser(dir)
		assert.True(t, found)
		assert.Equal(t, AnalyserLarastan, analyser)
	})

	t.Run("detects nunomaduro/larastan", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "phpstan.neon"), []byte(""), 0644)
		require.NoError(t, err)

		larastanDir := filepath.Join(dir, "vendor", "nunomaduro", "larastan")
		err = os.MkdirAll(larastanDir, 0755)
		require.NoError(t, err)

		analyser, found := DetectAnalyser(dir)
		assert.True(t, found)
		assert.Equal(t, AnalyserLarastan, analyser)
	})
}

func TestBuildPintCommand_Good(t *testing.T) {
	t.Run("basic command", func(t *testing.T) {
		dir := t.TempDir()
		opts := FormatOptions{Dir: dir}
		cmd, args := buildPintCommand(opts)
		assert.Equal(t, "pint", cmd)
		assert.Contains(t, args, "--test")
	})

	t.Run("fix enabled", func(t *testing.T) {
		dir := t.TempDir()
		opts := FormatOptions{Dir: dir, Fix: true}
		_, args := buildPintCommand(opts)
		assert.NotContains(t, args, "--test")
	})

	t.Run("diff enabled", func(t *testing.T) {
		dir := t.TempDir()
		opts := FormatOptions{Dir: dir, Diff: true}
		_, args := buildPintCommand(opts)
		assert.Contains(t, args, "--diff")
	})

	t.Run("with specific paths", func(t *testing.T) {
		dir := t.TempDir()
		paths := []string{"app", "tests"}
		opts := FormatOptions{Dir: dir, Paths: paths}
		_, args := buildPintCommand(opts)
		assert.Equal(t, paths, args[len(args)-2:])
	})

	t.Run("uses vendor binary if exists", func(t *testing.T) {
		dir := t.TempDir()
		binDir := filepath.Join(dir, "vendor", "bin")
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)
		pintPath := filepath.Join(binDir, "pint")
		err = os.WriteFile(pintPath, []byte(""), 0755)
		require.NoError(t, err)

		opts := FormatOptions{Dir: dir}
		cmd, _ := buildPintCommand(opts)
		assert.Equal(t, pintPath, cmd)
	})
}

func TestBuildPHPStanCommand_Good(t *testing.T) {
	t.Run("basic command", func(t *testing.T) {
		dir := t.TempDir()
		opts := AnalyseOptions{Dir: dir}
		cmd, args := buildPHPStanCommand(opts)
		assert.Equal(t, "phpstan", cmd)
		assert.Equal(t, []string{"analyse"}, args)
	})

	t.Run("with level", func(t *testing.T) {
		dir := t.TempDir()
		opts := AnalyseOptions{Dir: dir, Level: 5}
		_, args := buildPHPStanCommand(opts)
		assert.Contains(t, args, "--level")
		assert.Contains(t, args, "5")
	})

	t.Run("with memory limit", func(t *testing.T) {
		dir := t.TempDir()
		opts := AnalyseOptions{Dir: dir, Memory: "2G"}
		_, args := buildPHPStanCommand(opts)
		assert.Contains(t, args, "--memory-limit")
		assert.Contains(t, args, "2G")
	})

	t.Run("uses vendor binary if exists", func(t *testing.T) {
		dir := t.TempDir()
		binDir := filepath.Join(dir, "vendor", "bin")
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)
		phpstanPath := filepath.Join(binDir, "phpstan")
		err = os.WriteFile(phpstanPath, []byte(""), 0755)
		require.NoError(t, err)

		opts := AnalyseOptions{Dir: dir}
		cmd, _ := buildPHPStanCommand(opts)
		assert.Equal(t, phpstanPath, cmd)
	})
}
