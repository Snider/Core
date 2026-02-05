#!/bin/bash
# setup-ubuntu.sh - Turn a fresh Ubuntu into a native tool building machine
#
# Installs and configures:
# - System dependencies (WebKitGTK, build tools)
# - Development tools (Go, Node.js, Git, gh CLI)
# - Claude Code CLI
# - core-ide with system tray integration
#
# Usage:
#   curl -fsSL https://host.uk.com/setup-ubuntu | bash
#   # or
#   ./scripts/setup-ubuntu.sh
#
# Environment variables (optional):
#   SKIP_GUI=1        - Skip GUI components (headless server)
#   SKIP_CLAUDE=1     - Skip Claude Code installation
#   GITHUB_TOKEN=xxx  - Pre-configure GitHub token

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if running as root
if [[ $EUID -eq 0 ]]; then
    log_error "Don't run this script as root. It will use sudo when needed."
    exit 1
fi

# Check Ubuntu version
if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    if [[ "$ID" != "ubuntu" ]] && [[ "$ID_LIKE" != *"ubuntu"* ]]; then
        log_warn "This script is designed for Ubuntu. Your distro: $ID"
        read -p "Continue anyway? [y/N] " -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Yy]$ ]] && exit 1
    fi
fi

log_info "Setting up Ubuntu as a native tool building machine..."

# ============================================================================
# Step 1: System Dependencies
# ============================================================================
log_info "Installing system dependencies..."

sudo apt-get update

# Build essentials
sudo apt-get install -y \
    build-essential \
    curl \
    wget \
    git \
    jq \
    unzip

# GUI dependencies (skip for headless)
if [[ -z "${SKIP_GUI:-}" ]]; then
    log_info "Installing GUI dependencies (WebKitGTK, GTK3)..."

    # Check Ubuntu version for correct WebKitGTK package
    UBUNTU_VERSION=$(lsb_release -rs 2>/dev/null || echo "22.04")

    # WebKitGTK 4.1 for Ubuntu 22.04+, 4.0 for older
    if dpkg --compare-versions "$UBUNTU_VERSION" "ge" "22.04"; then
        WEBKIT_PKG="libwebkit2gtk-4.1-dev"
    else
        WEBKIT_PKG="libwebkit2gtk-4.0-dev"
    fi

    sudo apt-get install -y \
        libgtk-3-dev \
        "$WEBKIT_PKG" \
        libappindicator3-dev \
        gir1.2-appindicator3-0.1

    log_success "GUI dependencies installed"
else
    log_info "Skipping GUI dependencies (SKIP_GUI=1)"
fi

log_success "System dependencies installed"

# ============================================================================
# Step 2: Go
# ============================================================================
GO_VERSION="1.25.6"

if command -v go &>/dev/null && [[ "$(go version 2>/dev/null | grep -oP 'go\d+\.\d+' | head -1)" == "go1.25" ]]; then
    log_success "Go $GO_VERSION already installed"
else
    log_info "Installing Go $GO_VERSION..."

    ARCH=$(dpkg --print-architecture)
    case $ARCH in
        amd64) GO_ARCH="amd64" ;;
        arm64) GO_ARCH="arm64" ;;
        *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac

    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" -o /tmp/go.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz

    # Add to path
    if ! grep -q '/usr/local/go/bin' ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
    fi
    export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

    log_success "Go $GO_VERSION installed"
fi

# ============================================================================
# Step 3: Node.js (via fnm for version management)
# ============================================================================
NODE_VERSION="22"

if command -v node &>/dev/null && [[ "$(node -v 2>/dev/null | cut -d. -f1)" == "v${NODE_VERSION}" ]]; then
    log_success "Node.js $NODE_VERSION already installed"
else
    log_info "Installing Node.js $NODE_VERSION via fnm..."

    # Install fnm
    if ! command -v fnm &>/dev/null; then
        curl -fsSL https://fnm.vercel.app/install | bash -s -- --skip-shell
        export PATH="$HOME/.local/share/fnm:$PATH"
        eval "$(fnm env)"
    fi

    # Install Node.js
    fnm install $NODE_VERSION
    fnm use $NODE_VERSION
    fnm default $NODE_VERSION

    # Add fnm to bashrc
    if ! grep -q 'fnm env' ~/.bashrc; then
        echo 'eval "$(fnm env --use-on-cd)"' >> ~/.bashrc
    fi

    log_success "Node.js $NODE_VERSION installed"
fi

# ============================================================================
# Step 4: GitHub CLI
# ============================================================================
if command -v gh &>/dev/null; then
    log_success "GitHub CLI already installed"
else
    log_info "Installing GitHub CLI..."

    curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
        sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
    sudo chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | \
        sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
    sudo apt-get update
    sudo apt-get install -y gh

    log_success "GitHub CLI installed"
fi

