package container

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/host-uk/core/pkg/io"
)

// State manages persistent container state.
type State struct {
	// Containers is a map of container ID to Container.
	Containers map[string]*Container `json:"containers"`

	mu       sync.RWMutex
	medium   io.Medium
	filePath string
}

// DefaultStateDir returns the default directory for state files (~/.core).
func DefaultStateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".core"), nil
}

// DefaultStatePath returns the default path for the state file.
func DefaultStatePath() (string, error) {
	dir, err := DefaultStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "containers.json"), nil
}

// DefaultLogsDir returns the default directory for container logs.
func DefaultLogsDir() (string, error) {
	dir, err := DefaultStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "logs"), nil
}

// NewState creates a new State instance.
func NewState(m io.Medium, filePath string) *State {
	return &State{
		Containers: make(map[string]*Container),
		medium:     m,
		filePath:   filePath,
	}
}

// LoadState loads the state from the given file path.
// If the file doesn't exist, returns an empty state.
func LoadState(m io.Medium, filePath string) (*State, error) {
	state := NewState(m, filePath)

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	content, err := m.Read(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(content), state); err != nil {
		return nil, err
	}

	return state, nil
}

// SaveState persists the state to the configured file path.
func (s *State) SaveState() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	absPath, err := filepath.Abs(s.filePath)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// s.medium.Write creates parent directories automatically
	return s.medium.Write(absPath, string(data))
}

// Add adds a container to the state and persists it.
func (s *State) Add(c *Container) error {
	s.mu.Lock()
	s.Containers[c.ID] = c
	s.mu.Unlock()

	return s.SaveState()
}

// Get retrieves a copy of a container by ID.
// Returns a copy to prevent data races when the container is modified.
func (s *State) Get(id string) (*Container, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.Containers[id]
	if !ok {
		return nil, false
	}
	// Return a copy to prevent data races
	copy := *c
	return &copy, true
}

// Update updates a container in the state and persists it.
func (s *State) Update(c *Container) error {
	s.mu.Lock()
	s.Containers[c.ID] = c
	s.mu.Unlock()

	return s.SaveState()
}

// Remove removes a container from the state and persists it.
func (s *State) Remove(id string) error {
	s.mu.Lock()
	delete(s.Containers, id)
	s.mu.Unlock()

	return s.SaveState()
}

// All returns copies of all containers in the state.
// Returns copies to prevent data races when containers are modified.
func (s *State) All() []*Container {
	s.mu.RLock()
	defer s.mu.RUnlock()

	containers := make([]*Container, 0, len(s.Containers))
	for _, c := range s.Containers {
		copy := *c
		containers = append(containers, &copy)
	}
	return containers
}

// FilePath returns the path to the state file.
func (s *State) FilePath() string {
	return s.filePath
}

// LogPath returns the log file path for a given container ID.
func LogPath(id string) (string, error) {
	logsDir, err := DefaultLogsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(logsDir, id+".log"), nil
}

// EnsureLogsDir ensures the logs directory exists.
func EnsureLogsDir(m io.Medium) error {
	logsDir, err := DefaultLogsDir()
	if err != nil {
		return err
	}
	return m.EnsureDir(logsDir)
}
