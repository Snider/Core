package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAgentID_Good(t *testing.T) {
	provider, name, err := parseAgentID("claude-opus/qa")
	require.NoError(t, err)
	assert.Equal(t, "claude-opus", provider)
	assert.Equal(t, "qa", name)
}

func TestParseAgentID_Bad(t *testing.T) {
	tests := []string{
		"noslash",
		"/missing-provider",
		"missing-name/",
		"",
	}
	for _, id := range tests {
		_, _, err := parseAgentID(id)
		assert.Error(t, err, "expected error for: %q", id)
	}
}

func TestAgentContextPath(t *testing.T) {
	path := agentContextPath("/ws/p101/i343", "claude-opus", "qa")
	assert.Equal(t, "/ws/p101/i343/agents/claude-opus/qa", path)
}

func TestUpdateAgentManifest_Good(t *testing.T) {
	tmp := t.TempDir()
	agentDir := filepath.Join(tmp, "agents", "test-provider", "test-agent")
	require.NoError(t, os.MkdirAll(agentDir, 0755))

	updateAgentManifest(agentDir, "test-provider", "test-agent")

	data, err := os.ReadFile(filepath.Join(agentDir, "manifest.json"))
	require.NoError(t, err)

	var m AgentManifest
	require.NoError(t, json.Unmarshal(data, &m))
	assert.Equal(t, "test-provider", m.Provider)
	assert.Equal(t, "test-agent", m.Name)
	assert.False(t, m.CreatedAt.IsZero())
	assert.False(t, m.LastSeen.IsZero())
}

func TestUpdateAgentManifest_PreservesCreatedAt(t *testing.T) {
	tmp := t.TempDir()
	agentDir := filepath.Join(tmp, "agents", "p", "a")
	require.NoError(t, os.MkdirAll(agentDir, 0755))

	// First call sets created_at
	updateAgentManifest(agentDir, "p", "a")

	data, err := os.ReadFile(filepath.Join(agentDir, "manifest.json"))
	require.NoError(t, err)
	var first AgentManifest
	require.NoError(t, json.Unmarshal(data, &first))

	// Second call should preserve created_at
	updateAgentManifest(agentDir, "p", "a")

	data, err = os.ReadFile(filepath.Join(agentDir, "manifest.json"))
	require.NoError(t, err)
	var second AgentManifest
	require.NoError(t, json.Unmarshal(data, &second))

	assert.Equal(t, first.CreatedAt, second.CreatedAt)
	assert.True(t, second.LastSeen.After(first.CreatedAt) || second.LastSeen.Equal(first.CreatedAt))
}
