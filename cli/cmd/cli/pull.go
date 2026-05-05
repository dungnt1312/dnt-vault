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

var pullProfileName string

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull SSH config from vault",
	RunE:  runPull,
}

func init() {
	pullCmd.Flags().StringVar(&pullProfileName, "profile", "", "Profile name to pull")
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
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

	fmt.Println(color.CyanString("Fetching profiles from vault..."))

	profiles, err := c.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %v", err)
	}

	if len(profiles) == 0 {
		fmt.Println(color.YellowString("No profiles found in vault"))
		return nil
	}

	var selectedProfile client.Profile
	if pullProfileName != "" {
		found := false
		for _, p := range profiles {
			if p.Name == pullProfileName {
				selectedProfile = p
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("profile '%s' not found", pullProfileName)
		}
	} else {
		fmt.Println(color.CyanString("\nAvailable profiles:"))
		var items []string
		for i, p := range profiles {
			keysInfo := "no keys"
			if p.HasKeys {
				keysInfo = fmt.Sprintf("%d keys", p.KeyCount)
			}
			updatedAgo := time.Since(p.UpdatedAt)
			items = append(items, fmt.Sprintf("%d. %s (%s) - Updated: %s ago - %s",
				i+1, p.Name, p.Hostname, formatDuration(updatedAgo), keysInfo))
		}

		idx, _, err := interactive.PromptSelect("Select profile", items)
		if err != nil {
			return err
		}
		selectedProfile = profiles[idx]
	}

	fmt.Println(color.CyanString("\nDownloading profile '%s'...", selectedProfile.Name))

	profileData, err := c.GetProfile(selectedProfile.Name)
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
		if profileData.Verify == "" {
			return fmt.Errorf("failed to decrypt config: %v\n\n⚠ This profile was pushed before v1.1.2 (no verify token).\n   Re-push from your client to fix this issue.", err)
		}
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

	if profileData.Profile.HasKeys && len(profileData.Keys) > 0 {
		fmt.Println(color.YellowString("\n⚠ This profile includes %d private keys.", len(profileData.Keys)))
		decrypt, err := interactive.PromptConfirm("Decrypt and restore keys?", false)
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

	fmt.Println(color.GreenString("\n✓ Profile '%s' pulled successfully", selectedProfile.Name))
	fmt.Println("  Config restored")
	if profileData.Profile.HasKeys {
		fmt.Printf("  %d keys available\n", profileData.Profile.KeyCount)
	}

	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
