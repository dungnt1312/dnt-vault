#!/bin/bash

# Release script for DNT-Vault
# Usage: ./scripts/release.sh <version>

set -e

VERSION=${1:-"1.0.0"}
REPO_ROOT=$(cd "$(dirname "$0")/.." && pwd)
RELEASE_DIR="$REPO_ROOT/releases"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}Building DNT-Vault v$VERSION${NC}"
echo ""

# Clean and create release directory
rm -rf "$RELEASE_DIR"
mkdir -p "$RELEASE_DIR"

# Platforms to build
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "linux/arm"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/386"
)

export PATH=$PATH:/usr/local/go/bin

# Build server
echo -e "${YELLOW}Building server...${NC}"
cd "$REPO_ROOT/server"

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r -a array <<< "$platform"
    GOOS="${array[0]}"
    GOARCH="${array[1]}"
    
    output_name="dnt-vault-server-$GOOS-$GOARCH"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo -e "  Building $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$RELEASE_DIR/$output_name" ./cmd/server
    
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} $output_name"
    else
        echo -e "  ${RED}✗${NC} Failed to build $output_name"
        exit 1
    fi
done

# Build CLI
echo ""
echo -e "${YELLOW}Building CLI...${NC}"
cd "$REPO_ROOT/cli"

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r -a array <<< "$platform"
    GOOS="${array[0]}"
    GOARCH="${array[1]}"
    
    output_name="dnt-vault-$GOOS-$GOARCH"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo -e "  Building $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$RELEASE_DIR/$output_name" ./cmd/cli
    
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} $output_name"
    else
        echo -e "  ${RED}✗${NC} Failed to build $output_name"
        exit 1
    fi
done

# Generate checksums
echo ""
echo -e "${YELLOW}Generating checksums...${NC}"
cd "$RELEASE_DIR"
sha256sum * > checksums.txt
echo -e "${GREEN}✓${NC} checksums.txt"

# Create release notes
echo ""
echo -e "${YELLOW}Creating release notes...${NC}"
cat > "$RELEASE_DIR/RELEASE_NOTES.md" << EOF
# DNT-Vault v$VERSION

## Installation

### Linux / macOS
\`\`\`bash
curl -fsSL https://raw.githubusercontent.com/dungnt1312/dnt-vault/master/install.sh | bash
\`\`\`

### Windows (PowerShell)
\`\`\`powershell
irm https://raw.githubusercontent.com/dungnt1312/dnt-vault/master/install.ps1 | iex
\`\`\`

### Manual Installation

Download the appropriate binary for your platform:

**Server:**
- Linux (amd64): \`dnt-vault-server-linux-amd64\`
- Linux (arm64): \`dnt-vault-server-linux-arm64\`
- macOS (Intel): \`dnt-vault-server-darwin-amd64\`
- macOS (Apple Silicon): \`dnt-vault-server-darwin-arm64\`
- Windows (64-bit): \`dnt-vault-server-windows-amd64.exe\`

**CLI:**
- Linux (amd64): \`dnt-vault-linux-amd64\`
- Linux (arm64): \`dnt-vault-linux-arm64\`
- macOS (Intel): \`dnt-vault-darwin-amd64\`
- macOS (Apple Silicon): \`dnt-vault-darwin-arm64\`
- Windows (64-bit): \`dnt-vault-windows-amd64.exe\`

Make the binary executable:
\`\`\`bash
chmod +x dnt-vault-server-*
chmod +x dnt-vault-*
\`\`\`

Move to PATH:
\`\`\`bash
sudo mv dnt-vault-server-* /usr/local/bin/dnt-vault-server
sudo mv dnt-vault-* /usr/local/bin/dnt-vault
\`\`\`

## Checksums

See \`checksums.txt\` for SHA256 checksums of all binaries.

## Documentation

- [README](https://github.com/dungnt1312/dnt-vault/blob/master/README.md)
- [Quick Start](https://github.com/dungnt1312/dnt-vault/blob/master/QUICKSTART.md)
- [Demo](https://github.com/dungnt1312/dnt-vault/blob/master/DEMO.md)

## What's Changed

See [CHANGELOG.md](https://github.com/dungnt1312/dnt-vault/blob/master/CHANGELOG.md)
EOF

echo -e "${GREEN}✓${NC} RELEASE_NOTES.md"

# Summary
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ Build complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${CYAN}Release directory:${NC} $RELEASE_DIR"
echo ""
echo -e "${CYAN}Files created:${NC}"
ls -lh "$RELEASE_DIR" | tail -n +2 | awk '{printf "  %s (%s)\n", $9, $5}'
echo ""
echo -e "${CYAN}Next steps:${NC}"
echo "  1. Review RELEASE_NOTES.md"
echo "  2. Create GitHub release:"
echo "     gh release create v$VERSION --title \"v$VERSION\" --notes-file releases/RELEASE_NOTES.md releases/*"
echo ""
