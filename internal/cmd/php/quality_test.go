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

// =============================================================================
// Psalm Detection Tests
// =============================================================================

func TestDetectPsalm_Good(t *testing.T) {
	t.Run("detects psalm.xml", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "psalm.xml"), []byte(""), 0644)
		require.NoError(t, err)

		// Also need vendor binary for it to return true
		binDir := filepath.Join(dir, "vendor", "bin")
		err = os.MkdirAll(binDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(binDir, "psalm"), []byte(""), 0755)
		require.NoError(t, err)

		psalmType, found := DetectPsalm(dir)
		assert.True(t, found)
		assert.Equal(t, PsalmStandard, psalmType)
	})

	t.Run("detects psalm.xml.dist", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "psalm.xml.dist"), []byte(""), 0644)
		require.NoError(t, err)

		binDir := filepath.Join(dir, "vendor", "bin")
		err = os.MkdirAll(binDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(binDir, "psalm"), []byte(""), 0755)
		require.NoError(t, err)

		_, found := DetectPsalm(dir)
		assert.True(t, found)
	})

	t.Run("detects vendor binary only", func(t *testing.T) {
		dir := t.TempDir()
		binDir := filepath.Join(dir, "vendor", "bin")
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(binDir, "psalm"), []byte(""), 0755)
		require.NoError(t, err)

		_, found := DetectPsalm(dir)
		assert.True(t, found)
	})
}

func TestDetectPsalm_Bad(t *testing.T) {
	t.Run("no psalm", func(t *testing.T) {
		dir := t.TempDir()
		_, found := DetectPsalm(dir)
		assert.False(t, found)
	})
}

// =============================================================================
// Rector Detection Tests
// =============================================================================

func TestDetectRector_Good(t *testing.T) {
	t.Run("detects rector.php", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "rector.php"), []byte("<?php"), 0644)
		require.NoError(t, err)

		found := DetectRector(dir)
		assert.True(t, found)
	})

	t.Run("detects vendor binary", func(t *testing.T) {
		dir := t.TempDir()
		binDir := filepath.Join(dir, "vendor", "bin")
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(binDir, "rector"), []byte(""), 0755)
		require.NoError(t, err)

		found := DetectRector(dir)
		assert.True(t, found)
	})
}

func TestDetectRector_Bad(t *testing.T) {
	t.Run("no rector", func(t *testing.T) {
		dir := t.TempDir()
		found := DetectRector(dir)
		assert.False(t, found)
	})
}

// =============================================================================
// Infection Detection Tests
// =============================================================================

func TestDetectInfection_Good(t *testing.T) {
	t.Run("detects infection.json", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "infection.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		found := DetectInfection(dir)
		assert.True(t, found)
	})

	t.Run("detects infection.json5", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "infection.json5"), []byte("{}"), 0644)
		require.NoError(t, err)

		found := DetectInfection(dir)
		assert.True(t, found)
	})

	t.Run("detects vendor binary", func(t *testing.T) {
		dir := t.TempDir()
		binDir := filepath.Join(dir, "vendor", "bin")
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(binDir, "infection"), []byte(""), 0755)
		require.NoError(t, err)

		found := DetectInfection(dir)
		assert.True(t, found)
	})
}

func TestDetectInfection_Bad(t *testing.T) {
	t.Run("no infection", func(t *testing.T) {
		dir := t.TempDir()
		found := DetectInfection(dir)
		assert.False(t, found)
	})
}

// =============================================================================
// QA Pipeline Tests
// =============================================================================

func TestGetQAStages_Good(t *testing.T) {
	t.Run("default stages", func(t *testing.T) {
		opts := QAOptions{}
		stages := GetQAStages(opts)
		assert.Equal(t, []QAStage{QAStageQuick, QAStageStandard}, stages)
	})

	t.Run("quick only", func(t *testing.T) {
		opts := QAOptions{Quick: true}
		stages := GetQAStages(opts)
		assert.Equal(t, []QAStage{QAStageQuick}, stages)
	})

	t.Run("full stages", func(t *testing.T) {
		opts := QAOptions{Full: true}
		stages := GetQAStages(opts)
		assert.Equal(t, []QAStage{QAStageQuick, QAStageStandard, QAStageFull}, stages)
	})
}

