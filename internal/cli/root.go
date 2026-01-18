package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

var (
	dryRun  bool
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:     "restic-helpers",
	Short:   "Restic backup helper utilities",
	Long:    `A CLI tool for managing restic backups with scheduling, notifications, and configuration management.`,
	Version: Version,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show commands without executing")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

func Execute() error {
	return rootCmd.Execute()
}

// IsDryRun returns whether dry-run mode is enabled
func IsDryRun() bool {
	return dryRun
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verbose
}

// LogVerbose prints a message if verbose mode is enabled
func LogVerbose(format string, args ...interface{}) {
	if verbose {
		fmt.Printf(format+"\n", args...)
	}
}
