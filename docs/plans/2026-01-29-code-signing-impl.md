# Code Signing Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add GPG checksums signing and macOS codesign/notarization to the build pipeline.

**Architecture:** `pkg/build/signing/` package with Signer interface. GPG signs CHECKSUMS.txt. macOS codesign runs after binary compilation, before archiving. Config in `.core/build.yaml` with env var fallbacks.

**Tech Stack:** Go, os/exec for gpg/codesign/xcrun CLI tools

---

### Task 1: Create Signing Package Structure

**Files:**
- Create: `pkg/build/signing/signer.go`

**Step 1: Create signer.go with interface and config types**

```go
// Package signing provides code signing for build artifacts.
package signing

import (
	"context"
	"os"
	"strings"
)

// Signer defines the interface for code signing implementations.
type Signer interface {
	// Name returns the signer's identifier.
	Name() string
	// Available checks if this signer can be used.
	Available() bool
	// Sign signs the artifact at the given path.
	Sign(ctx context.Context, path string) error
}

// SignConfig holds signing configuration from .core/build.yaml.
type SignConfig struct {
	Enabled bool        `yaml:"enabled"`
	GPG     GPGConfig   `yaml:"gpg,omitempty"`
	MacOS   MacOSConfig `yaml:"macos,omitempty"`
	Windows WindowsConfig `yaml:"windows,omitempty"`
}

// GPGConfig holds GPG signing configuration.
type GPGConfig struct {
	Key string `yaml:"key"` // Key ID or fingerprint, supports $ENV
}

// MacOSConfig holds macOS codesign configuration.
type MacOSConfig struct {
	Identity    string `yaml:"identity"`      // Developer ID Application: ...
	Notarize    bool   `yaml:"notarize"`      // Submit to Apple for notarization
	AppleID     string `yaml:"apple_id"`      // Apple account email
	TeamID      string `yaml:"team_id"`       // Team ID
	AppPassword string `yaml:"app_password"`  // App-specific password
}

// WindowsConfig holds Windows signtool configuration (placeholder).
type WindowsConfig struct {
	Certificate string `yaml:"certificate"` // Path to .pfx
	Password    string `yaml:"password"`    // Certificate password
}

// DefaultSignConfig returns sensible defaults.
func DefaultSignConfig() SignConfig {
	return SignConfig{
		Enabled: true,
		GPG: GPGConfig{
			Key: os.Getenv("GPG_KEY_ID"),
		},
		MacOS: MacOSConfig{
			Identity:    os.Getenv("CODESIGN_IDENTITY"),
			AppleID:     os.Getenv("APPLE_ID"),
			TeamID:      os.Getenv("APPLE_TEAM_ID"),
			AppPassword: os.Getenv("APPLE_APP_PASSWORD"),
		},
	}
}

// ExpandEnv expands environment variables in config values.
func (c *SignConfig) ExpandEnv() {
	c.GPG.Key = expandEnv(c.GPG.Key)
	c.MacOS.Identity = expandEnv(c.MacOS.Identity)
	c.MacOS.AppleID = expandEnv(c.MacOS.AppleID)
	c.MacOS.TeamID = expandEnv(c.MacOS.TeamID)
	c.MacOS.AppPassword = expandEnv(c.MacOS.AppPassword)
	c.Windows.Certificate = expandEnv(c.Windows.Certificate)
	c.Windows.Password = expandEnv(c.Windows.Password)
}

// expandEnv expands $VAR or ${VAR} in a string.
func expandEnv(s string) string {
	if strings.HasPrefix(s, "$") {
		return os.ExpandEnv(s)
	}
	return s
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/build/signing/...`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/build/signing/signer.go
git commit -m "feat(signing): add Signer interface and config types

Defines interface for GPG, macOS, and Windows signing.
Config supports env var expansion for secrets.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 2: Implement GPG Signer

**Files:**
- Create: `pkg/build/signing/gpg.go`
- Create: `pkg/build/signing/gpg_test.go`

