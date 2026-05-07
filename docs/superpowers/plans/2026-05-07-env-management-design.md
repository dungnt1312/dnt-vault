# Environment Variables Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a complete `dnt-vault env` feature set for encrypted environment variable sync (client + server), isolated from existing SSH sync.

**Architecture:** Reuse the current SSH sync architecture patterns: client-side encryption, authenticated REST API, and filesystem-backed server storage. Add a parallel env subsystem with its own API namespace (`/api/v1/env`), CLI command group (`dnt-vault env`), separate client master key file, and dedicated storage subtree under each user.

**Tech Stack:** Go 1.22, cobra (CLI), gorilla/mux (server routing), AES-256-GCM + PBKDF2 (crypto), YAML config, Go testing package, shell integration test.

---

### File Structure Map

**CLI files**
- Modify: `cli/cmd/cli/main.go` (root command description update)
- Modify: `cli/cmd/cli/init.go` (SSH key file rename migration)
- Create: `cli/cmd/cli/env.go` (env command group)
- Create: `cli/cmd/cli/env_init.go`
- Create: `cli/cmd/cli/env_push.go`
- Create: `cli/cmd/cli/env_pull.go`
- Create: `cli/cmd/cli/env_list.go`
- Create: `cli/cmd/cli/env_get.go`
- Create: `cli/cmd/cli/env_set.go`
- Create: `cli/cmd/cli/env_delete.go`
- Create: `cli/internal/envmanager/envmanager.go` (encryption/decryption + merge helpers)
- Create: `cli/internal/envmanager/parser.go` (.env parser)
- Create: `cli/internal/envmanager/validator.go` (scope/name validation + URL scope conversion)
- Modify: `cli/internal/client/client.go` (env API methods + DTOs + retry helper)
- Modify: `cli/internal/config/app_config.go` (separate SSH/Env encryption config + backup dirs)
- Modify: `cli/internal/backup/backup.go` (support env file backup)
- Create: `cli/internal/envmanager/envmanager_test.go`
- Create: `cli/internal/envmanager/parser_test.go`
- Create: `cli/internal/envmanager/validator_test.go`

**Server files**
- Modify: `server/internal/models/models.go` (env models)
- Modify: `server/internal/storage/storage.go` (env storage interface)
- Modify: `server/internal/storage/filesystem.go` (env storage implementation)
- Modify: `server/internal/api/router.go` (env routes)
- Modify: `server/internal/api/handlers.go` (env handlers)
- Create: `server/internal/envapi/helpers.go` (optional extraction for env handler logic if handlers grow too large)
- Create: `server/internal/api/handlers_env_test.go`
- Create: `server/internal/storage/filesystem_env_test.go`

**Test/scripts/docs**
- Modify: `test.sh` (add end-to-end env workflow)
- Create: `test-env.sh` (focused env integration path)
- Modify: `README.md` (document env commands + migration behavior)

---

### Task 1: Add Env Config Model and SSH Key Migration

**Files:**
- Modify: `cli/internal/config/app_config.go`
- Modify: `cli/cmd/cli/init.go`
- Test: `cli/internal/config/app_config_test.go` (create)

- [ ] **Step 1: Write failing tests for config schema + migration behavior**

```go
func TestLoadAppConfig_WithLegacyMasterKey_PopulatesSSHMasterKey(t *testing.T) {
    // Arrange: config.yaml has encryption.master_key_file only
    // Act: load config and run migration helper
    // Assert: SSH master key path preserved, env key path defaulted
}

func TestInit_CreatesSSHMasterKeyFile(t *testing.T) {
    // Assert init writes ssh-master.key and not master.key for new installs
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./internal/config -run 'TestLoadAppConfig_WithLegacyMasterKey|TestInit_CreatesSSHMasterKeyFile' -v`

Expected: FAIL (missing new config fields/migration behavior).

- [ ] **Step 3: Implement config schema split + migration support**

