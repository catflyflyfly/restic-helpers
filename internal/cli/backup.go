package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/catflyflyfly/restic-helpers/internal/config"
	"github.com/catflyflyfly/restic-helpers/internal/notify"
	"github.com/catflyflyfly/restic-helpers/internal/retry"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup <repo-name>",
	Short: "Run a backup for a repository",
	Long:  `Executes a restic backup for the specified repository and prunes old snapshots.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBackup,
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

func runBackup(cmd *cobra.Command, args []string) error {
	repoName := args[0]

	LogVerbose("Starting backup for repository: %s", repoName)

	LogVerbose("Loading paths configuration...")
	paths, err := config.GetPaths()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	LogVerbose("Loading global configuration...")
	cfg, err := config.LoadWithVerbose(IsVerbose())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	LogVerbose("Loading repository configuration...")
	repoCfg, err := config.LoadRepo(repoName)
	if err != nil {
		return fmt.Errorf("failed to load repository config: %w", err)
	}
	if IsVerbose() {
		repoCfg.PrettyPrint()
	}

	// Check required files
	LogVerbose("Checking required files...")
	for _, f := range []string{repoCfg.RepoFile, repoCfg.PasswordFile, repoCfg.PathsFile} {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("required file missing: %s", f)
		}
		LogVerbose("  %s: ok", f)
	}

	// Create notifier with dry-run and verbose awareness
	notifier := notify.New(&cfg.Telegram, repoCfg.Healthcheck, IsDryRun(), IsVerbose())

	// Build backup command
	LogVerbose("Building backup command...")
	backupArgs := []string{
		"backup",
		fmt.Sprintf("--repository-file=%s", repoCfg.RepoFile),
		fmt.Sprintf("--password-file=%s", repoCfg.PasswordFile),
		fmt.Sprintf("--files-from=%s", repoCfg.PathsFile),
		"--exclude-caches",
	}

	// Add core exclude file if it exists
	coreExcludeFile := filepath.Join(paths.ConfigDir, "core.exclude.txt")
	if _, err := os.Stat(coreExcludeFile); err == nil {
		LogVerbose("  Adding core exclude file: %s", coreExcludeFile)
		backupArgs = append(backupArgs, fmt.Sprintf("--exclude-file=%s", coreExcludeFile))
	}

	// Add repo exclude file if it exists
	if repoCfg.ExcludeFile != "" {
		LogVerbose("  Adding repo exclude file: %s", repoCfg.ExcludeFile)
		backupArgs = append(backupArgs, fmt.Sprintf("--exclude-file=%s", repoCfg.ExcludeFile))
	}

	if IsVerbose() {
		backupArgs = append(backupArgs, "--verbose")
	}

	// Get prune config
	pruneConfig := cfg.Prune
	if repoCfg.Prune != nil {
		LogVerbose("Using repository-specific prune config")
		pruneConfig = *repoCfg.Prune
	} else {
		LogVerbose("Using global prune config: keep_daily=%d, keep_weekly=%d, keep_monthly=%d",
			pruneConfig.KeepDaily, pruneConfig.KeepWeekly, pruneConfig.KeepMonthly)
	}

	// Build prune command
	LogVerbose("Building prune command...")
	pruneArgs := []string{
		"forget",
		fmt.Sprintf("--repository-file=%s", repoCfg.RepoFile),
		fmt.Sprintf("--password-file=%s", repoCfg.PasswordFile),
		fmt.Sprintf("--keep-daily=%d", pruneConfig.KeepDaily),
		fmt.Sprintf("--keep-weekly=%d", pruneConfig.KeepWeekly),
		fmt.Sprintf("--keep-monthly=%d", pruneConfig.KeepMonthly),
		"--prune",
	}

	if IsVerbose() {
		pruneArgs = append(pruneArgs, "--verbose")
	}

	if IsDryRun() {
		fmt.Println("[dry-run] Backup command:")
		fmt.Printf("restic %s\n", formatCmd(backupArgs))
		fmt.Println()
		// Let notifier print its own summary
		notifier.PrintDryRunSummary()
		fmt.Println()
		fmt.Println("[dry-run] Prune command:")
		fmt.Printf("restic %s\n", formatCmd(pruneArgs))

		return nil
	}

	// Ping healthcheck start
	LogVerbose("Pinging healthcheck (start)...")
	if err := notifier.PingHealthcheck("start"); err != nil {
		LogVerbose("Warning: failed to ping healthcheck: %v", err)
	}

	// Run backup with retry
	LogVerbose("Running backup...")
	LogVerbose("Executing: restic %s", strings.Join(backupArgs, " "))
	if err := retry.RunWithRetry("backup", func() error { return runResticCommand(backupArgs) }, cfg.Retry, LogVerbose); err != nil {
		LogVerbose("Backup failed after retries, sending notifications...")
		_ = notifier.SendTelegram(fmt.Sprintf("Backup failed for %s: %v", repoName, err))
		_ = notifier.PingHealthcheck("fail")
		return fmt.Errorf("backup failed: %w", err)
	}
	LogVerbose("Backup completed successfully")

	// Run prune with retry
	LogVerbose("Pruning old snapshots...")
	LogVerbose("Executing: restic %s", strings.Join(pruneArgs, " "))
	if err := retry.RunWithRetry("prune", func() error { return runResticCommand(pruneArgs) }, cfg.Retry, LogVerbose); err != nil {
		LogVerbose("Prune failed after retries, sending notifications...")
		_ = notifier.SendTelegram(fmt.Sprintf("Prune failed for %s: %v", repoName, err))
		return fmt.Errorf("prune failed: %w", err)
	}
	LogVerbose("Prune completed successfully")

	// Ping healthcheck success
	LogVerbose("Pinging healthcheck (success)...")
	if err := notifier.PingHealthcheck("success"); err != nil {
		LogVerbose("Warning: failed to ping healthcheck: %v", err)
	}

	fmt.Printf("Backup completed successfully for %s\n", repoName)
	return nil
}

func runResticCommand(args []string) error {
	cmd := exec.Command("restic", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func formatCmd(args []string) string {
	return strings.Join(args, " \\\n  ")
}
