package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	core "github.com/host-uk/core/pkg/framework/core"
	"golang.org/x/crypto/chacha20poly1305"
)

// ChaCha20Encrypt encrypts plaintext using ChaCha20-Poly1305.
// The key must be 32 bytes. The nonce is randomly generated and prepended
// to the ciphertext.
func ChaCha20Encrypt(plaintext, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, core.E("crypt.ChaCha20Encrypt", "failed to create cipher", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, core.E("crypt.ChaCha20Encrypt", "failed to generate nonce", err)
	}

	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// ChaCha20Decrypt decrypts ciphertext encrypted with ChaCha20Encrypt.
// The key must be 32 bytes. Expects the nonce prepended to the ciphertext.
func ChaCha20Decrypt(ciphertext, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, core.E("crypt.ChaCha20Decrypt", "failed to create cipher", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, core.E("crypt.ChaCha20Decrypt", "ciphertext too short", nil)
	}

	nonce, encrypted := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, core.E("crypt.ChaCha20Decrypt", "failed to decrypt", err)
	}

	return plaintext, nil
}

// AESGCMEncrypt encrypts plaintext using AES-256-GCM.
// The key must be 32 bytes. The nonce is randomly generated and prepended
// to the ciphertext.
func AESGCMEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, core.E("crypt.AESGCMEncrypt", "failed to create cipher", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, core.E("crypt.AESGCMEncrypt", "failed to create GCM", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, core.E("crypt.AESGCMEncrypt", "failed to generate nonce", err)
	}

	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// AESGCMDecrypt decrypts ciphertext encrypted with AESGCMEncrypt.
// The key must be 32 bytes. Expects the nonce prepended to the ciphertext.
func AESGCMDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, core.E("crypt.AESGCMDecrypt", "failed to create cipher", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, core.E("crypt.AESGCMDecrypt", "failed to create GCM", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, core.E("crypt.AESGCMDecrypt", "ciphertext too short", nil)
	}

	nonce, encrypted := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, core.E("crypt.AESGCMDecrypt", "failed to decrypt", err)
	}

	return plaintext, nil
}
