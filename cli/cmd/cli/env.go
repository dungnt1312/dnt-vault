package main

import "github.com/spf13/cobra"

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage encrypted environment variables",
}

func init() {
	rootCmd.AddCommand(envCmd)
}
