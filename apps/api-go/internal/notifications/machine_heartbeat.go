package notifications

import (
	"context"
	"fmt"
	"log"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// MachineHeartbeatNotifier sends notifications for machine heartbeat events
type MachineHeartbeatNotifier struct {
	store            storage.Store
	webhookNotifier  *Notifier
	telegramNotifier *TelegramNotifier
	logger           *log.Logger
}

// NewMachineHeartbeatNotifier creates a new machine heartbeat notifier
func NewMachineHeartbeatNotifier(store storage.Store, webhookNotifier *Notifier, telegramNotifier *TelegramNotifier, logger *log.Logger) *MachineHeartbeatNotifier {
	return &MachineHeartbeatNotifier{
		store:            store,
		webhookNotifier:  webhookNotifier,
		telegramNotifier: telegramNotifier,
		logger:           logger,
	}
}

// NotifyMachineOffline sends notifications when a machine goes offline
func (n *MachineHeartbeatNotifier) NotifyMachineOffline(ctx context.Context, machine storage.Machine) error {
	// Get user for this machine
	user, err := n.store.GetUserByID(ctx, machine.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Send webhook notifications
	if n.webhookNotifier != nil {
		webhooks, err := n.store.ListWebhooks(ctx, user.ID)
		if err != nil {
			n.logger.Printf("Failed to list webhooks for user %d: %v", user.ID, err)
		} else {
			for _, webhook := range webhooks {
				if !webhook.IsActive {
					continue
				}

				event := WebhookMachineEvent{
					Event: "machine.offline",
					Machine: WebhookMachine{
						ID:          machine.ID,
						Name:        machine.Name,
						Hostname:    machine.Hostname,
						Description: machine.Description,
						Status:      "offline",
						LastSeen:    machine.LastSeen,
					},
				}

				if err := n.webhookNotifier.SendMachineEvent(ctx, webhook, event); err != nil {
					n.logger.Printf("Failed to send webhook notification for machine %d: %v", machine.ID, err)
				}
			}
		}
	}

	// Send Telegram notifications
	if n.telegramNotifier != nil {
		recipients, err := n.store.ListTelegramRecipients(ctx, user.ID)
		if err != nil {
			n.logger.Printf("Failed to list Telegram recipients for user %d: %v", user.ID, err)
		} else {
			message := fmt.Sprintf("🔴 *Machine Offline Alert*\n\n"+
				"Machine: `%s`\n"+
				"Hostname: `%s`\n"+
				"Status: Offline\n"+
				"Last Seen: %s",
				machine.Name,
				machine.Hostname,
				machine.LastSeen.Format("2006-01-02 15:04:05"))

			for _, recipient := range recipients {
				if !recipient.IsActive {
					continue
				}

				if err := n.telegramNotifier.sendToRecipient(ctx, recipient, message); err != nil {
					n.logger.Printf("Failed to send Telegram notification for machine %d: %v", machine.ID, err)
				}
			}
		}
	}

	return nil
}

// NotifyMachineOnline sends notifications when a machine comes back online
func (n *MachineHeartbeatNotifier) NotifyMachineOnline(ctx context.Context, machine storage.Machine) error {
	// Get user for this machine
	user, err := n.store.GetUserByID(ctx, machine.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Send webhook notifications
	if n.webhookNotifier != nil {
		webhooks, err := n.store.ListWebhooks(ctx, user.ID)
		if err != nil {
			n.logger.Printf("Failed to list webhooks for user %d: %v", user.ID, err)
		} else {
			for _, webhook := range webhooks {
				if !webhook.IsActive {
					continue
				}

				event := WebhookMachineEvent{
					Event: "machine.online",
					Machine: WebhookMachine{
						ID:          machine.ID,
						Name:        machine.Name,
						Hostname:    machine.Hostname,
						Description: machine.Description,
						Status:      "online",
						LastSeen:    machine.LastSeen,
					},
				}

				if err := n.webhookNotifier.SendMachineEvent(ctx, webhook, event); err != nil {
					n.logger.Printf("Failed to send webhook notification for machine %d: %v", machine.ID, err)
				}
			}
		}
	}

	// Send Telegram notifications
	if n.telegramNotifier != nil {
		recipients, err := n.store.ListTelegramRecipients(ctx, user.ID)
		if err != nil {
			n.logger.Printf("Failed to list Telegram recipients for user %d: %v", user.ID, err)
		} else {
			message := fmt.Sprintf("🟢 *Machine Recovery Alert*\n\n"+
				"Machine: `%s`\n"+
				"Hostname: `%s`\n"+
				"Status: Back Online\n"+
				"Recovered At: %s",
				machine.Name,
				machine.Hostname,
				machine.LastSeen.Format("2006-01-02 15:04:05"))

			for _, recipient := range recipients {
				if !recipient.IsActive {
					continue
				}

				if err := n.telegramNotifier.sendToRecipient(ctx, recipient, message); err != nil {
					n.logger.Printf("Failed to send Telegram notification for machine %d: %v", machine.ID, err)
				}
			}
		}
	}

	return nil
}
