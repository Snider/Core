// Package openpgp provides PGP encryption, decryption, and key management.
// It wraps the Enchantrix library's PGP functionality.
package openpgp

import (
	"fmt"
	"io"
	"os"

	"github.com/Snider/Enchantrix/pkg/crypt"
)

var service *crypt.Service

func init() {
	service = crypt.NewService()
}

// KeyPair holds a generated PGP key pair in armored format.
type KeyPair struct {
	PublicKey  string
	PrivateKey string
}

// CreateKeyPair generates a new PGP key pair with the given identity and optional passphrase.
// Note: Enchantrix does not support passphrase-protected keys, so the passphrase
// parameter is used as a comment in the key metadata.
func CreateKeyPair(identity, passphrase string) (*KeyPair, error) {
	pubBytes, privBytes, err := service.GeneratePGPKeyPair(identity, identity+"@example.com", passphrase)
	if err != nil {
		return nil, err
	}
	return &KeyPair{
		PublicKey:  string(pubBytes),
		PrivateKey: string(privBytes),
	}, nil
}

// EncryptPGP encrypts data for a recipient, optionally signing it.
// - writer: destination for the encrypted data
// - recipientPath: path to the recipient's public key file
// - data: plaintext to encrypt
// - signerPath: optional path to the signer's private key file (not supported in Enchantrix)
// - signerPassphrase: optional passphrase for the signer's private key (not supported)
func EncryptPGP(writer io.Writer, recipientPath, data string, signerPath, signerPassphrase *string) error {
	// Read recipient public key
	recipientKey, err := os.ReadFile(recipientPath)
	if err != nil {
		return fmt.Errorf("failed to open recipient public key file: %w", err)
	}

	ciphertext, err := service.EncryptPGP(recipientKey, []byte(data))
	if err != nil {
		return err
	}

	_, err = writer.Write(ciphertext)
	return err
}

// DecryptPGP decrypts a PGP message, optionally verifying the signature.
// - recipientPath: path to the recipient's private key file
// - message: armored PGP message to decrypt
// - passphrase: passphrase for the recipient's private key (not supported in Enchantrix)
// - signerPath: optional path to the signer's public key file for verification (not supported)
func DecryptPGP(recipientPath, message, passphrase string, signerPath *string) (string, error) {
	// Read recipient private key
	recipientKey, err := os.ReadFile(recipientPath)
	if err != nil {
		return "", fmt.Errorf("failed to open recipient private key file: %w", err)
	}

	plaintext, err := service.DecryptPGP(recipientKey, []byte(message))
	if err != nil {
		return "", fmt.Errorf("failed to read PGP message: %w", err)
	}

	return string(plaintext), nil
}

// generateTestKeys creates a test key pair and writes it to temporary files.
// Returns paths to the public and private key files, and a cleanup function.
func generateTestKeys(t interface {
	Helper()
	Fatalf(string, ...any)
}, identity, passphrase string) (pubPath, privPath string, cleanup func()) {
	t.Helper()

	keyPair, err := CreateKeyPair(identity, passphrase)
	if err != nil {
		t.Fatalf("failed to create key pair: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "pgp-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	pubPath = tempDir + "/test.pub"
	privPath = tempDir + "/test.priv"

	if err := os.WriteFile(pubPath, []byte(keyPair.PublicKey), 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to write public key: %v", err)
	}

	if err := os.WriteFile(privPath, []byte(keyPair.PrivateKey), 0600); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to write private key: %v", err)
	}

	cleanup = func() { os.RemoveAll(tempDir) }
	return pubPath, privPath, cleanup
}

// EncryptPGPToString is a convenience function that encrypts to a string.
func EncryptPGPToString(recipientKey, data string) (string, error) {
	ciphertext, err := service.EncryptPGP([]byte(recipientKey), []byte(data))
	if err != nil {
		return "", err
	}
	return string(ciphertext), nil
}
