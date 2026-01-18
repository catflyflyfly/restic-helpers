package notify

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/catfly/restic-helpers/internal/config"
)

var (
	ErrTelegramDisabled  = errors.New("telegram notifications disabled")
	ErrTelegramNoToken   = errors.New("telegram bot_token not configured")
	ErrTelegramNoChatID  = errors.New("telegram chat_id not configured")
	ErrHealthcheckNotSet = errors.New("healthcheck URL not configured")
)

// Notifier handles sending notifications
type Notifier struct {
	telegram       *config.TelegramConfig
	healthcheckURL string
	dryRun         bool
	verbose        bool
	client         *http.Client
}

// New creates a new Notifier
func New(telegram *config.TelegramConfig, healthcheckURL string, dryRun, verbose bool) *Notifier {
	return &Notifier{
		telegram:       telegram,
		healthcheckURL: healthcheckURL,
		dryRun:         dryRun,
		verbose:        verbose,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// logVerbose prints a message if verbose mode is enabled
func (n *Notifier) logVerbose(format string, args ...interface{}) {
	if n.verbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// validateTelegram checks if telegram is properly configured
func (n *Notifier) validateTelegram() error {
	if n.telegram == nil || !n.telegram.Enabled {
		return ErrTelegramDisabled
	}
	if n.telegram.BotToken == "" {
		return ErrTelegramNoToken
	}
	if n.telegram.ChatID == "" {
		return ErrTelegramNoChatID
	}
	return nil
}

// validateHealthcheck checks if healthcheck is configured
func (n *Notifier) validateHealthcheck() error {
	if n.healthcheckURL == "" {
		return ErrHealthcheckNotSet
	}
	return nil
}

// SendTelegram sends a message via Telegram
func (n *Notifier) SendTelegram(message string) error {
	if err := n.validateTelegram(); err != nil {
		n.logVerbose("Skipping telegram: %v", err)
		return nil
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.telegram.BotToken)

	if n.dryRun {
		fmt.Printf("curl -fsS -X POST %s -d chat_id=%s -d text='%s'\n", apiURL, n.telegram.ChatID, message)
		return nil
	}

	resp, err := n.client.PostForm(apiURL, url.Values{
		"chat_id": {n.telegram.ChatID},
		"text":    {message},
	})
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}

// PingHealthcheck pings a healthchecks.io URL with the given status
func (n *Notifier) PingHealthcheck(status string) error {
	if err := n.validateHealthcheck(); err != nil {
		n.logVerbose("Skipping healthcheck: %v", err)
		return nil
	}

	var pingURL string
	switch status {
	case "start":
		pingURL = n.healthcheckURL + "/start"
	case "fail":
		pingURL = n.healthcheckURL + "/fail"
	default:
		pingURL = n.healthcheckURL
	}

	if n.dryRun {
		fmt.Printf("curl -fsS -m 10 --retry 5 -o /dev/null %s\n", pingURL)
		return nil
	}

	resp, err := n.client.Get(pingURL)
	if err != nil {
		return fmt.Errorf("failed to ping healthcheck: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// PrintDryRunSummary prints a summary of notification config for dry-run mode
func (n *Notifier) PrintDryRunSummary() {
	if !n.dryRun {
		return
	}

	fmt.Println("[dry-run] Notifications:")

	if err := n.validateHealthcheck(); err != nil {
		fmt.Printf("[dry-run]   Healthcheck: %v\n", err)
	} else {
		fmt.Printf("[dry-run]   Healthcheck: %s\n", n.healthcheckURL)
		fmt.Println("[dry-run]     On start:")
		n.PingHealthcheck("start")
		fmt.Println("[dry-run]     On success:")
		n.PingHealthcheck("success")
		fmt.Println("[dry-run]     On failure:")
		n.PingHealthcheck("fail")
	}

	fmt.Println()

	if err := n.validateTelegram(); err != nil {
		fmt.Printf("[dry-run]   Telegram: %v\n", err)
	} else {
		fmt.Printf("[dry-run]   Telegram: enabled (chat_id: %s)\n", n.telegram.ChatID)
		fmt.Println("[dry-run]     On failure:")
		n.SendTelegram("<error message>")
	}
}
