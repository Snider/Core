package daemon

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAddDaemonCommand(t *testing.T) {
	root := &cobra.Command{Use: "core"}
	AddDaemonCommand(root)

	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "daemon" {
			found = true
			assert.NotEmpty(t, cmd.Short)
			assert.NotNil(t, cmd.RunE)

			// Check flags
			assert.NotNil(t, cmd.Flags().Lookup("mcp-transport"))
			assert.NotNil(t, cmd.Flags().Lookup("mcp-addr"))
		}
	}
	assert.True(t, found, "daemon command should be registered")
}
