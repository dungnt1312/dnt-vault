#!/bin/bash

set -e

export PATH=$PATH:/usr/local/go/bin

VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
COMMIT_SHA="$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"

SERVER_PKG="main"
CLI_PKG="main"

LDFLAGS="-X '${SERVER_PKG}.Version=${VERSION}' \
         -X '${SERVER_PKG}.BuildTime=${BUILD_TIME}' \
         -X '${SERVER_PKG}.CommitSHA=${COMMIT_SHA}'"

echo "Building DNT-Vault ${VERSION} (${COMMIT_SHA})..."

# Build server
echo "Building server..."
cd server
go build -ldflags "${LDFLAGS}" -o bin/dnt-vault-server ./cmd/server
echo "✓ Server built: server/bin/dnt-vault-server"

# Build CLI
echo "Building CLI..."
cd ../cli
go build -ldflags "${LDFLAGS}" -o bin/ssh-sync ./cmd/cli
echo "✓ CLI built: cli/bin/ssh-sync"

cd ..
echo ""
echo "Build complete! Version: ${VERSION}"
echo ""
echo "Server: ./server/bin/dnt-vault-server"
echo "CLI:    ./cli/bin/ssh-sync"
