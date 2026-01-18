package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/catfly/restic-helpers/internal/config"
	"github.com/catfly/restic-helpers/internal/launchd"
	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule <repo-name> <cron-expression>",
	Short: "Schedule automated backups (macOS only)",
	Long: `Creates a launchd job to run backups on a schedule.

Examples:
  restic-helpers schedule myrepo "0 2 * * *"     # Daily at 2 AM
  restic-helpers schedule myrepo "0 */6 * * *"  # Every 6 hours`,
	Args: cobra.ExactArgs(2),
	RunE: runSchedule,
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
}

func runSchedule(cmd *cobra.Command, args []string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("scheduling is only supported on macOS")
	}

	repoName := args[0]
	cronExpr := args[1]

	LogVerbose("Loading paths configuration")
	paths, err := config.GetPaths()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	LogVerbose("Loading repository config: %s", repoName)
	if _, err := config.LoadRepo(repoName); err != nil {
		return fmt.Errorf("failed to load repository config: %w", err)
	}

	LogVerbose("Getting executable path")
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	LogVerbose("Binary path: %s", binaryPath)

	LogVerbose("Parsing cron expression: %s", cronExpr)
	job, err := launchd.CreateJob(repoName, cronExpr, binaryPath)
	if err != nil {
		return fmt.Errorf("failed to create launchd job: %w", err)
	}
	LogVerbose("Created %d calendar intervals", len(job.StartCalendarInterval))

	plistPath, _ := launchd.GetPlistPath(repoName)

	if IsDryRun() {
		plistContent, err := launchd.EncodePlist(job)
		if err != nil {
			return fmt.Errorf("failed to encode plist: %w", err)
		}
		fmt.Printf("[dry-run] Would create launchd job at %s\n\n", plistPath)
		fmt.Println(plistContent)
		return nil
	}

	LogVerbose("Creating state directory: %s", paths.StateDir)
	if err := os.MkdirAll(paths.StateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	LogVerbose("Uninstalling existing job if present")
	_ = launchd.Uninstall(repoName)

	LogVerbose("Installing launchd job")
	if err := launchd.Install(job, repoName); err != nil {
		return fmt.Errorf("failed to install launchd job: %w", err)
	}

	fmt.Printf("Scheduled backup for %s\n", repoName)
	fmt.Printf("  Schedule: %s\n", cronExpr)
	fmt.Printf("  Plist: %s\n", plistPath)

	return nil
}