```go
type AppConfig struct {
    // ...existing...
    Encryption struct {
        SSHMasterKeyFile string `yaml:"ssh_master_key_file"`
        EnvMasterKeyFile string `yaml:"env_master_key_file"`
        // Legacy fallback
        MasterKeyFile string `yaml:"master_key_file,omitempty"`
    } `yaml:"encryption"`
    Env struct {
        BackupDir string `yaml:"backup_dir"`
    } `yaml:"env"`
}

func (cfg *AppConfig) NormalizePaths(homeDir string) {
    if cfg.Encryption.SSHMasterKeyFile == "" && cfg.Encryption.MasterKeyFile != "" {
        cfg.Encryption.SSHMasterKeyFile = cfg.Encryption.MasterKeyFile
    }
    if cfg.Encryption.SSHMasterKeyFile == "" {
        cfg.Encryption.SSHMasterKeyFile = filepath.Join(homeDir, ".dnt-vault", "ssh-master.key")
    }
    if cfg.Encryption.EnvMasterKeyFile == "" {
        cfg.Encryption.EnvMasterKeyFile = filepath.Join(homeDir, ".dnt-vault", "env-master.key")
    }
}
```

- [ ] **Step 4: Update init flow for SSH key rename behavior**

```go
sshMasterKeyFile := filepath.Join(configDir, "ssh-master.key")
if err := os.WriteFile(sshMasterKeyFile, []byte(masterPassword), 0600); err != nil {
    return err
}
cfg.Encryption.SSHMasterKeyFile = sshMasterKeyFile
cfg.Encryption.EnvMasterKeyFile = filepath.Join(configDir, "env-master.key")
cfg.Backup.Dir = filepath.Join(configDir, "backups", "ssh")
cfg.Env.BackupDir = filepath.Join(configDir, "backups", "env")
```

- [ ] **Step 5: Run tests to verify pass**

Run: `cd cli && go test ./internal/config ./cmd/cli -v`

Expected: PASS for new tests and no regression on existing code.

- [ ] **Step 6: Commit**

```bash
git add cli/internal/config/app_config.go cli/cmd/cli/init.go cli/internal/config/app_config_test.go
git commit -m "refactor: split encryption key config for ssh and env"
```

---

### Task 2: Build Env Domain Utilities (Validation, Parser, Crypto Workflow)

**Files:**
- Create: `cli/internal/envmanager/validator.go`
- Create: `cli/internal/envmanager/parser.go`
- Create: `cli/internal/envmanager/envmanager.go`
- Test: `cli/internal/envmanager/validator_test.go`
- Test: `cli/internal/envmanager/parser_test.go`
- Test: `cli/internal/envmanager/envmanager_test.go`

- [ ] **Step 1: Write failing validator tests**

```go
func TestValidateScopeName(t *testing.T) {
    valid := []string{"global", "global/aws", "myapp/production", "my_app/staging"}
    invalid := []string{"", "../escape", "myapp production", "myapp#prod"}
}

func TestScopeURLConversion(t *testing.T) {
    require.Equal(t, "myapp-production", ScopeToURL("myapp/production"))
    require.Equal(t, "myapp/production", ScopeFromURL("myapp-production"))
}
```

- [ ] **Step 2: Write failing parser tests**

```go
func TestParseEnvFile(t *testing.T) {
    // Covers comments, empty lines, quoted values, duplicate key error
}
```

- [ ] **Step 3: Write failing encryption roundtrip tests**

```go
func TestEncryptDecryptVariablesRoundtrip(t *testing.T) {}
func TestDecryptVariables_WrongPasswordFails(t *testing.T) {}
func TestDecryptVariables_TamperedCiphertextFails(t *testing.T) {}
```

- [ ] **Step 4: Run tests to verify they fail**

Run: `cd cli && go test ./internal/envmanager -v`

Expected: FAIL due to missing implementation.

- [ ] **Step 5: Implement validator/parser/envmanager minimal code**

```go
var scopePattern = regexp.MustCompile(`^[a-zA-Z0-9_\-/]+$`)

func ValidateScope(scope string) error { /* enforce non-empty + regex */ }

func ParseEnvFile(path string) (map[string]string, error) { /* parse KEY=VALUE lines */ }

func EncryptVariables(vars map[string]string, password string) (string, error) {
    payload, _ := json.Marshal(vars)
    return crypto.Encrypt(string(payload), password)
}

func DecryptVariables(blob, password string) (map[string]string, error) {
    plaintext, _ := crypto.Decrypt(blob, password)
    var vars map[string]string
    json.Unmarshal([]byte(plaintext), &vars)
    return vars, nil
}
```

- [ ] **Step 6: Run envmanager tests to pass**

