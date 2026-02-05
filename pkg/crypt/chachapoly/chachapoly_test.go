package chachapoly

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return key
}

func TestEncryptDecrypt_Good(t *testing.T) {
	key := generateKey(t)
	plaintext := []byte("hello, XChaCha20-Poly1305!")

	ciphertext, err := Encrypt(plaintext, key)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)
	// Ciphertext should be longer than plaintext (nonce + overhead)
	assert.Greater(t, len(ciphertext), len(plaintext))

	decrypted, err := Decrypt(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecrypt_Bad(t *testing.T) {
	key1 := generateKey(t)
	key2 := generateKey(t)
	plaintext := []byte("secret data")

	ciphertext, err := Encrypt(plaintext, key1)
	require.NoError(t, err)

	// Decrypting with a different key should fail
	_, err = Decrypt(ciphertext, key2)
	assert.Error(t, err)
}

func TestEncryptDecrypt_Ugly(t *testing.T) {
	// Invalid key length should fail
	shortKey := []byte("too-short")
	_, err := Encrypt([]byte("data"), shortKey)
	assert.Error(t, err)

	_, err = Decrypt([]byte("data"), shortKey)
	assert.Error(t, err)

	// Ciphertext too short should fail
	key := generateKey(t)
	_, err = Decrypt([]byte("short"), key)
	assert.Error(t, err)
}

func TestEncryptDecryptEmpty_Good(t *testing.T) {
	key := generateKey(t)
	plaintext := []byte{}

	ciphertext, err := Encrypt(plaintext, key)
	require.NoError(t, err)

	decrypted, err := Decrypt(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptNonDeterministic_Good(t *testing.T) {
	key := generateKey(t)
	plaintext := []byte("same input")

	ct1, err := Encrypt(plaintext, key)
	require.NoError(t, err)

	ct2, err := Encrypt(plaintext, key)
	require.NoError(t, err)

	// Different nonces mean different ciphertexts
	assert.NotEqual(t, ct1, ct2, "each encryption should produce unique ciphertext due to random nonce")

	// Both should decrypt to the same plaintext
	d1, err := Decrypt(ct1, key)
	require.NoError(t, err)
	d2, err := Decrypt(ct2, key)
	require.NoError(t, err)
	assert.Equal(t, d1, d2)
}
