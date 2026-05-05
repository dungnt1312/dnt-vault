package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dnt/vault-cli/internal/backup"
	"github.com/dnt/vault-cli/internal/client"
	appconfig "github.com/dnt/vault-cli/internal/config"
	"github.com/dnt/vault-cli/internal/crypto"
	"github.com/dnt/vault-cli/internal/interactive"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Profile management commands",
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles in vault",
	RunE:  runProfileList,
}

var profileUseCmd = &cobra.Command{
	Use:   "use <profile-name>",
	Short: "Pull and apply a profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileUse,
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileUseCmd)
}

func runProfileList(cmd *cobra.Command, args []string) error {
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
		marker := " "
		if cfg.Profiles.Current == p.Name {
			marker = "*"
			fmt.Println(color.GreenString("  %s %s", marker, p.Name) + color.WhiteString(" (%s) [current]", p.Hostname))
		} else {
			fmt.Println(color.GreenString("  %s %s", marker, p.Name) + color.WhiteString(" (%s)", p.Hostname))
		}

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

func runProfileUse(cmd *cobra.Command, args []string) error {
	profileName := args[0]
	
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

	fmt.Println(color.CyanString("Fetching profile '%s'...", profileName))

	profileData, err := c.GetProfile(profileName)
	if err != nil {
		return fmt.Errorf("failed to download profile: %v", err)
	}

	masterPassword, err := os.ReadFile(cfg.Encryption.MasterKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read master key: %v", err)
	}

	if profileData.Verify != "" {
		verified, err := crypto.Decrypt(profileData.Verify, string(masterPassword))
		if err != nil || verified != "dnt-vault-ok" {
			return fmt.Errorf("wrong master password — use the same password that was set when this profile was pushed")
		}
	}

	fmt.Println(color.CyanString("Decrypting config..."))
	decryptedConfig, err := crypto.Decrypt(profileData.Config, string(masterPassword))
	if err != nil {
		return fmt.Errorf("failed to decrypt config: %v", err)
	}

	localConfigExists := false
	var localConfig string
	if data, err := os.ReadFile(cfg.SSH.ConfigPath); err == nil {
		localConfig = string(data)
		localConfigExists = true
	}

	if localConfigExists && appconfig.HasDifference(localConfig, decryptedConfig) {
		fmt.Println(color.YellowString("\n⚠ Local SSH config exists and differs from vault.\n"))
		appconfig.ShowDiff(localConfig, decryptedConfig)

		abort, err := interactive.PromptConfirm("\nAbort pull?", true)
		if err != nil {
			return err
		}
		if abort {
			fmt.Println(color.YellowString("Pull aborted"))
			return nil
		}
	}

	if cfg.Backup.Enabled && localConfigExists {
		backupMgr := backup.NewBackupManager(cfg.Backup.Dir, cfg.Backup.MaxBackups)
		backupPath, err := backupMgr.Backup(cfg.SSH.ConfigPath)
		if err != nil {
			fmt.Println(color.YellowString("⚠ Failed to create backup: %v", err))
		} else {
			fmt.Println(color.GreenString("Creating backup: %s", backupPath))
		}
	}

	fmt.Println(color.CyanString("Writing to %s...", cfg.SSH.ConfigPath))
	if err := appconfig.WriteSSHConfig(cfg.SSH.ConfigPath, decryptedConfig); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	// Update current profile in config
	cfg.Profiles.Current = profileName
	if err := appconfig.SaveAppConfig(cfg); err != nil {
		fmt.Println(color.YellowString("⚠ Failed to set current profile: %v", err))
	} else {
		fmt.Println(color.GreenString("✓ Current profile set to: %s", profileName))
	}

	if profileData.Profile.HasKeys && len(profileData.Keys) > 0 {
		fmt.Println(color.YellowString("\n⚠ This profile includes %d private keys.", len(profileData.Keys)))
		decrypt, err := interactive.PromptConfirm("Decrypt and restore keys?", true)
		if err != nil {
			return err
		}

		if decrypt {
			keyPassphrase, err := interactive.PromptPassword("\nEnter passphrase for key decryption")
			if err != nil {
				return err
			}

			fmt.Println(color.CyanString("\nDecrypting keys..."))
			for keyName, encryptedKey := range profileData.Keys {
				decryptedKey, err := crypto.Decrypt(encryptedKey, keyPassphrase)
				if err != nil {
					fmt.Println(color.RedString("  ✗ Failed to decrypt %s: %v", keyName, err))
					continue
				}

				keyPath := filepath.Join(cfg.SSH.KeysDir, keyName)
				if err := os.WriteFile(keyPath, []byte(decryptedKey), 0600); err != nil {
					fmt.Println(color.RedString("  ✗ Failed to write %s: %v", keyName, err))
					continue
				}

				fmt.Println(color.GreenString("  ✓ %s -> %s", keyName, keyPath))
			}
		}
	}

	fmt.Println(color.GreenString("\n✓ Profile '%s' applied successfully", profileName))
	return nil
}