Run: `cd cli && go test ./internal/envmanager -v`

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add cli/internal/envmanager
git commit -m "feat: add env parsing validation and encryption helpers"
```

---

### Task 3: Extend Server Models + Storage Interface for Env Scopes

**Files:**
- Modify: `server/internal/models/models.go`
- Modify: `server/internal/storage/storage.go`
- Modify: `server/internal/storage/filesystem.go`
- Test: `server/internal/storage/filesystem_env_test.go`

- [ ] **Step 1: Write failing storage tests for env scope CRUD**

```go
func TestFilesystemStorage_SaveAndGetEnvScope(t *testing.T) {}
func TestFilesystemStorage_ListEnvScopes(t *testing.T) {}
func TestFilesystemStorage_DeleteEnvScope(t *testing.T) {}
func TestFilesystemStorage_DeleteEnvKey(t *testing.T) {}
```

- [ ] **Step 2: Run tests to confirm fail**

Run: `cd server && go test ./internal/storage -run Env -v`

Expected: FAIL (methods/types missing).

- [ ] **Step 3: Add env models + storage interface methods**

```go
type EnvScopeMetadata struct {
    Name          string    `json:"name"`
    VariableCount int       `json:"variable_count"`
    UpdatedAt     time.Time `json:"updated_at"`
    Hostname      string    `json:"hostname"`
}

type EnvScopeData struct {
    Metadata  EnvScopeMetadata   `json:"metadata"`
    Variables map[string]string  `json:"variables"`
    Verify    string             `json:"verify,omitempty"`
}

type Storage interface {
    // existing profile methods...
    SaveEnvScope(username, scope string, data models.EnvScopeData) error
    GetEnvScope(username, scope string) (*models.EnvScopeData, error)
    ListEnvScopes(username string) ([]models.EnvScopeMetadata, error)
    DeleteEnvScope(username, scope string) error
    SetEnvVariable(username, scope, key, value string, metadata models.EnvScopeMetadata) error
    DeleteEnvVariable(username, scope, key string, metadata models.EnvScopeMetadata) error
    EnvScopeExists(username, scope string) bool
}
```

- [ ] **Step 4: Implement filesystem env storage layout**

```go
func (fs *FilesystemStorage) getEnvScopePath(username, scope string) string {
    return filepath.Join(fs.getUserPath(username), "env", scope)
}

// Persist:
// metadata.json
// variables.enc (JSON map key->encrypted-value OR encrypted full blob per design choice)
// verify.enc (optional)
```

- [ ] **Step 5: Run storage tests to pass**

Run: `cd server && go test ./internal/storage -v`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add server/internal/models/models.go server/internal/storage/storage.go server/internal/storage/filesystem.go server/internal/storage/filesystem_env_test.go
git commit -m "feat: add env scope storage model and filesystem implementation"
```

---

### Task 4: Add Env API Endpoints and Handler Logic

**Files:**
- Modify: `server/internal/api/router.go`
- Modify: `server/internal/api/handlers.go`
- Test: `server/internal/api/handlers_env_test.go`

- [ ] **Step 1: Write failing handler tests for all env endpoints**

```go
func TestListEnvScopes_AuthRequired(t *testing.T) {}
func TestGetEnvScope_ReturnsScopeData(t *testing.T) {}
func TestSaveEnvScope_Upserts(t *testing.T) {}
func TestSetEnvVariable_UpdatesSingleKey(t *testing.T) {}
func TestDeleteEnvVariable_RemovesKey(t *testing.T) {}
func TestDeleteEnvScope_RemovesAll(t *testing.T) {}
```

- [ ] **Step 2: Run API tests to verify fail**

Run: `cd server && go test ./internal/api -run Env -v`

Expected: FAIL due to missing routes/handlers.

- [ ] **Step 3: Register env routes in protected router**

```go
protected.HandleFunc("/env/scopes", handler.ListEnvScopes).Methods("GET")
protected.HandleFunc("/env/scopes/{scope}", handler.GetEnvScope).Methods("GET")
protected.HandleFunc("/env/scopes/{scope}", handler.SaveEnvScope).Methods("POST")
protected.HandleFunc("/env/scopes/{scope}", handler.DeleteEnvScope).Methods("DELETE")
protected.HandleFunc("/env/scopes/{scope}/{key}", handler.GetEnvVariable).Methods("GET")
protected.HandleFunc("/env/scopes/{scope}/{key}", handler.SetEnvVariable).Methods("PUT")
protected.HandleFunc("/env/scopes/{scope}/{key}", handler.DeleteEnvVariable).Methods("DELETE")
```

- [ ] **Step 4: Implement handlers with consistent error style**

```go
func (h *Handler) SaveEnvScope(w http.ResponseWriter, r *http.Request) {
    // decode body, set UpdatedAt, compute VariableCount, persist via storage
}

func (h *Handler) GetEnvVariable(w http.ResponseWriter, r *http.Request) {
    // load scope, return one key, 404 if missing
}
```

