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
		SSHMasterKeyFile string `yaml:"ssh_master_key_file"`
		EnvMasterKeyFile string `yaml:"env_master_key_file"`
		MasterKeyFile    string `yaml:"master_key_file,omitempty"`
	} `yaml:"encryption"`
	Env struct {
		BackupDir string `yaml:"backup_dir"`
	} `yaml:"env"`
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

	cfg.NormalizePaths(homeDir)

	return &cfg, nil
}

func (cfg *AppConfig) NormalizePaths(homeDir string) {
	configDir := filepath.Join(homeDir, ".dnt-vault")

	if cfg.Encryption.SSHMasterKeyFile == "" && cfg.Encryption.MasterKeyFile != "" {
		cfg.Encryption.SSHMasterKeyFile = cfg.Encryption.MasterKeyFile
	}

	if cfg.Encryption.SSHMasterKeyFile == "" {
		cfg.Encryption.SSHMasterKeyFile = filepath.Join(configDir, "ssh-master.key")
	}

	if cfg.Encryption.EnvMasterKeyFile == "" {
		cfg.Encryption.EnvMasterKeyFile = filepath.Join(configDir, "env-master.key")
	}

	if cfg.Backup.Dir == "" {
		cfg.Backup.Dir = filepath.Join(configDir, "backups", "ssh")
	}

	if cfg.Env.BackupDir == "" {
		cfg.Env.BackupDir = filepath.Join(configDir, "backups", "env")
	}
}

func (cfg *AppConfig) SSHMasterKeyPath() string {
	if cfg.Encryption.SSHMasterKeyFile != "" {
		return cfg.Encryption.SSHMasterKeyFile
	}
	return cfg.Encryption.MasterKeyFile
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
