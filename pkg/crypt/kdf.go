// Package crypt provides cryptographic utilities including encryption,
// hashing, key derivation, HMAC, and checksum functions.
package crypt

import (
	"crypto/rand"
	"crypto/sha256"
	"io"

	core "github.com/host-uk/core/pkg/framework/core"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/scrypt"
)

// Argon2id default parameters.
const (
	argon2Memory      = 64 * 1024 // 64 MB
	argon2Time        = 3
	argon2Parallelism = 4
	argon2KeyLen      = 32
	argon2SaltLen     = 16
)

// DeriveKey derives a key from a passphrase using Argon2id with default parameters.
// The salt must be argon2SaltLen bytes. keyLen specifies the desired key length.
func DeriveKey(passphrase, salt []byte, keyLen uint32) []byte {
	return argon2.IDKey(passphrase, salt, argon2Time, argon2Memory, argon2Parallelism, keyLen)
}

// DeriveKeyScrypt derives a key from a passphrase using scrypt.
// Uses recommended parameters: N=32768, r=8, p=1.
func DeriveKeyScrypt(passphrase, salt []byte, keyLen int) ([]byte, error) {
	key, err := scrypt.Key(passphrase, salt, 32768, 8, 1, keyLen)
	if err != nil {
		return nil, core.E("crypt.DeriveKeyScrypt", "failed to derive key", err)
	}
	return key, nil
}

// HKDF derives a key using HKDF-SHA256.
// secret is the input keying material, salt is optional (can be nil),
// info is optional context, and keyLen is the desired output length.
func HKDF(secret, salt, info []byte, keyLen int) ([]byte, error) {
	reader := hkdf.New(sha256.New, secret, salt, info)
	key := make([]byte, keyLen)
	if _, err := io.ReadFull(reader, key); err != nil {
		return nil, core.E("crypt.HKDF", "failed to derive key", err)
	}
	return key, nil
}

// generateSalt creates a random salt of the given length.
func generateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	if _, err := rand.Read(salt); err != nil {
		return nil, core.E("crypt.generateSalt", "failed to generate random salt", err)
	}
	return salt, nil
}
