package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/catfly/restic-helpers/internal/retry"
)

const (
	AppName   = "restic-helpers"
	EnvPrefix = "X_RESTIC_"
)

// Paths holds standard application paths
type Paths struct {
	ConfigDir string
	DataDir   string
	StateDir  string
	ReposDir  string
}

// TelegramConfig holds Telegram notification settings
type TelegramConfig struct {
	Enabled  bool   `toml:"enabled" json:"enabled"`
	BotToken string `toml:"bot_token" json:"bot_token,omitempty"`
	ChatID   string `toml:"chat_id" json:"chat_id,omitempty"`
}

// PruneConfig holds snapshot retention settings
type PruneConfig struct {
	KeepDaily   int `toml:"keep_daily" json:"keep_daily"`
	KeepWeekly  int `toml:"keep_weekly" json:"keep_weekly"`
	KeepMonthly int `toml:"keep_monthly" json:"keep_monthly"`
}

// RetryConfig is an alias for retry.Config
type RetryConfig = retry.Config

// Config holds the global configuration
type Config struct {
	Telegram TelegramConfig `toml:"telegram" json:"telegram"`
	Prune    PruneConfig    `toml:"prune" json:"prune"`
	Retry    RetryConfig    `toml:"retry" json:"retry"`
}

// RepoConfig holds per-repository configuration
type RepoConfig struct {
	Name         string       `json:"name"`
	RepoFile     string       `json:"repo_file"`
	PasswordFile string       `json:"password_file"`
	PathsFile    string       `json:"paths_file"`
	ExcludeFile  string       `json:"exclude_file,omitempty"`
	Healthcheck  string       `json:"healthcheck,omitempty"`
	Prune        *PruneConfig `json:"prune,omitempty"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Telegram: TelegramConfig{
			Enabled: true,
		},
		Prune: PruneConfig{
			KeepDaily:   7,
			KeepWeekly:  4,
			KeepMonthly: 6,
		},
		Retry: retry.DefaultConfig(),
	}
}

// GetPaths returns the standard application paths
func GetPaths() (*Paths, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".config", AppName)
	reposDir := filepath.Join(configDir, "repositories")

	return &Paths{
		ConfigDir: configDir,
		ReposDir:  reposDir,
	}, nil
}

// Load loads the configuration from files and environment
func Load() (*Config, error) {
	return LoadWithVerbose(false)
}

// LoadWithVerbose loads the configuration and optionally prints it
func LoadWithVerbose(verbose bool) (*Config, error) {
	cfg := DefaultConfig()

	paths, err := GetPaths()
	if err != nil {
		return cfg, err
	}

	// Load from config.toml
	configFile := filepath.Join(paths.ConfigDir, "config.toml")
	if err := loadTOMLFile(configFile, cfg); err != nil && !os.IsNotExist(err) {
		return cfg, fmt.Errorf("failed to load config.toml: %w", err)
	}

	// Override with secret.toml
	secretFile := filepath.Join(paths.ConfigDir, "secret.toml")
	if err := loadTOMLFile(secretFile, cfg); err != nil && !os.IsNotExist(err) {
		return cfg, fmt.Errorf("failed to load secret.toml: %w", err)
	}

	// Override with environment variables
	applyEnvOverrides(cfg)

	if verbose {
		cfg.PrettyPrint()
	}

	return cfg, nil
}

// PrettyPrint prints the config as formatted JSON (hides sensitive values)
func (c *Config) PrettyPrint() {
	// Create a copy with masked sensitive values
	masked := *c
	if masked.Telegram.BotToken != "" {
		masked.Telegram.BotToken = "***"
	}

	data, err := json.MarshalIndent(masked, "", "  ")
	if err != nil {
		fmt.Printf("Config: %+v\n", masked)
		return
	}
	fmt.Printf("Config loaded:\n%s\n", string(data))
}

// PrettyPrint prints the repo config as formatted JSON
func (r *RepoConfig) PrettyPrint() {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		fmt.Printf("RepoConfig: %+v\n", r)
		return
	}
	fmt.Printf("Repository config:\n%s\n", string(data))
}

// loadTOMLFile loads a TOML file into the config struct
func loadTOMLFile(path string, cfg *Config) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	_, err := toml.DecodeFile(path, cfg)
	return err
}

// applyEnvOverrides applies environment variable overrides to the config
func applyEnvOverrides(cfg *Config) {
	applyEnvOverridesTelegramConfig(&cfg.Telegram)
	applyEnvOverridesRetryConfig(&cfg.Retry)
	applyEnvOverridesPruneConfig(&cfg.Prune)
}

func applyEnvOverridesTelegramConfig(cfg *TelegramConfig) {
	setEnvBool(&cfg.Enabled, EnvPrefix+"TELEGRAM_ENABLED")
	setEnvString(&cfg.BotToken, EnvPrefix+"TELEGRAM_BOT_TOKEN")
	setEnvString(&cfg.ChatID, EnvPrefix+"TELEGRAM_CHAT_ID")
}

func applyEnvOverridesRetryConfig(cfg *RetryConfig) {
	setEnvInt(&cfg.Multiplier, EnvPrefix+"RETRY_MULTIPLIER")
	setEnvInt(&cfg.MaxAttempts, EnvPrefix+"RETRY_MAX_ATTEMPTS")
	setEnvInt(&cfg.BackoffMin, EnvPrefix+"RETRY_BACKOFF_MIN")
	setEnvInt(&cfg.BackoffMax, EnvPrefix+"RETRY_BACKOFF_MAX")
	setEnvInt(&cfg.ExpBase, EnvPrefix+"RETRY_EXP_BASE")
}

func applyEnvOverridesPruneConfig(cfg *PruneConfig) {
	setEnvInt(&cfg.KeepDaily, EnvPrefix+"PRUNE_KEEP_DAILY")
	setEnvInt(&cfg.KeepWeekly, EnvPrefix+"PRUNE_KEEP_WEEKLY")
	setEnvInt(&cfg.KeepMonthly, EnvPrefix+"PRUNE_KEEP_MONTHLY")
}

// setEnvString sets a string value from environment variable if present
func setEnvString(target *string, envKey string) {
	if value := os.Getenv(envKey); value != "" {
		*target = value
	}
}

// setEnvInt sets an int value from environment variable if present
func setEnvInt(target *int, envKey string) {
	if value := os.Getenv(envKey); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			*target = intVal
		}
	}
}

// setEnvBool sets a bool value from environment variable if present
func setEnvBool(target *bool, envKey string) {
	if value := os.Getenv(envKey); value != "" {
		lower := strings.ToLower(value)
		*target = lower == "true" || lower == "1" || lower == "yes"
	}
}

// LoadRepo loads a repository configuration
func LoadRepo(name string) (*RepoConfig, error) {
	paths, err := GetPaths()
	if err != nil {
		return nil, err
	}

	repoDir := filepath.Join(paths.ReposDir, name)

	// Check if repository exists
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository %q does not exist", name)
	}

	repo := &RepoConfig{
		Name:         name,
		RepoFile:     filepath.Join(repoDir, "name.txt"),
		PasswordFile: filepath.Join(repoDir, "password.txt"),
		PathsFile:    filepath.Join(repoDir, "paths.txt"),
	}

	// Set exclude file if it exists
	excludeFile := filepath.Join(repoDir, "exclude.txt")
	if _, err := os.Stat(excludeFile); err == nil {
		repo.ExcludeFile = excludeFile
	}

	// Read healthcheck URL
	healthcheckFile := filepath.Join(repoDir, "healthcheck.txt")
	if data, err := os.ReadFile(healthcheckFile); err == nil {
		repo.Healthcheck = strings.TrimSpace(string(data))
	}

	// Load per-repo prune config
	pruneFile := filepath.Join(repoDir, "prune.toml")
	if _, err := os.Stat(pruneFile); err == nil {
		var pruneConfig PruneConfig
		if _, err := toml.DecodeFile(pruneFile, &pruneConfig); err != nil {
			return nil, fmt.Errorf("failed to load prune.toml: %w", err)
		}
		repo.Prune = &pruneConfig
	}

	return repo, nil
}