# ============================================================================
# Step 5: Claude Code CLI
# ============================================================================
if [[ -z "${SKIP_CLAUDE:-}" ]]; then
    if command -v claude &>/dev/null; then
        log_success "Claude Code already installed"
    else
        log_info "Installing Claude Code CLI..."

        # Install via npm (requires Node.js)
        npm install -g @anthropic-ai/claude-code

        log_success "Claude Code installed"
    fi
else
    log_info "Skipping Claude Code (SKIP_CLAUDE=1)"
fi

# ============================================================================
# Step 6: core CLI
# ============================================================================
if command -v core &>/dev/null; then
    log_success "core CLI already installed"
else
    log_info "Installing core CLI..."

    # Install from releases
    ARCH=$(dpkg --print-architecture)
    CORE_URL="https://github.com/host-uk/core/releases/latest/download/core-linux-${ARCH}"

    curl -fsSL "$CORE_URL" -o /tmp/core
    chmod +x /tmp/core
    sudo mv /tmp/core /usr/local/bin/core

    log_success "core CLI installed"
fi

# ============================================================================
# Step 7: core-ide (GUI mode)
# ============================================================================
if [[ -z "${SKIP_GUI:-}" ]]; then
    if command -v core-ide &>/dev/null; then
        log_success "core-ide already installed"
    else
        log_info "Installing core-ide..."

        ARCH=$(dpkg --print-architecture)
        IDE_URL="https://github.com/host-uk/core/releases/latest/download/core-ide-linux-${ARCH}.deb"

        curl -fsSL "$IDE_URL" -o /tmp/core-ide.deb
        sudo dpkg -i /tmp/core-ide.deb || sudo apt-get install -f -y
        rm /tmp/core-ide.deb

        log_success "core-ide installed"
    fi

    # Setup autostart
    log_info "Configuring autostart..."

    mkdir -p ~/.config/autostart
    cat > ~/.config/autostart/core-ide.desktop << 'EOF'
[Desktop Entry]
Type=Application
Name=Core IDE
Comment=Development Environment
Exec=/usr/local/bin/core-ide
Icon=core-ide
Terminal=false
Categories=Development;
X-GNOME-Autostart-enabled=true
EOF

    log_success "Autostart configured"
fi

# ============================================================================
# Step 8: GitHub Authentication
# ============================================================================
if gh auth status &>/dev/null; then
    log_success "GitHub already authenticated"
else
    log_info "GitHub authentication required..."

    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        echo "$GITHUB_TOKEN" | gh auth login --with-token
        log_success "GitHub authenticated via token"
    else
        log_warn "Run 'gh auth login' to authenticate with GitHub"
    fi
fi

# ============================================================================
# Step 9: SSH Key Setup
# ============================================================================
if [[ -f ~/.ssh/id_ed25519 ]]; then
    log_success "SSH key already exists"
else
    log_info "Generating SSH key..."

    read -p "Enter email for SSH key: " EMAIL
    ssh-keygen -t ed25519 -C "$EMAIL" -f ~/.ssh/id_ed25519 -N ""

    eval "$(ssh-agent -s)"
    ssh-add ~/.ssh/id_ed25519

    log_success "SSH key generated"
    log_warn "Add this key to GitHub: https://github.com/settings/keys"
    echo ""
    cat ~/.ssh/id_ed25519.pub
    echo ""
fi

# ============================================================================
# Step 10: Create workspace directory
# ============================================================================
WORKSPACE="$HOME/Code"

if [[ -d "$WORKSPACE" ]]; then
    log_success "Workspace directory exists: $WORKSPACE"
else
    log_info "Creating workspace directory..."
    mkdir -p "$WORKSPACE"
    log_success "Created: $WORKSPACE"
fi

# ============================================================================
# Summary
# ============================================================================
echo ""
echo "============================================================"
echo -e "${GREEN}Setup complete!${NC}"
echo "============================================================"
echo ""
echo "Installed:"
echo "  - Go $(go version 2>/dev/null | grep -oP 'go\d+\.\d+\.\d+' || echo 'not in path yet')"
echo "  - Node.js $(node -v 2>/dev/null || echo 'not in path yet')"
echo "  - GitHub CLI $(gh --version 2>/dev/null | head -1 || echo 'installed')"
echo "  - core CLI $(core --version 2>/dev/null || echo 'installed')"

if [[ -z "${SKIP_GUI:-}" ]]; then
    echo "  - core-ide (GUI mode)"
fi

if [[ -z "${SKIP_CLAUDE:-}" ]]; then
    echo "  - Claude Code CLI"
fi

echo ""
echo "Next steps:"
echo "  1. Restart your terminal (or run: source ~/.bashrc)"
echo "  2. Run 'gh auth login' if not already authenticated"

if [[ ! -f ~/.ssh/id_ed25519.pub ]] || ! gh auth status &>/dev/null; then
    echo "  3. Add your SSH key to GitHub"
fi

echo ""
echo "To start developing:"
echo "  cd ~/Code"
echo "  gh repo clone host-uk/core"
echo "  cd core && core doctor"
echo ""
