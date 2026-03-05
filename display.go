package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	bold   = color.New(color.Bold)
	dim    = color.New(color.Faint)
	cyan   = color.New(color.FgCyan)
	green  = color.New(color.FgGreen, color.Bold)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed)
)

func display(root string, repos []RepoInfo) {
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

	if flagGroup {
		displayGrouped(repos)
	} else {
		displayFlat(repos)
	}
}

// repoRow renders one repo line (and optional URL line) into the flat layout.
func repoRow(r RepoInfo, maxName, maxBranch int, showURL bool, cleanCount, dirtyCount *int) {
	name := bold.Sprintf("%-*s", maxName, r.Name)
	branch := cyan.Sprintf("%-*s", maxBranch, r.Branch)

	var statusStr string
	if r.Error != "" {
		statusStr = red.Sprint("! error")
	} else if r.IsClean {
		statusStr = green.Sprint("✓ clean")
		*cleanCount++
	} else {
		word := "change"
		if r.Changes != 1 {
			word = "changes"
		}
		statusStr = yellow.Sprintf("● %d %s", r.Changes, word)
		*dirtyCount++
	}

	syncStr := ""
	if r.Ahead > 0 && r.Behind > 0 {
		syncStr = dim.Sprintf("  ↑%d ↓%d", r.Ahead, r.Behind)
	} else if r.Ahead > 0 {
		syncStr = dim.Sprintf("  ↑%d", r.Ahead)
	} else if r.Behind > 0 {
		syncStr = red.Sprintf("  ↓%d", r.Behind)
	}

	fmt.Printf("  %s  %s  %s%s\n", name, branch, statusStr, syncStr)
	if showURL && r.RemoteURL != "" {
		fmt.Printf("  %s  %s\n", strings.Repeat(" ", maxName), dim.Sprint(r.RemoteURL))
	}
}

func displayFlat(repos []RepoInfo) {
	maxName, maxBranch := columnWidths(repos)
	cleanCount, dirtyCount, nonRepoCount := 0, 0, 0

	for _, r := range repos {
		if !r.IsRepo {
			nonRepoCount++
			if flagAll {
				fmt.Printf("  %-*s  %s\n", maxName, r.Name, dim.Sprint("—"))
			}
			continue
		}
		repoRow(r, maxName, maxBranch, flagURL, &cleanCount, &dirtyCount)
	}

	printSummary(cleanCount, dirtyCount, nonRepoCount)
}

func displayGrouped(repos []RepoInfo) {
	// Build ordered groups (insertion order = first repo seen with that URL).
	const noRemote = "(no remote)"
	type group struct {
		url   string
		repos []RepoInfo
	}
	var order []string
	groups := map[string]*group{}

	for _, r := range repos {
		if !r.IsRepo {
			continue
		}
		key := r.RemoteURL
		if key == "" {
			key = noRemote
		}
		if _, exists := groups[key]; !exists {
			order = append(order, key)
			groups[key] = &group{url: key}
		}
		groups[key].repos = append(groups[key].repos, r)
	}

	// Column widths scoped per group for tighter alignment.
	cleanCount, dirtyCount := 0, 0

	for i, key := range order {
		g := groups[key]

		// Group header
		if key == noRemote {
			bold.Printf("  %s\n", dim.Sprint(noRemote))
		} else {
			bold.Printf("  %s\n", key)
		}

		maxName, maxBranch := columnWidths(g.repos)
		for _, r := range g.repos {
			// Indent one extra level inside the group
			// We reuse repoRow but wrap the output manually via indent prefix.
			name := bold.Sprintf("    %-*s", maxName, r.Name)
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

			syncStr := ""
			if r.Ahead > 0 && r.Behind > 0 {
				syncStr = dim.Sprintf("  ↑%d ↓%d", r.Ahead, r.Behind)
			} else if r.Ahead > 0 {
				syncStr = dim.Sprintf("  ↑%d", r.Ahead)
			} else if r.Behind > 0 {
				syncStr = red.Sprintf("  ↓%d", r.Behind)
			}

			fmt.Printf("%s  %s  %s%s\n", name, branch, statusStr, syncStr)
		}

		if i < len(order)-1 {
			fmt.Println()
		}
	}

	printSummary(cleanCount, dirtyCount, 0)
}

func columnWidths(repos []RepoInfo) (maxName, maxBranch int) {
	for _, r := range repos {
		if len(r.Name) > maxName {
			maxName = len(r.Name)
		}
		if r.IsRepo && len(r.Branch) > maxBranch {
			maxBranch = len(r.Branch)
		}
	}
	return
}

func printSummary(cleanCount, dirtyCount, nonRepoCount int) {
	fmt.Println()
	repoCount := cleanCount + dirtyCount
	var parts []string
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
