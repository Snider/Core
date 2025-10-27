package io

import "github.com/stretchr/testify/assert"

// MockMedium implements the Medium interface for testing purposes.
type MockMedium struct {
	Files map[string]string
	Dirs  map[string]bool
}

func NewMockMedium() *MockMedium {
	return &MockMedium{
		Files: make(map[string]string),
		Dirs:  make(map[string]bool),
	}
}

func (m *MockMedium) Read(path string) (string, error) {
	content, ok := m.Files[path]
	if !ok {
		return "", assert.AnError // Simulate file not found error
	}
	return content, nil
}

func (m *MockMedium) Write(path, content string) error {
	m.Files[path] = content
	return nil
}

func (m *MockMedium) EnsureDir(path string) error {
	m.Dirs[path] = true
	return nil
}

func (m *MockMedium) IsFile(path string) bool {
	_, ok := m.Files[path]
	return ok
}

func (m *MockMedium) FileGet(path string) (string, error) {
	return m.Read(path)
}

func (m *MockMedium) FileSet(path, content string) error {
	return m.Write(path, content)
}
