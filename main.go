package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagAll      bool
	flagDepth    int
	flagNoColor  bool
	flagFetch    bool
	flagURL      bool
	flagGroup    bool
)

var rootCmd = &cobra.Command{
	Use:   "lsgit [path]",
	Short: "ls for git repositories — list subdirs with branch and status",
	Long: `lsgit scans subdirectories and displays each git repository's
current branch and working-tree status at a glance.`,
	Args:    cobra.MaximumNArgs(1),
	RunE:    run,
	Version: version,
}

func init() {
	rootCmd.Flags().BoolVarP(&flagAll, "all", "a", false, "show non-git directories too")
	rootCmd.Flags().IntVarP(&flagDepth, "depth", "d", 1, "max depth to scan for nested repos (0 = unlimited)")
	rootCmd.Flags().BoolVar(&flagNoColor, "no-color", false, "disable color output")
	rootCmd.Flags().BoolVarP(&flagFetch, "fetch", "f", false, "run git fetch before checking status (slower)")
	rootCmd.Flags().BoolVarP(&flagURL, "url", "u", false, "show remote origin URL for each repository")
	rootCmd.Flags().BoolVarP(&flagGroup, "group", "g", false, "group repositories by remote origin URL (implies --url)")
}

func run(cmd *cobra.Command, args []string) error {
	root := "."
	if len(args) == 1 {
		root = args[0]
	}

	abs, err := absolutePath(root)
	if err != nil {
		return fmt.Errorf("cannot resolve path %q: %w", root, err)
	}

	if stat, err := os.Stat(abs); err != nil || !stat.IsDir() {
		return fmt.Errorf("%q is not a directory", abs)
	}

	if flagGroup {
		flagURL = true // grouping requires URL data
	}

	results := scan(abs, 0)
	display(abs, results)
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
