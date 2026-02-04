package devops

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/io"
	"gopkg.in/yaml.v3"
)

// TestConfig holds test configuration from .core/test.yaml.
type TestConfig struct {
	Version  int               `yaml:"version"`
	Command  string            `yaml:"command,omitempty"`
	Commands []TestCommand     `yaml:"commands,omitempty"`
	Env      map[string]string `yaml:"env,omitempty"`
}

// TestCommand is a named test command.
type TestCommand struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

// TestOptions configures test execution.
type TestOptions struct {
	Name    string   // Run specific named command from .core/test.yaml
	Command []string // Override command (from -- args)
}

// Test runs tests in the dev environment.
func (d *DevOps) Test(ctx context.Context, projectDir string, opts TestOptions) error {
	running, err := d.IsRunning(ctx)
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("dev environment not running (run 'core dev boot' first)")
	}

	var cmd string

	// Priority: explicit command > named command > auto-detect
	if len(opts.Command) > 0 {
		cmd = strings.Join(opts.Command, " ")
	} else if opts.Name != "" {
		cfg, err := LoadTestConfig(d.medium, projectDir)
		if err != nil {
			return err
		}
		for _, c := range cfg.Commands {
			if c.Name == opts.Name {
				cmd = c.Run
				break
			}
		}
		if cmd == "" {
			return fmt.Errorf("test command %q not found in .core/test.yaml", opts.Name)
		}
	} else {
		cmd = DetectTestCommand(d.medium, projectDir)
		if cmd == "" {
			return fmt.Errorf("could not detect test command (create .core/test.yaml)")
		}
	}

	// Run via SSH - construct command as single string for shell execution
	return d.sshShell(ctx, []string{"cd", "/app", "&&", cmd})
}

// DetectTestCommand auto-detects the test command for a project.
func DetectTestCommand(m io.Medium, projectDir string) string {
	// 1. Check .core/test.yaml
	cfg, err := LoadTestConfig(m, projectDir)
	if err == nil && cfg.Command != "" {
		return cfg.Command
	}

	// 2. Check composer.json for test script
	if hasFile(m, projectDir, "composer.json") {
		if hasComposerScript(m, projectDir, "test") {
			return "composer test"
		}
	}

	// 3. Check package.json for test script
	if hasFile(m, projectDir, "package.json") {
		if hasPackageScript(m, projectDir, "test") {
			return "npm test"
		}
	}

	// 4. Check go.mod
	if hasFile(m, projectDir, "go.mod") {
		return "go test ./..."
	}

	// 5. Check pytest
	if hasFile(m, projectDir, "pytest.ini") || hasFile(m, projectDir, "pyproject.toml") {
		return "pytest"
	}

	// 6. Check Taskfile
	if hasFile(m, projectDir, "Taskfile.yaml") || hasFile(m, projectDir, "Taskfile.yml") {
		return "task test"
	}

	return ""
}

// LoadTestConfig loads .core/test.yaml.
func LoadTestConfig(m io.Medium, projectDir string) (*TestConfig, error) {
	path := filepath.Join(projectDir, ".core", "test.yaml")
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	content, err := m.Read(absPath)
	if err != nil {
		return nil, err
	}

	var cfg TestConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func hasFile(m io.Medium, dir, name string) bool {
	path := filepath.Join(dir, name)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return m.IsFile(absPath)
}

func hasPackageScript(m io.Medium, projectDir, script string) bool {
	path := filepath.Join(projectDir, "package.json")
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	content, err := m.Read(absPath)
	if err != nil {
		return false
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal([]byte(content), &pkg); err != nil {
		return false
	}

	_, ok := pkg.Scripts[script]
	return ok
}

func hasComposerScript(m io.Medium, projectDir, script string) bool {
	path := filepath.Join(projectDir, "composer.json")
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	content, err := m.Read(absPath)
	if err != nil {
		return false
	}

	var pkg struct {
		Scripts map[string]interface{} `json:"scripts"`
	}
	if err := json.Unmarshal([]byte(content), &pkg); err != nil {
		return false
	}

	_, ok := pkg.Scripts[script]
	return ok
}
