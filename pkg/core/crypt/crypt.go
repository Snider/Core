package crypt

import (
	"context"
	"fmt"
	"io"

	"github.com/Snider/Core"
	"github.com/Snider/Core/crypt/openpgp"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Options holds configuration for the crypt service.
type Options struct{}

// Service provides cryptographic functions to the application.
type Service struct {
	*core.Runtime[Options]
}

// HashType defines the supported hashing algorithms.
type HashType string

const (
	LTHN   HashType = "lthn"
	SHA512 HashType = "sha512"
	SHA256 HashType = "sha256"
	SHA1   HashType = "sha1"
	MD5    HashType = "md5"
)

// Service provides cryptographic functions.
// It is the main entry point for all cryptographic operations
// and is bound to the frontend.

type API struct {
	core *core.Core
}

// New is a factory function that creates a new crypt Service.
func New(c *core.Core) (any, error) {
	s := &Service{
		Runtime: core.NewRuntime(c, Options{}),
	}
	return s, nil
}

// handleIPCEvents is the central IPC handler for the crypt service.
func (s *Service) handleIPCEvents(c *core.Core, msg core.Message) error {
	switch msg.(type) {
	case core.ActionServiceStartup:
		return s.ServiceStartup(context.Background(), application.ServiceOptions{})
	default:
		c.App.Logger.Error("Crypt: Unknown message type", "type", fmt.Sprintf("%T", msg))
	}
	return nil
}

// ServiceStartup is called when the app starts. It handles one-time cryptographic setup.
func (s *Service) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.Core().App.Logger.Info("Crypt service started")
	// Key generation logic will be implemented here, likely depending on the config service.
	return nil
}

// EncryptPGP encrypts data for a recipient, optionally signing it.
// It acts as a wrapper around the underlying openpgp library function.
func (s *Service) EncryptPGP(writer io.Writer, recipientPath, data string, signerPath, signerPassphrase *string) (string, error) {
	return openpgp.EncryptPGP(writer, recipientPath, data, signerPath, signerPassphrase)
}

// DecryptPGP decrypts a PGP message, optionally verifying the signature.
// It acts as a wrapper around the underlying openpgp library function.
func (s *Service) DecryptPGP(recipientPath, message, passphrase string, signerPath *string) (string, error) {
	return openpgp.DecryptPGP(recipientPath, message, passphrase, signerPath)
}