func TestGetQAChecks_Good(t *testing.T) {
	t.Run("quick stage checks", func(t *testing.T) {
		dir := t.TempDir()
		checks := GetQAChecks(dir, QAStageQuick)
		assert.Contains(t, checks, "audit")
		assert.Contains(t, checks, "fmt")
		assert.Contains(t, checks, "stan")
	})

	t.Run("standard stage includes test", func(t *testing.T) {
		dir := t.TempDir()
		checks := GetQAChecks(dir, QAStageStandard)
		assert.Contains(t, checks, "test")
	})

	t.Run("standard stage includes psalm if available", func(t *testing.T) {
		dir := t.TempDir()
		binDir := filepath.Join(dir, "vendor", "bin")
		err := os.MkdirAll(binDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(binDir, "psalm"), []byte(""), 0755)
		require.NoError(t, err)

		checks := GetQAChecks(dir, QAStageStandard)
		assert.Contains(t, checks, "psalm")
	})

	t.Run("full stage includes rector if available", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "rector.php"), []byte("<?php"), 0644)
		require.NoError(t, err)

		checks := GetQAChecks(dir, QAStageFull)
		assert.Contains(t, checks, "rector")
	})
}

// =============================================================================
// Security Checks Tests
// =============================================================================

func TestRunEnvSecurityChecks_Good(t *testing.T) {
	t.Run("detects debug mode enabled", func(t *testing.T) {
		dir := t.TempDir()
		envContent := "APP_DEBUG=true\nAPP_KEY=base64:abcdefghijklmnopqrstuvwxyz123456\n"
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		checks := runEnvSecurityChecks(dir)

		var debugCheck *SecurityCheck
		for i := range checks {
			if checks[i].ID == "debug_mode" {
				debugCheck = &checks[i]
				break
			}
		}

		require.NotNil(t, debugCheck)
		assert.False(t, debugCheck.Passed)
		assert.Equal(t, "critical", debugCheck.Severity)
	})

	t.Run("passes with debug disabled", func(t *testing.T) {
		dir := t.TempDir()
		envContent := "APP_DEBUG=false\nAPP_KEY=base64:abcdefghijklmnopqrstuvwxyz123456\n"
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		checks := runEnvSecurityChecks(dir)

		var debugCheck *SecurityCheck
		for i := range checks {
			if checks[i].ID == "debug_mode" {
				debugCheck = &checks[i]
				break
			}
		}

		require.NotNil(t, debugCheck)
		assert.True(t, debugCheck.Passed)
	})

	t.Run("detects weak app key", func(t *testing.T) {
		dir := t.TempDir()
		envContent := "APP_KEY=short\n"
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		checks := runEnvSecurityChecks(dir)

		var keyCheck *SecurityCheck
		for i := range checks {
			if checks[i].ID == "app_key_set" {
				keyCheck = &checks[i]
				break
			}
		}

		require.NotNil(t, keyCheck)
		assert.False(t, keyCheck.Passed)
	})

	t.Run("detects non-https app url", func(t *testing.T) {
		dir := t.TempDir()
		envContent := "APP_URL=http://example.com\n"
		err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644)
		require.NoError(t, err)

		checks := runEnvSecurityChecks(dir)

		var urlCheck *SecurityCheck
		for i := range checks {
			if checks[i].ID == "https_enforced" {
				urlCheck = &checks[i]
				break
			}
		}

		require.NotNil(t, urlCheck)
		assert.False(t, urlCheck.Passed)
	})
}

func TestRunFilesystemSecurityChecks_Good(t *testing.T) {
	t.Run("detects .env in public", func(t *testing.T) {
		dir := t.TempDir()
		publicDir := filepath.Join(dir, "public")
		err := os.MkdirAll(publicDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(publicDir, ".env"), []byte(""), 0644)
		require.NoError(t, err)

		checks := runFilesystemSecurityChecks(dir)

		found := false
		for _, check := range checks {
			if check.ID == "env_not_public" && !check.Passed {
				found = true
				break
			}
		}
		assert.True(t, found, "should detect .env in public directory")
	})

	t.Run("detects .git in public", func(t *testing.T) {
		dir := t.TempDir()
		gitDir := filepath.Join(dir, "public", ".git")
		err := os.MkdirAll(gitDir, 0755)
		require.NoError(t, err)

		checks := runFilesystemSecurityChecks(dir)

		found := false
		for _, check := range checks {
			if check.ID == "git_not_public" && !check.Passed {
				found = true
				break
			}
		}
		assert.True(t, found, "should detect .git in public directory")
	})

	t.Run("passes with clean public directory", func(t *testing.T) {
		dir := t.TempDir()
		publicDir := filepath.Join(dir, "public")
		err := os.MkdirAll(publicDir, 0755)
		require.NoError(t, err)
		// Add only safe files
		err = os.WriteFile(filepath.Join(publicDir, "index.php"), []byte("<?php"), 0644)
		require.NoError(t, err)

		checks := runFilesystemSecurityChecks(dir)
		assert.Empty(t, checks, "should not report issues for clean public directory")
	})
}