**Step 1: Write the failing test**

```go
package signing

import (
	"testing"
)

func TestGPGSigner_Good_Name(t *testing.T) {
	s := NewGPGSigner("ABCD1234")
	if s.Name() != "gpg" {
		t.Errorf("expected name 'gpg', got %q", s.Name())
	}
}

func TestGPGSigner_Good_Available(t *testing.T) {
	s := NewGPGSigner("ABCD1234")
	// Available depends on gpg being installed
	_ = s.Available()
}

func TestGPGSigner_Bad_NoKey(t *testing.T) {
	s := NewGPGSigner("")
	if s.Available() {
		t.Error("expected Available() to be false when key is empty")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/build/signing/... -run TestGPGSigner -v`
Expected: FAIL (NewGPGSigner not defined)

**Step 3: Write implementation**

```go
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
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/build/signing/... -run TestGPGSigner -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/build/signing/gpg.go pkg/build/signing/gpg_test.go
git commit -m "feat(signing): add GPG signer

Signs files with detached ASCII-armored signatures (.asc).

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 3: Implement macOS Codesign

**Files:**
- Create: `pkg/build/signing/codesign.go`
- Create: `pkg/build/signing/codesign_test.go`

**Step 1: Write the failing test**

```go
package signing

import (
	"runtime"
	"testing"
)

func TestMacOSSigner_Good_Name(t *testing.T) {
	s := NewMacOSSigner(MacOSConfig{Identity: "Developer ID Application: Test"})
	if s.Name() != "codesign" {
		t.Errorf("expected name 'codesign', got %q", s.Name())
	}
}

func TestMacOSSigner_Good_Available(t *testing.T) {
	s := NewMacOSSigner(MacOSConfig{Identity: "Developer ID Application: Test"})

	// Only available on macOS with identity set
	if runtime.GOOS == "darwin" {
		// May or may not be available depending on Xcode
		_ = s.Available()
	} else {
		if s.Available() {
			t.Error("expected Available() to be false on non-macOS")
		}
	}
}

