# Code Signing Design (S3.3)

## Summary

Integrate standard code signing tools into the build pipeline. GPG signs checksums by default. macOS codesign + notarization for Apple binaries. Windows signtool deferred.

## Design Decisions

- **Sign during build**: Signing happens in `pkg/build/signing/` after compilation, before archiving
- **Config location**: `.core/build.yaml` with environment variable fallbacks for secrets
- **GPG scope**: Signs `checksums.txt` only (standard pattern like Go, Terraform)
- **macOS flow**: Codesign always when identity configured, notarize optional with flag/config
- **Windows**: Placeholder for later implementation

## Package Structure

```
pkg/build/signing/
├── signer.go      # Signer interface + SignConfig
├── gpg.go         # GPG checksums signing
├── codesign.go    # macOS codesign + notarize
└── signtool.go    # Windows placeholder
```

## Signer Interface

```go
// pkg/build/signing/signer.go
type Signer interface {
    Name() string
    Available() bool
    Sign(ctx context.Context, artifact string) error
}

type SignConfig struct {
    Enabled  bool        `yaml:"enabled"`
    GPG      GPGConfig   `yaml:"gpg,omitempty"`
    MacOS    MacOSConfig `yaml:"macos,omitempty"`
    Windows  WindowsConfig `yaml:"windows,omitempty"`
}

type GPGConfig struct {
    Key string `yaml:"key"` // Key ID or fingerprint, supports $ENV
}

type MacOSConfig struct {
    Identity    string `yaml:"identity"`     // Developer ID Application: ...
    Notarize    bool   `yaml:"notarize"`     // Submit to Apple
    AppleID     string `yaml:"apple_id"`     // Apple account email
    TeamID      string `yaml:"team_id"`      // Team ID
    AppPassword string `yaml:"app_password"` // App-specific password
}

type WindowsConfig struct {
    Certificate string `yaml:"certificate"` // Path to .pfx
    Password    string `yaml:"password"`    // Certificate password
}
```

## Config Schema

In `.core/build.yaml`:

```yaml
sign:
  enabled: true

  gpg:
    key: $GPG_KEY_ID

  macos:
    identity: "Developer ID Application: Your Name (TEAM_ID)"
    notarize: false
    apple_id: $APPLE_ID
    team_id: $APPLE_TEAM_ID
    app_password: $APPLE_APP_PASSWORD

  # windows: (deferred)
  #   certificate: $WINDOWS_CERT_PATH
  #   password: $WINDOWS_CERT_PASSWORD
```

## Build Pipeline Integration

```
Build() in pkg/build/builders/go.go
    ↓
compile binaries
    ↓
Sign macOS binaries (codesign)     ← NEW
    ↓
Notarize if enabled (wait)         ← NEW
    ↓
Create archives (tar.gz, zip)
    ↓
Generate checksums.txt
    ↓
GPG sign checksums.txt             ← NEW
    ↓
Return artifacts
```

## GPG Signer

```go
// pkg/build/signing/gpg.go
type GPGSigner struct {
    KeyID string
}

func (s *GPGSigner) Name() string { return "gpg" }

func (s *GPGSigner) Available() bool {
    _, err := exec.LookPath("gpg")
    return err == nil && s.KeyID != ""
}

func (s *GPGSigner) Sign(ctx context.Context, file string) error {
    cmd := exec.CommandContext(ctx, "gpg",
        "--detach-sign",
        "--armor",
        "--local-user", s.KeyID,
        "--output", file+".asc",
        file,
    )
    return cmd.Run()
}
```

**Output:** `checksums.txt.asc` (ASCII armored detached signature)

**User verification:**
```bash
gpg --verify checksums.txt.asc checksums.txt
sha256sum -c checksums.txt
```

## macOS Codesign

```go
// pkg/build/signing/codesign.go
type MacOSSigner struct {
    Identity    string
    Notarize    bool
    AppleID     string
    TeamID      string
    AppPassword string
}

func (s *MacOSSigner) Name() string { return "codesign" }

func (s *MacOSSigner) Available() bool {
    if runtime.GOOS != "darwin" {
        return false
    }
    _, err := exec.LookPath("codesign")
    return err == nil && s.Identity != ""
}

func (s *MacOSSigner) Sign(ctx context.Context, binary string) error {
    cmd := exec.CommandContext(ctx, "codesign",
        "--sign", s.Identity,
        "--timestamp",
        "--options", "runtime",
        "--force",
        binary,
    )
    return cmd.Run()
}

func (s *MacOSSigner) NotarizeAndStaple(ctx context.Context, binary string) error {
    // 1. Create ZIP for submission
    zipPath := binary + ".zip"
    exec.CommandContext(ctx, "zip", "-j", zipPath, binary).Run()
    defer os.Remove(zipPath)

    // 2. Submit and wait
    cmd := exec.CommandContext(ctx, "xcrun", "notarytool", "submit",
        zipPath,
        "--apple-id", s.AppleID,
        "--team-id", s.TeamID,
        "--password", s.AppPassword,
        "--wait",
    )
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("notarization failed: %w", err)
    }

    // 3. Staple ticket
    return exec.CommandContext(ctx, "xcrun", "stapler", "staple", binary).Run()
}
```

## CLI Flags

```bash
core build                    # Sign with defaults (GPG + codesign if configured)
core build --no-sign          # Skip all signing
core build --notarize         # Enable macOS notarization (overrides config)
```

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `GPG_KEY_ID` | GPG key ID or fingerprint |
| `CODESIGN_IDENTITY` | macOS Developer ID (fallback) |
| `APPLE_ID` | Apple account email |
| `APPLE_TEAM_ID` | Apple Developer Team ID |
| `APPLE_APP_PASSWORD` | App-specific password for notarization |

## Deferred

- **Windows signtool**: Placeholder implementation returning nil
- **Sigstore/keyless signing**: Future consideration
- **Binary-level GPG signatures**: Only checksums.txt signed

## Implementation Steps

1. Create `pkg/build/signing/` package structure
2. Implement Signer interface and SignConfig
3. Implement GPGSigner
4. Implement MacOSSigner with codesign
5. Add notarization support to MacOSSigner
6. Add SignConfig to build.Config
7. Integrate signing into build pipeline
8. Add CLI flags (--no-sign, --notarize)
9. Add Windows placeholder
10. Tests with mocked exec

## Dependencies

- `gpg` CLI (system)
- `codesign` CLI (macOS Xcode Command Line Tools)
- `xcrun notarytool` (macOS Xcode Command Line Tools)
- `xcrun stapler` (macOS Xcode Command Line Tools)
