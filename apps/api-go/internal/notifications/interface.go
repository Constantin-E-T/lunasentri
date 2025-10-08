package notifications

import (
	"context"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// AlertNotifier defines the interface for sending alert notifications
type AlertNotifier interface {
	// Notify sends notifications for an alert event
	Notify(ctx context.Context, rule storage.AlertRule, event *storage.AlertEvent) error
}

// WebhookTester defines the interface for testing webhook delivery
type WebhookTester interface {
	// SendTest sends a test notification to a webhook
	SendTest(ctx context.Context, webhook storage.Webhook) error
}

// Notify implements AlertNotifier interface for the webhook notifier
func (n *Notifier) Notify(ctx context.Context, rule storage.AlertRule, event *storage.AlertEvent) error {
	if event == nil {
		return nil
	}
	return n.Send(ctx, rule, *event)
}