func TestMacOSSigner_Bad_NoIdentity(t *testing.T) {
	s := NewMacOSSigner(MacOSConfig{})
	if s.Available() {
		t.Error("expected Available() to be false when identity is empty")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/snider/Code/Core && go test ./pkg/build/signing/... -run TestMacOSSigner -v`
Expected: FAIL (NewMacOSSigner not defined)

**Step 3: Write implementation**

```go
package signing

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// MacOSSigner signs binaries using macOS codesign.
type MacOSSigner struct {
	config MacOSConfig
}

// Compile-time interface check.
var _ Signer = (*MacOSSigner)(nil)

// NewMacOSSigner creates a new macOS signer.
func NewMacOSSigner(cfg MacOSConfig) *MacOSSigner {
	return &MacOSSigner{config: cfg}
}

// Name returns "codesign".
func (s *MacOSSigner) Name() string {
	return "codesign"
}

// Available checks if running on macOS with codesign and identity configured.
func (s *MacOSSigner) Available() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	if s.config.Identity == "" {
		return false
	}
	_, err := exec.LookPath("codesign")
	return err == nil
}

// Sign codesigns a binary with hardened runtime.
func (s *MacOSSigner) Sign(ctx context.Context, binary string) error {
	if !s.Available() {
		return fmt.Errorf("codesign.Sign: codesign not available")
	}

	cmd := exec.CommandContext(ctx, "codesign",
		"--sign", s.config.Identity,
		"--timestamp",
		"--options", "runtime", // Hardened runtime for notarization
		"--force",
		binary,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("codesign.Sign: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Notarize submits binary to Apple for notarization and staples the ticket.
// This blocks until Apple responds (typically 1-5 minutes).
func (s *MacOSSigner) Notarize(ctx context.Context, binary string) error {
	if s.config.AppleID == "" || s.config.TeamID == "" || s.config.AppPassword == "" {
		return fmt.Errorf("codesign.Notarize: missing Apple credentials (apple_id, team_id, app_password)")
	}

	// Create ZIP for submission
	zipPath := binary + ".zip"
	zipCmd := exec.CommandContext(ctx, "zip", "-j", zipPath, binary)
	if output, err := zipCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("codesign.Notarize: failed to create zip: %w\nOutput: %s", err, string(output))
	}
	defer os.Remove(zipPath)

	// Submit to Apple and wait
	submitCmd := exec.CommandContext(ctx, "xcrun", "notarytool", "submit",
		zipPath,
		"--apple-id", s.config.AppleID,
		"--team-id", s.config.TeamID,
		"--password", s.config.AppPassword,
		"--wait",
	)
	if output, err := submitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("codesign.Notarize: notarization failed: %w\nOutput: %s", err, string(output))
	}

	// Staple the ticket
	stapleCmd := exec.CommandContext(ctx, "xcrun", "stapler", "staple", binary)
	if output, err := stapleCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("codesign.Notarize: failed to staple: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// ShouldNotarize returns true if notarization is enabled.
func (s *MacOSSigner) ShouldNotarize() bool {
	return s.config.Notarize
}
```

**Step 4: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/build/signing/... -run TestMacOSSigner -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/build/signing/codesign.go pkg/build/signing/codesign_test.go
git commit -m "feat(signing): add macOS codesign + notarization

Signs binaries with Developer ID and hardened runtime.
Notarization submits to Apple and staples ticket.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 4: Add Windows Placeholder

**Files:**
- Create: `pkg/build/signing/signtool.go`

**Step 1: Create placeholder implementation**

```go
package signing

import (
	"context"
)

// WindowsSigner signs binaries using Windows signtool (placeholder).
type WindowsSigner struct {
	config WindowsConfig
}

// Compile-time interface check.
var _ Signer = (*WindowsSigner)(nil)

// NewWindowsSigner creates a new Windows signer.
func NewWindowsSigner(cfg WindowsConfig) *WindowsSigner {
	return &WindowsSigner{config: cfg}
}

// Name returns "signtool".
func (s *WindowsSigner) Name() string {
	return "signtool"
}

// Available returns false (not yet implemented).
func (s *WindowsSigner) Available() bool {
	return false
}

// Sign is a placeholder that does nothing.
func (s *WindowsSigner) Sign(ctx context.Context, binary string) error {
	// TODO: Implement Windows signing
	return nil
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/build/signing/...`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/build/signing/signtool.go
git commit -m "feat(signing): add Windows signtool placeholder

Placeholder for future Windows code signing support.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 5: Add SignConfig to BuildConfig

**Files:**
- Modify: `pkg/build/config.go`
- Modify: `pkg/build/config_test.go`

**Step 1: Add Sign field to BuildConfig**

In `pkg/build/config.go`, add to the `BuildConfig` struct:

```go
// Add import
import "github.com/host-uk/core/pkg/build/signing"

// Add to BuildConfig struct after Targets field:
	// Sign contains code signing configuration.
	Sign signing.SignConfig `yaml:"sign,omitempty"`
```

**Step 2: Update DefaultConfig**

In `DefaultConfig()`, add:

```go
	Sign: signing.DefaultSignConfig(),
```

**Step 3: Update applyDefaults**

In `applyDefaults()`, add:

```go
	// Expand environment variables in sign config
	cfg.Sign.ExpandEnv()
```

**Step 4: Add test for sign config loading**

Add to `pkg/build/config_test.go`:

```go
func TestLoadConfig_Good_SignConfig(t *testing.T) {
	tmpDir := t.TempDir()
	coreDir := filepath.Join(tmpDir, ".core")
	os.MkdirAll(coreDir, 0755)

	configContent := `version: 1
sign:
  enabled: true
  gpg:
    key: "ABCD1234"
  macos:
    identity: "Developer ID Application: Test"
    notarize: true
`
	os.WriteFile(filepath.Join(coreDir, "build.yaml"), []byte(configContent), 0644)

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Sign.Enabled {
		t.Error("expected Sign.Enabled to be true")
	}
	if cfg.Sign.GPG.Key != "ABCD1234" {
		t.Errorf("expected GPG.Key 'ABCD1234', got %q", cfg.Sign.GPG.Key)
	}
	if cfg.Sign.MacOS.Identity != "Developer ID Application: Test" {
		t.Errorf("expected MacOS.Identity, got %q", cfg.Sign.MacOS.Identity)
	}
	if !cfg.Sign.MacOS.Notarize {
		t.Error("expected MacOS.Notarize to be true")
	}
}
```

**Step 5: Run tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/build/... -run TestLoadConfig -v`
Expected: PASS

**Step 6: Commit**

```bash
git add pkg/build/config.go pkg/build/config_test.go
git commit -m "feat(build): add SignConfig to BuildConfig

Loads signing configuration from .core/build.yaml.
Expands environment variables for secrets.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 6: Create Sign Helper Functions

**Files:**
- Create: `pkg/build/signing/sign.go`

**Step 1: Create orchestration helpers**

```go
package signing

import (
	"context"
	"fmt"
	"runtime"

	"github.com/host-uk/core/pkg/build"
)

// SignBinaries signs macOS binaries in the artifacts list.
// Only signs darwin binaries when running on macOS with a configured identity.
func SignBinaries(ctx context.Context, cfg SignConfig, artifacts []build.Artifact) error {
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
		if err := signer.Sign(ctx, artifact.Path); err != nil {
			return fmt.Errorf("failed to sign %s: %w", artifact.Path, err)
		}
	}

	return nil
}

// NotarizeBinaries notarizes macOS binaries if enabled.
func NotarizeBinaries(ctx context.Context, cfg SignConfig, artifacts []build.Artifact) error {
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
		if err := signer.Notarize(ctx, artifact.Path); err != nil {
			return fmt.Errorf("failed to notarize %s: %w", artifact.Path, err)
		}
	}

	return nil
}

// SignChecksums signs the checksums file with GPG.
func SignChecksums(ctx context.Context, cfg SignConfig, checksumFile string) error {
	if !cfg.Enabled {
		return nil
	}

	signer := NewGPGSigner(cfg.GPG.Key)
	if !signer.Available() {
		return nil // Silently skip if not configured
	}

	fmt.Printf("  Signing %s with GPG...\n", checksumFile)
	if err := signer.Sign(ctx, checksumFile); err != nil {
		return fmt.Errorf("failed to sign checksums: %w", err)
	}

	return nil
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./pkg/build/signing/...`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/build/signing/sign.go
git commit -m "feat(signing): add orchestration helpers

SignBinaries, NotarizeBinaries, SignChecksums for pipeline integration.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 7: Integrate Signing into CLI

**Files:**
- Modify: `cmd/core/cmd/build.go`

**Step 1: Add --no-sign and --notarize flags**

After the existing flag declarations (around line 74), add:

```go
	var noSign bool
	var notarize bool

	buildCmd.BoolFlag("no-sign", "Skip all code signing", &noSign)
	buildCmd.BoolFlag("notarize", "Enable macOS notarization (requires Apple credentials)", &notarize)
```

**Step 2: Update runProjectBuild signature**

Update the function signature and call:

```go
// Update function signature:
func runProjectBuild(buildType string, ciMode bool, targetsFlag string, outputDir string, doArchive bool, doChecksum bool, configPath string, format string, push bool, imageName string, noSign bool, notarize bool) error {

// Update the Action call:
buildCmd.Action(func() error {
	return runProjectBuild(buildType, ciMode, targets, outputDir, doArchive, doChecksum, configPath, format, push, imageName, noSign, notarize)
})
```

**Step 3: Add signing import**

Add to imports:

```go
	"github.com/host-uk/core/pkg/build/signing"
```

**Step 4: Add signing after build, before archive**

After the build succeeds (around line 228), add:

```go
	// Sign macOS binaries if enabled
	signCfg := buildCfg.Sign
	if notarize {
		signCfg.MacOS.Notarize = true
	}
	if noSign {
		signCfg.Enabled = false
	}

	if signCfg.Enabled && runtime.GOOS == "darwin" {
		if !ciMode {
			fmt.Println()
			fmt.Printf("%s Signing binaries...\n", buildHeaderStyle.Render("Sign:"))
		}

		if err := signing.SignBinaries(ctx, signCfg, artifacts); err != nil {
			if !ciMode {
				fmt.Printf("%s Signing failed: %v\n", buildErrorStyle.Render("Error:"), err)
			}
			return err
		}

		if signCfg.MacOS.Notarize {
			if err := signing.NotarizeBinaries(ctx, signCfg, artifacts); err != nil {
				if !ciMode {
					fmt.Printf("%s Notarization failed: %v\n", buildErrorStyle.Render("Error:"), err)
				}
				return err
			}
		}
	}
```

**Step 5: Add GPG signing after checksums**

After WriteChecksumFile (around line 297), add:

```go
		// Sign checksums with GPG
		if signCfg.Enabled {
			if err := signing.SignChecksums(ctx, signCfg, checksumPath); err != nil {
				if !ciMode {
					fmt.Printf("%s GPG signing failed: %v\n", buildErrorStyle.Render("Error:"), err)
				}
				return err
			}
		}
```

**Step 6: Verify it compiles**

Run: `cd /Users/snider/Code/Core && go build ./cmd/core/...`
Expected: No errors

**Step 7: Commit**

```bash
git add cmd/core/cmd/build.go
git commit -m "feat(cli): integrate signing into build command

Adds --no-sign and --notarize flags.
Signs macOS binaries after build, GPG signs checksums.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 8: Add Integration Test

**Files:**
- Create: `pkg/build/signing/signing_test.go`

**Step 1: Create integration test**

```go
package signing

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/host-uk/core/pkg/build"
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
	artifacts := []build.Artifact{
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

	artifacts := []build.Artifact{
		{Path: "/tmp/test-binary", OS: "darwin", Arch: "arm64"},
	}

	err := SignBinaries(ctx, cfg, artifacts)
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
```

**Step 2: Run all signing tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/build/signing/... -v`
Expected: All tests pass

**Step 3: Commit**

```bash
git add pkg/build/signing/signing_test.go
git commit -m "test(signing): add integration tests

Tests for skip conditions and disabled configs.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

### Task 9: Update TODO.md and Final Verification

**Step 1: Build CLI**

Run: `cd /Users/snider/Code/Core && go build -o bin/core ./cmd/core`
Expected: No errors

**Step 2: Test help output**

Run: `./bin/core build --help`
Expected: Shows --no-sign and --notarize flags

**Step 3: Run all tests**

Run: `cd /Users/snider/Code/Core && go test ./pkg/build/... -v`
Expected: All tests pass

**Step 4: Update TODO.md**

Mark S3.3 tasks as complete in `tasks/TODO.md`:

```markdown
### S3.3 Code Signing (Standard) ✅
- [x] macOS codesign integration
- [x] macOS notarization
- [ ] Windows signtool integration (placeholder added)
- [x] GPG signing (standard tools)
```

**Step 5: Final commit**

```bash
git add tasks/TODO.md
git commit -m "chore(signing): finalize S3.3 code signing

Implemented:
- GPG signing of CHECKSUMS.txt
- macOS codesign with hardened runtime
- macOS notarization via notarytool
- Windows signtool placeholder

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Summary

9 tasks covering:
1. Signing package structure (Signer interface, SignConfig)
2. GPG signer implementation
3. macOS codesign + notarization
4. Windows signtool placeholder
5. Add SignConfig to BuildConfig
6. Orchestration helpers (SignBinaries, SignChecksums)
7. CLI integration (--no-sign, --notarize)
8. Integration tests
9. Final verification and TODO update
