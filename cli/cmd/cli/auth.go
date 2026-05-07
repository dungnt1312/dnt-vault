package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dnt/vault-cli/internal/client"
	"github.com/dnt/vault-cli/internal/config"
	"github.com/dnt/vault-cli/internal/interactive"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to vault server",
	RunE:  runLogin,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from vault server",
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadAppConfig()
	if err != nil {
		return fmt.Errorf("config not found. Run 'dnt-vault init' first")
	}

	homeDir, _ := os.UserHomeDir()
	tokenFile := filepath.Join(homeDir, ".dnt-vault", "token")

	c := client.NewClient(cfg.Server.URL, tokenFile)

	fmt.Println(color.CyanString("Vault Server: %s", cfg.Server.URL))

	username := os.Getenv("DNT_VAULT_USERNAME")
	password := os.Getenv("DNT_VAULT_PASSWORD")
	if username == "" {
		username, err = interactive.PromptString("Username", "admin")
		if err != nil {
			return err
		}
	}
	if password == "" {
		password, err = interactive.PromptPassword("Password")
		if err != nil {
			return err
		}
	}

	if err := c.Login(username, password); err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	fmt.Println(color.GreenString("\n✓ Logged in successfully"))
	fmt.Println(color.GreenString("Token saved to %s", tokenFile))

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	homeDir, _ := os.UserHomeDir()
	tokenFile := filepath.Join(homeDir, ".dnt-vault", "token")

	if err := os.Remove(tokenFile); err != nil {
		if os.IsNotExist(err) {
			fmt.Println(color.YellowString("Already logged out"))
			return nil
		}
		return err
	}

	fmt.Println(color.GreenString("✓ Logged out successfully"))
	return nil
}
