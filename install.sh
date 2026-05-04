#!/bin/bash

# DNT-Vault Installation Script
# Supports: Linux, macOS, Windows (Git Bash)

set -e

VERSION="1.1.1"
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

    # For Windows, use current directory
    if [ "$OS" = "windows" ]; then
        INSTALL_DIR="."
        echo -e "${YELLOW}Windows detected: Installing to current directory${NC}"
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
    echo -e "${GREEN}✓ Installation complete!${NC}"
    echo ""
    echo -e "${CYAN}Installed binaries:${NC}"
    echo -e "  Server: $SERVER_OUT"
    echo -e "  CLI:    $CLI_OUT"
    echo ""

    if [ "$OS" = "windows" ]; then
        echo -e "${YELLOW}Add to PATH or run from current directory:${NC}"
        echo -e "  ./dnt-vault-server.exe"
        echo -e "  ./dnt-vault.exe"
    else
        echo -e "${CYAN}Usage:${NC}"
        echo -e "  dnt-vault-server    # Start server"
        echo -e "  dnt-vault init      # Initialize client"
    fi

    echo ""
    echo -e "${CYAN}Quick start:${NC}"
    echo -e "  1. Start server: dnt-vault-server"
    echo -e "  2. Init client:  dnt-vault init"
    echo -e "  3. Login:        dnt-vault login"
    echo -e "  4. Push config:  dnt-vault push"
    echo ""
    echo -e "${CYAN}Documentation: https://github.com/$REPO${NC}"
}

# Uninstall
uninstall() {
    echo -e "${YELLOW}Uninstalling DNT-Vault...${NC}"
    
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
        detect_os
        if [ "$OS" = "windows" ]; then
            INSTALL_DIR="."
        fi
        uninstall
        ;;
    *)
        echo "Usage: $0 {install|uninstall}"
        exit 1
        ;;
esac
