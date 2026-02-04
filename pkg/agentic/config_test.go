package agentic

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Good_FromEnvFile(t *testing.T) {
	m := io.NewMockMedium()
	tmpDir := "/tmp/agentic-test"

	envContent := `
AGENTIC_BASE_URL=https://test.api.com
AGENTIC_TOKEN=test-token-123
AGENTIC_PROJECT=my-project
AGENTIC_AGENT_ID=agent-001
`
	err := m.Write(filepath.Join(tmpDir, ".env"), envContent)
	require.NoError(t, err)

	cfg, err := LoadConfig(m, tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "https://test.api.com", cfg.BaseURL)
	assert.Equal(t, "test-token-123", cfg.Token)
	assert.Equal(t, "my-project", cfg.DefaultProject)
	assert.Equal(t, "agent-001", cfg.AgentID)
}

func TestLoadConfig_Good_FromEnvVars(t *testing.T) {
	m := io.NewMockMedium()
	tmpDir := "/tmp/agentic-test"

	envContent := `
AGENTIC_TOKEN=env-file-token
`
	err := m.Write(filepath.Join(tmpDir, ".env"), envContent)
	require.NoError(t, err)

	// Set environment variables that should override
	_ = os.Setenv("AGENTIC_BASE_URL", "https://env-override.com")
	_ = os.Setenv("AGENTIC_TOKEN", "env-override-token")
	defer func() {
		_ = os.Unsetenv("AGENTIC_BASE_URL")
		_ = os.Unsetenv("AGENTIC_TOKEN")
	}()

	cfg, err := LoadConfig(m, tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "https://env-override.com", cfg.BaseURL)
	assert.Equal(t, "env-override-token", cfg.Token)
}

func TestLoadConfig_Bad_NoToken(t *testing.T) {
	m := io.NewMockMedium()
	tmpDir := "/tmp/agentic-test"

	// Create empty .env
	err := m.Write(filepath.Join(tmpDir, ".env"), "")
	require.NoError(t, err)

	// Ensure no env vars are set
	_ = os.Unsetenv("AGENTIC_TOKEN")
	_ = os.Unsetenv("AGENTIC_BASE_URL")

	_, err = LoadConfig(m, tmpDir)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no authentication token")
}

func TestLoadConfig_Good_EnvFileWithQuotes(t *testing.T) {
	m := io.NewMockMedium()
	tmpDir := "/tmp/agentic-test"

	// Test with quoted values
	envContent := `
AGENTIC_TOKEN="quoted-token"
AGENTIC_BASE_URL='single-quoted-url'
`
	err := m.Write(filepath.Join(tmpDir, ".env"), envContent)
	require.NoError(t, err)

	cfg, err := LoadConfig(m, tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "quoted-token", cfg.Token)
	assert.Equal(t, "single-quoted-url", cfg.BaseURL)
}

func TestLoadConfig_Good_EnvFileWithComments(t *testing.T) {
	m := io.NewMockMedium()
	tmpDir := "/tmp/agentic-test"

	envContent := `
# This is a comment
AGENTIC_TOKEN=token-with-comments

# Another comment
AGENTIC_PROJECT=commented-project
`
	err := m.Write(filepath.Join(tmpDir, ".env"), envContent)
	require.NoError(t, err)

	cfg, err := LoadConfig(m, tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "token-with-comments", cfg.Token)
	assert.Equal(t, "commented-project", cfg.DefaultProject)
}

func TestSaveConfig_Good(t *testing.T) {
	m := io.NewMockMedium()

	// Create temp home directory
	tmpHome, err := os.MkdirTemp("", "agentic-home")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpHome) }()

	// Override HOME for the test
	originalHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpHome)
	defer func() { _ = os.Setenv("HOME", originalHome) }()

	cfg := &Config{
		BaseURL:        "https://saved.api.com",
		Token:          "saved-token",
		DefaultProject: "saved-project",
		AgentID:        "saved-agent",
	}

	err = SaveConfig(m, cfg)
	require.NoError(t, err)

	// Verify file was created
	configPath := filepath.Join(tmpHome, ".core", "agentic.yaml")
	assert.True(t, m.Exists(configPath))

	// Read back the config
	data, err := m.Read(configPath)
	require.NoError(t, err)
	assert.Contains(t, data, "saved.api.com")
	assert.Contains(t, data, "saved-token")
}

func TestConfigPath_Good(t *testing.T) {
	m := io.NewMockMedium()
	path, err := ConfigPath(m)

	require.NoError(t, err)
	assert.Contains(t, path, ".core")
	assert.Contains(t, path, "agentic.yaml")
}

func TestLoadConfig_Good_DefaultBaseURL(t *testing.T) {
	m := io.NewMockMedium()
	tmpDir := "/tmp/agentic-test"

	// Only provide token, should use default base URL
	envContent := `
AGENTIC_TOKEN=test-token
`
	err := m.Write(filepath.Join(tmpDir, ".env"), envContent)
	require.NoError(t, err)

	// Clear any env overrides
	_ = os.Unsetenv("AGENTIC_BASE_URL")

	cfg, err := LoadConfig(m, tmpDir)

	require.NoError(t, err)
	assert.Equal(t, DefaultBaseURL, cfg.BaseURL)
}
