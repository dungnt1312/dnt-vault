package envmanager

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "# comment\nA=1\nB=\"two\"\n\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	vars, err := ParseEnvFile(path)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if vars["A"] != "1" || vars["B"] != "two" {
		t.Fatalf("unexpected parse result: %#v", vars)
	}
}

func TestParseEnvFile_DuplicateKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "A=1\nA=2\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	if _, err := ParseEnvFile(path); err == nil {
		t.Fatalf("expected duplicate key error")
	}
}
