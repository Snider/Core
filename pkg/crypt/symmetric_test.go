package crypt

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChaCha20_Good(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.NoError(t, err)

	plaintext := []byte("ChaCha20-Poly1305 test data")

	encrypted, err := ChaCha20Encrypt(plaintext, key)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := ChaCha20Decrypt(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestChaCha20_Bad(t *testing.T) {
	key := make([]byte, 32)
	wrongKey := make([]byte, 32)
	_, _ = rand.Read(key)
	_, _ = rand.Read(wrongKey)

	plaintext := []byte("secret message")

	encrypted, err := ChaCha20Encrypt(plaintext, key)
	assert.NoError(t, err)

	_, err = ChaCha20Decrypt(encrypted, wrongKey)
	assert.Error(t, err)
}

func TestAESGCM_Good(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	assert.NoError(t, err)

	plaintext := []byte("AES-256-GCM test data")

	encrypted, err := AESGCMEncrypt(plaintext, key)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := AESGCMDecrypt(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
