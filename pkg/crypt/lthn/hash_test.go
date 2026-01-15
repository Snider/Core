package lthn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "hashes simple string",
			payload: "hello",
		},
		{
			name:    "hashes empty string",
			payload: "",
		},
		{
			name:    "hashes unicode",
			payload: "héllo wörld 日本語",
		},
		{
			name:    "hashes long string",
			payload: "the quick brown fox jumps over the lazy dog",
		},
		{
			name:    "hashes special characters",
			payload: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash(tt.payload)

			// Should return a non-empty hash
			assert.NotEmpty(t, result)

			// Should be consistent (same input = same output)
			assert.Equal(t, result, Hash(tt.payload))
		})
	}
}

func TestHash_Uniqueness(t *testing.T) {
	// Different inputs should produce different hashes
	hash1 := Hash("input1")
	hash2 := Hash("input2")
	hash3 := Hash("input3")

	assert.NotEqual(t, hash1, hash2)
	assert.NotEqual(t, hash2, hash3)
	assert.NotEqual(t, hash1, hash3)
}

func TestHash_Consistency(t *testing.T) {
	// Same input should always produce the same hash
	payload := "consistent-test-payload"

	results := make([]string, 10)
	for i := 0; i < 10; i++ {
		results[i] = Hash(payload)
	}

	for i := 1; i < len(results); i++ {
		assert.Equal(t, results[0], results[i], "hash should be consistent across calls")
	}
}
