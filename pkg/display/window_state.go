package display

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// WindowState holds the persisted state of a window.
type WindowState struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Maximized bool   `json:"maximized"`
	Screen    string `json:"screen,omitempty"` // Screen identifier for multi-monitor
	URL       string `json:"url,omitempty"`    // Last URL/route
	UpdatedAt int64  `json:"updatedAt"`
}

// WindowStateManager handles saving and restoring window positions.
type WindowStateManager struct {
	states    map[string]*WindowState
	filePath  string
	mu        sync.RWMutex
	dirty     bool
	saveTimer *time.Timer
}

// NewWindowStateManager creates a new window state manager.
// It loads existing state from the config directory.
func NewWindowStateManager() *WindowStateManager {
	m := &WindowStateManager{
		states: make(map[string]*WindowState),
	}

	// Determine config path
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	m.filePath = filepath.Join(configDir, "Core", "window_state.json")

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(m.filePath), 0755)

	// Load existing state
	m.load()

	return m
}

// load reads window states from disk.
func (m *WindowStateManager) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No saved state yet
		}
		return err
	}

	return json.Unmarshal(data, &m.states)
}

// save writes window states to disk.
func (m *WindowStateManager) save() error {
	m.mu.RLock()
	data, err := json.MarshalIndent(m.states, "", "  ")
	m.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}

// scheduleSave debounces saves to avoid excessive disk writes.
func (m *WindowStateManager) scheduleSave() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.dirty = true

	// Cancel existing timer
	if m.saveTimer != nil {
		m.saveTimer.Stop()
	}

	// Schedule save after 500ms of no changes
	m.saveTimer = time.AfterFunc(500*time.Millisecond, func() {
		m.mu.Lock()
		if m.dirty {
			m.dirty = false
			m.mu.Unlock()
			m.save()
		} else {
			m.mu.Unlock()
		}
	})
}

// GetState returns the saved state for a window, or nil if none.
func (m *WindowStateManager) GetState(name string) *WindowState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.states[name]
}

// SetState saves the state for a window.
func (m *WindowStateManager) SetState(name string, state *WindowState) {
	m.mu.Lock()
	state.UpdatedAt = time.Now().Unix()
	m.states[name] = state
	m.mu.Unlock()

	m.scheduleSave()
}

// UpdatePosition updates just the position of a window.
func (m *WindowStateManager) UpdatePosition(name string, x, y int) {
	m.mu.Lock()
	state, ok := m.states[name]
	if !ok {
		state = &WindowState{}
		m.states[name] = state
	}
	state.X = x
	state.Y = y
	state.UpdatedAt = time.Now().Unix()
	m.mu.Unlock()

	m.scheduleSave()
}

// UpdateSize updates just the size of a window.
func (m *WindowStateManager) UpdateSize(name string, width, height int) {
	m.mu.Lock()
	state, ok := m.states[name]
	if !ok {
		state = &WindowState{}
		m.states[name] = state
	}
	state.Width = width
	state.Height = height
	state.UpdatedAt = time.Now().Unix()
	m.mu.Unlock()

	m.scheduleSave()
}

// UpdateMaximized updates the maximized state of a window.
func (m *WindowStateManager) UpdateMaximized(name string, maximized bool) {
	m.mu.Lock()
	state, ok := m.states[name]
	if !ok {
		state = &WindowState{}
		m.states[name] = state
	}
	state.Maximized = maximized
	state.UpdatedAt = time.Now().Unix()
	m.mu.Unlock()

	m.scheduleSave()
}

// CaptureState captures the current state from a window.
func (m *WindowStateManager) CaptureState(name string, window *application.WebviewWindow) {
	if window == nil {
		return
	}

	x, y := window.Position()
	width, height := window.Size()

	m.mu.Lock()
	state, ok := m.states[name]
	if !ok {
		state = &WindowState{}
		m.states[name] = state
	}
	state.X = x
	state.Y = y
	state.Width = width
	state.Height = height
	state.Maximized = window.IsMaximised()
	state.UpdatedAt = time.Now().Unix()
	m.mu.Unlock()

	m.scheduleSave()
}

// ApplyState applies saved state to window options.
// Returns the modified options with position/size restored.
func (m *WindowStateManager) ApplyState(opts application.WebviewWindowOptions) application.WebviewWindowOptions {
	state := m.GetState(opts.Name)
	if state == nil {
		return opts
	}

	// Only apply if we have valid saved dimensions
	if state.Width > 0 && state.Height > 0 {
		opts.Width = state.Width
		opts.Height = state.Height
	}

	// Apply position (check for reasonable values)
	if state.X != 0 || state.Y != 0 {
		opts.X = state.X
		opts.Y = state.Y
	}

	// Apply maximized state
	if state.Maximized {
		opts.StartState = application.WindowStateMaximised
	}

	return opts
}

// ForceSync immediately saves all state to disk.
func (m *WindowStateManager) ForceSync() error {
	m.mu.Lock()
	if m.saveTimer != nil {
		m.saveTimer.Stop()
		m.saveTimer = nil
	}
	m.dirty = false
	m.mu.Unlock()

	return m.save()
}

// Clear removes all saved window states.
func (m *WindowStateManager) Clear() error {
	m.mu.Lock()
	m.states = make(map[string]*WindowState)
	m.mu.Unlock()

	return m.save()
}

// ListStates returns all saved window names.
func (m *WindowStateManager) ListStates() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.states))
	for name := range m.states {
		names = append(names, name)
	}
	return names
}
