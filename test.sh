#!/bin/bash

set -e

echo "🚀 Testing DNT-Vault SSH Config Sync"
echo "===================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Paths
SERVER_BIN="./server/bin/dnt-vault-server"
CLI_BIN="./cli/bin/ssh-sync"
TEST_DIR="/tmp/dnt-vault-test"
SERVER_DATA="$TEST_DIR/server-data"
SERVER_CONFIG="$TEST_DIR/server-config"
CLIENT_HOME="$TEST_DIR/client"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
    rm -rf "$TEST_DIR"
}

trap cleanup EXIT

# Setup
echo -e "${YELLOW}Setting up test environment...${NC}"
rm -rf "$TEST_DIR"
mkdir -p "$SERVER_DATA" "$SERVER_CONFIG" "$CLIENT_HOME/.ssh" "$CLIENT_HOME/.ssh-sync"

# Create test SSH config
cat > "$CLIENT_HOME/.ssh/config" << 'EOF'
Host example
    HostName example.com
    User ubuntu
    Port 22

Host github
    HostName github.com
    User git
EOF

# Create test SSH key
ssh-keygen -t ed25519 -f "$CLIENT_HOME/.ssh/id_test" -N "" -q
echo -e "${GREEN}✓ Test environment created${NC}"

# Start server
echo -e "\n${YELLOW}Starting server...${NC}"
export PORT=18443
export DATA_PATH="$SERVER_DATA"
export CONFIG_PATH="$SERVER_CONFIG"
export HOME="$CLIENT_HOME"

$SERVER_BIN > "$TEST_DIR/server.log" 2>&1 &
SERVER_PID=$!

sleep 2

if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo -e "${RED}✗ Server failed to start${NC}"
    cat "$TEST_DIR/server.log"
    exit 1
fi

echo -e "${GREEN}✓ Server started (PID: $SERVER_PID)${NC}"

# Test CLI
echo -e "\n${YELLOW}Testing CLI commands...${NC}"

# Test 1: Init
echo -e "\n${YELLOW}Test 1: Initialize client${NC}"
echo -e "http://localhost:18443\ntestpass\ntestpass" | $CLI_BIN init
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Init successful${NC}"
else
    echo -e "${RED}✗ Init failed${NC}"
    exit 1
fi

# Test 2: Login
echo -e "\n${YELLOW}Test 2: Login${NC}"
echo -e "admin\nadmin" | $CLI_BIN login
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Login successful${NC}"
else
    echo -e "${RED}✗ Login failed${NC}"
    exit 1
fi

# Test 3: Push without keys
echo -e "\n${YELLOW}Test 3: Push config (without keys)${NC}"
echo "test-profile" | $CLI_BIN push
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Push successful${NC}"
else
    echo -e "${RED}✗ Push failed${NC}"
    exit 1
fi

# Test 4: List profiles
echo -e "\n${YELLOW}Test 4: List profiles${NC}"
$CLI_BIN list
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ List successful${NC}"
else
    echo -e "${RED}✗ List failed${NC}"
    exit 1
fi

# Test 5: Pull profile
echo -e "\n${YELLOW}Test 5: Pull profile${NC}"
mv "$CLIENT_HOME/.ssh/config" "$CLIENT_HOME/.ssh/config.backup"
echo -e "1\nn" | $CLI_BIN pull
if [ $? -eq 0 ] && [ -f "$CLIENT_HOME/.ssh/config" ]; then
    echo -e "${GREEN}✓ Pull successful${NC}"
else
    echo -e "${RED}✗ Pull failed${NC}"
    exit 1
fi

# Test 6: Verify content
echo -e "\n${YELLOW}Test 6: Verify pulled content${NC}"
if diff -q "$CLIENT_HOME/.ssh/config" "$CLIENT_HOME/.ssh/config.backup" > /dev/null; then
    echo -e "${GREEN}✓ Content matches${NC}"
else
    echo -e "${RED}✗ Content mismatch${NC}"
    exit 1
fi

# Test 7: Delete profile
echo -e "\n${YELLOW}Test 7: Delete profile${NC}"
echo "y" | $CLI_BIN delete --profile test-profile
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Delete successful${NC}"
else
    echo -e "${RED}✗ Delete failed${NC}"
    exit 1
fi

# Summary
echo -e "\n${GREEN}===================================="
echo "✓ All tests passed!"
echo "====================================${NC}"
echo ""
echo "Server log: $TEST_DIR/server.log"
echo "Test data: $TEST_DIR"
