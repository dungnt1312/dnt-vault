# Quick Start Guide

## 1. Build

```bash
./build.sh
```

## 2. Start Server

```bash
# Terminal 1: Start server
export PORT=8443
export DATA_PATH=/tmp/dnt-vault-data
export CONFIG_PATH=/tmp/dnt-vault-config

./server/bin/dnt-vault-server
```

Default credentials: `admin/admin`

## 3. Setup Client

```bash
# Terminal 2: Initialize client
./cli/bin/ssh-sync init
# Enter server URL: http://localhost:8443
# Enter master password: (your choice)

# Login
./cli/bin/ssh-sync login
# Username: admin
# Password: admin
```

## 4. Push Your Config

```bash
# Push SSH config only
./cli/bin/ssh-sync push

# Or push with private keys
./cli/bin/ssh-sync push --include-keys
```

## 5. Pull on Another Machine

```bash
# List available profiles
./cli/bin/ssh-sync list

# Pull interactively
./cli/bin/ssh-sync pull

# Or pull specific profile
./cli/bin/ssh-sync pull --profile your-hostname
```

## Common Commands

```bash
# List all profiles
./cli/bin/ssh-sync list

# Delete a profile
./cli/bin/ssh-sync delete --profile old-laptop

# Logout
./cli/bin/ssh-sync logout
```

## Testing

Run integration tests:

```bash
./test.sh
```

## Troubleshooting

### Server won't start
```bash
# Check if port is in use
lsof -i :8443

# Check logs
tail -f /tmp/dnt-vault-config/server.log
```

### Login fails
```bash
# Verify server is running
curl http://localhost:8443/api/v1/profiles

# Check client config
cat ~/.ssh-sync/config.yaml
```

### Decryption fails
- Verify master password is correct
- For keys: ensure passphrase matches what was used during push
