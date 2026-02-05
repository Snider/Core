package signing

import (
	"context"
	"fmt"
	"os/exec"
)

// GPGSigner signs files using GPG.
type GPGSigner struct {
	KeyID string
}

// Compile-time interface check.
var _ Signer = (*GPGSigner)(nil)

// NewGPGSigner creates a new GPG signer.
func NewGPGSigner(keyID string) *GPGSigner {
	return &GPGSigner{KeyID: keyID}
}

// Name returns "gpg".
func (s *GPGSigner) Name() string {
	return "gpg"
}

// Available checks if gpg is installed and key is configured.
func (s *GPGSigner) Available() bool {
	if s.KeyID == "" {
		return false
	}
	_, err := exec.LookPath("gpg")
	return err == nil
}

// Sign creates a detached ASCII-armored signature.
// For file.txt, creates file.txt.asc
func (s *GPGSigner) Sign(ctx context.Context, file string) error {
	if !s.Available() {
		return fmt.Errorf("gpg.Sign: gpg not available or key not configured")
	}

	cmd := exec.CommandContext(ctx, "gpg",
		"--detach-sign",
		"--armor",
		"--local-user", s.KeyID,
		"--output", file+".asc",
		file,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gpg.Sign: %w\nOutput: %s", err, string(output))
	}

	return nil
}
