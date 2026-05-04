# Release Process

This document describes the release process for DNT-Vault.

## Prerequisites

- Go 1.22+
- GitHub CLI (`gh`)
- Git
- Write access to the repository

## Release Steps

### 1. Prepare Release

```bash
# Update version in files
VERSION="1.0.1"

# Update CHANGELOG.md
vim CHANGELOG.md

# Commit changes
git add CHANGELOG.md
git commit -m "Release: v$VERSION"
git push origin master
```

### 2. Build Binaries

Run the release script:

```bash
./scripts/release.sh $VERSION
```

This will:
- Build binaries for all platforms
- Create checksums
- Package releases

### 3. Create GitHub Release

```bash
# Create release with binaries
gh release create v$VERSION \
  --title "v$VERSION" \
  --notes-file RELEASE_NOTES.md \
  releases/*
```

### 4. Verify Release

```bash
# Test installation
curl -fsSL https://raw.githubusercontent.com/dungnt1312/dnt-vault/master/install.sh | bash

# Test binaries
ssh-sync --version
dnt-vault-server --version
```

## Supported Platforms

### Linux
- `linux-amd64`
- `linux-arm64`
- `linux-arm`

### macOS
- `darwin-amd64`
- `darwin-arm64`

### Windows
- `windows-amd64.exe`
- `windows-386.exe`

## Version Numbering

Follow Semantic Versioning (semver):
- `MAJOR.MINOR.PATCH`
- Example: `1.0.0`, `1.1.0`, `2.0.0`

### When to increment:

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

## Release Checklist

- [ ] Update CHANGELOG.md
- [ ] Update version in code
- [ ] Run tests: `./test.sh`
- [ ] Build all platforms: `./scripts/release.sh`
- [ ] Create git tag
- [ ] Push to GitHub
- [ ] Create GitHub release
- [ ] Upload binaries
- [ ] Test installation script
- [ ] Update documentation if needed
- [ ] Announce release

## Hotfix Process

For urgent fixes:

```bash
# Create hotfix branch
git checkout -b hotfix/v1.0.1

# Make fixes
git commit -m "Fix: critical bug"

# Merge to master
git checkout master
git merge hotfix/v1.0.1

# Release
./scripts/release.sh 1.0.1

# Delete hotfix branch
git branch -d hotfix/v1.0.1
```

## Rollback

If a release has issues:

```bash
# Delete release
gh release delete v1.0.1

# Delete tag
git tag -d v1.0.1
git push origin :refs/tags/v1.0.1

# Revert commit if needed
git revert <commit-hash>
git push origin master
```
