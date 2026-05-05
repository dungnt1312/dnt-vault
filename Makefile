.PHONY: all build build-cli build-server test clean release

VERSION ?= $(shell git describe --tags --always --dirty | sed 's/^v//')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X 'main.Version=v$(VERSION)' \
	-X 'main.BuildTime=$(BUILD_TIME)' \
	-X 'main.CommitSHA=$(COMMIT)'

all: build

build: build-cli build-server
	@echo "✓ Built version $(VERSION)"

build-cli:
	cd cli && go build -ldflags "$(LDFLAGS)" -o ../bin/dnt-vault ./cmd/cli
	@echo "✓ CLI built: bin/dnt-vault"

build-server:
	cd server && go build -ldflags "$(LDFLAGS)" -o ../bin/dnt-vault-server ./cmd/server
	@echo "✓ Server built: bin/dnt-vault-server"

test:
	cd cli && go test ./...
	cd server && go test ./...
	@echo "✓ Tests passed"

clean:
	rm -rf bin/ releases/*
	@echo "✓ Cleaned"

release: build
	@bash scripts/release.sh $(VERSION)
