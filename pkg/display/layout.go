package display

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Layout represents a saved window arrangement.
type Layout struct {
	Name      string                 `json:"name"`
	Windows   map[string]WindowState `json:"windows"`
	CreatedAt int64                  `json:"createdAt"`
	UpdatedAt int64                  `json:"updatedAt"`
}

// LayoutManager handles saving and restoring window layouts.
type LayoutManager struct {
	layouts  map[string]*Layout
	filePath string
	mu       sync.RWMutex
}

// NewLayoutManager creates a new layout manager.
func NewLayoutManager() *LayoutManager {
	m := &LayoutManager{
		layouts: make(map[string]*Layout),
	}

	// Determine config path
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	m.filePath = filepath.Join(configDir, "Core", "layouts.json")

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(m.filePath), 0755)

	// Load existing layouts
	m.load()

	return m
}

// load reads layouts from disk.
func (m *LayoutManager) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No saved layouts yet
		}
		return err
	}

	return json.Unmarshal(data, &m.layouts)
}

// save writes layouts to disk.
func (m *LayoutManager) save() error {
	m.mu.RLock()
	data, err := json.MarshalIndent(m.layouts, "", "  ")
	m.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}

// SaveLayout saves a new layout or updates an existing one.
func (m *LayoutManager) SaveLayout(name string, windows map[string]WindowState) error {
	if name == "" {
		return fmt.Errorf("layout name is required")
	}

	m.mu.Lock()
	now := time.Now().Unix()

	existing, ok := m.layouts[name]
	if ok {
		// Update existing layout
		existing.Windows = windows
		existing.UpdatedAt = now
	} else {
		// Create new layout
		m.layouts[name] = &Layout{
			Name:      name,
			Windows:   windows,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	m.mu.Unlock()

	return m.save()
}

// GetLayout returns a layout by name.
func (m *LayoutManager) GetLayout(name string) *Layout {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.layouts[name]
}

// ListLayouts returns all saved layout names with metadata.
func (m *LayoutManager) ListLayouts() []LayoutInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]LayoutInfo, 0, len(m.layouts))
	for _, layout := range m.layouts {
		result = append(result, LayoutInfo{
			Name:        layout.Name,
			WindowCount: len(layout.Windows),
			CreatedAt:   layout.CreatedAt,
			UpdatedAt:   layout.UpdatedAt,
		})
	}
	return result
}

// DeleteLayout removes a layout by name.
func (m *LayoutManager) DeleteLayout(name string) error {
	m.mu.Lock()
	if _, ok := m.layouts[name]; !ok {
		m.mu.Unlock()
		return fmt.Errorf("layout not found: %s", name)
	}
	delete(m.layouts, name)
	m.mu.Unlock()

	return m.save()
}

// LayoutInfo contains summary information about a layout.
type LayoutInfo struct {
	Name        string `json:"name"`
	WindowCount int    `json:"windowCount"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}
