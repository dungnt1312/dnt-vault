dnt-vault/
├── README.md                    # Documentation
├── .gitignore                   # Git ignore rules
├── build.sh                     # Build script
├── test.sh                      # Integration test script
│
├── server/                      # Vault server
│   ├── go.mod
│   ├── go.sum
│   ├── bin/
│   │   └── dnt-vault-server     # Server binary
│   ├── cmd/
│   │   └── server/
│   │       └── main.go          # Server entry point
│   └── internal/
│       ├── api/
│       │   ├── handlers.go      # HTTP request handlers
│       │   ├── middleware.go    # Auth, logging, CORS
│       │   └── router.go        # Route definitions
│       ├── auth/
│       │   └── auth.go          # JWT authentication
│       ├── storage/
│       │   ├── storage.go       # Storage interface
│       │   └── filesystem.go    # File-based storage
│       └── models/
│           └── models.go        # Data structures
│
├── cli/                         # Client CLI
│   ├── go.mod
│   ├── go.sum
│   ├── bin/
│   │   └── ssh-sync             # CLI binary
│   ├── cmd/
│   │   └── cli/
│   │       ├── main.go          # CLI entry point
│   │       ├── init.go          # init command
│   │       ├── auth.go          # login/logout commands
│   │       ├── push.go          # push command
│   │       ├── pull.go          # pull command
│   │       ├── list.go          # list command
│   │       └── delete.go        # delete command
│   ├── internal/
│   │   ├── client/
│   │   │   └── client.go        # HTTP client
│   │   ├── config/
│   │   │   ├── parser.go        # SSH config parser
│   │   │   └── differ.go        # Diff generator
│   │   ├── crypto/
│   │   │   └── crypto.go        # Encryption/decryption
│   │   ├── backup/
│   │   │   └── backup.go        # Backup management
│   │   └── interactive/
│   │       └── prompt.go        # Interactive prompts
│   └── pkg/
│       └── models/              # Shared models
│
└── shared/                      # Shared code
    └── models/
        └── types.go             # Common types
