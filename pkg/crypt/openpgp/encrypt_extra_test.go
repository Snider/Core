package openpgp

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDecryptMalformedMessage checks that DecryptPGP handles non-PGP or malformed input gracefully.
func TestDecryptMalformedMessage(t *testing.T) {
	// Generate a key pair for this test
	_, recipientPriv, cleanup := generateTestKeys(t, "recipient", "")
	defer cleanup()

	malformedMessage := "This is not a PGP message."

	// The passphrase parameter is ignored by Enchantrix
	_, err := DecryptPGP(recipientPriv, malformedMessage, "", nil)
	assert.Error(t, err, "Decryption should fail for a malformed message, but it did not.")
	// Enchantrix returns a different error message
	assert.Contains(t, err.Error(), "failed to read PGP message", "Expected error about failing to read PGP message")
}

// TestEncryptWithNonexistentRecipient checks that EncryptPGP fails when the recipient's public key file does not exist.
func TestEncryptWithNonexistentRecipient(t *testing.T) {
	var encryptedBuf bytes.Buffer
	err := EncryptPGP(&encryptedBuf, "/path/to/nonexistent/key.pub", "message", nil, nil)
	assert.Error(t, err, "Encryption should fail if recipient key does not exist, but it succeeded.")
	assert.Contains(t, err.Error(), "failed to open recipient public key file", "Expected file open error for recipient key")
}

// TestEncryptDecryptRoundtrip verifies that encryption and decryption work correctly.
func TestEncryptDecryptRoundtrip(t *testing.T) {
	recipientPub, recipientPriv, cleanup := generateTestKeys(t, "recipient", "")
	defer cleanup()

	originalMessage := "Hello, PGP World!"

	var encryptedBuf bytes.Buffer
	err := EncryptPGP(&encryptedBuf, recipientPub, originalMessage, nil, nil)
	assert.NoError(t, err, "Encryption failed unexpectedly")

	encryptedMessage := encryptedBuf.String()
	assert.NotEmpty(t, encryptedMessage, "Encrypted message should not be empty")
	assert.NotEqual(t, originalMessage, encryptedMessage, "Encrypted message should differ from original")

	// Decrypt the message
	decryptedMessage, err := DecryptPGP(recipientPriv, encryptedMessage, "", nil)
	assert.NoError(t, err, "Decryption failed unexpectedly")
	assert.Equal(t, originalMessage, decryptedMessage, "Decrypted message should match original")
}

// TestEncryptToStringAndDecrypt tests the EncryptPGPToString convenience function.
func TestEncryptToStringAndDecrypt(t *testing.T) {
	keyPair, err := CreateKeyPair("test-user", "")
	assert.NoError(t, err, "Key pair creation failed")
	assert.NotNil(t, keyPair)

	originalMessage := "Test message for string encryption"

	encrypted, err := EncryptPGPToString(keyPair.PublicKey, originalMessage)
	assert.NoError(t, err, "EncryptPGPToString failed")
	assert.NotEmpty(t, encrypted)

	// Write private key to temp file for DecryptPGP which expects file path
	tempDir := t.TempDir()
	privKeyPath := tempDir + "/key.priv"
	err = writeFile(privKeyPath, keyPair.PrivateKey)
	assert.NoError(t, err)

	decrypted, err := DecryptPGP(privKeyPath, encrypted, "", nil)
	assert.NoError(t, err, "Decryption failed")
	assert.Equal(t, originalMessage, decrypted)
}

// TestCreateKeyPair tests key pair generation.
func TestCreateKeyPair(t *testing.T) {
	t.Run("creates valid key pair", func(t *testing.T) {
		keyPair, err := CreateKeyPair("test-identity", "")
		assert.NoError(t, err)
		assert.NotNil(t, keyPair)
		assert.NotEmpty(t, keyPair.PublicKey)
		assert.NotEmpty(t, keyPair.PrivateKey)
		assert.Contains(t, keyPair.PublicKey, "BEGIN PGP PUBLIC KEY BLOCK")
		assert.Contains(t, keyPair.PrivateKey, "BEGIN PGP PRIVATE KEY BLOCK")
	})

	t.Run("different identities produce different keys", func(t *testing.T) {
		keyPair1, err1 := CreateKeyPair("identity1", "")
		keyPair2, err2 := CreateKeyPair("identity2", "")
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, keyPair1.PublicKey, keyPair2.PublicKey)
		assert.NotEqual(t, keyPair1.PrivateKey, keyPair2.PrivateKey)
	})
}

// Helper to write a file
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0600)
}

// TestDecryptWithNonexistentKey tests DecryptPGP with a non-existent key file.
func TestDecryptWithNonexistentKey(t *testing.T) {
	_, err := DecryptPGP("/path/to/nonexistent/key.priv", "encrypted message", "", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open recipient private key file")
}

// TestEncryptPGPToStringWithInvalidKey tests EncryptPGPToString with an invalid key.
func TestEncryptPGPToStringWithInvalidKey(t *testing.T) {
	_, err := EncryptPGPToString("not-a-valid-key", "test message")
	assert.Error(t, err)
}

// TestCreateEncryptedKeyFile tests the helper function for creating encrypted key files.
func TestCreateEncryptedKeyFile(t *testing.T) {
	path, cleanup := createEncryptedKeyFile(t)
	defer cleanup()

	assert.NotEmpty(t, path)

	// Verify file exists and has correct content
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "BEGIN PGP PRIVATE KEY BLOCK")
}

// errorWriter is a mock writer that always returns an error.
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, os.ErrPermission
}

// TestEncryptPGPWriteError tests that EncryptPGP handles write errors correctly.
func TestEncryptPGPWriteError(t *testing.T) {
	recipientPub, _, cleanup := generateTestKeys(t, "recipient", "")
	defer cleanup()

	err := EncryptPGP(&errorWriter{}, recipientPub, "test message", nil, nil)
	assert.Error(t, err)
}

// TestGenerateTestKeys tests the generateTestKeys helper function.
func TestGenerateTestKeys(t *testing.T) {
	pubPath, privPath, cleanup := generateTestKeys(t, "test-user", "test-pass")
	defer cleanup()

	assert.NotEmpty(t, pubPath)
	assert.NotEmpty(t, privPath)

	// Verify files exist
	pubContent, err := os.ReadFile(pubPath)
	assert.NoError(t, err)
	assert.Contains(t, string(pubContent), "BEGIN PGP PUBLIC KEY BLOCK")

	privContent, err := os.ReadFile(privPath)
	assert.NoError(t, err)
	assert.Contains(t, string(privContent), "BEGIN PGP PRIVATE KEY BLOCK")
}
