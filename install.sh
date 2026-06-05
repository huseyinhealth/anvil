#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

# Banner
echo -e "${CYAN}${BOLD}"
echo "    ___              _ __"
echo "   /   |  ____ _  __(_) /"
echo "  / /| | / __ \ \/ / / /"
echo " / ___ |/ / / /\  / / /"
echo "/_/  |_/_/ /_/ /_/_/_/"
echo -e "${RESET}"
echo -e "${BOLD}Instance + Mod Manager for FabricMC${RESET}"
echo ""

# Helpers
info()    { echo -e "${CYAN}::${RESET} $1"; }
success() { echo -e "${GREEN}✔${RESET} $1"; }
warn()    { echo -e "${YELLOW}⚠${RESET} $1"; }
error()   { echo -e "${RED}✖${RESET} $1"; exit 1; }

# Parse args
SUDO=""
BIN_DIR="$HOME/.local/bin"
if [[ "$1" == "--system" ]]; then
    BIN_DIR="/usr/local/bin"
    SUDO="sudo"
fi

# Check dependencies
info "Checking dependencies..."
command -v go &>/dev/null  || error "Go is not installed. Install it from https://go.dev/dl/"
command -v git &>/dev/null || error "Git is not installed."
success "Dependencies OK"

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED="1.21"
if [[ "$(printf '%s\n' "$REQUIRED" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED" ]]; then
    error "Go $REQUIRED or higher is required (found $GO_VERSION)"
fi
success "Go $GO_VERSION"

# Clone or update
INSTALL_DIR="$HOME/.local/share/anvil-src"
if [ -d "$INSTALL_DIR" ]; then
    info "Updating existing source..."
    git -C "$INSTALL_DIR" pull --quiet --rebase
else
    info "Cloning repository..."
    git clone --quiet https://github.com/huseyinhealth/anvil "$INSTALL_DIR"
fi
success "Source ready"

# Client ID
echo ""
warn "A Microsoft Azure Client ID is required for login."
echo -e "  Create one at ${CYAN}https://portal.azure.com${RESET} and follow the instructions in the README."
echo -ne "${BOLD}Enter your Azure Client ID: ${RESET}"
read -r CLIENT_ID
if [[ -z "$CLIENT_ID" ]]; then
    warn "No client ID provided. Login will not work."
    CLIENT_ID=""
fi

# Build
info "Building anvil..."
cd "$INSTALL_DIR"
go build -ldflags "-s -w -X 'anvil/internal.LauncherClientID=$CLIENT_ID'" -o anvil . 2>/dev/null || error "Build failed."
success "Build complete"

# Install binary
mkdir -p "$BIN_DIR"
$SUDO mv anvil "$BIN_DIR/anvil"
$SUDO chmod +x "$BIN_DIR/anvil"
success "Installed to $BIN_DIR/anvil"

# PATH check
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    warn "$BIN_DIR is not in your PATH."
    echo -e "  Add this to your shell config:"
    echo -e "  ${BOLD}export PATH=\"\$HOME/.local/bin:\$PATH\"${RESET}"
fi

echo ""
echo -e "${GREEN}${BOLD}Anvil installed successfully!${RESET}"
echo -e "Run ${CYAN}anvil help${RESET} to get started."
