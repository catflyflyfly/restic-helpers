package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/catflyflyfly/restic-helpers/internal/config"
	"github.com/catflyflyfly/restic-helpers/internal/notify"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check <repo-name>",
	Short: "Verify repository integrity",
	Long:  `Runs restic check to verify the integrity of a repository.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	repoName := args[0]

	LogVerbose("Starting check for repository: %s", repoName)

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
	for _, f := range []string{repoCfg.RepoFile, repoCfg.PasswordFile} {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("required file missing: %s", f)
		}
		LogVerbose("  %s: ok", f)
	}

	notifier := notify.New(&cfg.Telegram, repoCfg.Healthcheck, IsDryRun(), IsVerbose())

	// Build check command
	LogVerbose("Building check command...")
	checkArgs := []string{
		"check",
		fmt.Sprintf("--repository-file=%s", repoCfg.RepoFile),
		fmt.Sprintf("--password-file=%s", repoCfg.PasswordFile),
	}

	if IsVerbose() {
		checkArgs = append(checkArgs, "--verbose")
	}

	if IsDryRun() {
		fmt.Println("[dry-run] Check command:")
		fmt.Printf("restic %s\n", formatCmd(checkArgs))
		notifier.PrintDryRunSummary()
		return nil
	}

	LogVerbose("Running check...")
	LogVerbose("Executing: restic %s", strings.Join(checkArgs, " "))

	resticCmd := exec.Command("restic", checkArgs...)
	resticCmd.Stdout = os.Stdout
	resticCmd.Stderr = os.Stderr

	if err := resticCmd.Run(); err != nil {
		LogVerbose("Check failed, sending notifications...")
		_ = notifier.SendTelegram(fmt.Sprintf("Check failed for %s: %v", repoName, err))
		return fmt.Errorf("restic check failed: %w", err)
	}

	LogVerbose("Check completed successfully")
	fmt.Printf("Repository %s is healthy\n", repoName)
	return nil
}
