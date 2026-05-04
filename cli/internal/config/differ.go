package config

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func Diff(local, remote string) string {
	localLines := strings.Split(local, "\n")
	remoteLines := strings.Split(remote, "\n")

	var result strings.Builder
	result.WriteString(color.RedString("--- Local\n"))
	result.WriteString(color.GreenString("+++ Remote\n"))

	maxLen := len(localLines)
	if len(remoteLines) > maxLen {
		maxLen = len(remoteLines)
	}

	for i := 0; i < maxLen; i++ {
		var localLine, remoteLine string
		if i < len(localLines) {
			localLine = localLines[i]
		}
		if i < len(remoteLines) {
			remoteLine = remoteLines[i]
		}

		if localLine != remoteLine {
			if localLine != "" {
				result.WriteString(color.RedString("- %s\n", localLine))
			}
			if remoteLine != "" {
				result.WriteString(color.GreenString("+ %s\n", remoteLine))
			}
		}
	}

	return result.String()
}

func HasDifference(local, remote string) bool {
	return strings.TrimSpace(local) != strings.TrimSpace(remote)
}

func ShowDiff(local, remote string) {
	if !HasDifference(local, remote) {
		fmt.Println(color.GreenString("✓ No differences"))
		return
	}

	fmt.Println(color.YellowString("\n⚠ Local SSH config differs from vault:\n"))
	fmt.Println(Diff(local, remote))
}
