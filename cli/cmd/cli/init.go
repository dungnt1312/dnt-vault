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
	Short: "Initialize ssh-sync configuration",
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

	configDir := filepath.Join(homeDir, ".ssh-sync")
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
	serverURL, err := interactive.PromptString("Server URL", "http://localhost:8443")
	if err != nil {
		return err
	}

	fmt.Println(color.GreenString("\n✓ Server configured: %s\n", serverURL))

	fmt.Println(color.YellowString("Master Password Setup:"))
	fmt.Println("This password encrypts your SSH configs.")
	masterPassword, err := interactive.PromptPassword("Enter master password")
	if err != nil {
		return err
	}

	confirmPassword, err := interactive.PromptPassword("Confirm password")
	if err != nil {
		return err
	}

	if masterPassword != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	masterKeyFile := filepath.Join(configDir, "master.key")
	if err := os.WriteFile(masterKeyFile, []byte(masterPassword), 0600); err != nil {
		return err
	}

	cfg := config.AppConfig{}
	cfg.Server.URL = serverURL
	cfg.Server.TLSVerify = true
	cfg.SSH.ConfigPath = filepath.Join(homeDir, ".ssh", "config")
	cfg.SSH.KeysDir = filepath.Join(homeDir, ".ssh")
	cfg.Profiles.DefaultNameFormat = "{hostname}"
	cfg.Backup.Enabled = true
	cfg.Backup.Dir = filepath.Join(configDir, "backups")
	cfg.Backup.MaxBackups = 10
	cfg.Encryption.MasterKeyFile = masterKeyFile

	if err := config.SaveAppConfig(&cfg); err != nil {
		return err
	}

	fmt.Println(color.GreenString("\n✓ Master key generated and saved to %s", masterKeyFile))
	fmt.Println(color.GreenString("✓ Configuration saved to %s", filepath.Join(configDir, "config.yaml")))
	fmt.Println(color.CyanString("\nRun 'ssh-sync login' to authenticate with the vault."))

	return nil
}
