package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dnt/vault-cli/internal/client"
	appconfig "github.com/dnt/vault-cli/internal/config"
	"github.com/dnt/vault-cli/internal/envmanager"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var envListCmd = &cobra.Command{
	Use:   "list [scope]",
	Short: "List env scopes or keys",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runEnvList,
}

func init() {
	envCmd.AddCommand(envListCmd)
}

func runEnvList(cmd *cobra.Command, args []string) error {
	cfg, err := appconfig.LoadAppConfig()
	if err != nil {
		return fmt.Errorf("config not found. Run 'dnt-vault init' first")
	}
	homeDir, _ := os.UserHomeDir()
	tokenFile := filepath.Join(homeDir, ".dnt-vault", "token")
	c := client.NewClient(cfg.Server.URL, tokenFile)
	if err := c.LoadToken(); err != nil {
		return fmt.Errorf("not logged in. Run 'dnt-vault login' first")
	}

	if len(args) == 0 {
		scopes, err := c.ListEnvScopes()
		if err != nil {
			return err
		}
		for _, s := range scopes {
			fmt.Println(envmanager.ScopeFromURL(s.Name))
		}
		return nil
	}

	master, err := os.ReadFile(cfg.Encryption.EnvMasterKeyFile)
	if err != nil {
		return fmt.Errorf("env encryption not initialized. Run: dnt-vault env init")
	}
	data, err := c.GetEnvScope(envmanager.ScopeToURL(args[0]))
	if err != nil {
		return err
	}
	vars, err := envmanager.DecryptVariables(data.Variables, string(master))
	if err != nil {
		return fmt.Errorf("decryption failed. incorrect env master password")
	}
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println(color.GreenString(k))
	}
	return nil
}
