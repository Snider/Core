// Package rsa provides RSA key generation, encryption, and decryption
// using OAEP with SHA-256.
//
// Ported from Enchantrix (github.com/Snider/Enchantrix/pkg/crypt/std/rsa).
package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// KeyPair holds PEM-encoded RSA public and private keys.
type KeyPair struct {
	PublicKey  string
	PrivateKey string
}

// GenerateKeyPair creates a new RSA key pair of the given bit size.
// The minimum accepted key size is 2048 bits.
// Returns a KeyPair with PEM-encoded public and private keys.
func GenerateKeyPair(bits int) (*KeyPair, error) {
	if bits < 2048 {
		return nil, fmt.Errorf("rsa: key size too small: %d (minimum 2048)", bits)
	}

	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("rsa: failed to generate private key: %w", err)
	}

	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("rsa: failed to marshal public key: %w", err)
	}
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return &KeyPair{
		PublicKey:  string(pubKeyPEM),
		PrivateKey: string(privKeyPEM),
	}, nil
}

// Encrypt encrypts data with the given PEM-encoded public key using RSA-OAEP
// with SHA-256.
func Encrypt(data []byte, publicKeyPEM string) ([]byte, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("rsa: failed to decode public key PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("rsa: failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("rsa: not an RSA public key")
	}

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, data, nil)
	if err != nil {
		return nil, fmt.Errorf("rsa: failed to encrypt data: %w", err)
	}

	return ciphertext, nil
}

// Decrypt decrypts data with the given PEM-encoded private key using RSA-OAEP
// with SHA-256.
func Decrypt(data []byte, privateKeyPEM string) ([]byte, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("rsa: failed to decode private key PEM")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("rsa: failed to parse private key: %w", err)
	}

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, data, nil)
	if err != nil {
		return nil, fmt.Errorf("rsa: failed to decrypt data: %w", err)
	}

	return plaintext, nil
}
