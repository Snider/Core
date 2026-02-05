package signing

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMacOSSigner_Good_Name(t *testing.T) {
	s := NewMacOSSigner(MacOSConfig{Identity: "Developer ID Application: Test"})
	assert.Equal(t, "codesign", s.Name())
}

func TestMacOSSigner_Good_Available(t *testing.T) {
	s := NewMacOSSigner(MacOSConfig{Identity: "Developer ID Application: Test"})

	if runtime.GOOS == "darwin" {
		// Just verify it doesn't panic
		_ = s.Available()
	} else {
		assert.False(t, s.Available())
	}
}

func TestMacOSSigner_Bad_NoIdentity(t *testing.T) {
	s := NewMacOSSigner(MacOSConfig{})
	assert.False(t, s.Available())
}

func TestMacOSSigner_Sign_Bad(t *testing.T) {
	t.Run("fails when not available", func(t *testing.T) {
		if runtime.GOOS == "darwin" {
			t.Skip("skipping on macOS")
		}
		s := NewMacOSSigner(MacOSConfig{Identity: "test"})
		err := s.Sign(context.Background(), "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available")
	})
}

func TestMacOSSigner_Notarize_Bad(t *testing.T) {
	t.Run("fails with missing credentials", func(t *testing.T) {
		s := NewMacOSSigner(MacOSConfig{})
		err := s.Notarize(context.Background(), "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing Apple credentials")
	})
}

func TestMacOSSigner_ShouldNotarize(t *testing.T) {
	s := NewMacOSSigner(MacOSConfig{Notarize: true})
	assert.True(t, s.ShouldNotarize())

	s2 := NewMacOSSigner(MacOSConfig{Notarize: false})
	assert.False(t, s2.ShouldNotarize())
}
