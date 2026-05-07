# Environment Variables Management - Design Spec

## Overview

Thêm feature quản lý environment variables vào dnt-vault. Cho phép users sync env variables (API keys, database URLs, tokens) giữa các machines, tách biệt với SSH config hiện tại.

**Key decisions:**
- Hierarchical organization: `project/environment` structure
- Variable-level sync: Sync từng biến, export ra shell hoặc file
- Separate encryption: Env có master password riêng, tách biệt SSH
- Separate command group: `dnt-vault env <subcommand>`

---

## 1. Architecture & Data Model

### Hierarchical Structure

```
global/                          # Shared secrets
├── aws/
│   ├── AWS_ACCESS_KEY_ID
│   ├── AWS_SECRET_ACCESS_KEY
│   └── AWS_REGION
├── github/
│   └── GITHUB_TOKEN
└── npm/
    └── NPM_TOKEN

<project-name>/                  # Project-specific
├── production/
│   ├── DATABASE_URL
│   ├── API_KEY
│   └── REDIS_URL
├── staging/
│   ├── DATABASE_URL
│   └── API_KEY
└── development/
    └── DATABASE_URL
```

**Scope format:** `<project>[/<environment>]`
- `global` - shared secrets
- `global/aws` - AWS credentials
- `myapp/production` - myapp production env
- `myapp` - all myapp environments

### Server Storage Structure

```
/var/lib/dnt-vault/data/
└── {username}/
    ├── ssh/                     # Existing SSH profiles
    │   └── {profile-name}/
    └── env/                     # Environment variables
        └── {scope}/             # e.g., "global", "myapp-production"
            ├── metadata.json    # Scope metadata
            └── variables.enc    # Encrypted env variables (JSON)
```

### Client Storage Structure

```
~/.dnt-vault/
├── config.yaml              # Existing
├── ssh-master.key           # Renamed from master.key
├── env-master.key           # Separate encryption for env
├── token                    # Existing
└── backups/
    ├── ssh/                 # SSH config backups
    └── env/                 # Env backups
```

---

## 2. API Endpoints

All endpoints require JWT authentication (`Authorization: Bearer <token>`).

```
GET    /api/v1/env/scopes                    # List all scopes
GET    /api/v1/env/scopes/:scope             # Get all variables in scope
POST   /api/v1/env/scopes/:scope             # Create/update scope with variables
DELETE /api/v1/env/scopes/:scope             # Delete entire scope
GET    /api/v1/env/scopes/:scope/:key        # Get specific variable
PUT    /api/v1/env/scopes/:scope/:key        # Set specific variable
DELETE /api/v1/env/scopes/:scope/:key        # Delete specific variable
```

### Request/Response Formats

```json
// POST /api/v1/env/scopes/myapp-production
{
  "variables": {
    "DATABASE_URL": "encrypted_value_base64",
    "API_KEY": "encrypted_value_base64",
    "REDIS_URL": "encrypted_value_base64"
  },
  "metadata": {
    "hostname": "macbook-pro",
    "updated_at": "2026-05-07T14:22:11Z"
  }
}

// GET /api/v1/env/scopes
{
  "scopes": [
    {
      "name": "global",
      "variable_count": 5,
      "updated_at": "2026-05-07T10:00:00Z",
      "hostname": "macbook-pro"
    },
    {
      "name": "myapp-production",
      "variable_count": 10,
      "updated_at": "2026-05-07T14:22:11Z",
      "hostname": "server-01"
    }
  ]
}
```

**Note:** Scope names sử dụng `-` thay vì `/` trong URL (e.g., `myapp/production` → `myapp-production`).

---

## 3. CLI Commands & Workflows

### Commands

```bash
# Initialize env encryption (separate from SSH)
dnt-vault env init
# → Prompts for env master password
# → Generates env-master.key
# → Updates config.yaml with env settings

# Push variables to a scope (merge: update existing, add new, keep untouched)
dnt-vault env push <scope>
# → Prompts: Enter variables (KEY=VALUE, empty line to finish)
# → If scope exists: merges with existing variables
# → If scope doesn't exist: creates new scope

# Push from file (merge behavior, use --replace to overwrite all)
dnt-vault env push <scope> --file .env
# → Reads .env file, encrypts, uploads
# → --replace flag: overwrite entire scope instead of merge

# Pull and inject to current shell
dnt-vault env pull <scope>
# → Outputs: export KEY1="value1"\nexport KEY2="value2"
# → Usage: eval $(dnt-vault env pull myapp/production)

# Pull to file
dnt-vault env pull <scope> --output .env
# → Downloads and writes to .env file
# → Creates backup before overwrite

# List all scopes
dnt-vault env list
# → Shows: global, myapp/production, myapp/staging

# List variables in a scope (keys only, no values)
dnt-vault env list <scope>
# → Shows: DATABASE_URL, API_KEY, REDIS_URL (updated 2 hours ago)

# Get specific variable
dnt-vault env get <scope> <key>
# → Outputs: value (plaintext, for piping)

# Set single variable
dnt-vault env set <scope> <key> <value>
# → Encrypts and uploads single variable

# Delete variable
dnt-vault env delete <scope> <key>
# → Deletes single variable from scope

# Delete entire scope
dnt-vault env delete <scope> --all
# → Prompts for confirmation, deletes all variables
```