- [ ] **Step 5: Run API tests to pass**

Run: `cd server && go test ./internal/api -v`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add server/internal/api/router.go server/internal/api/handlers.go server/internal/api/handlers_env_test.go
git commit -m "feat: add env scope API endpoints under /api/v1/env"
```

---

### Task 5: Extend CLI HTTP Client with Env Methods + Retry

**Files:**
- Modify: `cli/internal/client/client.go`
- Test: `cli/internal/client/client_env_test.go` (create)

- [ ] **Step 1: Write failing client tests for env API calls**

```go
func TestClient_ListEnvScopes(t *testing.T) {}
func TestClient_SaveEnvScope(t *testing.T) {}
func TestClient_SetEnvVariable(t *testing.T) {}
func TestClient_RetryOnTransientNetworkFailure(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd cli && go test ./internal/client -run Env -v`

Expected: FAIL (methods not defined).

- [ ] **Step 3: Add env DTOs + methods matching API contract**

```go
type EnvScope struct {
    Name          string    `json:"name"`
    VariableCount int       `json:"variable_count"`
    UpdatedAt     time.Time `json:"updated_at"`
    Hostname      string    `json:"hostname"`
}

type EnvScopeData struct {
    Metadata  EnvScope             `json:"metadata"`
    Variables map[string]string    `json:"variables"`
    Verify    string               `json:"verify,omitempty"`
}

func (c *Client) ListEnvScopes() ([]EnvScope, error) {}
func (c *Client) GetEnvScope(scope string) (*EnvScopeData, error) {}
func (c *Client) SaveEnvScope(scope string, data EnvScopeData) error {}
func (c *Client) SetEnvVariable(scope, key, value string) error {}
func (c *Client) DeleteEnvVariable(scope, key string) error {}
```

- [ ] **Step 4: Add retry wrapper for transient transport errors (3 attempts, backoff)**

```go
func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
    // retry on net errors / 5xx for idempotent operations, exponential backoff
}
```

- [ ] **Step 5: Run client tests to pass**

Run: `cd cli && go test ./internal/client -v`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cli/internal/client/client.go cli/internal/client/client_env_test.go
git commit -m "feat: add env API client methods with retry support"
```

---

### Task 6: Implement `dnt-vault env init` and Command Group Scaffolding

**Files:**
- Create: `cli/cmd/cli/env.go`
- Create: `cli/cmd/cli/env_init.go`
- Modify: `cli/cmd/cli/main.go`
- Test: `cli/cmd/cli/env_init_test.go` (create)

- [ ] **Step 1: Write failing CLI tests for env init**

```go
func TestEnvInit_CreatesEnvMasterKeyFile(t *testing.T) {}
func TestEnvInit_RejectsMismatchedPasswords(t *testing.T) {}
func TestEnvInit_RequiresExistingBaseConfig(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd cli && go test ./cmd/cli -run EnvInit -v`

Expected: FAIL (command missing).

- [ ] **Step 3: Implement env command root and init command**

```go
var envCmd = &cobra.Command{
    Use:   "env",
    Short: "Manage encrypted environment variables",
}

var envInitCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize env encryption key",
    RunE:  runEnvInit,
}

func runEnvInit(cmd *cobra.Command, args []string) error {
    // prompt password twice
    // write ~/.dnt-vault/env-master.key (0600)
    // update config with encryption.env_master_key_file
}
```

- [ ] **Step 4: Run command tests to pass**

Run: `cd cli && go test ./cmd/cli -run EnvInit -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cli/cmd/cli/main.go cli/cmd/cli/env.go cli/cmd/cli/env_init.go cli/cmd/cli/env_init_test.go
git commit -m "feat: add env command group and env init"
```

---

### Task 7: Implement `env push` (interactive + file + merge/replace)

**Files:**
- Create: `cli/cmd/cli/env_push.go`
- Modify: `cli/internal/envmanager/envmanager.go`
- Test: `cli/cmd/cli/env_push_test.go` (create)

- [ ] **Step 1: Write failing tests for push modes and merge behavior**

```go
func TestEnvPush_FromFile_MergesByDefault(t *testing.T) {}
func TestEnvPush_FromFile_ReplaceOverwritesWhenFlagSet(t *testing.T) {}
func TestEnvPush_InteractiveRejectsDuplicateKeys(t *testing.T) {}
func TestEnvPush_EmptyInputReturnsNoVariablesError(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd cli && go test ./cmd/cli -run EnvPush -v`

