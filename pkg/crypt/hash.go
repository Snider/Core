package crypt

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	core "github.com/host-uk/core/pkg/framework/core"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using Argon2id with default parameters.
// Returns a string in the format: $argon2id$v=19$m=65536,t=3,p=4$<base64salt>$<base64hash>
func HashPassword(password string) (string, error) {
	salt, err := generateSalt(argon2SaltLen)
	if err != nil {
		return "", core.E("crypt.HashPassword", "failed to generate salt", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Parallelism, argon2KeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Parallelism,
		b64Salt, b64Hash)

	return encoded, nil
}

// VerifyPassword verifies a password against an Argon2id hash string.
// The hash must be in the format produced by HashPassword.
func VerifyPassword(password, hash string) (bool, error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, core.E("crypt.VerifyPassword", "invalid hash format", nil)
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, core.E("crypt.VerifyPassword", "failed to parse version", err)
	}

	var memory uint32
	var time uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &parallelism); err != nil {
		return false, core.E("crypt.VerifyPassword", "failed to parse parameters", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, core.E("crypt.VerifyPassword", "failed to decode salt", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, core.E("crypt.VerifyPassword", "failed to decode hash", err)
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, parallelism, uint32(len(expectedHash)))

	return subtle.ConstantTimeCompare(computedHash, expectedHash) == 1, nil
}

// HashBcrypt hashes a password using bcrypt with the given cost.
// Cost must be between bcrypt.MinCost and bcrypt.MaxCost.
func HashBcrypt(password string, cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", core.E("crypt.HashBcrypt", "failed to hash password", err)
	}
	return string(hash), nil
}

// VerifyBcrypt verifies a password against a bcrypt hash.
func VerifyBcrypt(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	if err != nil {
		return false, core.E("crypt.VerifyBcrypt", "failed to verify password", err)
	}
	return true, nil
}
