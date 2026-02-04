package crypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveKey_Good(t *testing.T) {
	passphrase := []byte("test-passphrase")
	salt := []byte("1234567890123456") // 16 bytes

	key1 := DeriveKey(passphrase, salt, 32)
	key2 := DeriveKey(passphrase, salt, 32)

	assert.Len(t, key1, 32)
	assert.Equal(t, key1, key2, "same inputs should produce same output")

	// Different passphrase should produce different key
	key3 := DeriveKey([]byte("different-passphrase"), salt, 32)
	assert.NotEqual(t, key1, key3)
}

func TestDeriveKeyScrypt_Good(t *testing.T) {
	passphrase := []byte("test-passphrase")
	salt := []byte("1234567890123456")

	key, err := DeriveKeyScrypt(passphrase, salt, 32)
	assert.NoError(t, err)
	assert.Len(t, key, 32)

	// Deterministic
	key2, err := DeriveKeyScrypt(passphrase, salt, 32)
	assert.NoError(t, err)
	assert.Equal(t, key, key2)
}

func TestHKDF_Good(t *testing.T) {
	secret := []byte("input-keying-material")
	salt := []byte("optional-salt")
	info := []byte("context-info")

	key1, err := HKDF(secret, salt, info, 32)
	assert.NoError(t, err)
	assert.Len(t, key1, 32)

	// Deterministic
	key2, err := HKDF(secret, salt, info, 32)
	assert.NoError(t, err)
	assert.Equal(t, key1, key2)

	// Different info should produce different key
	key3, err := HKDF(secret, salt, []byte("different-info"), 32)
	assert.NoError(t, err)
	assert.NotEqual(t, key1, key3)
}
