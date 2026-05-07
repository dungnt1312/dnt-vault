package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dnt/vault-server/internal/models"
	"github.com/gorilla/mux"
)

type mockStorage struct {
	scopes map[string]*models.EnvScopeData
}

func newMockStorage() *mockStorage {
	return &mockStorage{scopes: make(map[string]*models.EnvScopeData)}
}

func (m *mockStorage) SaveProfile(_ string, _ models.ProfileData) error          { return nil }
func (m *mockStorage) GetProfile(_, _ string) (*models.ProfileData, error)       { return nil, nil }
func (m *mockStorage) ListProfiles(_ string) ([]models.Profile, error)           { return nil, nil }
func (m *mockStorage) DeleteProfile(_, _ string) error                           { return nil }
func (m *mockStorage) ProfileExists(_, _ string) bool                            { return false }

func (m *mockStorage) SaveEnvScope(username, scope string, data models.EnvScopeData) error {
	key := username + "/" + scope
	m.scopes[key] = &data
	return nil
}

func (m *mockStorage) GetEnvScope(username, scope string) (*models.EnvScopeData, error) {
	key := username + "/" + scope
	d, ok := m.scopes[key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return d, nil
}

func (m *mockStorage) ListEnvScopes(username string) ([]models.EnvScope, error) {
	var out []models.EnvScope
	prefix := username + "/"
	for k, v := range m.scopes {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			out = append(out, v.Metadata)
		}
	}
	return out, nil
}

func (m *mockStorage) DeleteEnvScope(username, scope string) error {
	key := username + "/" + scope
	if _, ok := m.scopes[key]; !ok {
		return fmt.Errorf("not found")
	}
	delete(m.scopes, key)
	return nil
}

func (m *mockStorage) EnvScopeExists(username, scope string) bool {
	key := username + "/" + scope
	_, ok := m.scopes[key]
	return ok
}

func makeAuthRequest(method, path string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), usernameKey, "testuser")
	return req.WithContext(ctx)
}

func TestListEnvScopes_Empty(t *testing.T) {
	store := newMockStorage()
	h := NewHandler(store, nil)

	req := makeAuthRequest("GET", "/api/v1/env/scopes", nil)
	rr := httptest.NewRecorder()
	h.ListEnvScopes(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp models.EnvScopeListResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Scopes) != 0 {
		t.Fatalf("expected empty scopes, got %d", len(resp.Scopes))
	}
}

func TestSaveAndGetEnvScope(t *testing.T) {
	store := newMockStorage()
	h := NewHandler(store, nil)

	data := models.EnvScopeData{
		Metadata:  models.EnvScope{Hostname: "host1"},
		Variables: map[string]string{"A": "enc1"},
	}

	req := makeAuthRequest("POST", "/api/v1/env/scopes/myapp-prod", data)
	req = mux.SetURLVars(req, map[string]string{"scope": "myapp-prod"})
	rr := httptest.NewRecorder()
	h.SaveEnvScope(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("save: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	req2 := makeAuthRequest("GET", "/api/v1/env/scopes/myapp-prod", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"scope": "myapp-prod"})
	rr2 := httptest.NewRecorder()
	h.GetEnvScope(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", rr2.Code)
	}
	var got models.EnvScopeData
	json.NewDecoder(rr2.Body).Decode(&got)
	if got.Variables["A"] != "enc1" {
		t.Fatalf("unexpected variables: %#v", got.Variables)
	}
}

func TestSetAndGetEnvVariable(t *testing.T) {
	store := newMockStorage()
	h := NewHandler(store, nil)

	payload := map[string]string{"value": "enc_val", "hostname": "h1"}
	req := makeAuthRequest("PUT", "/api/v1/env/scopes/global/TOKEN", payload)
	req = mux.SetURLVars(req, map[string]string{"scope": "global", "key": "TOKEN"})
	rr := httptest.NewRecorder()
	h.SetEnvVariable(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("set: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	req2 := makeAuthRequest("GET", "/api/v1/env/scopes/global/TOKEN", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"scope": "global", "key": "TOKEN"})
	rr2 := httptest.NewRecorder()
	h.GetEnvVariable(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", rr2.Code)
	}
	var out struct{ Value string }
	json.NewDecoder(rr2.Body).Decode(&out)
	if out.Value != "enc_val" {
		t.Fatalf("expected enc_val, got %q", out.Value)
	}
}

func TestDeleteEnvVariable(t *testing.T) {
	store := newMockStorage()
	h := NewHandler(store, nil)

	store.scopes["testuser/global"] = &models.EnvScopeData{
		Metadata:  models.EnvScope{Name: "global", VariableCount: 2, UpdatedAt: time.Now(), Hostname: "h"},
		Variables: map[string]string{"A": "1", "B": "2"},
	}

	req := makeAuthRequest("DELETE", "/api/v1/env/scopes/global/A", nil)
	req = mux.SetURLVars(req, map[string]string{"scope": "global", "key": "A"})
	rr := httptest.NewRecorder()
	h.DeleteEnvVariable(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	got := store.scopes["testuser/global"]
	if _, ok := got.Variables["A"]; ok {
		t.Fatalf("expected A deleted")
	}
	if got.Metadata.VariableCount != 1 {
		t.Fatalf("expected count 1, got %d", got.Metadata.VariableCount)
	}
}

func TestDeleteEnvScope(t *testing.T) {
	store := newMockStorage()
	h := NewHandler(store, nil)

	store.scopes["testuser/global"] = &models.EnvScopeData{
		Metadata:  models.EnvScope{Name: "global"},
		Variables: map[string]string{"A": "1"},
	}

	req := makeAuthRequest("DELETE", "/api/v1/env/scopes/global", nil)
	req = mux.SetURLVars(req, map[string]string{"scope": "global"})
	rr := httptest.NewRecorder()
	h.DeleteEnvScope(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if store.EnvScopeExists("testuser", "global") {
		t.Fatalf("scope should be deleted")
	}
}
