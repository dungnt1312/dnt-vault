# DNT-Vault v1.1.1

## Installation

### Linux / macOS
```bash
curl -fsSL https://raw.githubusercontent.com/dungnt1312/dnt-vault/master/install.sh | bash
```

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/dungnt1312/dnt-vault/master/install.ps1 | iex
```

### Manual Installation

Download the appropriate binary for your platform:

**Server:**
- Linux (amd64): `dnt-vault-server-linux-amd64`
- Linux (arm64): `dnt-vault-server-linux-arm64`
- macOS (Intel): `dnt-vault-server-darwin-amd64`
- macOS (Apple Silicon): `dnt-vault-server-darwin-arm64`
- Windows (64-bit): `dnt-vault-server-windows-amd64.exe`

**CLI:**
- Linux (amd64): `dnt-vault-linux-amd64`
- Linux (arm64): `dnt-vault-linux-arm64`
- macOS (Intel): `dnt-vault-darwin-amd64`
- macOS (Apple Silicon): `dnt-vault-darwin-arm64`
- Windows (64-bit): `dnt-vault-windows-amd64.exe`

Make the binary executable:
```bash
chmod +x dnt-vault-server-*
chmod +x dnt-vault-*
```

Move to PATH:
```bash
sudo mv dnt-vault-server-* /usr/local/bin/dnt-vault-server
sudo mv dnt-vault-* /usr/local/bin/dnt-vault
```

## Checksums

See `checksums.txt` for SHA256 checksums of all binaries.

## Documentation

- [README](https://github.com/dungnt1312/dnt-vault/blob/master/README.md)
- [Quick Start](https://github.com/dungnt1312/dnt-vault/blob/master/QUICKSTART.md)
- [Demo](https://github.com/dungnt1312/dnt-vault/blob/master/DEMO.md)

## What's Changed

See [CHANGELOG.md](https://github.com/dungnt1312/dnt-vault/blob/master/CHANGELOG.md)
