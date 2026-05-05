#!/bin/bash

# DNT-Vault Installation Script
# Supports: Linux, macOS, Windows (Git Bash)

set -e

REPO="dungnt1312/dnt-vault"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}"
cat << 'EOF'
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║              DNT-Vault SSH Config Sync Installer              ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)     OS="linux";;
        Darwin*)    OS="darwin";;
        MINGW*|MSYS*|CYGWIN*)    OS="windows";;
        *)          OS="unknown";;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   ARCH="amd64";;
        aarch64|arm64)  ARCH="arm64";;
        armv7l)         ARCH="arm";;
        *)              ARCH="unknown";;
    esac
}

# Check if running with sudo (for Linux/macOS)
check_sudo() {
    if [ "$OS" != "windows" ] && [ "$EUID" -ne 0 ]; then
        echo -e "${YELLOW}This script requires sudo privileges to install to $INSTALL_DIR${NC}"
        echo -e "${YELLOW}Please run: sudo $0${NC}"
        exit 1
    fi
}

# Fetch latest version from GitHub Releases API
fetch_latest_version() {
    local api_url="https://api.github.com/repos/$REPO/releases/latest"
    local tag

    if command -v curl &> /dev/null; then
        tag=$(curl -fsSL "$api_url" | grep -o '"tag_name": "[^"]*"' | cut -d'"' -f4)
    elif command -v wget &> /dev/null; then
        tag=$(wget -qO- "$api_url" | grep -o '"tag_name": "[^"]*"' | cut -d'"' -f4)
    else
        echo -e "${RED}Error: curl or wget is required${NC}"
        exit 1
    fi

    if [ -z "$tag" ]; then
        echo -e "${RED}Error: failed to fetch latest version from GitHub${NC}"
        exit 1
    fi

    echo "$tag"
}

# Download and install
install() {
    detect_os
    detect_arch

    echo -e "${CYAN}Detected OS: $OS${NC}"
    echo -e "${CYAN}Detected Architecture: $ARCH${NC}"
    echo ""

    if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
        echo -e "${RED}Unsupported OS or architecture${NC}"
        exit 1
    fi

    # Fetch latest version from GitHub
    echo -e "${CYAN}Fetching latest version...${NC}"
    VERSION=$(fetch_latest_version | sed 's/^v//')
    echo -e "${GREEN}Latest version: v$VERSION${NC}"
    echo ""

    # Set install dir
    if [ "$OS" = "windows" ]; then
        INSTALL_DIR="$HOME/bin"
        echo -e "${CYAN}Windows (Git Bash) detected: Installing to $INSTALL_DIR${NC}"
        mkdir -p "$INSTALL_DIR"
    else
        check_sudo
    fi

    # Download URLs
    SERVER_URL="https://github.com/$REPO/releases/download/v$VERSION/dnt-vault-server-$OS-$ARCH"
    CLI_URL="https://github.com/$REPO/releases/download/v$VERSION/dnt-vault-$OS-$ARCH"

    if [ "$OS" = "windows" ]; then
        SERVER_URL="${SERVER_URL}.exe"
        CLI_URL="${CLI_URL}.exe"
    fi

    SERVER_OUT="$INSTALL_DIR/dnt-vault-server"
    CLI_OUT="$INSTALL_DIR/dnt-vault"
    if [ "$OS" = "windows" ]; then
        SERVER_OUT="${SERVER_OUT}.exe"
        CLI_OUT="${CLI_OUT}.exe"
    fi

    echo -e "${CYAN}Downloading binaries...${NC}"

    # Download server
    echo -e "${YELLOW}Downloading server...${NC}"
    if command -v curl &> /dev/null; then
        curl -L -o "$SERVER_OUT" "$SERVER_URL"
    elif command -v wget &> /dev/null; then
        wget -O "$SERVER_OUT" "$SERVER_URL"
    else
        echo -e "${RED}Error: curl or wget is required${NC}"
        exit 1
    fi

    # Download CLI
    echo -e "${YELLOW}Downloading CLI...${NC}"
    if command -v curl &> /dev/null; then
        curl -L -o "$CLI_OUT" "$CLI_URL"
    else
        wget -O "$CLI_OUT" "$CLI_URL"
    fi

    # Make executable
    chmod +x "$SERVER_OUT"
    chmod +x "$CLI_OUT"

    echo ""
    echo -e "${GREEN}✓ Binaries installed:${NC}"
    echo -e "  Server: $SERVER_OUT"
    echo -e "  CLI:    $CLI_OUT"

    if [ "$OS" = "windows" ]; then
        setup_path_windows
    fi

    echo ""
    echo -e "${CYAN}Quick start:${NC}"
    echo -e "  1. dnt-vault-server   # Start server"
    echo -e "  2. dnt-vault init     # Initialize client"
    echo -e "  3. dnt-vault login    # Login"
    echo -e "  4. dnt-vault push     # Push SSH config"
    echo ""
    echo -e "${CYAN}Documentation: https://github.com/$REPO${NC}"
}

# Add ~/bin to PATH in ~/.bashrc and source it
setup_path_windows() {
    local BASHRC="$HOME/.bashrc"
    local BIN_DIR="$HOME/bin"

    # Add to .bashrc if not already there
    if ! grep -qF 'PATH="$HOME/bin' "$BASHRC" 2>/dev/null && ! grep -qF "PATH=\$HOME/bin" "$BASHRC" 2>/dev/null; then
        echo "" >> "$BASHRC"
        echo "# dnt-vault" >> "$BASHRC"
        echo 'export PATH="$HOME/bin:$PATH"' >> "$BASHRC"
        echo -e "${GREEN}✓ Added ~/bin to PATH in $BASHRC${NC}"
    else
        echo -e "${GREEN}✓ ~/bin already in PATH ($BASHRC)${NC}"
    fi

    # Source .bashrc to apply in current session
    # shellcheck disable=SC1090
    source "$BASHRC" 2>/dev/null || true
    export PATH="$BIN_DIR:$PATH"

    echo -e "${GREEN}✓ PATH updated — dnt-vault is ready to use${NC}"
}

# Uninstall
uninstall() {
    echo -e "${YELLOW}Uninstalling DNT-Vault...${NC}"

    detect_os
    if [ "$OS" = "windows" ]; then
        INSTALL_DIR="$HOME/bin"
    fi

    if [ "$OS" != "windows" ] && [ "$EUID" -ne 0 ]; then
        echo -e "${YELLOW}This script requires sudo privileges${NC}"
        exit 1
    fi

    rm -f "$INSTALL_DIR/dnt-vault-server" "$INSTALL_DIR/dnt-vault-server.exe"
    rm -f "$INSTALL_DIR/dnt-vault" "$INSTALL_DIR/dnt-vault.exe"

    echo -e "${GREEN}✓ Uninstalled successfully${NC}"
}

# Main
case "${1:-install}" in
    install)
        install
        ;;
    uninstall)
        uninstall
        ;;
    *)
        echo "Usage: $0 {install|uninstall}"
        exit 1
        ;;
esac
