package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dnt/vault-cli/internal/client"
	appconfig "github.com/dnt/vault-cli/internal/config"
	"github.com/dnt/vault-cli/internal/crypto"
	"github.com/dnt/vault-cli/internal/interactive"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	pushIncludeKeys bool
	pushProfileName string
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push SSH config to vault",
	RunE:  runPush,
}

func init() {
	pushCmd.Flags().BoolVar(&pushIncludeKeys, "include-keys", true, "Include private keys")
	pushCmd.Flags().StringVar(&pushProfileName, "profile", "", "Profile name (default: hostname)")
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
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

	fmt.Println(color.CyanString("Analyzing SSH config..."))

	sshConfig, err := appconfig.ParseSSHConfig(cfg.SSH.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to parse SSH config: %v", err)
	}

	fmt.Printf("  Config file: %s\n", cfg.SSH.ConfigPath)
	fmt.Printf("  Hosts found: %d\n", len(sshConfig.Hosts))

	identityFiles := sshConfig.GetIdentityFiles()
	if len(identityFiles) > 0 {
		fmt.Printf("  Referenced keys:\n")
		for _, keyFile := range identityFiles {
			fmt.Printf("    - %s\n", keyFile)
		}
	}

	profileName := pushProfileName
	if profileName == "" {
		hostname := interactive.GetHostname()
		profileName, err = interactive.PromptString("\nProfile name", hostname)
		if err != nil {
			return err
		}
	}

	if pushIncludeKeys && len(identityFiles) > 0 {
		fmt.Println(color.YellowString("\n⚠ You are about to upload private keys to the vault."))
		fmt.Println("  Keys will be encrypted with a passphrase.")

		confirm, err := interactive.PromptConfirm("\nContinue?", false)
		if err != nil {
			return err
		}
		if !confirm {
			return fmt.Errorf("aborted")
		}
	}

	masterPassword, err := os.ReadFile(cfg.Encryption.MasterKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read master key: %v", err)
	}

	fmt.Println(color.CyanString("\nEncrypting config..."))
	encryptedConfig, err := crypto.Encrypt(sshConfig.Raw, string(masterPassword))
	if err != nil {
		return fmt.Errorf("failed to encrypt config: %v", err)
	}

	verifyToken, err := crypto.Encrypt("dnt-vault-ok", string(masterPassword))
	if err != nil {
		return fmt.Errorf("failed to generate verify token: %v", err)
	}

	hash := sha256.Sum256([]byte(sshConfig.Raw))
	configHash := hex.EncodeToString(hash[:])

	profileData := client.ProfileData{
		Profile: client.Profile{
			Name:       profileName,
			Hostname:   interactive.GetHostname(),
			HasKeys:    false,
			KeyCount:   0,
			ConfigHash: configHash,
		},
		Config: encryptedConfig,
		Verify: verifyToken,
		Keys:   make(map[string]string),
		KeysIV: make(map[string]string),
	}

	if pushIncludeKeys && len(identityFiles) > 0 {
		keyPassphrase, err := interactive.PromptPassword("\nEnter passphrase for key encryption")
		if err != nil {
			return err
		}

		confirmPassphrase, err := interactive.PromptPassword("Confirm passphrase")
		if err != nil {
			return err
		}

		if keyPassphrase != confirmPassphrase {
			return fmt.Errorf("passphrases do not match")
		}

		fmt.Println(color.CyanString("\nEncrypting keys (%d files)...", len(identityFiles)))

		for _, keyFile := range identityFiles {
			keyData, err := os.ReadFile(keyFile)
			if err != nil {
				fmt.Printf(color.YellowString("  ⚠ Skipping %s: %v\n", keyFile, err))
				continue
			}

			encryptedKey, err := crypto.Encrypt(string(keyData), keyPassphrase)
			if err != nil {
				return fmt.Errorf("failed to encrypt key %s: %v", keyFile, err)
			}

			keyName := filepath.Base(keyFile)
			profileData.Keys[keyName] = encryptedKey
			profileData.KeysIV[keyName] = ""
		}

		profileData.Profile.HasKeys = true
		profileData.Profile.KeyCount = len(profileData.Keys)
	}

	fmt.Println(color.CyanString("Uploading to vault..."))

	if err := c.SaveProfile(profileName, profileData); err != nil {
		return fmt.Errorf("failed to upload: %v", err)
	}

	fmt.Println(color.GreenString("\n✓ Profile '%s' pushed successfully", profileName))
	fmt.Printf("  Updated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("  Config hash: %s\n", configHash[:8]+"...")
	if profileData.Profile.HasKeys {
		fmt.Printf("  Keys: %d files uploaded\n", profileData.Profile.KeyCount)
	}

	return nil
}
