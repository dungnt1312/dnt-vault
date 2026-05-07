#!/bin/bash

set -e

SERVER_BIN="./bin/dnt-vault-server"
CLI_BIN="./bin/dnt-vault"
TEST_DIR="/tmp/dnt-vault-test-env"
SERVER_DATA="$TEST_DIR/server-data"
SERVER_CONFIG="$TEST_DIR/server-config"
CLIENT_HOME="$TEST_DIR/client"

cleanup() {
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
    rm -rf "$TEST_DIR"
}

trap cleanup EXIT

rm -rf "$TEST_DIR"
mkdir -p "$SERVER_DATA" "$SERVER_CONFIG" "$CLIENT_HOME/.ssh" "$CLIENT_HOME/.dnt-vault"

cat > "$CLIENT_HOME/.env.test" << 'EOF'
DATABASE_URL=postgres://localhost/db
API_KEY=abc123
EOF

export PORT=18443
export DATA_PATH="$SERVER_DATA"
export CONFIG_PATH="$SERVER_CONFIG"
export HOME="$CLIENT_HOME"
export DNT_VAULT_SERVER_URL="http://localhost:18443"
export DNT_VAULT_MASTER_PASSWORD="sshpass"
export DNT_VAULT_ENV_MASTER_PASSWORD="envpass"
export DNT_VAULT_USERNAME="admin"
export DNT_VAULT_PASSWORD="admin"

$SERVER_BIN > "$TEST_DIR/server.log" 2>&1 &
SERVER_PID=$!
sleep 2

$CLI_BIN init
$CLI_BIN login
$CLI_BIN env init

$CLI_BIN env push myapp/production --file "$CLIENT_HOME/.env.test"
$CLI_BIN env list
$CLI_BIN env list myapp/production
$CLI_BIN env get myapp/production API_KEY
$CLI_BIN env set myapp/production API_KEY def456
$CLI_BIN env pull myapp/production --output "$CLIENT_HOME/.env.out"
$CLI_BIN env delete myapp/production API_KEY
$CLI_BIN env delete myapp/production --all --yes

echo "env workflow ok"
