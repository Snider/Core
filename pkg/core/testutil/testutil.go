package testutil

// MockConfig is a mock implementation of the core.Config interface.
type MockConfig struct{}

func (m *MockConfig) Get(key string, out any) error { return nil }
func (m *MockConfig) Set(key string, v any) error   { return nil }
