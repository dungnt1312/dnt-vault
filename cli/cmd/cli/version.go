package main

import (
	"fmt"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	CommitSHA = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Println(color.CyanString("DNT-Vault CLI"))
	fmt.Printf("  Version:    %s\n", color.GreenString(Version))
	fmt.Printf("  Build time: %s\n", BuildTime)
	fmt.Printf("  Commit:     %s\n", CommitSHA)
	fmt.Printf("  Go:         %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
