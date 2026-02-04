package signing

import (
	"context"
	"fmt"
	"runtime"

	"github.com/host-uk/core/pkg/io"
)

// Artifact represents a build output that can be signed.
// This mirrors build.Artifact to avoid import cycles.
type Artifact struct {
	Path string
	OS   string
	Arch string
}

// SignBinaries signs macOS binaries in the artifacts list.
// Only signs darwin binaries when running on macOS with a configured identity.
func SignBinaries(ctx context.Context, fs io.Medium, cfg SignConfig, artifacts []Artifact) error {
	if !cfg.Enabled {
		return nil
	}

	// Only sign on macOS
	if runtime.GOOS != "darwin" {
		return nil
	}

	signer := NewMacOSSigner(cfg.MacOS)
	if !signer.Available() {
		return nil // Silently skip if not configured
	}

	for _, artifact := range artifacts {
		if artifact.OS != "darwin" {
			continue
		}

		fmt.Printf("  Signing %s...\n", artifact.Path)
		if err := signer.Sign(ctx, fs, artifact.Path); err != nil {
			return fmt.Errorf("failed to sign %s: %w", artifact.Path, err)
		}
	}

	return nil
}

// NotarizeBinaries notarizes macOS binaries if enabled.
func NotarizeBinaries(ctx context.Context, fs io.Medium, cfg SignConfig, artifacts []Artifact) error {
	if !cfg.Enabled || !cfg.MacOS.Notarize {
		return nil
	}

	if runtime.GOOS != "darwin" {
		return nil
	}

	signer := NewMacOSSigner(cfg.MacOS)
	if !signer.Available() {
		return fmt.Errorf("notarization requested but codesign not available")
	}

	for _, artifact := range artifacts {
		if artifact.OS != "darwin" {
			continue
		}

		fmt.Printf("  Notarizing %s (this may take a few minutes)...\n", artifact.Path)
		if err := signer.Notarize(ctx, fs, artifact.Path); err != nil {
			return fmt.Errorf("failed to notarize %s: %w", artifact.Path, err)
		}
	}

	return nil
}

// SignChecksums signs the checksums file with GPG.
func SignChecksums(ctx context.Context, fs io.Medium, cfg SignConfig, checksumFile string) error {
	if !cfg.Enabled {
		return nil
	}

	signer := NewGPGSigner(cfg.GPG.Key)
	if !signer.Available() {
		return nil // Silently skip if not configured
	}

	fmt.Printf("  Signing %s with GPG...\n", checksumFile)
	if err := signer.Sign(ctx, fs, checksumFile); err != nil {
		return fmt.Errorf("failed to sign checksums: %w", err)
	}

	return nil
}
