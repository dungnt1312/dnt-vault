package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dnt/vault-server/internal/auth"
	"github.com/dnt/vault-server/internal/models"
	"github.com/dnt/vault-server/internal/storage"
	"github.com/gorilla/mux"
)

type Handler struct {
	storage storage.Storage
	auth    *auth.AuthService
}

func NewHandler(storage storage.Storage, auth *auth.AuthService) *Handler {
	return &Handler{
		storage: storage,
		auth:    auth,
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, expiresAt, err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials || err == auth.ErrUserNotFound {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

func (h *Handler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(usernameKey).(string)

	profiles, err := h.storage.ListProfiles(username)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list profiles")
		return
	}

	respondJSON(w, http.StatusOK, models.ProfileListResponse{
		Profiles: profiles,
	})
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(usernameKey).(string)
	vars := mux.Vars(r)
	profileName := vars["name"]

	data, err := h.storage.GetProfile(username, profileName)
	if err != nil {
		respondError(w, http.StatusNotFound, "profile not found")
		return
	}

	respondJSON(w, http.StatusOK, data)
}

func (h *Handler) SaveProfile(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(usernameKey).(string)
	vars := mux.Vars(r)
	profileName := vars["name"]

	var data models.ProfileData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data.Profile.Name = profileName
	data.Profile.UpdatedAt = time.Now()

	if !h.storage.ProfileExists(username, profileName) {
		data.Profile.CreatedAt = time.Now()
	}

	if err := h.storage.SaveProfile(username, data); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save profile")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"profile": data.Profile,
	})
}

func (h *Handler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(usernameKey).(string)
	vars := mux.Vars(r)
	profileName := vars["name"]

	if err := h.storage.DeleteProfile(username, profileName); err != nil {
		respondError(w, http.StatusNotFound, "profile not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListEnvScopes(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(usernameKey).(string)
	scopes, err := h.storage.ListEnvScopes(username)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list scopes")
		return
	}
	respondJSON(w, http.StatusOK, models.EnvScopeListResponse{Scopes: scopes})
}

func (h *Handler) GetEnvScope(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(usernameKey).(string)
	scope := mux.Vars(r)["scope"]
	data, err := h.storage.GetEnvScope(username, scope)
	if err != nil {
		respondError(w, http.StatusNotFound, "scope not found")
		return
	}
	respondJSON(w, http.StatusOK, data)
}

func (h *Handler) SaveEnvScope(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(usernameKey).(string)
	scope := mux.Vars(r)["scope"]

	var data models.EnvScopeData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data.Metadata.Name = scope
	data.Metadata.UpdatedAt = time.Now()
	data.Metadata.VariableCount = len(data.Variables)
	if data.Metadata.Hostname == "" {
		data.Metadata.Hostname = "unknown"
	}

	if err := h.storage.SaveEnvScope(username, scope, data); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save scope")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"success": true, "scope": data.Metadata})
}

func (h *Handler) DeleteEnvScope(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(usernameKey).(string)
	scope := mux.Vars(r)["scope"]
	if err := h.storage.DeleteEnvScope(username, scope); err != nil {
		respondError(w, http.StatusNotFound, "scope not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetEnvVariable(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(usernameKey).(string)
	vars := mux.Vars(r)
	scope := vars["scope"]
	key := vars["key"]
	data, err := h.storage.GetEnvScope(username, scope)
	if err != nil {
		respondError(w, http.StatusNotFound, "scope not found")
		return
	}
	value, ok := data.Variables[key]
	if !ok {
		respondError(w, http.StatusNotFound, "variable not found")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"value": value})
}

func (h *Handler) SetEnvVariable(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(usernameKey).(string)
	vars := mux.Vars(r)
	scope := vars["scope"]
	key := vars["key"]

	var payload struct {
		Value    string `json:"value"`
		Hostname string `json:"hostname"`
		Verify   string `json:"verify,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := h.storage.GetEnvScope(username, scope)
	if err != nil {
		data = &models.EnvScopeData{Metadata: models.EnvScope{Name: scope}, Variables: map[string]string{}}
	}
	if data.Variables == nil {
		data.Variables = map[string]string{}
	}
	data.Variables[key] = payload.Value
	data.Metadata.UpdatedAt = time.Now()
	data.Metadata.Name = scope
	data.Metadata.VariableCount = len(data.Variables)
	if payload.Hostname != "" {
		data.Metadata.Hostname = payload.Hostname
	}
	if data.Metadata.Hostname == "" {
		data.Metadata.Hostname = "unknown"
	}
	if payload.Verify != "" {
		data.Verify = payload.Verify
	}

	if err := h.storage.SaveEnvScope(username, scope, *data); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to set variable")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (h *Handler) DeleteEnvVariable(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(usernameKey).(string)
	vars := mux.Vars(r)
	scope := vars["scope"]
	key := vars["key"]

	data, err := h.storage.GetEnvScope(username, scope)
	if err != nil {
		respondError(w, http.StatusNotFound, "scope not found")
		return
	}
	if _, ok := data.Variables[key]; !ok {
		respondError(w, http.StatusNotFound, "variable not found")
		return
	}
	delete(data.Variables, key)
	data.Metadata.UpdatedAt = time.Now()
	data.Metadata.VariableCount = len(data.Variables)

	if err := h.storage.SaveEnvScope(username, scope, *data); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete variable")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("deleted %s", key)})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, models.ErrorResponse{Error: message})
}
