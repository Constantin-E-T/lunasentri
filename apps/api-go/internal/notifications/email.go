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
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/notifications/m365"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// EmailNotifier handles email notifications for alert events using Microsoft 365
type EmailNotifier struct {
	store      storage.Store
	config     *config.EmailConfig
	tokenCache *m365.TokenCache
	client     *http.Client
	logger     *log.Logger
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(store storage.Store, cfg *config.EmailConfig, logger *log.Logger) *EmailNotifier {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &EmailNotifier{
		store:      store,
		config:     cfg,
		tokenCache: m365.NewTokenCache(),
		client:     client,
		logger:     logger,
	}
}

// Send sends email notifications for an alert event to all active email recipients
func (e *EmailNotifier) Send(ctx context.Context, rule storage.AlertRule, event storage.AlertEvent) error {
	if !e.config.IsEnabled() {
		return nil // Email notifications disabled
	}

	// Fetch all active email recipients
	recipients, err := e.getAllActiveRecipients(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active email recipients: %w", err)
	}

	if len(recipients) == 0 {
		e.logger.Println("[EMAIL] No active email recipients found, skipping notification")
		return nil
	}

	// Build email subject and body
	subject := fmt.Sprintf("LunaSentri Alert: %s", rule.Name)
	htmlBody := e.buildAlertEmailHTML(rule, event)

	// Send to all recipients concurrently
	for _, recipient := range recipients {
		go func(r storage.EmailRecipient) {
			if err := e.sendToRecipient(ctx, r, subject, htmlBody); err != nil {
				e.logger.Printf("[EMAIL] failed to send to %s: %v", r.Email, err)
			}
		}(recipient)
	}

	return nil
}

// SendTest sends a test email to verify the configuration
func (e *EmailNotifier) SendTest(ctx context.Context, recipient storage.EmailRecipient) error {
	if !e.config.IsEnabled() {
		return fmt.Errorf("email notifications are disabled")
	}

	// Check rate limiting and cooldown
	if err := e.checkDeliveryPreconditions(recipient); err != nil {
		return err
	}

	subject := "LunaSentri Test Email"
	htmlBody := `
		<html>
		<body>
			<h2>🌙 LunaSentri Test Email</h2>
			<p>This is a test email from your LunaSentri monitoring system.</p>
			<p>If you received this, your email notifications are configured correctly!</p>
			<hr>
			<p style="color: #666; font-size: 12px;">This is a test message. No action required.</p>
		</body>
		</html>
	`

	return e.sendToRecipient(ctx, recipient, subject, htmlBody)
}

// getAllActiveRecipients fetches all active email recipients from all users
func (e *EmailNotifier) getAllActiveRecipients(ctx context.Context) ([]storage.EmailRecipient, error) {
	users, err := e.store.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var allRecipients []storage.EmailRecipient
	for _, user := range users {
		recipients, err := e.store.ListEmailRecipients(ctx, user.ID)
		if err != nil {
			e.logger.Printf("[EMAIL] Failed to fetch recipients for user %d: %v", user.ID, err)
			continue
		}

		// Filter active recipients
		for _, recipient := range recipients {
			if recipient.IsActive {
				allRecipients = append(allRecipients, recipient)
			}
		}
	}

	return allRecipients, nil
}

// sendToRecipient sends an email to a specific recipient with retry logic
func (e *EmailNotifier) sendToRecipient(ctx context.Context, recipient storage.EmailRecipient, subject, htmlBody string) error {
	// Check rate limiting and cooldown
	if err := e.checkDeliveryPreconditions(recipient); err != nil {
		return err
	}

	// Update last attempt time
	now := time.Now()
	if err := e.store.UpdateEmailDeliveryState(ctx, recipient.ID, now, recipient.CooldownUntil); err != nil {
		e.logger.Printf("[EMAIL] failed to update delivery state for recipient=%d: %v", recipient.ID, err)
	}

	// Retry logic with exponential backoff
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := e.sendViaGraph(ctx, recipient.Email, subject, htmlBody)
		if err == nil {
			// Success
			if markErr := e.store.MarkEmailSuccess(ctx, recipient.ID, time.Now()); markErr != nil {
				e.logger.Printf("[EMAIL] failed to mark recipient=%d as successful: %v", recipient.ID, markErr)
			}
			e.logger.Printf("[EMAIL] delivered recipient=%d email=%s attempt=%d",
				recipient.ID, recipient.Email, attempt)
			return nil
		}

		// Log failure
		e.logger.Printf("[EMAIL] attempt=%d/%d recipient=%d error=%v",
			attempt, maxRetries, recipient.ID, err)

		// If this is the last retry, record failure
		if attempt == maxRetries {
			if incrErr := e.store.IncrementEmailFailure(ctx, recipient.ID, time.Now()); incrErr != nil {
				e.logger.Printf("[EMAIL] failed to increment failure count for recipient=%d: %v",
					recipient.ID, incrErr)
			}

			// Check if we need to enter cooldown
			e.checkAndApplyCooldown(ctx, recipient)
			return fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
		}

		// Exponential backoff: 1s, 2s, 4s
		backoff := time.Duration(1<<uint(attempt-1)) * time.Second
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// sendViaGraph sends email via Microsoft Graph API
func (e *EmailNotifier) sendViaGraph(ctx context.Context, to, subject, htmlBody string) error {
	// Fetch access token
	token, _, err := m365.FetchToken(ctx, e.config.M365TenantID, e.config.M365ClientID,
		e.config.M365Secret, e.tokenCache)
	if err != nil {
		return fmt.Errorf("failed to fetch access token: %w", err)
	}

	// Build Graph API request
	graphURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", e.config.M365Sender)

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"subject": subject,
			"body": map[string]interface{}{
				"contentType": "HTML",
				"content":     htmlBody,
			},
			"toRecipients": []map[string]interface{}{
				{
					"emailAddress": map[string]string{
						"address": to,
					},
				},
			},
		},
		"saveToSentItems": false,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", graphURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read error response
		var errorBody bytes.Buffer
		errorBody.ReadFrom(resp.Body)

		// Determine if error is retryable
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			return fmt.Errorf("[EMAIL] retryable error status=%d body=%s", resp.StatusCode, errorBody.String())
		}
		return fmt.Errorf("[EMAIL] fatal error status=%d body=%s", resp.StatusCode, errorBody.String())
	}

	return nil
}

