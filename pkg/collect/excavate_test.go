package collect

import (
	"context"
	"fmt"
	"testing"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

// mockCollector is a simple collector for testing the Excavator.
type mockCollector struct {
	name   string
	items  int
	err    error
	called bool
}

func (m *mockCollector) Name() string { return m.name }

func (m *mockCollector) Collect(ctx context.Context, cfg *Config) (*Result, error) {
	m.called = true
	if m.err != nil {
		return &Result{Source: m.name, Errors: 1}, m.err
	}

	result := &Result{Source: m.name, Items: m.items}
	for i := 0; i < m.items; i++ {
		result.Files = append(result.Files, fmt.Sprintf("/output/%s/%d.md", m.name, i))
	}

	if cfg.DryRun {
		return &Result{Source: m.name}, nil
	}

	return result, nil
}

func TestExcavator_Name_Good(t *testing.T) {
	e := &Excavator{}
	assert.Equal(t, "excavator", e.Name())
}

func TestExcavator_Run_Good(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	c1 := &mockCollector{name: "source-a", items: 3}
	c2 := &mockCollector{name: "source-b", items: 5}

	e := &Excavator{
		Collectors: []Collector{c1, c2},
	}

	result, err := e.Run(context.Background(), cfg)

	assert.NoError(t, err)
	assert.True(t, c1.called)
	assert.True(t, c2.called)
	assert.Equal(t, 8, result.Items)
	assert.Len(t, result.Files, 8)
}

func TestExcavator_Run_Good_Empty(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")

	e := &Excavator{}
	result, err := e.Run(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Items)
}

func TestExcavator_Run_Good_DryRun(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	c1 := &mockCollector{name: "source-a", items: 10}
	c2 := &mockCollector{name: "source-b", items: 20}

	e := &Excavator{
		Collectors: []Collector{c1, c2},
	}

	result, err := e.Run(context.Background(), cfg)

	assert.NoError(t, err)
	assert.True(t, c1.called)
	assert.True(t, c2.called)
	// In dry run, mockCollector returns 0 items
	assert.Equal(t, 0, result.Items)
}

func TestExcavator_Run_Good_ScanOnly(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")

	c1 := &mockCollector{name: "source-a", items: 10}

	var progressMessages []string
	cfg.Dispatcher.On(EventProgress, func(e Event) {
		progressMessages = append(progressMessages, e.Message)
	})

	e := &Excavator{
		Collectors: []Collector{c1},
		ScanOnly:   true,
	}

	result, err := e.Run(context.Background(), cfg)

	assert.NoError(t, err)
	assert.False(t, c1.called, "Collector should not be called in scan-only mode")
	assert.Equal(t, 0, result.Items)
	assert.NotEmpty(t, progressMessages)
	assert.Contains(t, progressMessages[0], "source-a")
}

func TestExcavator_Run_Good_WithErrors(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	c1 := &mockCollector{name: "good", items: 5}
	c2 := &mockCollector{name: "bad", err: fmt.Errorf("network error")}
	c3 := &mockCollector{name: "also-good", items: 3}

	e := &Excavator{
		Collectors: []Collector{c1, c2, c3},
	}

	result, err := e.Run(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 8, result.Items)
	assert.Equal(t, 1, result.Errors) // c2 failed
	assert.True(t, c1.called)
	assert.True(t, c2.called)
	assert.True(t, c3.called)
}

func TestExcavator_Run_Good_CancelledContext(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	c1 := &mockCollector{name: "source-a", items: 5}

	e := &Excavator{
		Collectors: []Collector{c1},
	}

	_, err := e.Run(ctx, cfg)
	assert.Error(t, err)
}

func TestExcavator_Run_Good_SavesState(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	c1 := &mockCollector{name: "source-a", items: 5}

	e := &Excavator{
		Collectors: []Collector{c1},
	}

	_, err := e.Run(context.Background(), cfg)
	assert.NoError(t, err)

	// Verify state was saved
	entry, ok := cfg.State.Get("source-a")
	assert.True(t, ok)
	assert.Equal(t, 5, entry.Items)
	assert.Equal(t, "source-a", entry.Source)
}

func TestExcavator_Run_Good_Events(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	var startCount, completeCount int
	cfg.Dispatcher.On(EventStart, func(e Event) { startCount++ })
	cfg.Dispatcher.On(EventComplete, func(e Event) { completeCount++ })

	c1 := &mockCollector{name: "source-a", items: 1}
	e := &Excavator{
		Collectors: []Collector{c1},
	}

	_, err := e.Run(context.Background(), cfg)
	assert.NoError(t, err)
	assert.Equal(t, 1, startCount)
	assert.Equal(t, 1, completeCount)
}
