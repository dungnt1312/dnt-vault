# Changelog

All notable changes to DNT-Vault will be documented in this file.

## [1.1.3] - 2026-05-05

### Added
- `dnt-vault profile list` — list profiles with current profile marker
- `dnt-vault profile use <name>` — pull and apply a profile, set as current

### Changed
- `dnt-vault list` deprecated (use `dnt-vault profile list`)

## [1.1.2] - 2026-05-05

### Added
- `dnt-vault upgrade` — self-update CLI from GitHub Releases
- Master password verification token on push, validated before pull
- `install.sh` auto-detects latest version from GitHub API
- Makefile with build, test, release targets
- GitHub Actions CI/CD for automated cross-platform releases

### Fixed
- Pull now shows clear "wrong master password" error instead of cryptic cipher error
- Windows Git Bash install now auto-adds to PATH and sources `.bashrc`

## [1.0.0] - 2026-05-04

### Added
- Initial release of DNT-Vault SSH Config Sync
- REST API vault server with JWT authentication
- CLI tool for push/pull operations
- Client-side encryption (AES-256-GCM)
- Private key sync with passphrase protection
- Conflict detection with diff display
- Automatic backup before pull
- Interactive profile selection
- Multi-user support
- Profile management (list, delete)
- Filesystem-based storage
- Comprehensive documentation

### Security
- PBKDF2 key derivation (100k iterations)
- bcrypt password hashing
- JWT token authentication (24h expiry)
- Secure file permissions (0600/0700)
- Client-side encryption (server never sees plaintext)

### Documentation
- README.md - Main documentation
- QUICKSTART.md - Quick start guide
- DEMO.md - Detailed demo walkthrough
- FEATURES.md - Feature list and technical details
- STRUCTURE.md - Project structure
- LICENSE - MIT License

### Scripts
- build.sh - Build both server and CLI
- test.sh - Integration tests

## [1.1.0] - 2026-05-04

### Added
- Version command for CLI (`dnt-vault version`) with build-time version embedding via ldflags
- Version info logged at server startup
- Server graceful shutdown on SIGINT/SIGTERM with 30s drain timeout
- Per-IP login rate limiting (5 attempts/minute) to protect against brute-force
- Request body size limit (10MB) to prevent OOM attacks
- Structured request logging: method, path, status code, duration

### Changed
- `build.sh` now injects Version, BuildTime, CommitSHA via `-ldflags` automatically from git tag
- `AppConfig` struct and `LoadAppConfig()` moved to `cli/internal/config` package (previously duplicated across cmd files)
- `http.Server` now has ReadTimeout (15s), WriteTimeout (15s), IdleTimeout (60s)
- Diff algorithm replaced with LCS-based implementation (was positional, now correct)
- `context.WithValue` uses typed key instead of plain string

### Fixed
- **Critical**: `SaveProfile` now writes atomically via temp dir + `os.Rename` — prevents corrupted profiles on crash mid-write
- `ListProfiles` now skips leftover `.tmp-*` directories

## [Unreleased]

### Planned
- TLS/HTTPS support
- Profile versioning
- Config validation
- Web UI
- Additional storage backends
