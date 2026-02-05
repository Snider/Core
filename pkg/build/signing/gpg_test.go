package signing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPGSigner_Good_Name(t *testing.T) {
	s := NewGPGSigner("ABCD1234")
	assert.Equal(t, "gpg", s.Name())
}

func TestGPGSigner_Good_Available(t *testing.T) {
	s := NewGPGSigner("ABCD1234")
	_ = s.Available()
}

func TestGPGSigner_Bad_NoKey(t *testing.T) {
	s := NewGPGSigner("")
	assert.False(t, s.Available())
}

func TestGPGSigner_Sign_Bad(t *testing.T) {
	t.Run("fails when no key", func(t *testing.T) {
		s := NewGPGSigner("")
		err := s.Sign(context.Background(), "test.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not available or key not configured")
	})
}
