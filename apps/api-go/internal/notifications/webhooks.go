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
	"net/url"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

const (
	// TODO: Make these configurable later
	MinAttemptInterval = 30 * time.Second // Minimum interval between delivery attempts
	FailureThreshold   = 3                // Number of failures within window to trigger cooldown
	FailureWindow      = 10 * time.Minute // Time window for counting failures
	CooldownDuration   = 15 * time.Minute // Duration of cooldown after reaching failure threshold
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

// WebhookMachineEvent represents a machine status event payload for webhooks
type WebhookMachineEvent struct {
	Event   string         `json:"event"` // "machine.offline" or "machine.online"
	Machine WebhookMachine `json:"machine"`
}

// WebhookMachine represents machine data in webhook payloads
type WebhookMachine struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Hostname    string    `json:"hostname"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	LastSeen    time.Time `json:"last_seen"`
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
			// Create independent context to avoid cancellation from parent
			sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := n.sendToWebhook(sendCtx, w, payload); err != nil {
				n.logger.Printf("Failed to send webhook to %s: %v", w.URL, err)
			}
		}(webhook)
	}

	return nil
}

// SendMachineEvent sends a machine status event to a webhook
func (n *Notifier) SendMachineEvent(ctx context.Context, webhook storage.Webhook, event WebhookMachineEvent) error {
	// Check delivery preconditions
	if err := n.checkDeliveryPreconditions(webhook); err != nil {
		return err
	}

	// Mark attempt
	now := time.Now()
	if err := n.store.UpdateWebhookDeliveryState(ctx, webhook.ID, now, nil); err != nil {
		n.logger.Printf("Failed to update webhook delivery state: %v", err)
	}

	// Marshal payload
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create signature
	signature, err := n.createSignature(payload, webhook.SecretHash)
	if err != nil {
		return fmt.Errorf("failed to create signature: %w", err)
	}

	// Send HTTP request
	statusCode, err := n.makeHTTPRequest(ctx, webhook.URL, payload, signature)
	if err != nil {
		// Record failure
		if updateErr := n.store.IncrementWebhookFailure(ctx, webhook.ID, now); updateErr != nil {
			n.logger.Printf("Failed to record webhook failure: %v", updateErr)
		}

		// Check if we should enter cooldown
		webhook.FailureCount++ // Simulate increment for cooldown check
		webhook.LastErrorAt = &now
		if n.shouldEnterCooldown(webhook) {
			cooldownUntil := now.Add(CooldownDuration)
			if updateErr := n.store.UpdateWebhookDeliveryState(ctx, webhook.ID, now, &cooldownUntil); updateErr != nil {
				n.logger.Printf("Failed to set webhook cooldown: %v", updateErr)
			}
			n.logger.Printf("Webhook %d entered cooldown until %s", webhook.ID, cooldownUntil.Format(time.RFC3339))
		}

		return fmt.Errorf("webhook delivery failed (status=%d): %w", statusCode, err)
	}

	// Record success
	if err := n.store.MarkWebhookSuccess(ctx, webhook.ID, now); err != nil {
		n.logger.Printf("Failed to record webhook success: %v", err)
	}

	n.logger.Printf("Successfully delivered machine event to webhook %d (event=%s, machine=%d)", webhook.ID, event.Event, event.Machine.ID)
	return nil
}

// SendTest sends a test webhook payload to verify the webhook configuration
func (n *Notifier) SendTest(ctx context.Context, webhook storage.Webhook) error {
	// Check rate limiting and cooldown before proceeding
	if err := n.checkDeliveryPreconditions(webhook); err != nil {
		return err
	}

	// Build test payload with clearly marked test data
	payload := WebhookPayload{
		RuleID:       0,
		RuleName:     "Test Webhook",
		Metric:       "cpu_pct",
		Comparison:   "above",
		ThresholdPct: 80.0,
		TriggerAfter: 1,
		Value:        85.5,
		TriggeredAt:  time.Now().Format(time.RFC3339),
		EventID:      0,
	}

	// Send the test payload (reuses existing sendToWebhook logic)
	return n.sendToWebhook(ctx, webhook, payload)
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
	// Check rate limiting and cooldown before proceeding
	if err := n.checkDeliveryPreconditions(webhook); err != nil {
		return err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HMAC signature
	signature, err := n.createSignature(payloadBytes, webhook.SecretHash)
	if err != nil {
		return fmt.Errorf("failed to create signature: %w", err)
	}

	// Extract domain from URL for safe logging (never log full URL or secrets)
	webhookURL, _ := url.Parse(webhook.URL)
	domain := webhookURL.Host
	if domain == "" {
		domain = webhook.URL // fallback for invalid URLs
	}

	// Update last attempt time before making the request
	now := time.Now()
	if err := n.store.UpdateWebhookDeliveryState(ctx, webhook.ID, now, webhook.CooldownUntil); err != nil {
		n.logger.Printf("[WEBHOOK] failed to update delivery state for webhook=%d: %v", webhook.ID, err)
	}

	// Retry logic with exponential backoff
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		statusCode, err := n.makeHTTPRequest(ctx, webhook.URL, payloadBytes, signature)
		if err == nil {
			// Success - mark webhook as successful and log
			if markErr := n.store.MarkWebhookSuccess(ctx, webhook.ID, time.Now()); markErr != nil {
				n.logger.Printf("[WEBHOOK] failed to mark webhook=%d as successful: %v", webhook.ID, markErr)
			}
			n.logger.Printf("[WEBHOOK] delivered webhook=%d url=%s status=%d attempt=%d",
				webhook.ID, domain, statusCode, attempt)
			return nil
		}

		// Log failure with attempt number
		n.logger.Printf("[WEBHOOK] failed webhook=%d url=%s attempt=%d error=%v",
			webhook.ID, domain, attempt, err)

		// If this was the last attempt, mark as failure and potentially set cooldown
		if attempt == maxRetries {
			if markErr := n.store.IncrementWebhookFailure(ctx, webhook.ID, time.Now()); markErr != nil {
				n.logger.Printf("[WEBHOOK] failed to increment failure count for webhook=%d: %v",
					webhook.ID, markErr)
			}

			// Check if we should enter cooldown based on failure threshold
			updatedWebhook, fetchErr := n.store.GetWebhook(ctx, webhook.ID, webhook.UserID)
			if fetchErr != nil {
				n.logger.Printf("[WEBHOOK] failed to fetch updated webhook=%d: %v", webhook.ID, fetchErr)
			} else if n.shouldEnterCooldown(*updatedWebhook) {
				cooldownUntil := time.Now().Add(CooldownDuration)
				if cooldownErr := n.store.UpdateWebhookDeliveryState(ctx, webhook.ID, now, &cooldownUntil); cooldownErr != nil {
					n.logger.Printf("[WEBHOOK] failed to set cooldown for webhook=%d: %v", webhook.ID, cooldownErr)
				} else {
					n.logger.Printf("[WEBHOOK] cooldown webhook=%d until=%s",
						webhook.ID, cooldownUntil.Format(time.RFC3339))
				}
			}

			n.logger.Printf("[WEBHOOK] failed webhook=%d url=%s attempts=%d final_error=%v",
				webhook.ID, domain, maxRetries, err)
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
func (n *Notifier) makeHTTPRequest(ctx context.Context, url string, payload []byte, signature string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LunaSentri-Signature", signature)
	req.Header.Set("User-Agent", "LunaSentri-Webhook/1.0")

	resp, err := n.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	return resp.StatusCode, nil
}

// checkDeliveryPreconditions verifies if a webhook delivery should proceed based on rate limiting and cooldown
func (n *Notifier) checkDeliveryPreconditions(webhook storage.Webhook) error {
	now := time.Now()

	// Check if webhook is in cooldown
	if webhook.CooldownUntil != nil && now.Before(*webhook.CooldownUntil) {
		n.logger.Printf("[WEBHOOK] throttled webhook=%d reason=cooldown until=%s",
			webhook.ID, webhook.CooldownUntil.Format(time.RFC3339))
		return &RateLimitError{
			Type:    "cooldown",
			Message: fmt.Sprintf("Webhook in cooldown until %s", webhook.CooldownUntil.Format(time.RFC3339)),
			RetryAt: webhook.CooldownUntil,
		}
	}

	// Check minimum interval between attempts
	if webhook.LastAttemptAt != nil {
		timeSinceLastAttempt := now.Sub(*webhook.LastAttemptAt)
		if timeSinceLastAttempt < MinAttemptInterval {
			delay := MinAttemptInterval - timeSinceLastAttempt
			n.logger.Printf("[WEBHOOK] rate_limited webhook=%d delay=%s",
				webhook.ID, delay.String())
			retryAt := webhook.LastAttemptAt.Add(MinAttemptInterval)
			return &RateLimitError{
				Type:    "rate_limit",
				Message: fmt.Sprintf("Rate limit active, can retry in %s", delay.String()),
				RetryAt: &retryAt,
			}
		}
	}

	return nil
}

// shouldEnterCooldown determines if a webhook should enter cooldown based on recent failures
func (n *Notifier) shouldEnterCooldown(webhook storage.Webhook) bool {
	// If we haven't reached the failure threshold, no cooldown
	if webhook.FailureCount < FailureThreshold {
		return false
	}

	// If we don't have a last error timestamp, enter cooldown as a safety measure
	if webhook.LastErrorAt == nil {
		return true
	}

	// Check if the failures occurred within the failure window
	now := time.Now()
	failureWindowStart := now.Add(-FailureWindow)
	return webhook.LastErrorAt.After(failureWindowStart)
}

// RateLimitError represents an error due to rate limiting or cooldown
type RateLimitError struct {
	Type    string     // "cooldown" or "rate_limit"
	Message string     // Human-readable error message
	RetryAt *time.Time // When the operation can be retried
}

func (e *RateLimitError) Error() string {
	return e.Message
}
