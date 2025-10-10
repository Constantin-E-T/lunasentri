package machines

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// HeartbeatStore defines the storage operations needed for heartbeat monitoring
type HeartbeatStore interface {
	ListAllMachines(ctx context.Context) ([]storage.Machine, error)
	UpdateMachineStatus(ctx context.Context, id int, status string, lastSeen time.Time) error
	RecordMachineOfflineNotification(ctx context.Context, machineID int, notifiedAt time.Time) error
	GetMachineLastOfflineNotification(ctx context.Context, machineID int) (time.Time, error)
	ClearMachineOfflineNotification(ctx context.Context, machineID int) error
}

// HeartbeatNotifier defines the interface for sending heartbeat notifications
type HeartbeatNotifier interface {
	// NotifyMachineOffline sends notifications when a machine goes offline
	NotifyMachineOffline(ctx context.Context, machine storage.Machine) error
	// NotifyMachineOnline sends notifications when a machine comes back online
	NotifyMachineOnline(ctx context.Context, machine storage.Machine) error
}

// HeartbeatMonitor monitors machine heartbeats and triggers notifications
type HeartbeatMonitor struct {
	store            HeartbeatStore
	notifier         HeartbeatNotifier
	logger           *log.Logger
	checkInterval    time.Duration
	offlineThreshold time.Duration
	stopCh           chan struct{}
	doneCh           chan struct{}
}

// HeartbeatConfig holds configuration for the heartbeat monitor
type HeartbeatConfig struct {
	CheckInterval    time.Duration // How often to check machine statuses
	OfflineThreshold time.Duration // How old last_seen must be to consider offline
}

// NewHeartbeatMonitor creates a new heartbeat monitor
func NewHeartbeatMonitor(store HeartbeatStore, notifier HeartbeatNotifier, logger *log.Logger, cfg HeartbeatConfig) *HeartbeatMonitor {
	return &HeartbeatMonitor{
		store:            store,
		notifier:         notifier,
		logger:           logger,
		checkInterval:    cfg.CheckInterval,
		offlineThreshold: cfg.OfflineThreshold,
		stopCh:           make(chan struct{}),
		doneCh:           make(chan struct{}),
	}
}

// Start begins the heartbeat monitoring loop in a background goroutine
func (m *HeartbeatMonitor) Start(ctx context.Context) {
	go m.run(ctx)
	m.logger.Printf("Heartbeat monitor started (check interval: %v, offline threshold: %v)", m.checkInterval, m.offlineThreshold)
}

// Stop gracefully stops the heartbeat monitor
func (m *HeartbeatMonitor) Stop() {
	close(m.stopCh)
	<-m.doneCh
	m.logger.Println("Heartbeat monitor stopped")
}

// run is the main monitoring loop
func (m *HeartbeatMonitor) run(ctx context.Context) {
	defer close(m.doneCh)

	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	// Run initial check immediately
	m.checkAllMachines(ctx)

	for {
		select {
		case <-ticker.C:
			m.checkAllMachines(ctx)
		case <-m.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// checkAllMachines checks all machines and updates statuses + sends notifications
func (m *HeartbeatMonitor) checkAllMachines(ctx context.Context) {
	machines, err := m.store.ListAllMachines(ctx)
	if err != nil {
		m.logger.Printf("Error listing machines for heartbeat check: %v", err)
		return
	}

	now := time.Now()

	for _, machine := range machines {
		if err := m.checkMachine(ctx, &machine, now); err != nil {
			m.logger.Printf("Error checking machine %d (%s): %v", machine.ID, machine.Hostname, err)
		}
	}
}

// checkMachine checks a single machine and handles status transitions
func (m *HeartbeatMonitor) checkMachine(ctx context.Context, machine *storage.Machine, now time.Time) error {
	// Compute current status based on last_seen
	timeSinceLastSeen := now.Sub(machine.LastSeen)
	isCurrentlyOnline := !machine.LastSeen.IsZero() && timeSinceLastSeen <= m.offlineThreshold
	newStatus := "offline"
	if isCurrentlyOnline {
		newStatus = "online"
	}

	// Check for status transitions
	previousStatus := machine.Status

	// Offline → Online transition (recovery)
	if previousStatus == "offline" && newStatus == "online" {
		m.logger.Printf("Machine %d (%s) came back online", machine.ID, machine.Hostname)

		// Update status in database
		if err := m.store.UpdateMachineStatus(ctx, machine.ID, "online", machine.LastSeen); err != nil {
			return fmt.Errorf("failed to update machine status: %w", err)
		}

		// Clear offline notification tracking
		if err := m.store.ClearMachineOfflineNotification(ctx, machine.ID); err != nil {
			m.logger.Printf("Warning: failed to clear offline notification for machine %d: %v", machine.ID, err)
		}

		// Send recovery notification
		if m.notifier != nil {
			machine.Status = "online"
			if err := m.notifier.NotifyMachineOnline(ctx, *machine); err != nil {
				m.logger.Printf("Failed to send online notification for machine %d: %v", machine.ID, err)
			}
		}
		return nil
	}

	// Online → Offline transition (went down)
	if previousStatus == "online" && newStatus == "offline" {
		m.logger.Printf("Machine %d (%s) went offline (last seen: %v ago)", machine.ID, machine.Hostname, timeSinceLastSeen)

		// Update status in database
		if err := m.store.UpdateMachineStatus(ctx, machine.ID, "offline", machine.LastSeen); err != nil {
			return fmt.Errorf("failed to update machine status: %w", err)
		}

		// Check if we've already notified
		lastNotified, err := m.store.GetMachineLastOfflineNotification(ctx, machine.ID)
		if err != nil {
			// If error or no notification record, proceed with notification
			lastNotified = time.Time{}
		}

		// Only notify if we haven't notified recently (avoid duplicates)
		if lastNotified.IsZero() || now.Sub(lastNotified) > m.offlineThreshold {
			if m.notifier != nil {
				machine.Status = "offline"
				if err := m.notifier.NotifyMachineOffline(ctx, *machine); err != nil {
					m.logger.Printf("Failed to send offline notification for machine %d: %v", machine.ID, err)
				} else {
					// Record that we sent the notification
					if err := m.store.RecordMachineOfflineNotification(ctx, machine.ID, now); err != nil {
						m.logger.Printf("Warning: failed to record offline notification for machine %d: %v", machine.ID, err)
					}
				}
			}
		}
		return nil
	}

	// Still offline (no change)
	if previousStatus == "offline" && newStatus == "offline" {
		// Do nothing - already offline and already notified
		return nil
	}

	// Still online (no change)
	if previousStatus == "online" && newStatus == "online" {
		// Do nothing - machine is healthy
		return nil
	}

	// Handle edge case: machine was never seen before (status is empty/unknown)
	if previousStatus != "online" && previousStatus != "offline" {
		// Set initial status
		if err := m.store.UpdateMachineStatus(ctx, machine.ID, newStatus, machine.LastSeen); err != nil {
			return fmt.Errorf("failed to set initial machine status: %w", err)
		}
	}

	return nil
}
