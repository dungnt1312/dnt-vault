package main

import (
	"fmt"
	"os"

	appconfig "github.com/dnt/vault-cli/internal/config"
	"github.com/dnt/vault-cli/internal/interactive"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var envInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize env encryption key",
	RunE:  runEnvInit,
}

func init() {
	envCmd.AddCommand(envInitCmd)
}

func runEnvInit(cmd *cobra.Command, args []string) error {
	cfg, err := appconfig.LoadAppConfig()
	if err != nil {
		return fmt.Errorf("config not found. Run 'dnt-vault init' first")
	}

	fmt.Println(color.YellowString("Env Master Password Setup:"))
	password := os.Getenv("DNT_VAULT_ENV_MASTER_PASSWORD")
	confirm := os.Getenv("DNT_VAULT_ENV_MASTER_PASSWORD_CONFIRM")
	if password == "" {
		password, err = interactive.PromptPassword("Enter env master password")
		if err != nil {
			return err
		}
		confirm, err = interactive.PromptPassword("Confirm env master password")
		if err != nil {
			return err
		}
	} else if confirm == "" {
		confirm = password
	}
	if password != confirm {
		return fmt.Errorf("passwords do not match")
	}

	if err := os.WriteFile(cfg.Encryption.EnvMasterKeyFile, []byte(password), 0o600); err != nil {
		return err
	}
	if err := appconfig.SaveAppConfig(cfg); err != nil {
		return err
	}

	fmt.Println(color.GreenString("✓ Env master key generated and saved to %s", cfg.Encryption.EnvMasterKeyFile))
	return nil
}
