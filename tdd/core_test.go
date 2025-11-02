package tdd

import (
	"errors"
	"testing"

	"github.com/Snider/Core/pkg/core"
	"github.com/stretchr/testify/assert"
)

// TestCore_New_Good tests the successful creation of a Core instance.
func TestCore_New_Good(t *testing.T) {
	c, err := core.New()
	assert.NoError(t, err, "core.New() should not return an error on default initialization")
	assert.NotNil(t, c, "core.New() should return a non-nil Core instance")
}

// TestCore_New_Bad tests Core creation with an option that returns an error.
func TestCore_New_Bad(t *testing.T) {
	expectedErr := errors.New("test error")
	badOption := func(c *core.Core) error {
		return expectedErr
	}

	_, err := core.New(badOption)
	assert.Error(t, err, "core.New() should return an error when a bad option is provided")
	assert.Equal(t, expectedErr, err, "The error returned should be the one from the bad option")
}

// TestCore_New_Ugly tests Core creation with a nil option.
func TestCore_New_Ugly(t *testing.T) {
	assert.NotPanics(t, func() {
		_, err := core.New(nil)
		assert.NoError(t, err, "core.New() should not return an error when a nil option is provided")
	}, "core.New() should not panic when a nil option is provided")
}