// checkDeliveryPreconditions checks rate limiting and cooldown before delivery
func (e *EmailNotifier) checkDeliveryPreconditions(recipient storage.EmailRecipient) error {
	now := time.Now()

	// Check cooldown
	if recipient.CooldownUntil != nil && now.Before(*recipient.CooldownUntil) {
		e.logger.Printf("[EMAIL] throttled recipient=%d reason=cooldown until=%s",
			recipient.ID, recipient.CooldownUntil.Format(time.RFC3339))
		return &RateLimitError{
			Type:    "cooldown",
			Message: fmt.Sprintf("Email recipient in cooldown until %s", recipient.CooldownUntil.Format(time.RFC3339)),
			RetryAt: recipient.CooldownUntil,
		}
	}

	// Check rate limiting
	if recipient.LastAttemptAt != nil {
		timeSinceLastAttempt := now.Sub(*recipient.LastAttemptAt)
		if timeSinceLastAttempt < MinAttemptInterval {
			delay := MinAttemptInterval - timeSinceLastAttempt
			retryAt := now.Add(delay)
			e.logger.Printf("[EMAIL] rate_limited recipient=%d delay=%s", recipient.ID, delay)
			return &RateLimitError{
				Type:    "rate_limit",
				Message: fmt.Sprintf("Rate limit active, can retry in %s", delay.Round(time.Second)),
				RetryAt: &retryAt,
			}
		}
	}

	return nil
}

// checkAndApplyCooldown checks if recipient should enter cooldown based on failure pattern
func (e *EmailNotifier) checkAndApplyCooldown(ctx context.Context, recipient storage.EmailRecipient) {
	// Refetch recipient to get updated failure count
	updated, err := e.store.GetEmailRecipient(ctx, recipient.ID, recipient.UserID)
	if err != nil {
		e.logger.Printf("[EMAIL] failed to fetch recipient for cooldown check: %v", err)
		return
	}

	// Check if failures occurred within the failure window
	if updated.FailureCount >= FailureThreshold {
		if updated.LastErrorAt != nil {
			timeSinceFirstFailure := time.Since(*updated.LastErrorAt)
			if timeSinceFirstFailure <= FailureWindow {
				// Enter cooldown
				cooldownUntil := time.Now().Add(CooldownDuration)
				if err := e.store.UpdateEmailDeliveryState(ctx, recipient.ID, time.Now(), &cooldownUntil); err != nil {
					e.logger.Printf("[EMAIL] failed to set cooldown: %v", err)
				} else {
					e.logger.Printf("[EMAIL] cooldown recipient=%d until=%s",
						recipient.ID, cooldownUntil.Format(time.RFC3339))
				}
			}
		}
	}
}

// buildAlertEmailHTML generates HTML email body for an alert
func (e *EmailNotifier) buildAlertEmailHTML(rule storage.AlertRule, event storage.AlertEvent) string {
	comparisonText := "above"
	if rule.Comparison == "below" {
		comparisonText = "below"
	}

	return fmt.Sprintf(`
		<html>
		<body>
			<h2>🌙 LunaSentri Alert Triggered</h2>
			<p><strong>Rule:</strong> %s</p>
			<p><strong>Metric:</strong> %s</p>
			<p><strong>Condition:</strong> %s %.1f%%</p>
			<p><strong>Current Value:</strong> %.1f%%</p>
			<p><strong>Triggered At:</strong> %s</p>
			<hr>
			<p style="color: #666; font-size: 12px;">
				This alert was triggered after %d consecutive samples exceeded the threshold.
			</p>
		</body>
		</html>
	`, rule.Name, rule.Metric, comparisonText, rule.ThresholdPct, event.Value,
		event.TriggeredAt.Format(time.RFC3339), rule.TriggerAfter)
}

// Notify implements AlertNotifier interface for the email notifier
func (e *EmailNotifier) Notify(ctx context.Context, rule storage.AlertRule, event *storage.AlertEvent) error {
	if event == nil {
		return nil
	}
	return e.Send(ctx, rule, *event)
}
