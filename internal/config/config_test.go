package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Telegram.Enabled != true {
		t.Error("expected Telegram.Enabled to be true by default")
	}

	if cfg.Prune.KeepDaily != 7 {
		t.Errorf("expected KeepDaily=7, got %d", cfg.Prune.KeepDaily)
	}

	if cfg.Prune.KeepWeekly != 4 {
		t.Errorf("expected KeepWeekly=4, got %d", cfg.Prune.KeepWeekly)
	}

	if cfg.Prune.KeepMonthly != 6 {
		t.Errorf("expected KeepMonthly=6, got %d", cfg.Prune.KeepMonthly)
	}

	if cfg.Retry.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts=5, got %d", cfg.Retry.MaxAttempts)
	}
}

func TestEnvOverrides(t *testing.T) {
	cfg := DefaultConfig()

	// Set environment variables
	os.Setenv("X_RESTIC_TELEGRAM_ENABLED", "false")
	os.Setenv("X_RESTIC_TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("X_RESTIC_PRUNE_KEEP_DAILY", "14")
	defer func() {
		os.Unsetenv("X_RESTIC_TELEGRAM_ENABLED")
		os.Unsetenv("X_RESTIC_TELEGRAM_BOT_TOKEN")
		os.Unsetenv("X_RESTIC_PRUNE_KEEP_DAILY")
	}()

	applyEnvOverrides(cfg)

	if cfg.Telegram.Enabled != false {
		t.Error("expected Telegram.Enabled to be false after env override")
	}

	if cfg.Telegram.BotToken != "test-token" {
		t.Errorf("expected BotToken='test-token', got %q", cfg.Telegram.BotToken)
	}

	if cfg.Prune.KeepDaily != 14 {
		t.Errorf("expected KeepDaily=14, got %d", cfg.Prune.KeepDaily)
	}
}

func TestLoadRepo(t *testing.T) {
	// Create temp config structure
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, ".config", "restic-helpers", "repos", "testrepo")
	if err := os.MkdirAll(repoDir, 0700); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	// Write repo files
	files := map[string]string{
		"name.txt":        "sftp:user@host:/backup",
		"password.txt":    "secret123",
		"paths.txt":       "/home/user\n/etc",
		"exclude.txt":     "*.tmp\n*.log",
		"healthcheck.txt": "https://hc-ping.com/abc123",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(repoDir, name), []byte(content), 0600); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	// Override home dir for test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	repo, err := LoadRepo("testrepo")
	if err != nil {
		t.Fatalf("LoadRepo failed: %v", err)
	}

	if repo.Name != "testrepo" {
		t.Errorf("expected Name='testrepo', got %q", repo.Name)
	}

	expectedRepoFile := filepath.Join(repoDir, "name.txt")
	if repo.RepoFile != expectedRepoFile {
		t.Errorf("expected RepoFile=%q, got %q", expectedRepoFile, repo.RepoFile)
	}

	expectedPasswordFile := filepath.Join(repoDir, "password.txt")
	if repo.PasswordFile != expectedPasswordFile {
		t.Errorf("expected PasswordFile=%q, got %q", expectedPasswordFile, repo.PasswordFile)
	}

	expectedPathsFile := filepath.Join(repoDir, "paths.txt")
	if repo.PathsFile != expectedPathsFile {
		t.Errorf("expected PathsFile=%q, got %q", expectedPathsFile, repo.PathsFile)
	}

	expectedExcludeFile := filepath.Join(repoDir, "exclude.txt")
	if repo.ExcludeFile != expectedExcludeFile {
		t.Errorf("expected ExcludeFile=%q, got %q", expectedExcludeFile, repo.ExcludeFile)
	}

	if repo.Healthcheck != "https://hc-ping.com/abc123" {
		t.Errorf("expected Healthcheck='https://hc-ping.com/abc123', got %q", repo.Healthcheck)
	}
}
