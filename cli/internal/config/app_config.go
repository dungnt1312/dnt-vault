package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Server struct {
		URL       string `yaml:"url"`
		TLSVerify bool   `yaml:"tls_verify"`
	} `yaml:"server"`
	SSH struct {
		ConfigPath string `yaml:"config_path"`
		KeysDir    string `yaml:"keys_dir"`
	} `yaml:"ssh"`
	Profiles struct {
		Current           string `yaml:"current"`
		DefaultNameFormat string `yaml:"default_name_format"`
	} `yaml:"profiles"`
	Backup struct {
		Enabled    bool   `yaml:"enabled"`
		Dir        string `yaml:"dir"`
		MaxBackups int    `yaml:"max_backups"`
	} `yaml:"backup"`
	Encryption struct {
		MasterKeyFile string `yaml:"master_key_file"`
	} `yaml:"encryption"`
}

func LoadAppConfig() (*AppConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(homeDir, ".dnt-vault", "config.yaml")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func SaveAppConfig(cfg *AppConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".dnt-vault")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.yaml")
	return os.WriteFile(configFile, data, 0600)
}
