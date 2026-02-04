package crypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_Good(t *testing.T) {
	password := "my-secure-password"

	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "$argon2id$")

	match, err := VerifyPassword(password, hash)
	assert.NoError(t, err)
	assert.True(t, match)
}

func TestVerifyPassword_Bad(t *testing.T) {
	password := "my-secure-password"
	wrongPassword := "wrong-password"

	hash, err := HashPassword(password)
	assert.NoError(t, err)

	match, err := VerifyPassword(wrongPassword, hash)
	assert.NoError(t, err)
	assert.False(t, match)
}

func TestHashBcrypt_Good(t *testing.T) {
	password := "bcrypt-test-password"

	hash, err := HashBcrypt(password, bcrypt.DefaultCost)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	match, err := VerifyBcrypt(password, hash)
	assert.NoError(t, err)
	assert.True(t, match)

	// Wrong password should not match
	match, err = VerifyBcrypt("wrong-password", hash)
	assert.NoError(t, err)
	assert.False(t, match)
}
