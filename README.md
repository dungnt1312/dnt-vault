# DNT-Vault SSH Config Sync

Self-hosted SSH config and keys synchronization tool written in Go.

## Features

- 🔐 **Secure**: Client-side encryption with AES-256-GCM
- 🔑 **Private Keys**: Optional sync with passphrase protection
- 🏠 **Self-hosted**: No third-party services required
- 🚀 **Simple**: Easy setup and intuitive CLI
- 🔄 **Conflict Detection**: Shows diff before overwriting
- 💾 **Auto Backup**: Automatic backups before pulling

## Architecture

```
┌─────────────┐         HTTPS/HTTP        ┌─────────────┐
│   Client    │ ◄────────────────────────► │   Server    │
│  (CLI Tool) │    Encrypted Data Only     │   (Vault)   │
└─────────────┘                            └─────────────┘
```

- **Server**: REST API vault server (stores encrypted data)
- **Client**: CLI tool for push/pull operations
- **Encryption**: All data encrypted client-side before upload

## Quick Start

### 1. Start Server

```bash
# Set environment variables (optional)
export PORT=8443
export DATA_PATH=/var/lib/dnt-vault/data
export CONFIG_PATH=/etc/dnt-vault

# Run server
./server/bin/dnt-vault-server
```

Default credentials: `admin/admin` (change after first login)

### 2. Setup Client

```bash
# Initialize configuration
./cli/bin/ssh-sync init

# Login to vault
./cli/bin/ssh-sync login
```

### 3. Push/Pull Configs

```bash
# Push current SSH config to vault
./cli/bin/ssh-sync push

# Push with private keys
./cli/bin/ssh-sync push --include-keys

# List available profiles
./cli/bin/ssh-sync list

# Pull a profile
./cli/bin/ssh-sync pull

# Pull specific profile
./cli/bin/ssh-sync pull --profile work-laptop
```

## CLI Commands

### Setup & Authentication

```bash
ssh-sync init              # Initialize configuration
ssh-sync login             # Login to vault server
ssh-sync logout            # Logout
```

### Sync Operations

```bash
ssh-sync push                          # Push config to vault
ssh-sync push --include-keys           # Push config + private keys
ssh-sync push --profile custom-name    # Custom profile name

ssh-sync pull                          # Interactive pull
ssh-sync pull --profile work-laptop    # Pull specific profile
```

### Management

```bash
ssh-sync list                          # List all profiles
ssh-sync delete --profile old-laptop   # Delete profile
```

## Configuration

### Server Config

Environment variables:
- `PORT`: Server port (default: 8443)
- `DATA_PATH`: Data storage path (default: /var/lib/dnt-vault/data)
- `CONFIG_PATH`: Config path (default: /etc/dnt-vault)

### Client Config

Located at `~/.ssh-sync/config.yaml`:

```yaml
server:
  url: http://localhost:8443
  tls_verify: true

ssh:
  config_path: ~/.ssh/config
  keys_dir: ~/.ssh

profiles:
  current: ""
  default_name_format: "{hostname}"

backup:
  enabled: true
  dir: ~/.ssh-sync/backups
  max_backups: 10

encryption:
  master_key_file: ~/.ssh-sync/master.key
```

## Security

### Encryption Layers

1. **Transport**: HTTPS (TLS 1.3)
2. **Authentication**: JWT tokens (24h expiry)
3. **Data at Rest**: AES-256-GCM encryption
4. **Private Keys**: Separate passphrase protection

### Key Points

- Server never sees plaintext data
- Master password stored locally only
- Private keys encrypted with separate passphrase
- All sensitive files have 0600 permissions

## Building from Source

### Requirements

- Go 1.22+

### Build

```bash
# Build server
cd server
go build -o bin/dnt-vault-server ./cmd/server

# Build CLI
cd cli
go build -o bin/ssh-sync ./cmd/cli
```

## Deployment

### Server (VPS)

```bash
# Create directories
sudo mkdir -p /var/lib/dnt-vault/data
sudo mkdir -p /etc/dnt-vault

# Copy binary
sudo cp server/bin/dnt-vault-server /usr/local/bin/

# Create systemd service
sudo tee /etc/systemd/system/dnt-vault.service << EOF
[Unit]
Description=DNT-Vault SSH Config Sync Server
After=network.target

[Service]
Type=simple
User=dnt-vault
Environment="PORT=8443"
Environment="DATA_PATH=/var/lib/dnt-vault/data"
Environment="CONFIG_PATH=/etc/dnt-vault"
ExecStart=/usr/local/bin/dnt-vault-server
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

# Start service
sudo systemctl daemon-reload
sudo systemctl enable dnt-vault
sudo systemctl start dnt-vault
```

### Client

```bash
# Copy to PATH
sudo cp cli/bin/ssh-sync /usr/local/bin/

# Or create alias
echo 'alias ssh-sync="/path/to/cli/bin/ssh-sync"' >> ~/.bashrc
```

## Workflow Example

### Initial Setup (Machine A)

```bash
# Initialize and login
ssh-sync init
ssh-sync login

# Push current config
ssh-sync push --profile work-laptop --include-keys
```

### Pull on Another Machine (Machine B)

```bash
# Initialize and login
ssh-sync init
ssh-sync login

# List and pull
ssh-sync list
ssh-sync pull --profile work-laptop
```

## API Documentation

### Authentication

**POST /api/v1/auth/login**
```json
{
  "username": "admin",
  "password": "secret"
}
```

### Profiles

**GET /api/v1/profiles** - List all profiles  
**GET /api/v1/profiles/:name** - Get profile data  
**POST /api/v1/profiles/:name** - Create/update profile  
**DELETE /api/v1/profiles/:name** - Delete profile

All profile endpoints require `Authorization: Bearer <token>` header.

## Troubleshooting

### Server won't start
- Check if port is already in use: `lsof -i :8443`
- Verify permissions on data/config directories

### Login fails
- Verify server URL in `~/.ssh-sync/config.yaml`
- Check server is running: `curl http://localhost:8443/api/v1/profiles`

### Decryption fails
- Ensure master password is correct
- For keys: verify passphrase matches what was used during push

## License

MIT

## Author

DNT Team
