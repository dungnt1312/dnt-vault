package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListEnvScopes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/env/scopes" || r.Method != "GET" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(EnvScopeListResponse{
			Scopes: []EnvScope{{Name: "global", VariableCount: 2}},
		})
	}))
	defer srv.Close()

	c := &Client{baseURL: srv.URL, httpClient: srv.Client(), token: "tok"}
	scopes, err := c.ListEnvScopes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scopes) != 1 || scopes[0].Name != "global" {
		t.Fatalf("unexpected scopes: %#v", scopes)
	}
}

func TestSaveEnvScope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/env/scopes/myapp-prod" || r.Method != "POST" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var data EnvScopeData
		json.NewDecoder(r.Body).Decode(&data)
		if data.Variables["A"] != "enc" {
			t.Fatalf("unexpected body: %#v", data)
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer srv.Close()

	c := &Client{baseURL: srv.URL, httpClient: srv.Client(), token: "tok"}
	err := c.SaveEnvScope("myapp-prod", EnvScopeData{Variables: map[string]string{"A": "enc"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetEnvVariable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/env/scopes/global/TOKEN" || r.Method != "PUT" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer srv.Close()

	c := &Client{baseURL: srv.URL, httpClient: srv.Client(), token: "tok"}
	err := c.SetEnvVariable("global", "TOKEN", "enc_val", "host", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteEnvVariable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/env/scopes/global/TOKEN" || r.Method != "DELETE" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "deleted"})
	}))
	defer srv.Close()

	c := &Client{baseURL: srv.URL, httpClient: srv.Client(), token: "tok"}
	err := c.DeleteEnvVariable("global", "TOKEN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryOnServerError(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(EnvScopeListResponse{Scopes: []EnvScope{}})
	}))
	defer srv.Close()

	c := &Client{baseURL: srv.URL, httpClient: srv.Client(), token: "tok"}
	_, err := c.ListEnvScopes()
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
