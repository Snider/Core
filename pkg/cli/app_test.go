package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDaemonFlags(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectedT string
		expectedA string
	}{
		{
			name:      "no daemon command",
			args:      []string{"core", "help"},
			expectedT: "",
			expectedA: "",
		},
		{
			name:      "daemon with equals",
			args:      []string{"core", "daemon", "--mcp-transport=tcp", "--mcp-addr=:9100"},
			expectedT: "tcp",
			expectedA: ":9100",
		},
		{
			name:      "daemon with spaces",
			args:      []string{"core", "daemon", "--mcp-transport", "unix", "--mcp-addr", "/tmp/mcp.sock"},
			expectedT: "unix",
			expectedA: "/tmp/mcp.sock",
		},
		{
			name:      "daemon with global flags",
			args:      []string{"core", "--debug", "daemon", "--mcp-transport=socket"},
			expectedT: "socket",
			expectedA: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("CORE_MCP_TRANSPORT")
			os.Unsetenv("CORE_MCP_ADDR")

			parseDaemonFlags(tt.args)

			assert.Equal(t, tt.expectedT, os.Getenv("CORE_MCP_TRANSPORT"))
			assert.Equal(t, tt.expectedA, os.Getenv("CORE_MCP_ADDR"))
		})
	}
}
