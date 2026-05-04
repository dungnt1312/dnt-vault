package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dnt-vault",
	Short: "DNT-Vault SSH Config Sync Tool",
	Long:  `A CLI tool to sync SSH configs and keys across multiple machines using a self-hosted vault server.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
