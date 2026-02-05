// Package chachapoly provides XChaCha20-Poly1305 authenticated encryption.
//
// Encrypt prepends a random nonce to the ciphertext; Decrypt extracts it.
// The key must be 32 bytes (256 bits).
//
// Ported from Enchantrix (github.com/Snider/Enchantrix/pkg/crypt/std/chachapoly).
package chachapoly

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

// Encrypt encrypts plaintext using XChaCha20-Poly1305.
// The key must be exactly 32 bytes. A random 24-byte nonce is generated
// and prepended to the returned ciphertext.
func Encrypt(plaintext, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("chachapoly: failed to create AEAD: %w", err)
	}

	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(plaintext)+aead.Overhead())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("chachapoly: failed to generate nonce: %w", err)
	}

	return aead.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext produced by Encrypt using XChaCha20-Poly1305.
// The key must be exactly 32 bytes. The nonce is extracted from the first
// 24 bytes of the ciphertext.
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("chachapoly: failed to create AEAD: %w", err)
	}

	minLen := aead.NonceSize() + aead.Overhead()
	if len(ciphertext) < minLen {
		return nil, fmt.Errorf("chachapoly: ciphertext too short: got %d bytes, need at least %d bytes", len(ciphertext), minLen)
	}

	nonce, ciphertext := ciphertext[:aead.NonceSize()], ciphertext[aead.NonceSize():]

	decrypted, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("chachapoly: decryption failed: %w", err)
	}

	if len(decrypted) == 0 {
		return []byte{}, nil
	}

	return decrypted, nil
}
