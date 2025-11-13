package testutil

// MockConfig is a no-op mock implementation of the core.Config interface.
// Its methods do not perform any operations and always return nil.
type MockConfig struct{}

// Get is a no-op that always returns nil. The `out` parameter is not modified.
func (m *MockConfig) Get(key string, out any) error { return nil }

// Set is a no-op that always returns nil. The value `v` is not stored.
func (m *MockConfig) Set(key string, v any) error { return nil }
