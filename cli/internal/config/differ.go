package config

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func Diff(local, remote string) string {
	localLines := strings.Split(strings.TrimRight(local, "\n"), "\n")
	remoteLines := strings.Split(strings.TrimRight(remote, "\n"), "\n")

	var result strings.Builder
	result.WriteString(color.RedString("--- Local\n"))
	result.WriteString(color.GreenString("+++ Remote\n"))

	common := lcsLines(localLines, remoteLines)
	li, ri := 0, 0
	for _, c := range common {
		for li < len(localLines) && localLines[li] != c {
			result.WriteString(color.RedString("- %s\n", localLines[li]))
			li++
		}
		for ri < len(remoteLines) && remoteLines[ri] != c {
			result.WriteString(color.GreenString("+ %s\n", remoteLines[ri]))
			ri++
		}
		result.WriteString(fmt.Sprintf("  %s\n", c))
		li++
		ri++
	}
	for ; li < len(localLines); li++ {
		result.WriteString(color.RedString("- %s\n", localLines[li]))
	}
	for ; ri < len(remoteLines); ri++ {
		result.WriteString(color.GreenString("+ %s\n", remoteLines[ri]))
	}

	return result.String()
}

func lcsLines(a, b []string) []string {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] > dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}
	var lcs []string
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs = append([]string{a[i-1]}, lcs...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}
	return lcs
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
