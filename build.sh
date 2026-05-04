#!/bin/bash

echo "Building DNT-Vault..."

export PATH=$PATH:/usr/local/go/bin

# Build server
echo "Building server..."
cd server
go build -o bin/dnt-vault-server ./cmd/server
if [ $? -ne 0 ]; then
    echo "Server build failed"
    exit 1
fi
echo "✓ Server built: server/bin/dnt-vault-server"

# Build CLI
echo "Building CLI..."
cd ../cli
go build -o bin/ssh-sync ./cmd/cli
if [ $? -ne 0 ]; then
    echo "CLI build failed"
    exit 1
fi
echo "✓ CLI built: cli/bin/ssh-sync"

cd ..
echo ""
echo "Build complete!"
echo ""
echo "Server: ./server/bin/dnt-vault-server"
echo "CLI:    ./cli/bin/ssh-sync"
