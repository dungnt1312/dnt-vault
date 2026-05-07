package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dnt/vault-cli/internal/backup"
	"github.com/dnt/vault-cli/internal/client"
	appconfig "github.com/dnt/vault-cli/internal/config"
	"github.com/dnt/vault-cli/internal/envmanager"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var envPullOutput string

var envPullCmd = &cobra.Command{
	Use:   "pull <scope>",
	Short: "Pull env variables from a scope",
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvPull,
}

func init() {
	envPullCmd.Flags().StringVar(&envPullOutput, "output", "", "Write output to file")
	envCmd.AddCommand(envPullCmd)
}

func runEnvPull(cmd *cobra.Command, args []string) error {
	scope := args[0]
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

	data, err := c.GetEnvScope(envmanager.ScopeToURL(scope))
	if err != nil {
		return fmt.Errorf("scope '%s' not found", scope)
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

	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		if envPullOutput == "" {
			lines = append(lines, fmt.Sprintf("export %s=%q", k, vars[k]))
		} else {
			lines = append(lines, fmt.Sprintf("%s=%s", k, vars[k]))
		}
	}

	output := strings.Join(lines, "\n") + "\n"
	if envPullOutput == "" {
		fmt.Print(output)
		return nil
	}

	if _, err := os.Stat(envPullOutput); err == nil {
		backupMgr := backup.NewBackupManager(cfg.Env.BackupDir, cfg.Backup.MaxBackups)
		if _, err := backupMgr.Backup(envPullOutput); err != nil {
			fmt.Println(color.YellowString("⚠ Failed to create backup: %v", err))
		}
	}

	if err := os.WriteFile(envPullOutput, []byte(output), 0o600); err != nil {
		return err
	}
	fmt.Println(color.GreenString("✓ Env scope '%s' written to %s", scope, envPullOutput))
	return nil
}
