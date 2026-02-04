#!/bin/bash
# Core CLI unified installer
# Served via *.core.help with BunnyCDN edge transformation
#
# Usage:
#   curl -fsSL setup.core.help | bash       # Interactive setup (default)
#   curl -fsSL ci.core.help | bash          # CI/CD (minimal, fast)
#   curl -fsSL dev.core.help | bash         # Full development
#   curl -fsSL go.core.help | bash          # Go development variant
#   curl -fsSL php.core.help | bash         # PHP/Laravel variant
#   curl -fsSL agent.core.help | bash       # AI agent variant
#
# Version override:
#   curl -fsSL setup.core.help | bash -s -- v1.0.0
#
set -eo pipefail

# === BunnyCDN Edge Variables (transformed at edge based on subdomain) ===
MODE="{{CORE_MODE}}"           # setup, ci, dev, variant
VARIANT="{{CORE_VARIANT}}"     # go, php, agent (when MODE=variant)

# === User overrides (fallback for local testing) ===
[[ "$MODE" == "{{CORE_MODE}}" ]] && MODE="${CORE_MODE:-setup}"
[[ "$VARIANT" == "{{CORE_VARIANT}}" ]] && VARIANT="${CORE_VARIANT:-}"

# === Configuration ===
VERSION="${1:-latest}"
REPO="host-uk/core"
BINARY="core"

# === Colours ===
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
DIM='\033[2m'
BOLD='\033[1m'
NC='\033[0m'

info()    { echo -e "${BLUE}>>>${NC} $1"; }
success() { echo -e "${GREEN}>>>${NC} $1"; }
error()   { echo -e "${RED}>>>${NC} $1" >&2; exit 1; }
dim()     { echo -e "${DIM}$1${NC}"; }

# === Platform Detection ===
detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
  esac

  case "$OS" in
    darwin|linux) ;;
    *) error "Unsupported OS: $OS (use Windows installer for Windows)" ;;
  esac
}

# === Version Resolution ===
resolve_version() {
  if [ "$VERSION" = "latest" ]; then
    info "Fetching latest version..."
    VERSION=$(curl -fsSL --max-time 10 "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
      error "Failed to fetch latest version from GitHub API"
    fi
  fi
}

# === Download Helpers ===
url_exists() {
  curl -fsSLI "$1" 2>/dev/null | grep -qE "HTTP/.* [23][0-9][0-9]"
}

find_archive() {
  local base="$1"
  local variant="$2"

  # Build candidate list (prefer xz over gz, variant over full)
  local candidates=()
  if [ -n "$variant" ]; then
    candidates+=("${base}-${variant}-${OS}-${ARCH}.tar.xz")
    candidates+=("${base}-${variant}-${OS}-${ARCH}.tar.gz")
  fi
  candidates+=("${base}-${OS}-${ARCH}.tar.xz")
  candidates+=("${base}-${OS}-${ARCH}.tar.gz")

  for archive in "${candidates[@]}"; do
    local url="https://github.com/${REPO}/releases/download/${VERSION}/${archive}"
    if url_exists "$url"; then
      ARCHIVE="$archive"
      DOWNLOAD_URL="$url"
      return 0
    fi
  done

  error "No compatible archive found for ${OS}/${ARCH}"
}

download_and_extract() {
  WORK_DIR=$(mktemp -d)
  trap 'rm -rf "$WORK_DIR"' EXIT

  info "Downloading ${ARCHIVE}..."
  if ! curl -fsSL --max-time 120 "$DOWNLOAD_URL" -o "$WORK_DIR/$ARCHIVE"; then
    error "Failed to download ${DOWNLOAD_URL}"
  fi

  info "Extracting..."
  case "$ARCHIVE" in
    *.tar.xz) tar -xJf "$WORK_DIR/$ARCHIVE" -C "$WORK_DIR" || error "Failed to extract archive" ;;
    *.tar.gz) tar -xzf "$WORK_DIR/$ARCHIVE" -C "$WORK_DIR" || error "Failed to extract archive" ;;
    *) error "Unknown archive format: $ARCHIVE" ;;
  esac
}

install_binary() {
  local install_dir="${1:-/usr/local/bin}"

  info "Installing to ${install_dir}..."
  chmod +x "$WORK_DIR/${BINARY}"
  if [ -w "$install_dir" ]; then
    mv "$WORK_DIR/${BINARY}" "${install_dir}/${BINARY}"
  else
    sudo mv "$WORK_DIR/${BINARY}" "${install_dir}/${BINARY}"
  fi
}

verify_install() {
  if command -v "$BINARY" &>/dev/null; then
    success "Installed successfully!"
    dim "$($BINARY --version)"
  else
    success "Installed to ${1:-/usr/local/bin}/${BINARY}"
    dim "Add the directory to your PATH if not already present"
  fi
}

# === Installation Modes ===

install_setup() {
  echo -e "${BOLD}Core CLI Installer${NC}"
  echo ""

  detect_platform
  resolve_version

  local install_dir="/usr/local/bin"
  info "Installing ${BINARY} ${VERSION} for ${OS}/${ARCH}..."
  find_archive "$BINARY" ""
  download_and_extract
  install_binary "$install_dir"
  verify_install "$install_dir"
}

install_ci() {
  detect_platform
  resolve_version

  echo "Installing ${BINARY} ${VERSION} (${OS}/${ARCH})..."
  find_archive "$BINARY" ""
  download_and_extract

  # CI: prefer /usr/local/bin, no sudo prompts
  chmod +x "$WORK_DIR/${BINARY}"
  if [ -w /usr/local/bin ]; then
    mv "$WORK_DIR/${BINARY}" /usr/local/bin/
  else
    sudo mv "$WORK_DIR/${BINARY}" /usr/local/bin/
  fi

  /usr/local/bin/${BINARY} --version
}

install_dev() {
  detect_platform
  resolve_version

  local install_dir="/usr/local/bin"
  info "Installing ${BINARY} ${VERSION} (full) for ${OS}/${ARCH}..."
  find_archive "$BINARY" ""
  download_and_extract
  install_binary "$install_dir"
  verify_install "$install_dir"

  echo ""
  echo "Full development variant installed. Available commands:"
  echo "  core dev     - Multi-repo workflows"
  echo "  core build   - Cross-platform builds"
  echo "  core release - Build and publish releases"
}

install_variant() {
  local variant="$1"

  detect_platform
  resolve_version

  local install_dir="/usr/local/bin"
  info "Installing ${BINARY} ${VERSION} (${variant} variant) for ${OS}/${ARCH}..."
  find_archive "$BINARY" "$variant"

  if [[ "$ARCHIVE" == "${BINARY}-${OS}-${ARCH}"* ]]; then
    dim "Using full variant (${variant} variant not available for ${VERSION})"
  fi

  download_and_extract
  install_binary "$install_dir"
  verify_install "$install_dir"
}

# === Main ===
case "$MODE" in
  setup)   install_setup ;;
  ci)      install_ci ;;
  dev)     install_dev ;;
  variant)
    [ -z "$VARIANT" ] && error "VARIANT must be specified when MODE=variant"
    install_variant "$VARIANT"
    ;;
  *)       error "Unknown mode: $MODE" ;;
esac
