package signing

import (
	"context"
	"runtime"
	"testing"
)

func TestSignBinaries_Good_SkipsNonDarwin(t *testing.T) {
	ctx := context.Background()
	cfg := SignConfig{
		Enabled: true,
		MacOS: MacOSConfig{
			Identity: "Developer ID Application: Test",
		},
	}

	// Create fake artifact for linux
	artifacts := []Artifact{
		{Path: "/tmp/test-binary", OS: "linux", Arch: "amd64"},
	}

	// Should not error even though binary doesn't exist (skips non-darwin)
	err := SignBinaries(ctx, cfg, artifacts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSignBinaries_Good_DisabledConfig(t *testing.T) {
	ctx := context.Background()
	cfg := SignConfig{
		Enabled: false,
	}

	artifacts := []Artifact{
		{Path: "/tmp/test-binary", OS: "darwin", Arch: "arm64"},
	}

	err := SignBinaries(ctx, cfg, artifacts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSignBinaries_Good_SkipsOnNonMacOS(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping on macOS - this tests non-macOS behavior")
	}

	ctx := context.Background()
	cfg := SignConfig{
		Enabled: true,
		MacOS: MacOSConfig{
			Identity: "Developer ID Application: Test",
		},
	}

	artifacts := []Artifact{
		{Path: "/tmp/test-binary", OS: "darwin", Arch: "arm64"},
	}

	err := SignBinaries(ctx, cfg, artifacts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNotarizeBinaries_Good_DisabledConfig(t *testing.T) {
	ctx := context.Background()
	cfg := SignConfig{
		Enabled: false,
	}

	artifacts := []Artifact{
		{Path: "/tmp/test-binary", OS: "darwin", Arch: "arm64"},
	}

	err := NotarizeBinaries(ctx, cfg, artifacts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNotarizeBinaries_Good_NotarizeDisabled(t *testing.T) {
	ctx := context.Background()
	cfg := SignConfig{
		Enabled: true,
		MacOS: MacOSConfig{
			Notarize: false,
		},
	}

	artifacts := []Artifact{
		{Path: "/tmp/test-binary", OS: "darwin", Arch: "arm64"},
	}

	err := NotarizeBinaries(ctx, cfg, artifacts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSignChecksums_Good_SkipsNoKey(t *testing.T) {
	ctx := context.Background()
	cfg := SignConfig{
		Enabled: true,
		GPG: GPGConfig{
			Key: "", // No key configured
		},
	}

	// Should silently skip when no key
	err := SignChecksums(ctx, cfg, "/tmp/CHECKSUMS.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSignChecksums_Good_Disabled(t *testing.T) {
	ctx := context.Background()
	cfg := SignConfig{
		Enabled: false,
	}

	err := SignChecksums(ctx, cfg, "/tmp/CHECKSUMS.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
