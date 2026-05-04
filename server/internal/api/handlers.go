package api

import (
	"encoding/json"
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

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, models.ErrorResponse{Error: message})
}
