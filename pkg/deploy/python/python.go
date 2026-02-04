package python

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/host-uk/core/pkg/framework/core"
	"github.com/kluctl/go-embed-python/python"
)

var (
	once    sync.Once
	ep      *python.EmbeddedPython
	initErr error
)

// Init initializes the embedded Python runtime.
func Init() error {
	once.Do(func() {
		ep, initErr = python.NewEmbeddedPython("core-deploy")
	})
	return initErr
}

// GetPython returns the embedded Python instance.
func GetPython() *python.EmbeddedPython {
	return ep
}

// RunScript runs a Python script with the given code and returns stdout.
func RunScript(ctx context.Context, code string, args ...string) (string, error) {
	if err := Init(); err != nil {
		return "", err
	}

	// Write code to temp file
	tmpFile, err := os.CreateTemp("", "core-*.py")
	if err != nil {
		return "", core.E("python", "create temp file", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(code); err != nil {
		_ = tmpFile.Close()
		return "", core.E("python", "write script", err)
	}
	_ = tmpFile.Close()

	// Build args: script path + any additional args
	cmdArgs := append([]string{tmpFile.Name()}, args...)

	// Get the command
	cmd, err := ep.PythonCmd(cmdArgs...)
	if err != nil {
		return "", core.E("python", "create command", err)
	}

	// Run with context
	output, err := cmd.Output()
	if err != nil {
		// Try to get stderr for better error message
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", core.E("python", "run script", fmt.Errorf("%w: %s", err, string(exitErr.Stderr)))
		}
		return "", core.E("python", "run script", err)
	}

	return string(output), nil
}

// RunModule runs a Python module (python -m module_name).
func RunModule(ctx context.Context, module string, args ...string) (string, error) {
	if err := Init(); err != nil {
		return "", err
	}

	cmdArgs := append([]string{"-m", module}, args...)
	cmd, err := ep.PythonCmd(cmdArgs...)
	if err != nil {
		return "", core.E("python", "create command", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", core.E("python", fmt.Sprintf("run module %s", module), err)
	}

	return string(output), nil
}

// DevOpsPath returns the path to the DevOps repo.
func DevOpsPath() (string, error) {
	if path := os.Getenv("DEVOPS_PATH"); path != "" {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", core.E("python", "get user home", err)
	}
	return filepath.Join(home, "Code", "DevOps"), nil
}

// CoolifyModulePath returns the path to the Coolify module_utils.
func CoolifyModulePath() (string, error) {
	path, err := DevOpsPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(path, "playbooks", "roles", "coolify", "module_utils"), nil
}

// CoolifyScript generates Python code to call the Coolify API.
func CoolifyScript(baseURL, apiToken, operation string, params map[string]any) (string, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", core.E("python", "marshal params", err)
	}

	modulePath, err := CoolifyModulePath()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`
import sys
import json
sys.path.insert(0, %q)

from swagger.coolify_api import CoolifyClient

client = CoolifyClient(
    base_url=%q,
    api_token=%q,
    timeout=30,
    verify_ssl=True,
)

params = json.loads(%q)
result = client._call(%q, params, check_response=False)
print(json.dumps(result))
`, modulePath, baseURL, apiToken, string(paramsJSON), operation), nil
}
