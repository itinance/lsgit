package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func display(root string, repos []RepoInfo) {
	bold := color.New(color.Bold)
	dim := color.New(color.Faint)
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	if flagNoColor {
		color.NoColor = true
	}

	fmt.Println()
	bold.Printf("  %s\n", root)
	fmt.Println()

	if len(repos) == 0 {
		dim.Println("  No git repositories found.")
		fmt.Println()
		return
	}

	// Column widths
	maxName := 0
	maxBranch := 0
	for _, r := range repos {
		if len(r.Name) > maxName {
			maxName = len(r.Name)
		}
		if r.IsRepo && len(r.Branch) > maxBranch {
			maxBranch = len(r.Branch)
		}
	}

	cleanCount, dirtyCount, nonRepoCount := 0, 0, 0

	for _, r := range repos {
		if !r.IsRepo {
			nonRepoCount++
			if flagAll {
				name := fmt.Sprintf("%-*s", maxName, r.Name)
				fmt.Printf("  %s  %s\n", name, dim.Sprint("—"))
			}
			continue
		}

		name := bold.Sprintf("%-*s", maxName, r.Name)
		branch := cyan.Sprintf("%-*s", maxBranch, r.Branch)

		var statusStr string
		if r.Error != "" {
			statusStr = red.Sprint("! error")
		} else if r.IsClean {
			statusStr = green.Sprint("✓ clean")
			cleanCount++
		} else {
			word := "change"
			if r.Changes != 1 {
				word = "changes"
			}
			statusStr = yellow.Sprintf("● %d %s", r.Changes, word)
			dirtyCount++
		}

		// Ahead/behind indicator
		syncStr := ""
		if r.Ahead > 0 && r.Behind > 0 {
			syncStr = dim.Sprintf("  ↑%d ↓%d", r.Ahead, r.Behind)
		} else if r.Ahead > 0 {
			syncStr = dim.Sprintf("  ↑%d", r.Ahead)
		} else if r.Behind > 0 {
			syncStr = red.Sprintf("  ↓%d", r.Behind)
		}

		fmt.Printf("  %s  %s  %s%s\n", name, branch, statusStr, syncStr)
		if flagURL && r.RemoteURL != "" {
			fmt.Printf("  %s  %s\n", strings.Repeat(" ", maxName), dim.Sprint(r.RemoteURL))
		}
	}

	// Summary line
	fmt.Println()
	repoCount := cleanCount + dirtyCount
	parts := []string{}
	if cleanCount > 0 {
		parts = append(parts, green.Sprintf("%d clean", cleanCount))
	}
	if dirtyCount > 0 {
		parts = append(parts, yellow.Sprintf("%d dirty", dirtyCount))
	}
	if nonRepoCount > 0 && flagAll {
		parts = append(parts, dim.Sprintf("%d other", nonRepoCount))
	}

	summary := fmt.Sprintf("%d repo", repoCount)
	if repoCount != 1 {
		summary += "s"
	}
	if len(parts) > 0 {
		summary += "  " + strings.Join(parts, "  ")
	}
	fmt.Printf("  %s\n\n", summary)
}
