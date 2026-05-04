# dnt-vault

Self-hosted SSH config and key synchronization tool written in Go. Sync your `~/.ssh/config` and private keys across machines through your own vault server — encrypted client-side, no third-party services.

## Install

Linux / macOS:

```
curl -fsSL https://raw.githubusercontent.com/dungnt1312/dnt-vault/master/install.sh | sudo bash
```

Windows (PowerShell):

```
irm https://raw.githubusercontent.com/dungnt1312/dnt-vault/master/install.ps1 | iex
```

Or download binaries directly from [Releases](https://github.com/dungnt1312/dnt-vault/releases).

## Quick Start

### 1. Start the Server

```
dnt-vault-server
```

Starts on `0.0.0.0:8443` by default. Default credentials: `admin` / `admin`.

```
PORT=8443
DATA_PATH=~/dnt-vault/data
CONFIG_PATH=~/dnt-vault/config
```

### 2. Initialize Client

```
dnt-vault init
```

Enter your server URL and set a master password. Config saved to `~/.dnt-vault/config.yaml`.

### 3. Login

```
dnt-vault login
```

### 4. Push SSH Config

```
dnt-vault push
```

Push with private keys:

```
dnt-vault push --include-keys
```

### 5. Pull on Another Machine

```
dnt-vault init    # set same server URL + master password
dnt-vault login
dnt-vault pull
```

## CLI Commands

```
dnt-vault init                          # Initialize config + master password
dnt-vault login                         # Login to vault server
dnt-vault logout                        # Clear local token

dnt-vault push                          # Push SSH config to vault
dnt-vault push --include-keys           # Push config + private keys
dnt-vault push --profile <name>         # Custom profile name

dnt-vault pull                          # Interactive pull (select profile)
dnt-vault pull --profile <name>         # Pull specific profile

dnt-vault list                          # List all profiles
dnt-vault delete --profile <name>       # Delete a profile

dnt-vault version                       # Show version info
```

## Features

- Client-side encryption: AES-256-GCM with PBKDF2 key derivation — server never sees plaintext.
- Private key sync: Optional, encrypted with a separate passphrase.
- Conflict detection: LCS-based diff before overwriting local config.
- Auto backup: Timestamped backups before every pull.
- Multi-profile: Multiple named profiles per user.
- Multi-user: Each user has isolated encrypted storage.
- Rate limiting: 5 login attempts/minute per IP.
- Graceful shutdown: Drains in-flight requests on SIGINT/SIGTERM.

## Configuration

Client config at `~/.dnt-vault/config.yaml`:

```yaml
server:
  url: http://your-server:8443
  tls_verify: true
ssh:
  config_path: ~/.ssh/config
  keys_dir: ~/.ssh
backup:
  enabled: true
  dir: ~/.dnt-vault/backups
  max_backups: 10
encryption:
  master_key_file: ~/.dnt-vault/master.key
```

Server environment variables:

```
PORT=8443
DATA_PATH=~/dnt-vault/data
CONFIG_PATH=~/dnt-vault/config
```

## Run as a systemd Service

```bash
sudo tee /etc/systemd/system/dnt-vault.service << 'EOF'
[Unit]
Description=DNT-Vault SSH Config Sync Server
After=network.target

[Service]
Type=simple
Environment="PORT=8443"
ExecStart=/usr/local/bin/dnt-vault-server
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now dnt-vault
```

## Architecture

```
┌─────────────┐      HTTP/HTTPS       ┌──────────────────┐
│  dnt-vault  │ ────── encrypted ───► │ dnt-vault-server │
│    (CLI)    │ ◄──── data only ────  │   (REST API)     │
└─────────────┘                       └──────────────────┘
```

- `server/`: REST API vault — stores encrypted blobs, JWT auth, filesystem storage.
- `cli/`: CLI tool — encrypts locally, pushes/pulls via HTTP.
- `shared/`: Common types shared between server and CLI.

## Build from Source

Requirements: Go 1.22+

```bash
./build.sh
# server/bin/dnt-vault-server
# cli/bin/dnt-vault
```

## API

```
POST   /api/v1/auth/login           # Login → JWT token
GET    /api/v1/profiles             # List profiles        [auth]
GET    /api/v1/profiles/:name       # Get profile data     [auth]
POST   /api/v1/profiles/:name       # Save profile         [auth]
DELETE /api/v1/profiles/:name       # Delete profile       [auth]
```

## Troubleshooting

**Server won't start** — check port: `lsof -i :8443`

**Login fails** — verify URL in `~/.dnt-vault/config.yaml`, check server is up: `curl http://localhost:8443/api/v1/profiles`

**Decryption fails** — master password must match what was used during `push`

## License

MIT
