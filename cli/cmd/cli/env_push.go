package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dnt/vault-cli/internal/client"
	appconfig "github.com/dnt/vault-cli/internal/config"
	"github.com/dnt/vault-cli/internal/envmanager"
	"github.com/dnt/vault-cli/internal/interactive"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var envPushFile string
var envPushReplace bool

var envPushCmd = &cobra.Command{
	Use:   "push <scope>",
	Short: "Push env variables to a scope",
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvPush,
}

func init() {
	envPushCmd.Flags().StringVar(&envPushFile, "file", "", "Read variables from file")
	envPushCmd.Flags().BoolVar(&envPushReplace, "replace", false, "Replace entire scope")
	envCmd.AddCommand(envPushCmd)
}

func runEnvPush(cmd *cobra.Command, args []string) error {
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

	incoming := map[string]string{}
	if envPushFile != "" {
		vars, err := envmanager.ParseEnvFile(envPushFile)
		if err != nil {
			return err
		}
		incoming = vars
	} else {
		fmt.Println("Enter variables (KEY=VALUE), empty line to finish:")
		s := bufio.NewScanner(os.Stdin)
		for {
			if !s.Scan() {
				break
			}
			line := strings.TrimSpace(s.Text())
			if line == "" {
				break
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid input %q", line)
			}
			k := strings.TrimSpace(parts[0])
			if _, ok := incoming[k]; ok {
				return fmt.Errorf("duplicate variable %s in input", k)
			}
			incoming[k] = strings.TrimSpace(parts[1])
		}
	}

	if len(incoming) == 0 {
		return fmt.Errorf("no variables provided. scope not created")
	}

	finalVars := incoming
	urlScope := envmanager.ScopeToURL(scope)
	if !envPushReplace {
		existing, err := c.GetEnvScope(urlScope)
		if err == nil {
			dec, err := envmanager.DecryptVariables(existing.Variables, string(master))
			if err != nil {
				return fmt.Errorf("decryption failed. incorrect env master password")
			}
			finalVars = envmanager.MergeVariables(dec, incoming)
		}
	}

	enc, err := envmanager.EncryptVariables(finalVars, string(master))
	if err != nil {
		return err
	}
	verify, err := envmanager.EncryptVariables(map[string]string{"verify": "dnt-vault-env-ok"}, string(master))
	if err != nil {
		return err
	}

	data := client.EnvScopeData{
		Metadata: client.EnvScope{Name: urlScope, VariableCount: len(enc), UpdatedAt: time.Now(), Hostname: interactive.GetHostname()},
		Variables: enc,
		Verify:    verify["verify"],
	}

	if err := c.SaveEnvScope(urlScope, data); err != nil {
		return err
	}

	fmt.Println(color.GreenString("✓ Env scope '%s' pushed successfully", scope))
	return nil
}
