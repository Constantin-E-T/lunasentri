package notifications

import (
	"context"
	"log"
	"time"

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
	// Each notifier gets its own independent context to prevent cancellation interference
	for _, notifier := range c.notifiers {
		go func(n AlertNotifier) {
			// Create independent context with its own timeout
			// This prevents one notifier's context cancellation from affecting others
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := n.Notify(notifyCtx, rule, event); err != nil {
				c.logger.Printf("[COMPOSITE_NOTIFIER] failed to send notification: %v", err)
			}
		}(notifier)
	}

	return nil
}
