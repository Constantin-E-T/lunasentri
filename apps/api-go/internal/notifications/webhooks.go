package notifications

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// WebhookPayload represents the JSON payload sent to webhooks
type WebhookPayload struct {
	RuleID       int     `json:"rule_id"`
	RuleName     string  `json:"rule_name"`
	Metric       string  `json:"metric"`
	Comparison   string  `json:"comparison"`
	ThresholdPct float64 `json:"threshold_pct"`
	TriggerAfter int     `json:"trigger_after"`
	Value        float64 `json:"value"`
	TriggeredAt  string  `json:"triggered_at"`
	EventID      int     `json:"event_id"`
}

// Notifier handles webhook notifications for alert events
type Notifier struct {
	store  storage.Store
	client *http.Client
	logger *log.Logger
}

// NewNotifier creates a new webhook notifier
func NewNotifier(store storage.Store, logger *log.Logger) *Notifier {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	return &Notifier{
		store:  store,
		client: client,
		logger: logger,
	}
}

// Send sends webhook notifications for an alert event to all active webhooks
func (n *Notifier) Send(ctx context.Context, rule storage.AlertRule, event storage.AlertEvent) error {
	// For now, fetch all active webhooks since we don't have multi-tenant yet
	// TODO: Once multi-tenant lands, filter by rule owner
	webhooks, err := n.getAllActiveWebhooks(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active webhooks: %w", err)
	}

	if len(webhooks) == 0 {
		n.logger.Println("No active webhooks found, skipping notification")
		return nil
	}

	// Build payload
	payload := WebhookPayload{
		RuleID:       rule.ID,
		RuleName:     rule.Name,
		Metric:       rule.Metric,
		Comparison:   rule.Comparison,
		ThresholdPct: rule.ThresholdPct,
		TriggerAfter: rule.TriggerAfter,
		Value:        event.Value,
		TriggeredAt:  event.TriggeredAt.Format(time.RFC3339),
		EventID:      event.ID,
	}

	// Send to all webhooks concurrently
	for _, webhook := range webhooks {
		go func(w storage.Webhook) {
			if err := n.sendToWebhook(ctx, w, payload); err != nil {
				n.logger.Printf("Failed to send webhook to %s: %v", w.URL, err)
			}
		}(webhook)
	}

	return nil
}

// getAllActiveWebhooks fetches all active webhooks from all users
// TODO: Replace with user-scoped fetch once multi-tenant is implemented
func (n *Notifier) getAllActiveWebhooks(ctx context.Context) ([]storage.Webhook, error) {
	// For now, we need to get all users and then their webhooks
	// This is a temporary solution until we have proper multi-tenant support
	users, err := n.store.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var allWebhooks []storage.Webhook
	for _, user := range users {
		webhooks, err := n.store.ListWebhooks(ctx, user.ID)
		if err != nil {
			n.logger.Printf("Failed to fetch webhooks for user %d: %v", user.ID, err)
			continue
		}

		// Filter active webhooks
		for _, webhook := range webhooks {
			if webhook.IsActive {
				allWebhooks = append(allWebhooks, webhook)
			}
		}
	}

	return allWebhooks, nil
}

// sendToWebhook sends the payload to a specific webhook with retry logic
func (n *Notifier) sendToWebhook(ctx context.Context, webhook storage.Webhook, payload WebhookPayload) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HMAC signature
	signature, err := n.createSignature(payloadBytes, webhook.SecretHash)
	if err != nil {
		return fmt.Errorf("failed to create signature: %w", err)
	}

	// Retry logic with exponential backoff
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := n.makeHTTPRequest(ctx, webhook.URL, payloadBytes, signature)
		if err == nil {
			// Success - mark webhook as successful
			if markErr := n.store.MarkWebhookSuccess(ctx, webhook.ID, time.Now()); markErr != nil {
				n.logger.Printf("Failed to mark webhook %d as successful: %v", webhook.ID, markErr)
			}
			n.logger.Printf("Successfully sent webhook to %s (attempt %d)", webhook.URL, attempt)
			return nil
		}

		n.logger.Printf("Webhook attempt %d failed for %s: %v", attempt, webhook.URL, err)

		// If this was the last attempt, mark as failure
		if attempt == maxRetries {
			if markErr := n.store.IncrementWebhookFailure(ctx, webhook.ID, time.Now()); markErr != nil {
				n.logger.Printf("Failed to increment failure count for webhook %d: %v", webhook.ID, markErr)
			}
			return fmt.Errorf("webhook delivery failed after %d attempts: %w", maxRetries, err)
		}

		// Exponential backoff: 1s, 2s, 4s
		backoffDuration := time.Duration(1<<(attempt-1)) * time.Second
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoffDuration):
		}
	}

	return nil
}

// createSignature creates HMAC-SHA256 signature for the payload
func (n *Notifier) createSignature(payload []byte, secretHex string) (string, error) {
	// Decode hex secret
	secretBytes, err := hex.DecodeString(secretHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode secret hex: %w", err)
	}

	// Create HMAC
	h := hmac.New(sha256.New, secretBytes)
	h.Write(payload)
	signature := hex.EncodeToString(h.Sum(nil))

	return "sha256=" + signature, nil
}

// makeHTTPRequest makes the actual HTTP request to the webhook URL
func (n *Notifier) makeHTTPRequest(ctx context.Context, url string, payload []byte, signature string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LunaSentri-Signature", signature)
	req.Header.Set("User-Agent", "LunaSentri-Webhook/1.0")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}