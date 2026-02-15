#!/bin/bash
# playconsole-cli - Google Play Console CLI installer
# Usage: curl -fsSL https://raw.githubusercontent.com/AndroidPoet/playconsole-cli/main/install.sh | bash

set -e

REPO="${GPC_REPO:-AndroidPoet/playconsole-cli}"
INSTALL_DIR="${GPC_INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${GPC_VERSION:-latest}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() { echo -e "${GREEN}==>${NC} $1"; }
warn() { echo -e "${YELLOW}==>${NC} $1"; }
error() { echo -e "${RED}Error:${NC} $1"; exit 1; }

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) error "Unsupported OS: $OS" ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
esac

info "Installing playconsole-cli for ${OS}/${ARCH}..."

# GitHub API headers
AUTH_HEADER=""
if [ -n "$GITHUB_TOKEN" ]; then
    AUTH_HEADER="Authorization: token $GITHUB_TOKEN"
fi

# Get download URL
if [ "$VERSION" = "latest" ]; then
    RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
else
    RELEASE_URL="https://api.github.com/repos/${REPO}/releases/tags/${VERSION}"
fi

info "Fetching release info..."

if [ -n "$AUTH_HEADER" ]; then
    RELEASE_JSON=$(curl -fsSL -H "$AUTH_HEADER" "$RELEASE_URL" 2>/dev/null) || error "Failed to fetch release info. Check your GITHUB_TOKEN."
else
    RELEASE_JSON=$(curl -fsSL "$RELEASE_URL" 2>/dev/null) || error "Failed to fetch release info."
fi

# Find asset for this platform
ASSET_NAME="playconsole-cli_.*_${OS}_${ARCH}"
if [ "$OS" = "windows" ]; then
    ASSET_NAME="${ASSET_NAME}.zip"
else
    ASSET_NAME="${ASSET_NAME}.tar.gz"
fi

DOWNLOAD_URL=$(echo "$RELEASE_JSON" | grep -o "\"browser_download_url\": \"[^\"]*${ASSET_NAME}[^\"]*\"" | head -1 | cut -d'"' -f4)

if [ -z "$DOWNLOAD_URL" ]; then
    # Try to find any matching asset
    warn "No pre-built binary for ${OS}/${ARCH}. Attempting build from source..."

    if ! command -v go &> /dev/null; then
        error "Go is required to build from source. Install Go from https://go.dev"
    fi

    TMPDIR=$(mktemp -d)
    cd "$TMPDIR"

    if [ -n "$GITHUB_TOKEN" ]; then
        git clone "https://${GITHUB_TOKEN}@github.com/${REPO}.git" playconsole-cli
    else
        git clone "https://github.com/${REPO}.git" playconsole-cli
    fi

    cd playconsole-cli
    go build -o playconsole-cli ./cmd/playconsole-cli
    mkdir -p "$INSTALL_DIR"
    mv playconsole-cli "$INSTALL_DIR/playconsole-cli"
    cd /
    rm -rf "$TMPDIR"
else
    info "Downloading from: $DOWNLOAD_URL"

    TMPDIR=$(mktemp -d)
    cd "$TMPDIR"

    if [ -n "$AUTH_HEADER" ]; then
        curl -fsSL -H "$AUTH_HEADER" -H "Accept: application/octet-stream" -o archive "$DOWNLOAD_URL"
    else
        curl -fsSL -o archive "$DOWNLOAD_URL"
    fi

    # Extract
    if [ "$OS" = "windows" ]; then
        unzip -q archive
    else
        tar -xzf archive
    fi

    # Install
    mkdir -p "$INSTALL_DIR"
    mv playconsole-cli "$INSTALL_DIR/playconsole-cli"

    cd /
    rm -rf "$TMPDIR"
fi

chmod +x "$INSTALL_DIR/playconsole-cli"

info "Installed playconsole-cli to $INSTALL_DIR/playconsole-cli"

# Check PATH
if ! echo "$PATH" | tr ':' '\n' | grep -q "^$INSTALL_DIR$"; then
    warn "Add to your PATH by adding this to ~/.bashrc or ~/.zshrc:"
    echo ""
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    echo ""
fi

# Verify installation
if "$INSTALL_DIR/playconsole-cli" version &>/dev/null; then
    info "Installation complete!"
    echo ""
    "$INSTALL_DIR/playconsole-cli" version
else
    warn "Installed but could not verify. Try running: $INSTALL_DIR/playconsole-cli --help"
fi
