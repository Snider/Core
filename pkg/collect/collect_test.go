package collect

import (
	"testing"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig_Good(t *testing.T) {
	cfg := NewConfig("/tmp/output")

	assert.NotNil(t, cfg)
	assert.Equal(t, "/tmp/output", cfg.OutputDir)
	assert.NotNil(t, cfg.Output)
	assert.NotNil(t, cfg.Limiter)
	assert.NotNil(t, cfg.State)
	assert.NotNil(t, cfg.Dispatcher)
	assert.False(t, cfg.Verbose)
	assert.False(t, cfg.DryRun)
}

func TestNewConfigWithMedium_Good(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/data")

	assert.NotNil(t, cfg)
	assert.Equal(t, m, cfg.Output)
	assert.Equal(t, "/data", cfg.OutputDir)
	assert.NotNil(t, cfg.Limiter)
	assert.NotNil(t, cfg.State)
	assert.NotNil(t, cfg.Dispatcher)
}

func TestMergeResults_Good(t *testing.T) {
	r1 := &Result{
		Source: "a",
		Items:  5,
		Errors: 1,
		Files:  []string{"a.md", "b.md"},
	}
	r2 := &Result{
		Source:  "b",
		Items:   3,
		Skipped: 2,
		Files:   []string{"c.md"},
	}

	merged := MergeResults("combined", r1, r2)
	assert.Equal(t, "combined", merged.Source)
	assert.Equal(t, 8, merged.Items)
	assert.Equal(t, 1, merged.Errors)
	assert.Equal(t, 2, merged.Skipped)
	assert.Len(t, merged.Files, 3)
}

func TestMergeResults_Good_NilResults(t *testing.T) {
	r1 := &Result{Items: 3}
	merged := MergeResults("test", r1, nil, nil)
	assert.Equal(t, 3, merged.Items)
}

func TestMergeResults_Good_Empty(t *testing.T) {
	merged := MergeResults("empty")
	assert.Equal(t, 0, merged.Items)
	assert.Equal(t, 0, merged.Errors)
	assert.Nil(t, merged.Files)
}
