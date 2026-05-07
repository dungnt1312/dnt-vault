package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizePaths_WithLegacyMasterKey_PopulatesSSHMasterKey(t *testing.T) {
	home := "/tmp/test-home"
	cfg := &AppConfig{}
	cfg.Encryption.MasterKeyFile = "/legacy/master.key"

	cfg.NormalizePaths(home)

	if cfg.Encryption.SSHMasterKeyFile != "/legacy/master.key" {
		t.Fatalf("expected legacy key path to be used for SSH, got %q", cfg.Encryption.SSHMasterKeyFile)
	}

	wantEnv := filepath.Join(home, ".dnt-vault", "env-master.key")
	if cfg.Encryption.EnvMasterKeyFile != wantEnv {
		t.Fatalf("expected env master key path %q, got %q", wantEnv, cfg.Encryption.EnvMasterKeyFile)
	}
}

func TestInitStyleConfig_CreatesSSHMasterKeyFilePath(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".dnt-vault")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	cfg := &AppConfig{}
	cfg.NormalizePaths(home)

	wantSSH := filepath.Join(configDir, "ssh-master.key")
	if cfg.SSHMasterKeyPath() != wantSSH {
		t.Fatalf("expected ssh master key path %q, got %q", wantSSH, cfg.SSHMasterKeyPath())
	}
}
