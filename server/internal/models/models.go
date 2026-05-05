package models

import "time"

type Profile struct {
	Name       string    `json:"name"`
	Hostname   string    `json:"hostname"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	HasKeys    bool      `json:"has_keys"`
	KeyCount   int       `json:"key_count"`
	ConfigHash string    `json:"config_hash"`
}

type ProfileData struct {
	Profile Profile           `json:"profile"`
	Config  string            `json:"config"`
	Verify  string            `json:"verify,omitempty"`
	Keys    map[string]string `json:"keys,omitempty"`
	KeysIV  map[string]string `json:"keys_iv,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ProfileListResponse struct {
	Profiles []Profile `json:"profiles"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type User struct {
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserStore struct {
	Users []User `json:"users"`
}
