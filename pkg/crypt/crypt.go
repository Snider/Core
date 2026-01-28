// Package crypt provides cryptographic functions to the Core application.
// It wraps the Enchantrix library, providing a Core-compatible service layer
// for hashing, checksums, RSA, and PGP operations.
package crypt

import (
	"fmt"
	"io"

	"github.com/host-uk/core/pkg/core"
	"github.com/Snider/Enchantrix/pkg/crypt"
)

// HandleIPCEvents processes IPC messages for the crypt service.
func (s *Service) HandleIPCEvents(c *core.Core, msg core.Message) error {
	switch msg.(type) {
	case core.ActionServiceStartup:
		// Crypt is stateless, no startup needed.
		return nil
	default:
		if c.App != nil && c.App.Logger != nil {
			c.App.Logger.Debug("Crypt: Unhandled message type", "type", fmt.Sprintf("%T", msg))
		}
	}
	return nil
}

// Options holds configuration for the crypt service.
type Options struct{}

// Service provides cryptographic functions to the application.
// It delegates to Enchantrix for all cryptographic operations.
type Service struct {
	*core.ServiceRuntime[Options]
	enchantrix *crypt.Service
}

// HashType defines the supported hashing algorithms.
// Re-exported from Enchantrix for convenience.
type HashType = crypt.HashType

// Hash type constants re-exported from Enchantrix.
const (
	LTHN   HashType = crypt.LTHN
	SHA512 HashType = crypt.SHA512
	SHA256 HashType = crypt.SHA256
	SHA1   HashType = crypt.SHA1
	MD5    HashType = crypt.MD5
)

// newCryptService contains the common logic for initializing a Service struct.
func newCryptService() (*Service, error) {
	return &Service{
		enchantrix: crypt.NewService(),
	}, nil
}

// New is the constructor for static dependency injection.
// It creates a Service instance without initializing the core.ServiceRuntime field.
func New() (*Service, error) {
	return newCryptService()
}

// Register is the constructor for dynamic dependency injection (used with core.WithService).
// It creates a Service instance and initializes its core.ServiceRuntime field.
func Register(c *core.Core) (any, error) {
	s, err := newCryptService()
	if err != nil {
		return nil, err
	}
	s.ServiceRuntime = core.NewServiceRuntime(c, Options{})
	return s, nil
}

// --- Hashing ---

// Hash computes a hash of the payload using the specified algorithm.
func (s *Service) Hash(lib HashType, payload string) string {
	return s.enchantrix.Hash(lib, payload)
}

// IsHashAlgo checks if the given string is a valid hash algorithm.
func (s *Service) IsHashAlgo(algo string) bool {
	return s.enchantrix.IsHashAlgo(algo)
}

// --- Checksums ---

// Luhn validates a number using the Luhn algorithm.
func (s *Service) Luhn(payload string) bool {
	return s.enchantrix.Luhn(payload)
}

// Fletcher16 computes the Fletcher-16 checksum.
func (s *Service) Fletcher16(payload string) uint16 {
	return s.enchantrix.Fletcher16(payload)
}

// Fletcher32 computes the Fletcher-32 checksum.
func (s *Service) Fletcher32(payload string) uint32 {
	return s.enchantrix.Fletcher32(payload)
}

// Fletcher64 computes the Fletcher-64 checksum.
func (s *Service) Fletcher64(payload string) uint64 {
	return s.enchantrix.Fletcher64(payload)
}

// --- RSA ---

// GenerateRSAKeyPair generates an RSA key pair with the specified bit size.
// Returns PEM-encoded public and private keys.
func (s *Service) GenerateRSAKeyPair(bits int) (publicKey, privateKey string, err error) {
	pubBytes, privBytes, err := s.enchantrix.GenerateRSAKeyPair(bits)
	if err != nil {
		return "", "", err
	}
	return string(pubBytes), string(privBytes), nil
}

// EncryptRSA encrypts data using an RSA public key.
// Takes PEM-encoded public key and returns base64-encoded ciphertext.
func (s *Service) EncryptRSA(publicKeyPEM, plaintext string) (string, error) {
	ciphertext, err := s.enchantrix.EncryptRSA([]byte(publicKeyPEM), []byte(plaintext), nil)
	if err != nil {
		return "", err
	}
	return string(ciphertext), nil
}

// DecryptRSA decrypts data using an RSA private key.
// Takes PEM-encoded private key and ciphertext.
func (s *Service) DecryptRSA(privateKeyPEM, ciphertext string) (string, error) {
	plaintext, err := s.enchantrix.DecryptRSA([]byte(privateKeyPEM), []byte(ciphertext), nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// --- PGP ---

// GeneratePGPKeyPair generates a PGP key pair.
// Note: Enchantrix PGP keys are not passphrase-protected. The comment parameter
// is used instead of passphrase for key metadata.
func (s *Service) GeneratePGPKeyPair(name, email, comment string) (publicKey, privateKey string, err error) {
	pubBytes, privBytes, err := s.enchantrix.GeneratePGPKeyPair(name, email, comment)
	if err != nil {
		return "", "", err
	}
	return string(pubBytes), string(privBytes), nil
}

// EncryptPGP encrypts data for a recipient and writes to the provided writer.
func (s *Service) EncryptPGP(writer io.Writer, recipientPublicKey, data string) error {
	ciphertext, err := s.enchantrix.EncryptPGP([]byte(recipientPublicKey), []byte(data))
	if err != nil {
		return err
	}
	_, err = writer.Write(ciphertext)
	return err
}

// EncryptPGPToString encrypts data for a recipient and returns the ciphertext.
func (s *Service) EncryptPGPToString(recipientPublicKey, data string) (string, error) {
	ciphertext, err := s.enchantrix.EncryptPGP([]byte(recipientPublicKey), []byte(data))
	if err != nil {
		return "", err
	}
	return string(ciphertext), nil
}

// DecryptPGP decrypts a PGP message.
// Note: Enchantrix does not support passphrase-protected keys for decryption.
func (s *Service) DecryptPGP(privateKey, message string) (string, error) {
	plaintext, err := s.enchantrix.DecryptPGP([]byte(privateKey), []byte(message))
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// SignPGP signs data with a PGP private key.
func (s *Service) SignPGP(privateKey, data string) (string, error) {
	signature, err := s.enchantrix.SignPGP([]byte(privateKey), []byte(data))
	if err != nil {
		return "", err
	}
	return string(signature), nil
}

// VerifyPGP verifies a PGP signature.
func (s *Service) VerifyPGP(publicKey, data, signature string) error {
	return s.enchantrix.VerifyPGP([]byte(publicKey), []byte(data), []byte(signature))
}

// SymmetricallyEncryptPGP encrypts data using a passphrase and writes to the provided writer.
func (s *Service) SymmetricallyEncryptPGP(writer io.Writer, data, passphrase string) error {
	ciphertext, err := s.enchantrix.SymmetricallyEncryptPGP([]byte(passphrase), []byte(data))
	if err != nil {
		return err
	}
	_, err = writer.Write(ciphertext)
	return err
}