Expected: FAIL.

- [ ] **Step 3: Implement push command logic**

```go
var envPushCmd = &cobra.Command{
    Use:   "push <scope>",
    Short: "Push env variables to a scope",
    Args:  cobra.ExactArgs(1),
    RunE:  runEnvPush,
}

// Flags:
// --file <path>
// --replace

func runEnvPush(cmd *cobra.Command, args []string) error {
    // validate scope
    // load vars from file or interactive input
    // fetch existing scope for merge unless replace
    // encrypt variables via envmanager
    // send POST /api/v1/env/scopes/:scope
}
```

- [ ] **Step 4: Run push tests to pass**

Run: `cd cli && go test ./cmd/cli -run EnvPush -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cli/cmd/cli/env_push.go cli/internal/envmanager/envmanager.go cli/cmd/cli/env_push_test.go
git commit -m "feat: implement env push with merge and replace modes"
```

---

### Task 8: Implement `env pull` (shell export + output file + backup)

**Files:**
- Create: `cli/cmd/cli/env_pull.go`
- Modify: `cli/internal/backup/backup.go`
- Test: `cli/cmd/cli/env_pull_test.go` (create)

- [ ] **Step 1: Write failing tests for output modes**

```go
func TestEnvPull_PrintsExportLines(t *testing.T) {}
func TestEnvPull_OutputFileCreatesBackupBeforeOverwrite(t *testing.T) {}
func TestEnvPull_WrongPasswordShowsDecryptionError(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd cli && go test ./cmd/cli -run EnvPull -v`

Expected: FAIL.

- [ ] **Step 3: Implement pull command**

```go
var envPullCmd = &cobra.Command{
    Use:   "pull <scope>",
    Short: "Pull env variables from a scope",
    Args:  cobra.ExactArgs(1),
    RunE:  runEnvPull,
}

// Flag: --output <path>

func runEnvPull(cmd *cobra.Command, args []string) error {
    // fetch scope from server
    // decrypt using env master key
    // if output flag: backup file + write KEY=VALUE lines
    // else print export KEY="VALUE" lines
}
```

- [ ] **Step 4: Run pull tests to pass**

Run: `cd cli && go test ./cmd/cli -run EnvPull -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cli/cmd/cli/env_pull.go cli/internal/backup/backup.go cli/cmd/cli/env_pull_test.go
git commit -m "feat: implement env pull for shell exports and file output"
```

---

### Task 9: Implement `env list/get/set/delete`

**Files:**
- Create: `cli/cmd/cli/env_list.go`
- Create: `cli/cmd/cli/env_get.go`
- Create: `cli/cmd/cli/env_set.go`
- Create: `cli/cmd/cli/env_delete.go`
- Test: `cli/cmd/cli/env_commands_test.go` (create)

- [ ] **Step 1: Write failing tests for each command behavior**

```go
func TestEnvList_NoScope_ListsScopes(t *testing.T) {}
func TestEnvList_WithScope_ListsKeysOnly(t *testing.T) {}
func TestEnvGet_ReturnsPlaintextValue(t *testing.T) {}
func TestEnvSet_UpdatesSingleVariable(t *testing.T) {}
func TestEnvDelete_KeyOnly_DeletesVariable(t *testing.T) {}
func TestEnvDelete_All_ConfirmsAndDeletesScope(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd cli && go test ./cmd/cli -run Env(List|Get|Set|Delete) -v`

Expected: FAIL.

- [ ] **Step 3: Implement command handlers**

```go
// env list [scope]
// env get <scope> <key>
// env set <scope> <key> <value>
// env delete <scope> <key>
// env delete <scope> --all
```

- [ ] **Step 4: Run command tests to pass**

Run: `cd cli && go test ./cmd/cli -run Env -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cli/cmd/cli/env_list.go cli/cmd/cli/env_get.go cli/cmd/cli/env_set.go cli/cmd/cli/env_delete.go cli/cmd/cli/env_commands_test.go
git commit -m "feat: add env list get set delete commands"
```

---

### Task 10: Wire Server Single-Variable Operations

**Files:**
- Modify: `server/internal/api/handlers.go`
- Modify: `server/internal/storage/filesystem.go`
- Test: `server/internal/api/handlers_env_test.go`
- Test: `server/internal/storage/filesystem_env_test.go`

