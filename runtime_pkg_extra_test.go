package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWithFactories_EmptyName(t *testing.T) {
	factories := map[string]ServiceFactory{
		"": func() (any, error) {
			return &MockService{Name: "test"}, nil
		},
	}
	_, err := NewWithFactories(nil, factories)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service name cannot be empty")
}
