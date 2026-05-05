package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/dnt/vault-server/internal/models"
)

type FilesystemStorage struct {
	basePath string
}

func NewFilesystemStorage(basePath string) (*FilesystemStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	return &FilesystemStorage{
		basePath: basePath,
	}, nil
}

func (fs *FilesystemStorage) getUserPath(username string) string {
	return filepath.Join(fs.basePath, username)
}

func (fs *FilesystemStorage) getProfilePath(username, profileName string) string {
	return filepath.Join(fs.getUserPath(username), profileName)
}

func (fs *FilesystemStorage) SaveProfile(username string, data models.ProfileData) error {
	userPath := fs.getUserPath(username)
	if err := os.MkdirAll(userPath, 0755); err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp(userPath, ".tmp-"+data.Profile.Name+"-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	metadataData, err := json.MarshalIndent(data.Profile, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "metadata.json"), metadataData, 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "config.enc"), []byte(data.Config), 0644); err != nil {
		return err
	}

	if data.Verify != "" {
		if err := os.WriteFile(filepath.Join(tmpDir, "verify.enc"), []byte(data.Verify), 0644); err != nil {
			return err
		}
	}

	if len(data.Keys) > 0 {
		keysPath := filepath.Join(tmpDir, "keys")
		if err := os.MkdirAll(keysPath, 0755); err != nil {
			return err
		}

		for keyName, keyData := range data.Keys {
			keyPath := filepath.Join(keysPath, keyName+".enc")
			if err := os.WriteFile(keyPath, []byte(keyData), 0644); err != nil {
				return err
			}
		}

		keysIVData, err := json.Marshal(data.KeysIV)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "keys_iv.json"), keysIVData, 0644); err != nil {
			return err
		}
	}

	profilePath := fs.getProfilePath(username, data.Profile.Name)
	if _, err := os.Stat(profilePath); err == nil {
		if err := os.RemoveAll(profilePath); err != nil {
			return err
		}
	}

	return os.Rename(tmpDir, profilePath)
}

func (fs *FilesystemStorage) GetProfile(username, name string) (*models.ProfileData, error) {
	profilePath := fs.getProfilePath(username, name)

	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return nil, errors.New("profile not found")
	}

	metadataPath := filepath.Join(profilePath, "metadata.json")
	metadataData, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var profile models.Profile
	if err := json.Unmarshal(metadataData, &profile); err != nil {
		return nil, err
	}

	configPath := filepath.Join(profilePath, "config.enc")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	verifyPath := filepath.Join(profilePath, "verify.enc")
	verifyData, _ := os.ReadFile(verifyPath) // ignore error if file doesn't exist

	data := &models.ProfileData{
		Profile: profile,
		Config:  string(configData),
		Verify:  string(verifyData),
		Keys:    make(map[string]string),
		KeysIV:  make(map[string]string),
	}

	keysPath := filepath.Join(profilePath, "keys")
	if _, err := os.Stat(keysPath); err == nil {
		entries, err := os.ReadDir(keysPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			keyName := entry.Name()
			if len(keyName) > 4 && keyName[len(keyName)-4:] == ".enc" {
				keyName = keyName[:len(keyName)-4]
			}

			keyPath := filepath.Join(keysPath, entry.Name())
			keyData, err := os.ReadFile(keyPath)
			if err != nil {
				return nil, err
			}

			data.Keys[keyName] = string(keyData)
		}

		keysIVPath := filepath.Join(profilePath, "keys_iv.json")
		if keysIVData, err := os.ReadFile(keysIVPath); err == nil {
			if err := json.Unmarshal(keysIVData, &data.KeysIV); err != nil {
				return nil, err
			}
		}
	}

	return data, nil
}

func (fs *FilesystemStorage) ListProfiles(username string) ([]models.Profile, error) {
	userPath := fs.getUserPath(username)

	if _, err := os.Stat(userPath); os.IsNotExist(err) {
		return []models.Profile{}, nil
	}

	entries, err := os.ReadDir(userPath)
	if err != nil {
		return nil, err
	}

	var profiles []models.Profile
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".tmp-") {
			continue
		}

		metadataPath := filepath.Join(userPath, entry.Name(), "metadata.json")
		metadataData, err := os.ReadFile(metadataPath)
		if err != nil {
			continue
		}

		var profile models.Profile
		if err := json.Unmarshal(metadataData, &profile); err != nil {
			continue
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (fs *FilesystemStorage) DeleteProfile(username, name string) error {
	profilePath := fs.getProfilePath(username, name)

	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return errors.New("profile not found")
	}

	return os.RemoveAll(profilePath)
}

func (fs *FilesystemStorage) ProfileExists(username, name string) bool {
	profilePath := fs.getProfilePath(username, name)
	_, err := os.Stat(profilePath)
	return err == nil
}
