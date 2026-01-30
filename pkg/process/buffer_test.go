package process

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRingBuffer(t *testing.T) {
	t.Run("write and read", func(t *testing.T) {
		rb := NewRingBuffer(10)

		n, err := rb.Write([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, "hello", rb.String())
		assert.Equal(t, 5, rb.Len())
	})

	t.Run("overflow wraps around", func(t *testing.T) {
		rb := NewRingBuffer(5)

		rb.Write([]byte("hello"))
		assert.Equal(t, "hello", rb.String())

		rb.Write([]byte("world"))
		// Should contain "world" (overwrote "hello")
		assert.Equal(t, 5, rb.Len())
		assert.Equal(t, "world", rb.String())
	})

	t.Run("partial overflow", func(t *testing.T) {
		rb := NewRingBuffer(10)

		rb.Write([]byte("hello"))
		rb.Write([]byte("worldx"))
		// Should contain "lloworldx" (11 chars, buffer is 10)
		assert.Equal(t, 10, rb.Len())
	})

	t.Run("empty buffer", func(t *testing.T) {
		rb := NewRingBuffer(10)
		assert.Equal(t, "", rb.String())
		assert.Equal(t, 0, rb.Len())
		assert.Nil(t, rb.Bytes())
	})

	t.Run("reset", func(t *testing.T) {
		rb := NewRingBuffer(10)
		rb.Write([]byte("hello"))
		rb.Reset()
		assert.Equal(t, "", rb.String())
		assert.Equal(t, 0, rb.Len())
	})

	t.Run("cap", func(t *testing.T) {
		rb := NewRingBuffer(42)
		assert.Equal(t, 42, rb.Cap())
	})

	t.Run("bytes returns copy", func(t *testing.T) {
		rb := NewRingBuffer(10)
		rb.Write([]byte("hello"))

		bytes := rb.Bytes()
		assert.Equal(t, []byte("hello"), bytes)

		// Modifying returned bytes shouldn't affect buffer
		bytes[0] = 'x'
		assert.Equal(t, "hello", rb.String())
	})
}
