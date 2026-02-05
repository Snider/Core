package pgp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateKeyPair_Good(t *testing.T) {
	kp, err := CreateKeyPair("Test User", "test@example.com", "")
	require.NoError(t, err)
	require.NotNil(t, kp)
	assert.Contains(t, kp.PublicKey, "-----BEGIN PGP PUBLIC KEY BLOCK-----")
	assert.Contains(t, kp.PrivateKey, "-----BEGIN PGP PRIVATE KEY BLOCK-----")
}

func TestCreateKeyPair_Bad(t *testing.T) {
	// Empty name still works (openpgp allows it), but test with password
	kp, err := CreateKeyPair("Secure User", "secure@example.com", "strong-password")
	require.NoError(t, err)
	require.NotNil(t, kp)
	assert.Contains(t, kp.PublicKey, "-----BEGIN PGP PUBLIC KEY BLOCK-----")
	assert.Contains(t, kp.PrivateKey, "-----BEGIN PGP PRIVATE KEY BLOCK-----")
}

func TestCreateKeyPair_Ugly(t *testing.T) {
	// Minimal identity
	kp, err := CreateKeyPair("", "", "")
	require.NoError(t, err)
	require.NotNil(t, kp)
}

func TestEncryptDecrypt_Good(t *testing.T) {
	kp, err := CreateKeyPair("Test User", "test@example.com", "")
	require.NoError(t, err)

	plaintext := []byte("hello, OpenPGP!")
	ciphertext, err := Encrypt(plaintext, kp.PublicKey)
	require.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.Contains(t, string(ciphertext), "-----BEGIN PGP MESSAGE-----")

	decrypted, err := Decrypt(ciphertext, kp.PrivateKey, "")
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecrypt_Bad(t *testing.T) {
	kp1, err := CreateKeyPair("User One", "one@example.com", "")
	require.NoError(t, err)
	kp2, err := CreateKeyPair("User Two", "two@example.com", "")
	require.NoError(t, err)

	plaintext := []byte("secret data")
	ciphertext, err := Encrypt(plaintext, kp1.PublicKey)
	require.NoError(t, err)

	// Decrypting with wrong key should fail
	_, err = Decrypt(ciphertext, kp2.PrivateKey, "")
	assert.Error(t, err)
}

func TestEncryptDecrypt_Ugly(t *testing.T) {
	// Invalid public key for encryption
	_, err := Encrypt([]byte("data"), "not-a-pgp-key")
	assert.Error(t, err)

	// Invalid private key for decryption
	_, err = Decrypt([]byte("data"), "not-a-pgp-key", "")
	assert.Error(t, err)
}

func TestEncryptDecryptWithPassword_Good(t *testing.T) {
	password := "my-secret-passphrase"
	kp, err := CreateKeyPair("Secure User", "secure@example.com", password)
	require.NoError(t, err)

	plaintext := []byte("encrypted with password-protected key")
	ciphertext, err := Encrypt(plaintext, kp.PublicKey)
	require.NoError(t, err)

	decrypted, err := Decrypt(ciphertext, kp.PrivateKey, password)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestSignVerify_Good(t *testing.T) {
	kp, err := CreateKeyPair("Signer", "signer@example.com", "")
	require.NoError(t, err)

	data := []byte("message to sign")
	signature, err := Sign(data, kp.PrivateKey, "")
	require.NoError(t, err)
	assert.NotEmpty(t, signature)
	assert.Contains(t, string(signature), "-----BEGIN PGP SIGNATURE-----")

	err = Verify(data, signature, kp.PublicKey)
	assert.NoError(t, err)
}

func TestSignVerify_Bad(t *testing.T) {
	kp, err := CreateKeyPair("Signer", "signer@example.com", "")
	require.NoError(t, err)

	data := []byte("original message")
	signature, err := Sign(data, kp.PrivateKey, "")
	require.NoError(t, err)

	// Verify with tampered data should fail
	err = Verify([]byte("tampered message"), signature, kp.PublicKey)
	assert.Error(t, err)
}

func TestSignVerify_Ugly(t *testing.T) {
	// Invalid key for signing
	_, err := Sign([]byte("data"), "not-a-key", "")
	assert.Error(t, err)

	// Invalid key for verification
	kp, err := CreateKeyPair("Signer", "signer@example.com", "")
	require.NoError(t, err)

	data := []byte("message")
	sig, err := Sign(data, kp.PrivateKey, "")
	require.NoError(t, err)

	err = Verify(data, sig, "not-a-key")
	assert.Error(t, err)
}

func TestSignVerifyWithPassword_Good(t *testing.T) {
	password := "signing-password"
	kp, err := CreateKeyPair("Signer", "signer@example.com", password)
	require.NoError(t, err)

	data := []byte("signed with password-protected key")
	signature, err := Sign(data, kp.PrivateKey, password)
	require.NoError(t, err)

	err = Verify(data, signature, kp.PublicKey)
	assert.NoError(t, err)
}

func TestFullRoundTrip_Good(t *testing.T) {
	// Generate keys, encrypt, decrypt, sign, and verify - full round trip
	kp, err := CreateKeyPair("Full Test", "full@example.com", "")
	require.NoError(t, err)

	original := []byte("full round-trip test data")

	// Encrypt then decrypt
	ciphertext, err := Encrypt(original, kp.PublicKey)
	require.NoError(t, err)
	decrypted, err := Decrypt(ciphertext, kp.PrivateKey, "")
	require.NoError(t, err)
	assert.Equal(t, original, decrypted)

	// Sign then verify
	signature, err := Sign(original, kp.PrivateKey, "")
	require.NoError(t, err)
	err = Verify(original, signature, kp.PublicKey)
	assert.NoError(t, err)
}
