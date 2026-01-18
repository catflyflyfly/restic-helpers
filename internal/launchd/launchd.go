package launchd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/catflyflyfly/restic-helpers/internal/config"
	"github.com/catflyflyfly/restic-helpers/internal/cron"
	"howett.net/plist"
)

const (
	labelPrefix = "com.restic-helpers"
)

// Job represents a launchd job configuration
type Job struct {
	Label                 string                  `plist:"Label"`
	ProgramArguments      []string                `plist:"ProgramArguments"`
	StartCalendarInterval []cron.CalendarInterval `plist:"StartCalendarInterval,omitempty"`
	StandardOutPath       string                  `plist:"StandardOutPath,omitempty"`
	StandardErrorPath     string                  `plist:"StandardErrorPath,omitempty"`
	EnvironmentVariables  map[string]string       `plist:"EnvironmentVariables,omitempty"`
	RunAtLoad             bool                    `plist:"RunAtLoad"`
}

// GetLabel returns the launchd label for a repository
func GetLabel(repoName string) string {
	return fmt.Sprintf("%s.%s", labelPrefix, repoName)
}

// GetPlistPath returns the path to the plist file for a repository
func GetPlistPath(repoName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "Library", "LaunchAgents", GetLabel(repoName)+".plist"), nil
}

// CreateJob creates a launchd job for scheduled backups
func CreateJob(repoName string, cronExpr string, binaryPath string) (*Job, error) {
	paths, err := config.GetPaths()
	if err != nil {
		return nil, err
	}

	intervals, err := cron.ParseCron(cronExpr)
	if err != nil {
		return nil, err
	}

	stdoutPath := filepath.Join(paths.StateDir, repoName+".out.log")
	stderrPath := filepath.Join(paths.StateDir, repoName+".err.log")

	job := &Job{
		Label: GetLabel(repoName),
		ProgramArguments: []string{
			binaryPath,
			"backup",
			repoName,
		},
		StartCalendarInterval: intervals,
		StandardOutPath:       stdoutPath,
		StandardErrorPath:     stderrPath,
		RunAtLoad:             false, // backup only runs on schedule
	}

	return job, nil
}

// Install installs a launchd job
func Install(job *Job, repoName string) error {
	plistPath, err := GetPlistPath(repoName)
	if err != nil {
		return err
	}

	// Ensure LaunchAgents directory exists
	dir := filepath.Dir(plistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	// Encode plist
	var buf bytes.Buffer
	encoder := plist.NewEncoder(&buf)
	encoder.Indent("\t")
	if err := encoder.Encode(job); err != nil {
		return fmt.Errorf("failed to encode plist: %w", err)
	}

	// Write plist file
	if err := os.WriteFile(plistPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write plist: %w", err)
	}

	// Load the job
	cmd := exec.Command("launchctl", "load", plistPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to load launchd job: %w", err)
	}

	return nil
}

// EncodePlist encodes a job to plist XML string
func EncodePlist(job *Job) (string, error) {
	var buf bytes.Buffer
	encoder := plist.NewEncoder(&buf)
	encoder.Indent("\t")
	if err := encoder.Encode(job); err != nil {
		return "", fmt.Errorf("failed to encode plist: %w", err)
	}
	return buf.String(), nil
}

// Uninstall removes a launchd job
func Uninstall(repoName string) error {
	plistPath, err := GetPlistPath(repoName)
	if err != nil {
		return err
	}

	// Unload the job (ignore errors if not loaded)
	cmd := exec.Command("launchctl", "unload", plistPath)
	_ = cmd.Run()

	// Remove the plist file
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist: %w", err)
	}

	return nil
}
