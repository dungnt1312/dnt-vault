package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dnt/vault-cli/internal/client"
	appconfig "github.com/dnt/vault-cli/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles in vault",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := appconfig.LoadAppConfig()
	if err != nil {
		return fmt.Errorf("config not found. Run 'ssh-sync init' first")
	}

	homeDir, _ := os.UserHomeDir()
	tokenFile := filepath.Join(homeDir, ".ssh-sync", "token")

	c := client.NewClient(cfg.Server.URL, tokenFile)
	if err := c.LoadToken(); err != nil {
		return fmt.Errorf("not logged in. Run 'ssh-sync login' first")
	}

	profiles, err := c.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %v", err)
	}

	if len(profiles) == 0 {
		fmt.Println(color.YellowString("No profiles found in vault"))
		return nil
	}

	fmt.Println(color.CyanString("Profiles on vault (%s):\n", cfg.Server.URL))

	for _, p := range profiles {
		fmt.Println(color.GreenString("  %s", p.Name) + color.WhiteString(" (%s)", p.Hostname))
		
		updatedAgo := time.Since(p.UpdatedAt)
		fmt.Printf("    Updated: %s ago\n", formatDuration(updatedAgo))
		
		if p.HasKeys {
			fmt.Printf("    Keys: %d files\n", p.KeyCount)
		} else {
			fmt.Println("    Keys: none")
		}
		
		fmt.Printf("    Hash: %s...\n", p.ConfigHash[:8])
		fmt.Println()
	}

	fmt.Printf(color.CyanString("Total: %d profiles\n"), len(profiles))

	return nil
}
