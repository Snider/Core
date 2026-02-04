package collect

import (
	"testing"
	"time"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestState_SetGet_Good(t *testing.T) {
	m := io.NewMockMedium()
	s := NewState(m, "/state.json")

	entry := &StateEntry{
		Source:  "github:test",
		LastRun: time.Now(),
		Items:   42,
		LastID:  "abc123",
		Cursor:  "cursor-xyz",
	}

	s.Set("github:test", entry)

	got, ok := s.Get("github:test")
	assert.True(t, ok)
	assert.Equal(t, entry.Source, got.Source)
	assert.Equal(t, entry.Items, got.Items)
	assert.Equal(t, entry.LastID, got.LastID)
	assert.Equal(t, entry.Cursor, got.Cursor)
}

func TestState_Get_Bad(t *testing.T) {
	m := io.NewMockMedium()
	s := NewState(m, "/state.json")

	got, ok := s.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestState_SaveLoad_Good(t *testing.T) {
	m := io.NewMockMedium()
	s := NewState(m, "/state.json")

	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	entry := &StateEntry{
		Source:  "market:bitcoin",
		LastRun: now,
		Items:   100,
		LastID:  "btc-100",
	}

	s.Set("market:bitcoin", entry)

	// Save state
	err := s.Save()
	assert.NoError(t, err)

	// Verify file was written
	assert.True(t, m.IsFile("/state.json"))

	// Load into a new state instance
	s2 := NewState(m, "/state.json")
	err = s2.Load()
	assert.NoError(t, err)

	got, ok := s2.Get("market:bitcoin")
	assert.True(t, ok)
	assert.Equal(t, "market:bitcoin", got.Source)
	assert.Equal(t, 100, got.Items)
	assert.Equal(t, "btc-100", got.LastID)
	assert.True(t, now.Equal(got.LastRun))
}

func TestState_Load_Good_NoFile(t *testing.T) {
	m := io.NewMockMedium()
	s := NewState(m, "/nonexistent.json")

	// Loading when no file exists should not error
	err := s.Load()
	assert.NoError(t, err)

	// State should be empty
	_, ok := s.Get("anything")
	assert.False(t, ok)
}

func TestState_Load_Bad_InvalidJSON(t *testing.T) {
	m := io.NewMockMedium()
	m.Files["/state.json"] = "not valid json"

	s := NewState(m, "/state.json")
	err := s.Load()
	assert.Error(t, err)
}

func TestState_SaveLoad_Good_MultipleEntries(t *testing.T) {
	m := io.NewMockMedium()
	s := NewState(m, "/state.json")

	s.Set("source-a", &StateEntry{Source: "source-a", Items: 10})
	s.Set("source-b", &StateEntry{Source: "source-b", Items: 20})
	s.Set("source-c", &StateEntry{Source: "source-c", Items: 30})

	err := s.Save()
	assert.NoError(t, err)

	s2 := NewState(m, "/state.json")
	err = s2.Load()
	assert.NoError(t, err)

	a, ok := s2.Get("source-a")
	assert.True(t, ok)
	assert.Equal(t, 10, a.Items)

	b, ok := s2.Get("source-b")
	assert.True(t, ok)
	assert.Equal(t, 20, b.Items)

	c, ok := s2.Get("source-c")
	assert.True(t, ok)
	assert.Equal(t, 30, c.Items)
}

func TestState_Set_Good_Overwrite(t *testing.T) {
	m := io.NewMockMedium()
	s := NewState(m, "/state.json")

	s.Set("source", &StateEntry{Source: "source", Items: 5})
	s.Set("source", &StateEntry{Source: "source", Items: 15})

	got, ok := s.Get("source")
	assert.True(t, ok)
	assert.Equal(t, 15, got.Items)
}

func TestNewState_Good(t *testing.T) {
	m := io.NewMockMedium()
	s := NewState(m, "/test/state.json")

	assert.NotNil(t, s)
	assert.NotNil(t, s.entries)
}
