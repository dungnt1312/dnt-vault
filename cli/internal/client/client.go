package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	tokenFile  string
}

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

func NewClient(baseURL, tokenFile string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		tokenFile:  tokenFile,
	}
}

func (c *Client) LoadToken() error {
	data, err := os.ReadFile(c.tokenFile)
	if err != nil {
		return err
	}
	c.token = string(data)
	return nil
}

func (c *Client) SaveToken(token string) error {
	dir := filepath.Dir(c.tokenFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(c.tokenFile, []byte(token), 0600)
}

func (c *Client) Login(username, password string) error {
	req := LoginRequest{
		Username: username,
		Password: password,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/auth/login",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: %s", resp.Status)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	c.token = loginResp.Token
	return c.SaveToken(loginResp.Token)
}

func (c *Client) ListProfiles() ([]Profile, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/profiles", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list profiles: %s", resp.Status)
	}

	var listResp ProfileListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	return listResp.Profiles, nil
}

func (c *Client) GetProfile(name string) (*ProfileData, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/profiles/"+name, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get profile: %s", resp.Status)
	}

	var data ProfileData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (c *Client) SaveProfile(name string, data ProfileData) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		c.baseURL+"/api/v1/profiles/"+name,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to save profile: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func (c *Client) DeleteProfile(name string) error {
	req, err := http.NewRequest("DELETE", c.baseURL+"/api/v1/profiles/"+name, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete profile: %s", resp.Status)
	}

	return nil
}
