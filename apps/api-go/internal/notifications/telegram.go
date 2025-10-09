package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/config"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

const (
	TelegramAPIURL = "https://api.telegram.org/bot%s/sendMessage"
)

// TelegramNotifier handles Telegram notifications for alert events
type TelegramNotifier struct {
	store  storage.Store
	config *config.TelegramConfig
	client *http.Client
	logger *log.Logger
}

// NewTelegramNotifier creates a new Telegram notifier
func NewTelegramNotifier(store storage.Store, cfg *config.TelegramConfig, logger *log.Logger) *TelegramNotifier {
	return &TelegramNotifier{
		store:  store,
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
		logger: logger,
	}
}

// Send sends Telegram notifications for an alert event to all active recipients
func (t *TelegramNotifier) Send(ctx context.Context, rule storage.AlertRule, event storage.AlertEvent) error {
	if !t.config.IsEnabled() {
		return nil
	}

	recipients, err := t.getAllActiveRecipients(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active Telegram recipients: %w", err)
	}

	if len(recipients) == 0 {
		t.logger.Println("[TELEGRAM] No active recipients found")
		return nil
	}

	message := t.buildAlertMessage(rule, event)

	for _, recipient := range recipients {
		go func(r storage.TelegramRecipient) {
			// Create independent context to avoid cancellation from parent
			sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := t.sendToRecipient(sendCtx, r, message); err != nil {
				t.logger.Printf("[TELEGRAM] failed to send to chat_id=%s: %v", r.ChatID, err)
			}
		}(recipient)
	}

	return nil
}

// SendTest sends a test message to verify configuration
func (t *TelegramNotifier) SendTest(ctx context.Context, recipient storage.TelegramRecipient) error {
	if !t.config.IsEnabled() {
		return fmt.Errorf("Telegram notifications are disabled")
	}

	message := "🌙 *LunaSentri Test Message*\n\n" +
		"This is a test notification from your LunaSentri monitoring system.\n\n" +
		"If you received this, your Telegram notifications are configured correctly! ✅"

	return t.sendToRecipient(ctx, recipient, message)
}

// getAllActiveRecipients fetches all active Telegram recipients from all users
func (t *TelegramNotifier) getAllActiveRecipients(ctx context.Context) ([]storage.TelegramRecipient, error) {
	users, err := t.store.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var allRecipients []storage.TelegramRecipient
	for _, user := range users {
		recipients, err := t.store.ListTelegramRecipients(ctx, user.ID)
		if err != nil {
			t.logger.Printf("[TELEGRAM] Failed to fetch recipients for user %d: %v", user.ID, err)
			continue
		}

		for _, recipient := range recipients {
			if recipient.IsActive {
				allRecipients = append(allRecipients, recipient)
			}
		}
	}

	return allRecipients, nil
}

// sendToRecipient sends a message to a specific Telegram chat
func (t *TelegramNotifier) sendToRecipient(ctx context.Context, recipient storage.TelegramRecipient, message string) error {
	apiURL := fmt.Sprintf(TelegramAPIURL, t.config.BotToken)

	payload := map[string]interface{}{
		"chat_id": recipient.ChatID,
		"text":    message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	now := time.Now()
	if err := t.store.UpdateTelegramDeliveryState(ctx, recipient.ID, now, recipient.CooldownUntil); err != nil {
		t.logger.Printf("[TELEGRAM] failed to update delivery state: %v", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		t.store.IncrementTelegramFailure(ctx, recipient.ID, time.Now())
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)
		t.store.IncrementTelegramFailure(ctx, recipient.ID, time.Now())
		return fmt.Errorf("Telegram API error status=%d body=%s", resp.StatusCode, errorBody.String())
	}

	t.store.MarkTelegramSuccess(ctx, recipient.ID, time.Now())
	t.logger.Printf("[TELEGRAM] delivered to chat_id=%s", recipient.ChatID)

	return nil
}

// buildAlertMessage builds the Telegram message for an alert
func (t *TelegramNotifier) buildAlertMessage(rule storage.AlertRule, event storage.AlertEvent) string {
	comparisonText := "above"
	if rule.Comparison == "below" {
		comparisonText = "below"
	}

	// Use plain text instead of Markdown to avoid parsing issues
	return fmt.Sprintf(
		"🚨 LunaSentri Alert\n\n"+
			"Rule: %s\n"+
			"Metric: %s\n"+
			"Condition: %s %.1f%%\n"+
			"Current Value: %.1f%%\n"+
			"Triggered: %s\n\n"+
			"Alert triggered after %d consecutive samples",
		rule.Name,
		rule.Metric,
		comparisonText,
		rule.ThresholdPct,
		event.Value,
		event.TriggeredAt.Format("2006-01-02 15:04:05"),
		rule.TriggerAfter,
	)
}

// Notify implements AlertNotifier interface
func (t *TelegramNotifier) Notify(ctx context.Context, rule storage.AlertRule, event *storage.AlertEvent) error {
	if event == nil {
		return nil
	}
	return t.Send(ctx, rule, *event)
}
