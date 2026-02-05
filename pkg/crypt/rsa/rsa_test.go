package rsa

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeyPair_Good(t *testing.T) {
	kp, err := GenerateKeyPair(2048)
	require.NoError(t, err)
	require.NotNil(t, kp)
	assert.Contains(t, kp.PublicKey, "-----BEGIN PUBLIC KEY-----")
	assert.Contains(t, kp.PrivateKey, "-----BEGIN RSA PRIVATE KEY-----")
}

func TestGenerateKeyPair_Bad(t *testing.T) {
	// Key size too small
	_, err := GenerateKeyPair(1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key size too small")
}

func TestGenerateKeyPair_Ugly(t *testing.T) {
	// Zero bits
	_, err := GenerateKeyPair(0)
	assert.Error(t, err)
}

func TestEncryptDecrypt_Good(t *testing.T) {
	kp, err := GenerateKeyPair(2048)
	require.NoError(t, err)

	plaintext := []byte("hello, RSA-OAEP with SHA-256!")
	ciphertext, err := Encrypt(plaintext, kp.PublicKey)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := Decrypt(ciphertext, kp.PrivateKey)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecrypt_Bad(t *testing.T) {
	kp1, err := GenerateKeyPair(2048)
	require.NoError(t, err)
	kp2, err := GenerateKeyPair(2048)
	require.NoError(t, err)

	plaintext := []byte("secret data")
	ciphertext, err := Encrypt(plaintext, kp1.PublicKey)
	require.NoError(t, err)

	// Decrypting with wrong private key should fail
	_, err = Decrypt(ciphertext, kp2.PrivateKey)
	assert.Error(t, err)
}

func TestEncryptDecrypt_Ugly(t *testing.T) {
	// Invalid PEM for encryption
	_, err := Encrypt([]byte("data"), "not-a-pem-key")
	assert.Error(t, err)

	// Invalid PEM for decryption
	_, err = Decrypt([]byte("data"), "not-a-pem-key")
	assert.Error(t, err)
}

func TestEncryptDecryptRoundTrip_Good(t *testing.T) {
	kp, err := GenerateKeyPair(2048)
	require.NoError(t, err)

	messages := []string{
		"",
		"a",
		"short message",
		"a slightly longer message with some special chars: !@#$%^&*()",
	}

	for _, msg := range messages {
		ciphertext, err := Encrypt([]byte(msg), kp.PublicKey)
		require.NoError(t, err)

		decrypted, err := Decrypt(ciphertext, kp.PrivateKey)
		require.NoError(t, err)
		assert.Equal(t, msg, string(decrypted), "round-trip failed for: %q", msg)
	}
}
