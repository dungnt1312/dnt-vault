package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dnt/vault-cli/internal/client"
	"github.com/dnt/vault-cli/internal/interactive"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("config not found. Run 'ssh-sync init' first")
	}

	homeDir, _ := os.UserHomeDir()
	tokenFile := filepath.Join(homeDir, ".ssh-sync", "token")

	c := client.NewClient(config.Server.URL, tokenFile)

	fmt.Println(color.CyanString("Vault Server: %s", config.Server.URL))

	username, err := interactive.PromptString("Username", "admin")
	if err != nil {
		return err
	}

	password, err := interactive.PromptPassword("Password")
	if err != nil {
		return err
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
	tokenFile := filepath.Join(homeDir, ".ssh-sync", "token")

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

func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(homeDir, ".ssh-sync", "config.yaml")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