- [ ] **Step 1: Write failing tests for PUT/GET/DELETE key-level behavior**

```go
func TestSetEnvVariable_CreatesScopeIfMissing(t *testing.T) {}
func TestGetEnvVariable_NotFound(t *testing.T) {}
func TestDeleteEnvVariable_UpdatesVariableCount(t *testing.T) {}
```

- [ ] **Step 2: Run tests to verify fail**

Run: `cd server && go test ./internal/api ./internal/storage -run EnvVariable -v`

Expected: FAIL.

- [ ] **Step 3: Implement key-level mutation logic**

```go
func (h *Handler) SetEnvVariable(w http.ResponseWriter, r *http.Request) {
    // load existing scope, mutate one key, persist, return success
}

func (h *Handler) DeleteEnvVariable(w http.ResponseWriter, r *http.Request) {
    // remove key and persist, 404 if key absent
}
```

- [ ] **Step 4: Run tests to pass**

Run: `cd server && go test ./internal/api ./internal/storage -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add server/internal/api/handlers.go server/internal/storage/filesystem.go server/internal/api/handlers_env_test.go server/internal/storage/filesystem_env_test.go
git commit -m "feat: implement key-level env variable API operations"
```

---

### Task 11: Integration Test Coverage for End-to-End Env Workflow

**Files:**
- Create: `test-env.sh`
- Modify: `test.sh`

- [ ] **Step 1: Write integration script for env flow**

```bash
# test-env.sh high-level flow
# 1) build binaries
# 2) start server with temp DATA_PATH/CONFIG_PATH
# 3) dnt-vault init
# 4) dnt-vault login
# 5) dnt-vault env init
# 6) dnt-vault env push myapp/production --file .env.test
# 7) dnt-vault env list
# 8) eval $(dnt-vault env pull myapp/production)
# 9) dnt-vault env set myapp/production KEY updated
# 10) dnt-vault env get myapp/production KEY
# 11) dnt-vault env delete myapp/production KEY
# 12) dnt-vault env delete myapp/production --all
```

- [ ] **Step 2: Execute env integration test and verify fails initially**

Run: `bash test-env.sh`

Expected: FAIL at first missing feature.

- [ ] **Step 3: Update script expectations after implementation complete**

```bash
# assert command outputs and exit codes
# assert backup file created for --output overwrite scenario
```

- [ ] **Step 4: Run full integration tests**

Run: `bash test.sh && bash test-env.sh`

Expected: PASS both scripts.

- [ ] **Step 5: Commit**

```bash
git add test.sh test-env.sh
git commit -m "test: add end-to-end env management integration coverage"
```

---

### Task 12: Documentation and Final Verification

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Write README updates for env command usage**

```markdown
## Environment Variables Sync

### Initialize
`dnt-vault env init`

### Push
`dnt-vault env push <scope> --file .env`

### Pull
`eval $(dnt-vault env pull <scope>)`
`dnt-vault env pull <scope> --output .env`

### Manage Keys
`dnt-vault env list [scope]`
`dnt-vault env get <scope> <key>`
`dnt-vault env set <scope> <key> <value>`
`dnt-vault env delete <scope> <key>`
`dnt-vault env delete <scope> --all`
```

- [ ] **Step 2: Add migration note for `master.key` -> `ssh-master.key`**

```markdown
On upgrade, legacy `encryption.master_key_file` is treated as SSH key path.
New installs write `~/.dnt-vault/ssh-master.key` and `~/.dnt-vault/env-master.key`.
```

- [ ] **Step 3: Run full test and lint/build verification**

Run: `make test`

Run: `make build`

Expected: All tests pass and both binaries build.

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: document env management commands and key migration"
```

---

## Self-Review Checklist (completed)

- Spec coverage: all major sections mapped (architecture/model, API, CLI workflows, encryption separation, edge cases, testing, backward compatibility).
- Placeholder scan: removed generic TODO language; each task includes concrete files, commands, and expected outcomes.
- Type consistency: aligned naming to `EnvScope`, `EnvScopeData`, `variables`, `metadata`, and `/api/v1/env/scopes` route family.

## Notes and Decisions to Lock Before Execution

- Server encrypted payload format choice:
  - **Option A (recommended for this repo):** keep per-key encrypted values (`variables` map with encrypted base64 values) as spec API examples show.
  - **Option B:** encrypt whole JSON blob once and store one blob string; simpler crypto path but diverges from request/response examples.
- This plan assumes **Option A** to match API examples and variable-level operations naturally.
