package storage

import (
	"testing"
	"time"

	"github.com/dnt/vault-server/internal/models"
)

func TestFilesystemStorage_SaveAndGetEnvScope(t *testing.T) {
	fs, err := NewFilesystemStorage(t.TempDir())
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}

	data := models.EnvScopeData{
		Metadata: models.EnvScope{Name: "myapp-production", VariableCount: 1, UpdatedAt: time.Now(), Hostname: "host"},
		Variables: map[string]string{"A": "enc"},
		Verify:    "verify",
	}

	if err := fs.SaveEnvScope("u", "myapp-production", data); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := fs.GetEnvScope("u", "myapp-production")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Metadata.Name != "myapp-production" || got.Variables["A"] != "enc" {
		t.Fatalf("unexpected scope data: %#v", got)
	}
}

func TestFilesystemStorage_ListAndDeleteEnvScope(t *testing.T) {
	fs, err := NewFilesystemStorage(t.TempDir())
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}

	data := models.EnvScopeData{
		Metadata: models.EnvScope{Name: "global", VariableCount: 1, UpdatedAt: time.Now(), Hostname: "host"},
		Variables: map[string]string{"A": "enc"},
	}
	if err := fs.SaveEnvScope("u", "global", data); err != nil {
		t.Fatalf("save: %v", err)
	}

	scopes, err := fs.ListEnvScopes("u")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(scopes) != 1 || scopes[0].Name != "global" {
		t.Fatalf("unexpected scopes: %#v", scopes)
	}

	if err := fs.DeleteEnvScope("u", "global"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if fs.EnvScopeExists("u", "global") {
		t.Fatalf("scope should not exist")
	}
}
