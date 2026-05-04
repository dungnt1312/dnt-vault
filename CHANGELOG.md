# Changelog

All notable changes to DNT-Vault will be documented in this file.

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

## [Unreleased]

### Planned
- TLS/HTTPS support
- Profile versioning
- Config validation
- Web UI
- Additional storage backends
