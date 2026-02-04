package crypt

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHMACSHA256_Good(t *testing.T) {
	// RFC 4231 Test Case 2
	key := []byte("Jefe")
	message := []byte("what do ya want for nothing?")
	expected := "5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843"

	mac := HMACSHA256(message, key)
	assert.Equal(t, expected, hex.EncodeToString(mac))
}

func TestVerifyHMAC_Good(t *testing.T) {
	key := []byte("secret-key")
	message := []byte("test message")

	mac := HMACSHA256(message, key)

	valid := VerifyHMAC(message, key, mac, sha256.New)
	assert.True(t, valid)
}

func TestVerifyHMAC_Bad(t *testing.T) {
	key := []byte("secret-key")
	message := []byte("test message")
	tampered := []byte("tampered message")

	mac := HMACSHA256(message, key)

	valid := VerifyHMAC(tampered, key, mac, sha256.New)
	assert.False(t, valid)
}
