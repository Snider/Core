package crypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt_Good(t *testing.T) {
	plaintext := []byte("hello, world!")
	passphrase := []byte("correct-horse-battery-staple")

	encrypted, err := Encrypt(plaintext, passphrase)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := Decrypt(encrypted, passphrase)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecrypt_Bad(t *testing.T) {
	plaintext := []byte("secret data")
	passphrase := []byte("correct-passphrase")
	wrongPassphrase := []byte("wrong-passphrase")

	encrypted, err := Encrypt(plaintext, passphrase)
	assert.NoError(t, err)

	_, err = Decrypt(encrypted, wrongPassphrase)
	assert.Error(t, err)
}

func TestEncryptDecryptAES_Good(t *testing.T) {
	plaintext := []byte("hello, AES world!")
	passphrase := []byte("my-secure-passphrase")

	encrypted, err := EncryptAES(plaintext, passphrase)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := DecryptAES(encrypted, passphrase)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
