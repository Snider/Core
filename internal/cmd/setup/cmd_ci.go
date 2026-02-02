package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/host-uk/core/pkg/cli"
	coreio "github.com/host-uk/core/pkg/io"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// CIConfig holds CI setup configuration from .core/ci.yaml
type CIConfig struct {
	// Homebrew tap (e.g., "host-uk/tap")
	Tap string `yaml:"tap"`
	// Formula name (defaults to "core")
	Formula string `yaml:"formula"`
	// Scoop bucket URL
	ScoopBucket string `yaml:"scoop_bucket"`
	// Chocolatey package name
	ChocolateyPkg string `yaml:"chocolatey_pkg"`
	// GitHub repository for direct downloads
	Repository string `yaml:"repository"`
	// Default version to install
	DefaultVersion string `yaml:"default_version"`
}

// DefaultCIConfig returns the default CI configuration.
func DefaultCIConfig() *CIConfig {
	return &CIConfig{
		Tap:            "host-uk/tap",
		Formula:        "core",
		ScoopBucket:    "https://github.com/host-uk/scoop-bucket.git",
		ChocolateyPkg:  "core-cli",
		Repository:     "host-uk/core",
		DefaultVersion: "dev",
	}
}

// LoadCIConfig loads CI configuration from .core/ci.yaml
func LoadCIConfig() *CIConfig {
	cfg := DefaultCIConfig()

	// Try to find .core/ci.yaml in current directory or parents
	dir, err := os.Getwd()
	if err != nil {
		return cfg
	}

	for {
		configPath := filepath.Join(dir, ".core", "ci.yaml")
		data, err := coreio.Local.Read(configPath)
		if err == nil {
			if err := yaml.Unmarshal([]byte(data), cfg); err == nil {
				return cfg
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return cfg
}

// CI setup command flags
var (
	ciShell   string
	ciVersion string
)

func init() {
	ciCmd := &cobra.Command{
		Use:   "ci",
		Short: "Output CI installation commands for core CLI",
		Long: `Output installation commands for the core CLI in CI environments.

Generates shell commands to install the core CLI using the appropriate
package manager for each platform:

  macOS/Linux: Homebrew (brew install host-uk/tap/core)
  Windows:     Scoop or Chocolatey, or direct download

Configuration can be customized via .core/ci.yaml:

  tap: host-uk/tap           # Homebrew tap
  formula: core              # Homebrew formula name
  scoop_bucket: https://...  # Scoop bucket URL
  chocolatey_pkg: core-cli   # Chocolatey package name
  repository: host-uk/core   # GitHub repo for direct downloads
  default_version: dev       # Default version to install

Examples:
  # Output installation commands for current platform
  core setup ci

  # Output for specific shell (bash, powershell, yaml)
  core setup ci --shell=bash
  core setup ci --shell=powershell
  core setup ci --shell=yaml

  # Install specific version
  core setup ci --version=v1.0.0

  # Use in GitHub Actions (pipe to shell)
  eval "$(core setup ci --shell=bash)"`,
		RunE: runSetupCI,
	}

	ciCmd.Flags().StringVar(&ciShell, "shell", "", "Output format: bash, powershell, yaml (auto-detected if not specified)")
	ciCmd.Flags().StringVar(&ciVersion, "version", "", "Version to install (tag name or 'dev' for latest dev build)")

	setupCmd.AddCommand(ciCmd)
}

func runSetupCI(cmd *cobra.Command, args []string) error {
	cfg := LoadCIConfig()

	// Use flag version or config default
	version := ciVersion
	if version == "" {
		version = cfg.DefaultVersion
	}

	// Auto-detect shell if not specified
	shell := ciShell
	if shell == "" {
		if runtime.GOOS == "windows" {
			shell = "powershell"
		} else {
			shell = "bash"
		}
	}

	switch shell {
	case "bash", "sh":
		return outputBashInstall(cfg, version)
	case "powershell", "pwsh", "ps1":
		return outputPowershellInstall(cfg, version)
	case "yaml", "yml", "gha", "github":
		return outputGitHubActionsYAML(cfg, version)
	default:
		return cli.Err("unsupported shell: %s (use bash, powershell, or yaml)", shell)
	}
}

func outputBashInstall(cfg *CIConfig, version string) error {
	script := fmt.Sprintf(`#!/bin/bash
set -e

VERSION="%s"
REPO="%s"
TAP="%s"
FORMULA="%s"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Try Homebrew first on macOS/Linux
if command -v brew &>/dev/null; then
  echo "Installing via Homebrew..."
  brew tap "$TAP" 2>/dev/null || true
  if [ "$VERSION" = "dev" ]; then
    brew install "${TAP}/${FORMULA}" --HEAD 2>/dev/null || brew upgrade "${TAP}/${FORMULA}" --fetch-HEAD 2>/dev/null || brew install "${TAP}/${FORMULA}"
  else
    brew install "${TAP}/${FORMULA}"
  fi
  %s --version
  exit 0
fi

# Fall back to direct download
echo "Installing %s CLI ${VERSION} for ${OS}/${ARCH}..."

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/%s-${OS}-${ARCH}"

# Download binary
curl -fsSL "$DOWNLOAD_URL" -o /tmp/%s
chmod +x /tmp/%s

# Install to /usr/local/bin (requires sudo on most systems)
if [ -w /usr/local/bin ]; then
  mv /tmp/%s /usr/local/bin/%s
else
  sudo mv /tmp/%s /usr/local/bin/%s
fi

echo "Installed:"
%s --version
`, version, cfg.Repository, cfg.Tap, cfg.Formula,
		cfg.Formula, cfg.Formula, cfg.Formula,
		cfg.Formula, cfg.Formula, cfg.Formula, cfg.Formula, cfg.Formula, cfg.Formula, cfg.Formula)

	fmt.Print(script)
	return nil
}

func outputPowershellInstall(cfg *CIConfig, version string) error {
	script := fmt.Sprintf(`# PowerShell installation script for %s CLI
$ErrorActionPreference = "Stop"

$Version = "%s"
$Repo = "%s"
$ScoopBucket = "%s"
$ChocoPkg = "%s"
$BinaryName = "%s"
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

# Try Scoop first
if (Get-Command scoop -ErrorAction SilentlyContinue) {
    Write-Host "Installing via Scoop..."
    scoop bucket add host-uk $ScoopBucket 2>$null
    scoop install "host-uk/$BinaryName"
    & $BinaryName --version
    exit 0
}

# Try Chocolatey
if (Get-Command choco -ErrorAction SilentlyContinue) {
    Write-Host "Installing via Chocolatey..."
    choco install $ChocoPkg -y
    & $BinaryName --version
    exit 0
}

# Fall back to direct download
Write-Host "Installing $BinaryName CLI $Version for windows/$Arch..."

$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$BinaryName-windows-$Arch.exe"
$InstallDir = "$env:LOCALAPPDATA\Programs\$BinaryName"
$BinaryPath = "$InstallDir\$BinaryName.exe"

# Create install directory
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# Download binary
Invoke-WebRequest -Uri $DownloadUrl -OutFile $BinaryPath

# Add to PATH if not already there
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($CurrentPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
}

Write-Host "Installed:"
& $BinaryPath --version
`, cfg.Formula, version, cfg.Repository, cfg.ScoopBucket, cfg.ChocolateyPkg, cfg.Formula)

	fmt.Print(script)
	return nil
}

func outputGitHubActionsYAML(cfg *CIConfig, version string) error {
	yaml := fmt.Sprintf(`# GitHub Actions steps to install %s CLI
# Add these to your workflow file

# Option 1: Direct download (fastest, no extra dependencies)
- name: Install %s CLI
  shell: bash
  run: |
    VERSION="%s"
    REPO="%s"
    BINARY="%s"
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"
    case "$ARCH" in
      x86_64|amd64) ARCH="amd64" ;;
      arm64|aarch64) ARCH="arm64" ;;
    esac
    curl -fsSL "https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}-${OS}-${ARCH}" -o "${BINARY}"
    chmod +x "${BINARY}"
    sudo mv "${BINARY}" /usr/local/bin/
    %s --version

# Option 2: Homebrew (better for caching, includes dependencies)
- name: Install %s CLI (Homebrew)
  run: |
    brew tap %s
    brew install %s/%s
    %s --version
`, cfg.Formula, cfg.Formula, version, cfg.Repository, cfg.Formula, cfg.Formula,
		cfg.Formula, cfg.Tap, cfg.Tap, cfg.Formula, cfg.Formula)

	fmt.Print(yaml)
	return nil
}
