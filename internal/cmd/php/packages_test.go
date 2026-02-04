package php

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadComposerJSON_Good(t *testing.T) {
	t.Run("reads valid composer.json", func(t *testing.T) {
		dir := t.TempDir()
		composerJSON := `{
			"name": "test/project",
			"require": {
				"php": "^8.2"
			}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		raw, err := readComposerJSON(dir)
		assert.NoError(t, err)
		assert.NotNil(t, raw)
		assert.Contains(t, string(raw["name"]), "test/project")
	})

	t.Run("preserves all fields", func(t *testing.T) {
		dir := t.TempDir()
		composerJSON := `{
			"name": "test/project",
			"description": "Test project",
			"require": {"php": "^8.2"},
			"autoload": {"psr-4": {"App\\": "src/"}}
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		raw, err := readComposerJSON(dir)
		assert.NoError(t, err)
		assert.Contains(t, string(raw["autoload"]), "psr-4")
	})
}

func TestReadComposerJSON_Bad(t *testing.T) {
	t.Run("missing composer.json", func(t *testing.T) {
		dir := t.TempDir()
		_, err := readComposerJSON(dir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to read composer.json")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte("not json{"), 0644)
		require.NoError(t, err)

		_, err = readComposerJSON(dir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to parse composer.json")
	})
}

func TestWriteComposerJSON_Good(t *testing.T) {
	t.Run("writes valid composer.json", func(t *testing.T) {
		dir := t.TempDir()
		raw := make(map[string]json.RawMessage)
		raw["name"] = json.RawMessage(`"test/project"`)

		err := writeComposerJSON(dir, raw)
		assert.NoError(t, err)

		// Verify file was written
		content, err := os.ReadFile(filepath.Join(dir, "composer.json"))
		assert.NoError(t, err)
		assert.Contains(t, string(content), "test/project")
		// Verify trailing newline
		assert.True(t, content[len(content)-1] == '\n')
	})

	t.Run("pretty prints with indentation", func(t *testing.T) {
		dir := t.TempDir()
		raw := make(map[string]json.RawMessage)
		raw["name"] = json.RawMessage(`"test/project"`)
		raw["require"] = json.RawMessage(`{"php":"^8.2"}`)

		err := writeComposerJSON(dir, raw)
		assert.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(dir, "composer.json"))
		assert.NoError(t, err)
		// Should be indented
		assert.Contains(t, string(content), "    ")
	})
}

func TestWriteComposerJSON_Bad(t *testing.T) {
	t.Run("fails for non-existent directory", func(t *testing.T) {
		raw := make(map[string]json.RawMessage)
		raw["name"] = json.RawMessage(`"test/project"`)

		err := writeComposerJSON("/non/existent/path", raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to write composer.json")
	})
}
func TestGetRepositories_Good(t *testing.T) {
	t.Run("returns empty slice when no repositories", func(t *testing.T) {
		raw := make(map[string]json.RawMessage)
		raw["name"] = json.RawMessage(`"test/project"`)

		repos, err := getRepositories(raw)
		assert.NoError(t, err)
		assert.Empty(t, repos)
	})

	t.Run("parses existing repositories", func(t *testing.T) {
		raw := make(map[string]json.RawMessage)
		raw["name"] = json.RawMessage(`"test/project"`)
		raw["repositories"] = json.RawMessage(`[{"type":"path","url":"/path/to/package"}]`)

		repos, err := getRepositories(raw)
		assert.NoError(t, err)
		assert.Len(t, repos, 1)
		assert.Equal(t, "path", repos[0].Type)
		assert.Equal(t, "/path/to/package", repos[0].URL)
	})

	t.Run("parses repositories with options", func(t *testing.T) {
		raw := make(map[string]json.RawMessage)
		raw["repositories"] = json.RawMessage(`[{"type":"path","url":"/path","options":{"symlink":true}}]`)

		repos, err := getRepositories(raw)
		assert.NoError(t, err)
		assert.Len(t, repos, 1)
		assert.NotNil(t, repos[0].Options)
		assert.Equal(t, true, repos[0].Options["symlink"])
	})
}

func TestGetRepositories_Bad(t *testing.T) {
	t.Run("fails for invalid repositories JSON", func(t *testing.T) {
		raw := make(map[string]json.RawMessage)
		raw["repositories"] = json.RawMessage(`not valid json`)

		_, err := getRepositories(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to parse repositories")
	})
}

func TestSetRepositories_Good(t *testing.T) {
	t.Run("sets repositories", func(t *testing.T) {
		raw := make(map[string]json.RawMessage)
		repos := []composerRepository{
			{Type: "path", URL: "/path/to/package"},
		}

		err := setRepositories(raw, repos)
		assert.NoError(t, err)
		assert.Contains(t, string(raw["repositories"]), "/path/to/package")
	})

	t.Run("removes repositories key when empty", func(t *testing.T) {
		raw := make(map[string]json.RawMessage)
		raw["repositories"] = json.RawMessage(`[{"type":"path"}]`)

		err := setRepositories(raw, []composerRepository{})
		assert.NoError(t, err)
		_, exists := raw["repositories"]
		assert.False(t, exists)
	})
}

func TestGetPackageInfo_Good(t *testing.T) {
	t.Run("extracts package name and version", func(t *testing.T) {
		dir := t.TempDir()
		composerJSON := `{
			"name": "vendor/package",
			"version": "1.0.0"
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		name, version, err := getPackageInfo(dir)
		assert.NoError(t, err)
		assert.Equal(t, "vendor/package", name)
		assert.Equal(t, "1.0.0", version)
	})

	t.Run("works without version", func(t *testing.T) {
		dir := t.TempDir()
		composerJSON := `{
			"name": "vendor/package"
		}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		name, version, err := getPackageInfo(dir)
		assert.NoError(t, err)
		assert.Equal(t, "vendor/package", name)
		assert.Equal(t, "", version)
	})
}

func TestGetPackageInfo_Bad(t *testing.T) {
	t.Run("missing composer.json", func(t *testing.T) {
		dir := t.TempDir()
		_, _, err := getPackageInfo(dir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to read package composer.json")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		dir := t.TempDir()
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte("not json{"), 0644)
		require.NoError(t, err)

		_, _, err = getPackageInfo(dir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to parse package composer.json")
	})

	t.Run("missing name", func(t *testing.T) {
		dir := t.TempDir()
		composerJSON := `{"version": "1.0.0"}`
		err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		_, _, err = getPackageInfo(dir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "package name not found")
	})
}

func TestLinkPackages_Good(t *testing.T) {
	t.Run("links a package", func(t *testing.T) {
		// Create project directory
		projectDir := t.TempDir()
		err := os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(`{"name":"test/project"}`), 0644)
		require.NoError(t, err)

		// Create package directory
		packageDir := t.TempDir()
		err = os.WriteFile(filepath.Join(packageDir, "composer.json"), []byte(`{"name":"vendor/package"}`), 0644)
		require.NoError(t, err)

		err = LinkPackages(projectDir, []string{packageDir})
		assert.NoError(t, err)

		// Verify repository was added
		raw, err := readComposerJSON(projectDir)
		assert.NoError(t, err)
		repos, err := getRepositories(raw)
		assert.NoError(t, err)
		assert.Len(t, repos, 1)
		assert.Equal(t, "path", repos[0].Type)
	})

	t.Run("skips already linked package", func(t *testing.T) {
		// Create project with existing repository
		projectDir := t.TempDir()
		packageDir := t.TempDir()

		err := os.WriteFile(filepath.Join(packageDir, "composer.json"), []byte(`{"name":"vendor/package"}`), 0644)
		require.NoError(t, err)

		absPackagePath, _ := filepath.Abs(packageDir)
		composerJSON := `{
			"name": "test/project",
			"repositories": [{"type":"path","url":"` + absPackagePath + `"}]
		}`
		err = os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		// Link again - should not add duplicate
		err = LinkPackages(projectDir, []string{packageDir})
		assert.NoError(t, err)

		raw, err := readComposerJSON(projectDir)
		assert.NoError(t, err)
		repos, err := getRepositories(raw)
		assert.NoError(t, err)
		assert.Len(t, repos, 1) // Still only one
	})

	t.Run("links multiple packages", func(t *testing.T) {
		projectDir := t.TempDir()
		err := os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(`{"name":"test/project"}`), 0644)
		require.NoError(t, err)

		pkg1Dir := t.TempDir()
		err = os.WriteFile(filepath.Join(pkg1Dir, "composer.json"), []byte(`{"name":"vendor/pkg1"}`), 0644)
		require.NoError(t, err)

		pkg2Dir := t.TempDir()
		err = os.WriteFile(filepath.Join(pkg2Dir, "composer.json"), []byte(`{"name":"vendor/pkg2"}`), 0644)
		require.NoError(t, err)

		err = LinkPackages(projectDir, []string{pkg1Dir, pkg2Dir})
		assert.NoError(t, err)

		raw, err := readComposerJSON(projectDir)
		assert.NoError(t, err)
		repos, err := getRepositories(raw)
		assert.NoError(t, err)
		assert.Len(t, repos, 2)
	})
}

func TestLinkPackages_Bad(t *testing.T) {
	t.Run("fails for non-PHP project", func(t *testing.T) {
		dir := t.TempDir()
		err := LinkPackages(dir, []string{"/path/to/package"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a PHP project")
	})

	t.Run("fails for non-PHP package", func(t *testing.T) {
		projectDir := t.TempDir()
		err := os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(`{"name":"test/project"}`), 0644)
		require.NoError(t, err)

		packageDir := t.TempDir()
		// No composer.json in package

		err = LinkPackages(projectDir, []string{packageDir})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a PHP package")
	})
}

func TestUnlinkPackages_Good(t *testing.T) {
	t.Run("unlinks package by name", func(t *testing.T) {
		projectDir := t.TempDir()
		packageDir := t.TempDir()

		err := os.WriteFile(filepath.Join(packageDir, "composer.json"), []byte(`{"name":"vendor/package"}`), 0644)
		require.NoError(t, err)

		absPackagePath, _ := filepath.Abs(packageDir)
		composerJSON := `{
			"name": "test/project",
			"repositories": [{"type":"path","url":"` + absPackagePath + `"}]
		}`
		err = os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		err = UnlinkPackages(projectDir, []string{"vendor/package"})
		assert.NoError(t, err)

		raw, err := readComposerJSON(projectDir)
		assert.NoError(t, err)
		repos, err := getRepositories(raw)
		assert.NoError(t, err)
		assert.Len(t, repos, 0)
	})

	t.Run("unlinks package by path", func(t *testing.T) {
		projectDir := t.TempDir()
		packageDir := t.TempDir()

		absPackagePath, _ := filepath.Abs(packageDir)
		composerJSON := `{
			"name": "test/project",
			"repositories": [{"type":"path","url":"` + absPackagePath + `"}]
		}`
		err := os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		err = UnlinkPackages(projectDir, []string{absPackagePath})
		assert.NoError(t, err)

		raw, err := readComposerJSON(projectDir)
		assert.NoError(t, err)
		repos, err := getRepositories(raw)
		assert.NoError(t, err)
		assert.Len(t, repos, 0)
	})

	t.Run("keeps non-path repositories", func(t *testing.T) {
		projectDir := t.TempDir()
		composerJSON := `{
			"name": "test/project",
			"repositories": [
				{"type":"vcs","url":"https://github.com/vendor/package"},
				{"type":"path","url":"/local/path"}
			]
		}`
		err := os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		err = UnlinkPackages(projectDir, []string{"/local/path"})
		assert.NoError(t, err)

		raw, err := readComposerJSON(projectDir)
		assert.NoError(t, err)
		repos, err := getRepositories(raw)
		assert.NoError(t, err)
		assert.Len(t, repos, 1)
		assert.Equal(t, "vcs", repos[0].Type)
	})
}

func TestUnlinkPackages_Bad(t *testing.T) {
	t.Run("fails for non-PHP project", func(t *testing.T) {
		dir := t.TempDir()
		err := UnlinkPackages(dir, []string{"vendor/package"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a PHP project")
	})
}

func TestListLinkedPackages_Good(t *testing.T) {
	t.Run("lists linked packages", func(t *testing.T) {
		projectDir := t.TempDir()
		packageDir := t.TempDir()

		err := os.WriteFile(filepath.Join(packageDir, "composer.json"), []byte(`{"name":"vendor/package","version":"1.0.0"}`), 0644)
		require.NoError(t, err)

		absPackagePath, _ := filepath.Abs(packageDir)
		composerJSON := `{
			"name": "test/project",
			"repositories": [{"type":"path","url":"` + absPackagePath + `"}]
		}`
		err = os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		linked, err := ListLinkedPackages(projectDir)
		assert.NoError(t, err)
		assert.Len(t, linked, 1)
		assert.Equal(t, "vendor/package", linked[0].Name)
		assert.Equal(t, "1.0.0", linked[0].Version)
		assert.Equal(t, absPackagePath, linked[0].Path)
	})

	t.Run("returns empty list when no linked packages", func(t *testing.T) {
		projectDir := t.TempDir()
		err := os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(`{"name":"test/project"}`), 0644)
		require.NoError(t, err)

		linked, err := ListLinkedPackages(projectDir)
		assert.NoError(t, err)
		assert.Empty(t, linked)
	})

	t.Run("uses basename when package info unavailable", func(t *testing.T) {
		projectDir := t.TempDir()
		composerJSON := `{
			"name": "test/project",
			"repositories": [{"type":"path","url":"/nonexistent/package-name"}]
		}`
		err := os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		linked, err := ListLinkedPackages(projectDir)
		assert.NoError(t, err)
		assert.Len(t, linked, 1)
		assert.Equal(t, "package-name", linked[0].Name)
	})

	t.Run("ignores non-path repositories", func(t *testing.T) {
		projectDir := t.TempDir()
		composerJSON := `{
			"name": "test/project",
			"repositories": [
				{"type":"vcs","url":"https://github.com/vendor/package"}
			]
		}`
		err := os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(composerJSON), 0644)
		require.NoError(t, err)

		linked, err := ListLinkedPackages(projectDir)
		assert.NoError(t, err)
		assert.Empty(t, linked)
	})
}

func TestListLinkedPackages_Bad(t *testing.T) {
	t.Run("fails for non-PHP project", func(t *testing.T) {
		dir := t.TempDir()
		_, err := ListLinkedPackages(dir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a PHP project")
	})
}

func TestUpdatePackages_Bad(t *testing.T) {
	t.Run("fails for non-PHP project", func(t *testing.T) {
		dir := t.TempDir()
		err := UpdatePackages(dir, []string{"vendor/package"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a PHP project")
	})
}

func TestUpdatePackages_Good(t *testing.T) {
	t.Skip("requires Composer installed")

	t.Run("runs composer update", func(t *testing.T) {
		projectDir := t.TempDir()
		err := os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte(`{"name":"test/project"}`), 0644)
		require.NoError(t, err)

		_ = UpdatePackages(projectDir, []string{"vendor/package"})
		// This will fail because composer update needs real dependencies
		// but it validates the command runs
	})
}

func TestLinkedPackage_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		pkg := LinkedPackage{
			Name:    "vendor/package",
			Path:    "/path/to/package",
			Version: "1.0.0",
		}

		assert.Equal(t, "vendor/package", pkg.Name)
		assert.Equal(t, "/path/to/package", pkg.Path)
		assert.Equal(t, "1.0.0", pkg.Version)
	})
}

func TestComposerRepository_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		repo := composerRepository{
			Type: "path",
			URL:  "/path/to/package",
			Options: map[string]any{
				"symlink": true,
			},
		}

		assert.Equal(t, "path", repo.Type)
		assert.Equal(t, "/path/to/package", repo.URL)
		assert.Equal(t, true, repo.Options["symlink"])
	})
}
