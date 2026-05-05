package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const githubRepo = "dungnt1312/dnt-vault"

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade dnt-vault CLI to the latest version",
	RunE:  runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	fmt.Println(color.CyanString("Checking for updates..."))

	latest, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to check latest version: %v", err)
	}

	current := strings.TrimPrefix(Version, "v")
	latestClean := strings.TrimPrefix(latest.TagName, "v")

	fmt.Printf("  Current: %s\n", color.YellowString(Version))
	fmt.Printf("  Latest:  %s\n", color.GreenString(latest.TagName))

	if Version != "dev" && current == latestClean {
		fmt.Println(color.GreenString("\n✓ Already up to date!"))
		return nil
	}

	binaryURL := buildBinaryURL(latest.TagName)
	fmt.Printf("\nDownloading %s from:\n  %s\n", latest.TagName, binaryURL)

	selfPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to locate current binary: %v", err)
	}
	selfPath, err = filepath.EvalSymlinks(selfPath)
	if err != nil {
		return fmt.Errorf("failed to resolve binary path: %v", err)
	}

	tmpFile := selfPath + ".tmp"
	if err := downloadFile(binaryURL, tmpFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("download failed: %v", err)
	}

	if err := os.Chmod(tmpFile, 0755); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to set permissions: %v", err)
	}

	backupPath := selfPath + ".old"
	os.Remove(backupPath)
	if err := os.Rename(selfPath, backupPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to backup current binary: %v\nTry running with sudo/admin", err)
	}

	if err := os.Rename(tmpFile, selfPath); err != nil {
		os.Rename(backupPath, selfPath)
		return fmt.Errorf("failed to replace binary: %v", err)
	}

	os.Remove(backupPath)

	fmt.Println(color.GreenString("\n✓ Upgraded to %s successfully!", latest.TagName))
	fmt.Printf("  Binary: %s\n", selfPath)
	return nil
}

func fetchLatestVersion() (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func buildBinaryURL(tag string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	name := fmt.Sprintf("dnt-vault-%s-%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}

	return fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", githubRepo, tag, name)
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
