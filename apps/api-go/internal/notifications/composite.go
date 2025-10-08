package notifications

import (
	"context"
	"log"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// CompositeNotifier sends notifications to multiple channels (webhooks, email, etc.)
type CompositeNotifier struct {
	notifiers []AlertNotifier
	logger    *log.Logger
}

// NewCompositeNotifier creates a notifier that fans out to multiple channels
func NewCompositeNotifier(logger *log.Logger, notifiers ...AlertNotifier) *CompositeNotifier {
	// Filter out nil notifiers
	var activeNotifiers []AlertNotifier
	for _, n := range notifiers {
		if n != nil {
			activeNotifiers = append(activeNotifiers, n)
		}
	}

	return &CompositeNotifier{
		notifiers: activeNotifiers,
		logger:    logger,
	}
}

// Notify sends the alert notification to all configured channels
func (c *CompositeNotifier) Notify(ctx context.Context, rule storage.AlertRule, event *storage.AlertEvent) error {
	if event == nil {
		return nil
	}

	// Fan out to all notifiers concurrently
	// We don't want one channel's failure to block others
	for _, notifier := range c.notifiers {
		go func(n AlertNotifier) {
			if err := n.Notify(ctx, rule, event); err != nil {
				c.logger.Printf("[COMPOSITE_NOTIFIER] failed to send notification: %v", err)
			}
		}(notifier)
	}

	return nil
}
