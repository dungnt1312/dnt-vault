package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dnt/vault-cli/internal/client"
	"github.com/dnt/vault-cli/internal/interactive"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var deleteProfileName string

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a profile from vault",
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().StringVar(&deleteProfileName, "profile", "", "Profile name to delete (required)")
	deleteCmd.MarkFlagRequired("profile")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("config not found. Run 'ssh-sync init' first")
	}

	homeDir, _ := os.UserHomeDir()
	tokenFile := filepath.Join(homeDir, ".ssh-sync", "token")

	c := client.NewClient(cfg.Server.URL, tokenFile)
	if err := c.LoadToken(); err != nil {
		return fmt.Errorf("not logged in. Run 'ssh-sync login' first")
	}

	fmt.Println(color.YellowString("⚠ Delete profile '%s' from vault?", deleteProfileName))
	fmt.Println("  This action cannot be undone.")

	confirm, err := interactive.PromptConfirm("\nConfirm", false)
	if err != nil {
		return err
	}

	if !confirm {
		fmt.Println(color.YellowString("Aborted"))
		return nil
	}

	fmt.Println(color.CyanString("\nDeleting profile..."))

	if err := c.DeleteProfile(deleteProfileName); err != nil {
		return fmt.Errorf("failed to delete profile: %v", err)
	}

	fmt.Println(color.GreenString("✓ Profile '%s' deleted from vault", deleteProfileName))

	return nil
}
