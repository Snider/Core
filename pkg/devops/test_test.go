package devops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectTestCommand_Good_ComposerJSON(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(`{"scripts":{"test":"pest"}}`), 0644)

	cmd := DetectTestCommand(tmpDir)
	if cmd != "composer test" {
		t.Errorf("expected 'composer test', got %q", cmd)
	}
}

func TestDetectTestCommand_Good_PackageJSON(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"scripts":{"test":"vitest"}}`), 0644)

	cmd := DetectTestCommand(tmpDir)
	if cmd != "npm test" {
		t.Errorf("expected 'npm test', got %q", cmd)
	}
}

func TestDetectTestCommand_Good_GoMod(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example"), 0644)

	cmd := DetectTestCommand(tmpDir)
	if cmd != "go test ./..." {
		t.Errorf("expected 'go test ./...', got %q", cmd)
	}
}

func TestDetectTestCommand_Good_CoreTestYaml(t *testing.T) {
	tmpDir := t.TempDir()
	coreDir := filepath.Join(tmpDir, ".core")
	os.MkdirAll(coreDir, 0755)
	os.WriteFile(filepath.Join(coreDir, "test.yaml"), []byte("command: custom-test"), 0644)

	cmd := DetectTestCommand(tmpDir)
	if cmd != "custom-test" {
		t.Errorf("expected 'custom-test', got %q", cmd)
	}
}

func TestDetectTestCommand_Good_Pytest(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "pytest.ini"), []byte("[pytest]"), 0644)

	cmd := DetectTestCommand(tmpDir)
	if cmd != "pytest" {
		t.Errorf("expected 'pytest', got %q", cmd)
	}
}

func TestDetectTestCommand_Good_Taskfile(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "Taskfile.yaml"), []byte("version: '3'"), 0644)

	cmd := DetectTestCommand(tmpDir)
	if cmd != "task test" {
		t.Errorf("expected 'task test', got %q", cmd)
	}
}

func TestDetectTestCommand_Bad_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := DetectTestCommand(tmpDir)
	if cmd != "" {
		t.Errorf("expected empty string, got %q", cmd)
	}
}

func TestDetectTestCommand_Good_Priority(t *testing.T) {
	// .core/test.yaml should take priority over other detection methods
	tmpDir := t.TempDir()
	coreDir := filepath.Join(tmpDir, ".core")
	os.MkdirAll(coreDir, 0755)
	os.WriteFile(filepath.Join(coreDir, "test.yaml"), []byte("command: my-custom-test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example"), 0644)

	cmd := DetectTestCommand(tmpDir)
	if cmd != "my-custom-test" {
		t.Errorf("expected 'my-custom-test' (from .core/test.yaml), got %q", cmd)
	}
}

func TestLoadTestConfig_Good(t *testing.T) {
	tmpDir := t.TempDir()
	coreDir := filepath.Join(tmpDir, ".core")
	os.MkdirAll(coreDir, 0755)

	configYAML := `version: 1
command: default-test
commands:
  - name: unit
    run: go test ./...
  - name: integration
    run: go test -tags=integration ./...
env:
  CI: "true"
`
	os.WriteFile(filepath.Join(coreDir, "test.yaml"), []byte(configYAML), 0644)

	cfg, err := LoadTestConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
	if cfg.Command != "default-test" {
		t.Errorf("expected command 'default-test', got %q", cfg.Command)
	}
	if len(cfg.Commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(cfg.Commands))
	}
	if cfg.Commands[0].Name != "unit" {
		t.Errorf("expected first command name 'unit', got %q", cfg.Commands[0].Name)
	}
	if cfg.Env["CI"] != "true" {
		t.Errorf("expected env CI='true', got %q", cfg.Env["CI"])
	}
}

func TestLoadTestConfig_Bad_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := LoadTestConfig(tmpDir)
	if err == nil {
		t.Error("expected error for missing config, got nil")
	}
}

func TestHasPackageScript_Good(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"scripts":{"test":"jest","build":"webpack"}}`), 0644)

	if !hasPackageScript(tmpDir, "test") {
		t.Error("expected to find 'test' script")
	}
	if !hasPackageScript(tmpDir, "build") {
		t.Error("expected to find 'build' script")
	}
}

func TestHasPackageScript_Bad_MissingScript(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"scripts":{"build":"webpack"}}`), 0644)

	if hasPackageScript(tmpDir, "test") {
		t.Error("expected not to find 'test' script")
	}
}

func TestHasComposerScript_Good(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(`{"scripts":{"test":"pest","post-install-cmd":"@php artisan migrate"}}`), 0644)

	if !hasComposerScript(tmpDir, "test") {
		t.Error("expected to find 'test' script")
	}
}

func TestHasComposerScript_Bad_MissingScript(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "composer.json"), []byte(`{"scripts":{"build":"@php build.php"}}`), 0644)

	if hasComposerScript(tmpDir, "test") {
		t.Error("expected not to find 'test' script")
	}
}
