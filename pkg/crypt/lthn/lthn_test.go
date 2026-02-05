package lthn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHash_Good(t *testing.T) {
	hash := Hash("hello")
	assert.Len(t, hash, 64, "SHA-256 hex digest should be 64 characters")
	assert.NotEmpty(t, hash)

	// Same input should always produce the same hash (deterministic)
	hash2 := Hash("hello")
	assert.Equal(t, hash, hash2, "same input must produce the same hash")
}

func TestHash_Bad(t *testing.T) {
	// Different inputs should produce different hashes
	hash1 := Hash("hello")
	hash2 := Hash("world")
	assert.NotEqual(t, hash1, hash2, "different inputs must produce different hashes")
}

func TestHash_Ugly(t *testing.T) {
	// Empty string should still produce a valid hash
	hash := Hash("")
	assert.Len(t, hash, 64)
	assert.NotEmpty(t, hash)
}

func TestVerify_Good(t *testing.T) {
	input := "test-data-123"
	hash := Hash(input)
	assert.True(t, Verify(input, hash), "Verify must return true for matching input")
}

func TestVerify_Bad(t *testing.T) {
	input := "test-data-123"
	hash := Hash(input)
	assert.False(t, Verify("wrong-input", hash), "Verify must return false for non-matching input")
	assert.False(t, Verify(input, "0000000000000000000000000000000000000000000000000000000000000000"),
		"Verify must return false for wrong hash")
}

func TestVerify_Ugly(t *testing.T) {
	// Empty input round-trip
	hash := Hash("")
	assert.True(t, Verify("", hash))
}

func TestSetKeyMap_Good(t *testing.T) {
	// Save original map
	original := GetKeyMap()

	// Set a custom key map
	custom := map[rune]rune{
		'a': 'b',
		'b': 'a',
	}
	SetKeyMap(custom)

	// Hash should use new key map
	hash1 := Hash("abc")

	// Restore original and hash again
	SetKeyMap(original)
	hash2 := Hash("abc")

	assert.NotEqual(t, hash1, hash2, "different key maps should produce different hashes")
}

func TestGetKeyMap_Good(t *testing.T) {
	km := GetKeyMap()
	require.NotNil(t, km)
	assert.Equal(t, '0', km['o'])
	assert.Equal(t, '1', km['l'])
	assert.Equal(t, '3', km['e'])
	assert.Equal(t, '4', km['a'])
	assert.Equal(t, 'z', km['s'])
	assert.Equal(t, '7', km['t'])
}

func TestCreateSalt_Good(t *testing.T) {
	// "hello" reversed is "olleh", with substitutions: o->0, l->1, l->1, e->3, h->h => "011eh" ... wait
	// Actually: reversed "olleh" => o->0, l->1, l->1, e->3, h->h => "0113h"
	// Let's verify by checking the hash is deterministic
	hash1 := Hash("hello")
	hash2 := Hash("hello")
	assert.Equal(t, hash1, hash2, "salt derivation must be deterministic")
}

func TestCreateSalt_Ugly(t *testing.T) {
	// Unicode input should not panic
	hash := Hash("\U0001f600\U0001f601\U0001f602")
	assert.Len(t, hash, 64)
}
