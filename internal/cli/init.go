package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/catflyflyfly/restic-helpers/internal/assets"
	"github.com/catflyflyfly/restic-helpers/internal/config"
	"github.com/spf13/cobra"
)

var baseFiles = map[string]string{
	"config.toml":      assets.DefaultConfig,
	"secret.toml":      assets.DefaultSecret,
	"core.exclude.txt": assets.CoreExclude,
}

var RepoConfigFiles = map[string]string{
	"name.txt":        assets.RepoName,
	"password.txt":    assets.RepoPassword,
	"paths.txt":       assets.RepoPaths,
	"exclude.txt":     assets.RepoExclude,
	"healthcheck.txt": assets.RepoHealthcheck,
}

var initCmd = &cobra.Command{
	Use:   "init <repo-name>",
	Short: "Initialize a new repository configuration",
	Long:  `Creates the directory structure and configuration files for a new restic repository.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	repoName := args[0]

	paths, err := config.GetPaths()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	if err := ensureBaseConfig(paths); err != nil {
		return err
	}

	repoDir := filepath.Join(paths.ReposDir, repoName)

	if IsDryRun() {
		fmt.Printf("[dry-run] Would create directory: %s\n", repoDir)
		for filename := range RepoConfigFiles {
			fmt.Printf("[dry-run] Would create file: %s\n", filepath.Join(repoDir, filename))
		}
		return nil
	}

	if _, err := os.Stat(repoDir); err == nil {
		return fmt.Errorf("repository %q already exists", repoName)
	}

	if err := os.MkdirAll(repoDir, 0700); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	for filename, content := range RepoConfigFiles {
		filePath := filepath.Join(repoDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	fmt.Printf("Initialized repository configuration at %s\n", repoDir)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit name.txt with your repository URL/path")
	fmt.Println("  2. Edit password.txt with your repository password")
	fmt.Println("  3. Edit paths.txt with the paths you want to backup")
	fmt.Println("  4. (Optional) Edit exclude.txt with additional exclusion patterns")
	fmt.Println("  5. (Optional) Edit healthcheck.txt with your healthchecks.io URL")

	return nil
}

// ensureBaseConfig creates the base config directory and default files if they don't exist.
func ensureBaseConfig(paths *config.Paths) error {
	if IsDryRun() {
		if _, err := os.Stat(paths.ConfigDir); os.IsNotExist(err) {
			fmt.Printf("[dry-run] Would create directory: %s\n", paths.ConfigDir)
		}
		if _, err := os.Stat(paths.StateDir); os.IsNotExist(err) {
			fmt.Printf("[dry-run] Would create directory: %s\n", paths.StateDir)
		}
		for filename := range baseFiles {
			filePath := filepath.Join(paths.ConfigDir, filename)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				fmt.Printf("[dry-run] Would create file: %s\n", filePath)
			}
		}
		return nil
	}

	// Create config directory
	if err := os.MkdirAll(paths.ConfigDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create state directory
	if err := os.MkdirAll(paths.StateDir, 0700); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Create base config files (skip if already exist)
	for filename, content := range baseFiles {
		filePath := filepath.Join(paths.ConfigDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
				return fmt.Errorf("failed to create %s: %w", filename, err)
			}
			fmt.Printf("Created %s\n", filePath)
		}
	}

	return nil
}
