package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasePlugin_Good(t *testing.T) {
	p := &BasePlugin{
		PluginName:    "test-plugin",
		PluginVersion: "1.0.0",
	}

	assert.Equal(t, "test-plugin", p.Name())
	assert.Equal(t, "1.0.0", p.Version())

	ctx := context.Background()
	assert.NoError(t, p.Init(ctx))
	assert.NoError(t, p.Start(ctx))
	assert.NoError(t, p.Stop(ctx))
}

func TestBasePlugin_Good_EmptyFields(t *testing.T) {
	p := &BasePlugin{}

	assert.Equal(t, "", p.Name())
	assert.Equal(t, "", p.Version())

	ctx := context.Background()
	assert.NoError(t, p.Init(ctx))
	assert.NoError(t, p.Start(ctx))
	assert.NoError(t, p.Stop(ctx))
}

func TestBasePlugin_Good_ImplementsPlugin(t *testing.T) {
	var _ Plugin = &BasePlugin{}
}
