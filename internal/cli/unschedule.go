package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/catfly/restic-helpers/internal/launchd"
	"github.com/spf13/cobra"
)

var unscheduleCmd = &cobra.Command{
	Use:   "unschedule <repo-name>",
	Short: "Remove scheduled backups",
	Long:  `Removes the launchd job for the specified repository.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUnschedule,
}

func init() {
	rootCmd.AddCommand(unscheduleCmd)
}

func runUnschedule(cmd *cobra.Command, args []string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("scheduling is only supported on macOS")
	}

	repoName := args[0]

	LogVerbose("Getting plist path for: %s", repoName)
	plistPath, err := launchd.GetPlistPath(repoName)
	if err != nil {
		return fmt.Errorf("failed to get plist path: %w", err)
	}
	LogVerbose("Plist path: %s", plistPath)

	LogVerbose("Checking if plist exists")
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return fmt.Errorf("no scheduled job found for %s", repoName)
	}

	if IsDryRun() {
		fmt.Printf("[dry-run] Would remove launchd job at %s\n\n", plistPath)
		content, err := os.ReadFile(plistPath)
		if err == nil {
			fmt.Println(string(content))
		}
		return nil
	}

	LogVerbose("Unloading launchd job")
	if err := launchd.Uninstall(repoName); err != nil {
		return fmt.Errorf("failed to uninstall launchd job: %w", err)
	}

	fmt.Printf("Unscheduled backup for %s\n", repoName)
	return nil
}