### Typical Workflows

```bash
# Setup (one-time)
dnt-vault env init

# Developer workflow - push from local .env
cd ~/projects/myapp
dnt-vault env push myapp/production --file .env.production

# DevOps workflow - pull to CI/CD
eval $(dnt-vault env pull myapp/production)
npm run build

# Quick variable update
dnt-vault env set myapp/production DATABASE_URL "postgres://new-host/db"

# New machine setup
dnt-vault env pull myapp/production --output .env
```

---

## 4. Encryption & Security

### Separate Encryption Keys

```
SSH Config:     ssh master password → ssh-master.key → AES-256-GCM
Env Variables:  env master password → env-master.key → AES-256-GCM
```

### Encryption Flow

1. **Key Derivation:**
   - PBKDF2(password, salt, 100k iterations, SHA-256) → 32-byte key
   - Store derived key in `~/.dnt-vault/env-master.key` (0600 permissions)

2. **Variable Encryption (client-side):**
   - Collect all variables in scope as JSON: `{"KEY1": "value1", "KEY2": "value2"}`
   - Generate random 32-byte salt
   - Generate random 12-byte nonce
   - AES-256-GCM encrypt JSON with derived key
   - Format: `[salt(32)][nonce(12)][ciphertext][tag(16)]`
   - Base64 encode entire blob

3. **Server Storage:**
   - Server receives encrypted blob (cannot decrypt)
   - Stores in `{username}/env/{scope}/variables.enc`
   - Stores metadata separately (unencrypted): hostname, timestamp, variable count

4. **Decryption (client-side):**
   - Download encrypted blob
   - Base64 decode → extract salt, nonce, ciphertext, tag
   - Derive key from env master password + salt
   - AES-256-GCM decrypt → JSON → variables

### Security Properties

- Zero-knowledge: Server never sees plaintext
- Authenticated encryption: GCM tag prevents tampering
- Unique encryption per push: Random salt + nonce
- Separate blast radius: SSH and env use different keys
- Password verification: Store PBKDF2 hash on server for validation before pull

---

## 5. Error Handling & Edge Cases

| Scenario | Error Message |
|----------|--------------|
| Env not initialized | "Env encryption not initialized. Run: dnt-vault env init" |
| Wrong env password | "Decryption failed. Incorrect env master password." |
| Scope not found | "Scope 'myapp/production' not found. Available scopes: ..." |
| Duplicate variable keys | "Duplicate variable KEY_NAME in input" |
| Network failure | Auto-retry 3x with exponential backoff, then error |
| File overwrite | Prompt confirmation, create backup before overwrite |
| Invalid scope name | "Invalid scope name. Use alphanumeric, dash, underscore, slash only" |
| Empty variables | "No variables provided. Scope not created." |
| Token expired | "Session expired. Please run: dnt-vault login" |

---

## 6. Testing Strategy

### Unit Tests

```
cli/internal/envmanager/
├── envmanager_test.go          # Encryption/decryption logic
├── parser_test.go              # .env file parsing
└── validator_test.go           # Scope name validation

server/internal/envapi/
├── handlers_test.go            # API endpoint handlers
└── storage_test.go             # Env storage operations
```

### Key Test Cases

1. **Encryption/Decryption:** Encrypt → Decrypt roundtrip, wrong password fails, tampered data fails
2. **Scope Management:** Create, list, delete scopes, invalid names rejected
3. **Variable Operations:** Set/get/update/delete single variables
4. **.env File Parsing:** Comments, empty lines, quoted values, duplicate keys, multiline values
5. **API Endpoints:** Auth required, correct HTTP status codes, JSON format

### Integration Tests (test-env.sh)

Full workflow: server start → init → login → env init → push → pull → verify → set/get → delete → cleanup

---

## 7. Backward Compatibility

- Existing SSH commands (`push`, `pull`, `list`, `delete`) unchanged
- Rename `master.key` → `ssh-master.key` with migration on first run
- Old clients without env feature continue to work normally
- Server API versioned under `/api/v1/env/` namespace
