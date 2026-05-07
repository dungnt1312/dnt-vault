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

var envDeleteAll bool
var envDeleteYes bool

var envDeleteCmd = &cobra.Command{
	Use:   "delete <scope> [key]",
	Short: "Delete env variable or entire scope",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runEnvDelete,
}

func init() {
	envDeleteCmd.Flags().BoolVar(&envDeleteAll, "all", false, "Delete entire scope")
	envDeleteCmd.Flags().BoolVarP(&envDeleteYes, "yes", "y", false, "Skip confirmation")
	envCmd.AddCommand(envDeleteCmd)
}

func runEnvDelete(cmd *cobra.Command, args []string) error {
	scope := args[0]
	if err := envmanager.ValidateScope(scope); err != nil {
		return err
	}
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

	urlScope := envmanager.ScopeToURL(scope)
	if envDeleteAll {
		if !envDeleteYes {
			ok, err := interactive.PromptConfirm("Delete entire scope?", false)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}
		return c.DeleteEnvScope(urlScope)
	}

	if len(args) != 2 {
		return fmt.Errorf("key is required unless --all is used")
	}
	return c.DeleteEnvVariable(urlScope, args[1])
}
