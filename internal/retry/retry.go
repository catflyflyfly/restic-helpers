package retry

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// Config holds retry behavior settings
type Config struct {
	Multiplier  int `toml:"multiplier" json:"multiplier"`
	MaxAttempts int `toml:"max_attempts" json:"max_attempts"`
	BackoffMin  int `toml:"backoff_min" json:"backoff_min"`
	BackoffMax  int `toml:"backoff_max" json:"backoff_max"`
	ExpBase     int `toml:"exp_base" json:"exp_base"`
}

// DefaultConfig returns the default retry configuration
func DefaultConfig() Config {
	return Config{
		Multiplier:  1,
		MaxAttempts: 5,
		BackoffMin:  1,
		BackoffMax:  60,
		ExpBase:     2,
	}
}

// Backoff creates an exponential backoff from the retry config.
func (r Config) Backoff() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = time.Duration(r.BackoffMin*r.Multiplier) * time.Second
	b.MaxInterval = time.Duration(r.BackoffMax) * time.Second
	b.Multiplier = float64(r.ExpBase)
	return b
}

// LogFunc is a function type for logging retry attempts
type LogFunc func(format string, args ...any)

// RunWithRetry executes an operation with retry logic using exponential backoff.
// The operation function is called on each attempt.
// The logFn is called to log verbose messages about retry attempts.
func RunWithRetry(name string, operation func() error, cfg Config, logFn LogFunc) error {
	attempt := 0
	op := func() (struct{}, error) {
		attempt++
		logFn("Attempt %d/%d for %s", attempt, cfg.MaxAttempts, name)
		return struct{}{}, operation()
	}

	notify := func(err error, d time.Duration) {
		logFn("%s failed: %v, retrying in %v...", name, err, d)
	}

	_, err := backoff.Retry(
		context.Background(),
		op,
		backoff.WithBackOff(cfg.Backoff()),
		backoff.WithMaxTries(uint(cfg.MaxAttempts)),
		backoff.WithNotify(notify),
	)
	return err
}
