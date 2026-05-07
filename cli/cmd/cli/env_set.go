package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dnt/vault-cli/internal/client"
	appconfig "github.com/dnt/vault-cli/internal/config"
	"github.com/dnt/vault-cli/internal/envmanager"
	"github.com/dnt/vault-cli/internal/interactive"
	"github.com/spf13/cobra"
)

var envSetCmd = &cobra.Command{
	Use:   "set <scope> <key> <value>",
	Short: "Set a single env variable",
	Args:  cobra.ExactArgs(3),
	RunE:  runEnvSet,
}

func init() {
	envCmd.AddCommand(envSetCmd)
}

func runEnvSet(cmd *cobra.Command, args []string) error {
	scope, key, value := args[0], args[1], args[2]
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
	enc, err := envmanager.EncryptVariables(map[string]string{key: value}, string(master))
	if err != nil {
		return err
	}
	verify, err := envmanager.EncryptVariables(map[string]string{"verify": "dnt-vault-env-ok"}, string(master))
	if err != nil {
		return err
	}
	homeDir, _ := os.UserHomeDir()
	tokenFile := filepath.Join(homeDir, ".dnt-vault", "token")
	c := client.NewClient(cfg.Server.URL, tokenFile)
	if err := c.LoadToken(); err != nil {
		return fmt.Errorf("not logged in. Run 'dnt-vault login' first")
	}
	if err := c.SetEnvVariable(envmanager.ScopeToURL(scope), key, enc[key], interactive.GetHostname(), verify["verify"]); err != nil {
		return err
	}
	fmt.Println("ok")
	return nil
}
