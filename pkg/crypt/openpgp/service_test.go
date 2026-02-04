package openpgp

import (
	"bytes"
	"testing"

	core "github.com/host-uk/core/pkg/framework/core"
	"github.com/stretchr/testify/assert"
)

func TestCreateKeyPair(t *testing.T) {
	c, _ := core.New()
	s := &Service{core: c}

	privKey, err := s.CreateKeyPair("test user", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, privKey)
	assert.Contains(t, privKey, "-----BEGIN PGP PRIVATE KEY BLOCK-----")
}

func TestEncryptDecrypt(t *testing.T) {
	c, _ := core.New()
	s := &Service{core: c}

	passphrase := "secret"
	privKey, err := s.CreateKeyPair("test user", passphrase)
	assert.NoError(t, err)

	// In this simple test, the public key is also in the armored private key string
	// (openpgp.ReadArmoredKeyRing reads both)
	publicKey := privKey

	data := "hello openpgp"
	var buf bytes.Buffer
	armored, err := s.EncryptPGP(&buf, publicKey, data)
	assert.NoError(t, err)
	assert.NotEmpty(t, armored)
	assert.NotEmpty(t, buf.String())

	decrypted, err := s.DecryptPGP(privKey, armored, passphrase)
	assert.NoError(t, err)
	assert.Equal(t, data, decrypted)
}
