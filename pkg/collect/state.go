package collect

import (
	"encoding/json"
	"sync"
	"time"

	core "github.com/host-uk/core/pkg/framework/core"
	"github.com/host-uk/core/pkg/io"
)

// State tracks collection progress for incremental runs.
// It persists entries to disk so that subsequent runs can resume
// where they left off.
type State struct {
	mu      sync.Mutex
	medium  io.Medium
	path    string
	entries map[string]*StateEntry
}

// StateEntry tracks state for one source.
type StateEntry struct {
	// Source identifies the collector.
	Source string `json:"source"`

	// LastRun is the timestamp of the last successful run.
	LastRun time.Time `json:"last_run"`

	// LastID is an opaque identifier for the last item processed.
	LastID string `json:"last_id,omitempty"`

	// Items is the total number of items collected so far.
	Items int `json:"items"`

	// Cursor is an opaque pagination cursor for resumption.
	Cursor string `json:"cursor,omitempty"`
}

// NewState creates a state tracker that persists to the given path
// using the provided storage medium.
func NewState(m io.Medium, path string) *State {
	return &State{
		medium:  m,
		path:    path,
		entries: make(map[string]*StateEntry),
	}
}

// Load reads state from disk. If the file does not exist, the state
// is initialised as empty without error.
func (s *State) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.medium.IsFile(s.path) {
		return nil
	}

	data, err := s.medium.Read(s.path)
	if err != nil {
		return core.E("collect.State.Load", "failed to read state file", err)
	}

	var entries map[string]*StateEntry
	if err := json.Unmarshal([]byte(data), &entries); err != nil {
		return core.E("collect.State.Load", "failed to parse state file", err)
	}

	if entries == nil {
		entries = make(map[string]*StateEntry)
	}
	s.entries = entries
	return nil
}

// Save writes state to disk.
func (s *State) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return core.E("collect.State.Save", "failed to marshal state", err)
	}

	if err := s.medium.Write(s.path, string(data)); err != nil {
		return core.E("collect.State.Save", "failed to write state file", err)
	}

	return nil
}

// Get returns a copy of the state for a source. The second return value
// indicates whether the entry was found.
func (s *State) Get(source string) (*StateEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[source]
	if !ok {
		return nil, false
	}
	// Return a copy to avoid callers mutating internal state.
	cp := *entry
	return &cp, true
}

// Set updates state for a source.
func (s *State) Set(source string, entry *StateEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[source] = entry
}
