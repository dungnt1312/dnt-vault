package storage

import "github.com/dnt/vault-server/internal/models"

type Storage interface {
	SaveProfile(username string, data models.ProfileData) error
	GetProfile(username, name string) (*models.ProfileData, error)
	ListProfiles(username string) ([]models.Profile, error)
	DeleteProfile(username, name string) error
	ProfileExists(username, name string) bool

	SaveEnvScope(username, scope string, data models.EnvScopeData) error
	GetEnvScope(username, scope string) (*models.EnvScopeData, error)
	ListEnvScopes(username string) ([]models.EnvScope, error)
	DeleteEnvScope(username, scope string) error
	EnvScopeExists(username, scope string) bool
}
