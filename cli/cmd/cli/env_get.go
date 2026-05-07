package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dnt/vault-cli/internal/client"
	appconfig "github.com/dnt/vault-cli/internal/config"
	"github.com/dnt/vault-cli/internal/envmanager"
	"github.com/spf13/cobra"
)

var envGetCmd = &cobra.Command{
	Use:   "get <scope> <key>",
	Short: "Get specific env variable",
	Args:  cobra.ExactArgs(2),
	RunE:  runEnvGet,
}

func init() {
	envCmd.AddCommand(envGetCmd)
}

func runEnvGet(cmd *cobra.Command, args []string) error {
	scope, key := args[0], args[1]
	if err := envmanager.ValidateScope(scope); err != nil {
		return err
	}
	cfg, err := appconfig.LoadAppConfig()
	if err != nil {
		return fmt.Errorf("config not found. Run 'dnt-vault init' first")
	}
	master, err := os.ReadFile(cfg.Encryption.EnvMasterKeyFile)
	if err != nil {
		return fmt.Errorf("env encryption not initialized. Run: dnt-vault env init")
	}
	homeDir, _ := os.UserHomeDir()
	tokenFile := filepath.Join(homeDir, ".dnt-vault", "token")
	c := client.NewClient(cfg.Server.URL, tokenFile)
	if err := c.LoadToken(); err != nil {
		return fmt.Errorf("not logged in. Run 'dnt-vault login' first")
	}
	encValue, err := c.GetEnvVariable(envmanager.ScopeToURL(scope), key)
	if err != nil {
		return err
	}
	dec, err := envmanager.DecryptVariables(map[string]string{key: encValue}, string(master))
	if err != nil {
		return fmt.Errorf("decryption failed. incorrect env master password")
	}
	fmt.Println(dec[key])
	return nil
}
