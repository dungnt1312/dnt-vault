package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dnt/vault-cli/internal/config"
	"github.com/dnt/vault-cli/internal/interactive"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize dnt-vault configuration",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".dnt-vault")
	configFile := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configFile); err == nil {
		overwrite, err := interactive.PromptConfirm("Configuration already exists. Overwrite?", false)
		if err != nil {
			return err
		}
		if !overwrite {
			return nil
		}
	}

	fmt.Println(color.CyanString("Welcome to DNT-Vault SSH Config Sync!\n"))

	fmt.Println(color.YellowString("Server Setup:"))
	serverURL := os.Getenv("DNT_VAULT_SERVER_URL")
	if serverURL == "" {
		serverURL, err = interactive.PromptString("Server URL", "http://localhost:8443")
		if err != nil {
			return err
		}
	}

	fmt.Println(color.GreenString("\n✓ Server configured: %s\n", serverURL))

	fmt.Println(color.YellowString("Master Password Setup:"))
	fmt.Println("This password encrypts your SSH configs.")
	masterPassword := os.Getenv("DNT_VAULT_MASTER_PASSWORD")
	confirmPassword := os.Getenv("DNT_VAULT_MASTER_PASSWORD_CONFIRM")
	if masterPassword == "" {
		masterPassword, err = interactive.PromptPassword("Enter master password")
		if err != nil {
			return err
		}
		confirmPassword, err = interactive.PromptPassword("Confirm password")
		if err != nil {
			return err
		}
	} else if confirmPassword == "" {
		confirmPassword = masterPassword
	}

	if masterPassword != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	sshMasterKeyFile := filepath.Join(configDir, "ssh-master.key")
	if err := os.WriteFile(sshMasterKeyFile, []byte(masterPassword), 0600); err != nil {
		return err
	}

	cfg := config.AppConfig{}
	cfg.Server.URL = serverURL
	cfg.Server.TLSVerify = true
	cfg.SSH.ConfigPath = filepath.Join(homeDir, ".ssh", "config")
	cfg.SSH.KeysDir = filepath.Join(homeDir, ".ssh")
	cfg.Profiles.DefaultNameFormat = "{hostname}"
	cfg.Backup.Enabled = true
	cfg.Backup.Dir = filepath.Join(configDir, "backups", "ssh")
	cfg.Backup.MaxBackups = 10
	cfg.Encryption.SSHMasterKeyFile = sshMasterKeyFile
	cfg.Encryption.EnvMasterKeyFile = filepath.Join(configDir, "env-master.key")
	cfg.Env.BackupDir = filepath.Join(configDir, "backups", "env")

	if err := config.SaveAppConfig(&cfg); err != nil {
		return err
	}

	fmt.Println(color.GreenString("\n✓ SSH master key generated and saved to %s", sshMasterKeyFile))
	fmt.Println(color.GreenString("✓ Configuration saved to %s", filepath.Join(configDir, "config.yaml")))
	fmt.Println(color.CyanString("\nRun 'dnt-vault login' to authenticate with the vault."))

	return nil
}
