package crypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	s := &Service{}
	payload := "hello"
	hash := s.Hash(LTHN, payload)
	assert.NotEmpty(t, hash)
}

func TestLuhn(t *testing.T) {
	s := &Service{}
	assert.True(t, s.Luhn("79927398713"))
	assert.False(t, s.Luhn("79927398714"))
}
