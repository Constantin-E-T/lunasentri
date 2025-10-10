package machines

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// MockHeartbeatNotifier is a mock notifier for testing
type MockHeartbeatNotifier struct {
	OfflineNotifications []storage.Machine
	OnlineNotifications  []storage.Machine
}

func (m *MockHeartbeatNotifier) NotifyMachineOffline(ctx context.Context, machine storage.Machine) error {
	m.OfflineNotifications = append(m.OfflineNotifications, machine)
	return nil
}

func (m *MockHeartbeatNotifier) NotifyMachineOnline(ctx context.Context, machine storage.Machine) error {
	m.OnlineNotifications = append(m.OnlineNotifications, machine)
	return nil
}

// MockStore implements the minimal Store interface needed for heartbeat monitoring
type MockHeartbeatStore struct {
	machines             []storage.Machine
	offlineNotifications map[int]time.Time
	statusUpdates        map[int]string
}

func NewMockHeartbeatStore() *MockHeartbeatStore {
	return &MockHeartbeatStore{
		machines:             []storage.Machine{},
		offlineNotifications: make(map[int]time.Time),
		statusUpdates:        make(map[int]string),
	}
}

func (m *MockHeartbeatStore) ListAllMachines(ctx context.Context) ([]storage.Machine, error) {
	return m.machines, nil
}

func (m *MockHeartbeatStore) UpdateMachineStatus(ctx context.Context, id int, status string, lastSeen time.Time) error {
	m.statusUpdates[id] = status
	// Update the machine in our list
	for i := range m.machines {
		if m.machines[i].ID == id {
			m.machines[i].Status = status
			break
		}
	}
	return nil
}

func (m *MockHeartbeatStore) RecordMachineOfflineNotification(ctx context.Context, machineID int, notifiedAt time.Time) error {
	m.offlineNotifications[machineID] = notifiedAt
	return nil
}

func (m *MockHeartbeatStore) GetMachineLastOfflineNotification(ctx context.Context, machineID int) (time.Time, error) {
	if t, ok := m.offlineNotifications[machineID]; ok {
		return t, nil
	}
	return time.Time{}, nil
}

func (m *MockHeartbeatStore) ClearMachineOfflineNotification(ctx context.Context, machineID int) error {
	delete(m.offlineNotifications, machineID)
	return nil
}

// testLogger creates a logger for tests
func testLogger() *log.Logger {
	return log.New(os.Stderr, "[TEST] ", log.LstdFlags)
}

func TestHeartbeatMonitor_MachineGoesOffline(t *testing.T) {
	store := NewMockHeartbeatStore()
	notifier := &MockHeartbeatNotifier{}
	
	// Create a machine that is online but hasn't reported in 3 minutes
	now := time.Now()
	store.machines = []storage.Machine{
		{
			ID:       1,
			UserID:   1,
			Name:     "test-machine",
			Hostname: "test-01",
			Status:   "online",
			LastSeen: now.Add(-3 * time.Minute), // 3 minutes ago
		},
	}

	cfg := HeartbeatConfig{
		CheckInterval:    100 * time.Millisecond,
		OfflineThreshold: 2 * time.Minute,
	}

	monitor := NewHeartbeatMonitor(store, notifier, testLogger(), cfg)
	ctx := context.Background()

	// Run a single check
	monitor.checkAllMachines(ctx)

	// Verify the machine was marked offline
	if store.statusUpdates[1] != "offline" {
		t.Errorf("Expected machine to be marked offline, got: %s", store.statusUpdates[1])
	}

	// Verify offline notification was sent
	if len(notifier.OfflineNotifications) != 1 {
		t.Errorf("Expected 1 offline notification, got: %d", len(notifier.OfflineNotifications))
	}

	// Verify notification was recorded
	if _, ok := store.offlineNotifications[1]; !ok {
		t.Error("Expected offline notification to be recorded")
	}
}

func TestHeartbeatMonitor_MachineComesBackOnline(t *testing.T) {
	store := NewMockHeartbeatStore()
	notifier := &MockHeartbeatNotifier{}
	
	// Create a machine that is offline but just reported
	now := time.Now()
	store.machines = []storage.Machine{
		{
			ID:       1,
			UserID:   1,
			Name:     "test-machine",
			Hostname: "test-01",
			Status:   "offline",
			LastSeen: now.Add(-30 * time.Second), // 30 seconds ago
		},
	}

	cfg := HeartbeatConfig{
		CheckInterval:    100 * time.Millisecond,
		OfflineThreshold: 2 * time.Minute,
	}

	monitor := NewHeartbeatMonitor(store, notifier, testLogger(), cfg)
	ctx := context.Background()

	// Run a single check
	monitor.checkAllMachines(ctx)

	// Verify the machine was marked online
	if store.statusUpdates[1] != "online" {
		t.Errorf("Expected machine to be marked online, got: %s", store.statusUpdates[1])
	}

	// Verify online notification was sent
	if len(notifier.OnlineNotifications) != 1 {
		t.Errorf("Expected 1 online notification, got: %d", len(notifier.OnlineNotifications))
	}

	// Verify offline notification was cleared
	if _, ok := store.offlineNotifications[1]; ok {
		t.Error("Expected offline notification to be cleared")
	}
}

func TestHeartbeatMonitor_NoDuplicateNotifications(t *testing.T) {
	store := NewMockHeartbeatStore()
	notifier := &MockHeartbeatNotifier{}
	
	// Create a machine that is already offline and we've already notified
	now := time.Now()
	store.machines = []storage.Machine{
		{
			ID:       1,
			UserID:   1,
			Name:     "test-machine",
			Hostname: "test-01",
			Status:   "offline",
			LastSeen: now.Add(-5 * time.Minute),
		},
	}
	store.offlineNotifications[1] = now.Add(-4 * time.Minute)

	cfg := HeartbeatConfig{
		CheckInterval:    100 * time.Millisecond,
		OfflineThreshold: 2 * time.Minute,
	}

	monitor := NewHeartbeatMonitor(store, notifier, testLogger(), cfg)
	ctx := context.Background()

	// Run a single check
	monitor.checkAllMachines(ctx)

	// Verify no new notification was sent
	if len(notifier.OfflineNotifications) != 0 {
		t.Errorf("Expected 0 offline notifications (already notified), got: %d", len(notifier.OfflineNotifications))
	}
}

func TestHeartbeatMonitor_StaysOnline(t *testing.T) {
	store := NewMockHeartbeatStore()
	notifier := &MockHeartbeatNotifier{}
	
	// Create a machine that is online and recently reported
	now := time.Now()
	store.machines = []storage.Machine{
		{
			ID:       1,
			UserID:   1,
			Name:     "test-machine",
			Hostname: "test-01",
			Status:   "online",
			LastSeen: now.Add(-30 * time.Second),
		},
	}

	cfg := HeartbeatConfig{
		CheckInterval:    100 * time.Millisecond,
		OfflineThreshold: 2 * time.Minute,
	}

	monitor := NewHeartbeatMonitor(store, notifier, testLogger(), cfg)
	ctx := context.Background()

	// Run a single check
	monitor.checkAllMachines(ctx)

	// Verify no status changes
	if _, changed := store.statusUpdates[1]; changed {
		t.Error("Expected no status change for healthy machine")
	}

	// Verify no notifications sent
	if len(notifier.OfflineNotifications) != 0 {
		t.Errorf("Expected 0 offline notifications, got: %d", len(notifier.OfflineNotifications))
	}
	if len(notifier.OnlineNotifications) != 0 {
		t.Errorf("Expected 0 online notifications, got: %d", len(notifier.OnlineNotifications))
	}
}

func TestHeartbeatMonitor_StaysOffline(t *testing.T) {
	store := NewMockHeartbeatStore()
	notifier := &MockHeartbeatNotifier{}
	
	// Create a machine that is offline and stays offline
	now := time.Now()
	store.machines = []storage.Machine{
		{
			ID:       1,
			UserID:   1,
			Name:     "test-machine",
			Hostname: "test-01",
			Status:   "offline",
			LastSeen: now.Add(-10 * time.Minute),
		},
	}
	// Already notified
	store.offlineNotifications[1] = now.Add(-9 * time.Minute)

	cfg := HeartbeatConfig{
		CheckInterval:    100 * time.Millisecond,
		OfflineThreshold: 2 * time.Minute,
	}

	monitor := NewHeartbeatMonitor(store, notifier, testLogger(), cfg)
	ctx := context.Background()

	// Run a single check
	monitor.checkAllMachines(ctx)

	// Verify no new notifications
	if len(notifier.OfflineNotifications) != 0 {
		t.Errorf("Expected 0 offline notifications (already offline), got: %d", len(notifier.OfflineNotifications))
	}
	if len(notifier.OnlineNotifications) != 0 {
		t.Errorf("Expected 0 online notifications, got: %d", len(notifier.OnlineNotifications))
	}
}

func TestHeartbeatMonitor_MultipleMachines(t *testing.T) {
	store := NewMockHeartbeatStore()
	notifier := &MockHeartbeatNotifier{}
	
	now := time.Now()
	store.machines = []storage.Machine{
		{
			ID:       1,
			UserID:   1,
			Name:     "machine-1",
			Status:   "online",
			LastSeen: now.Add(-30 * time.Second), // Still online
		},
		{
			ID:       2,
			UserID:   1,
			Name:     "machine-2",
			Status:   "online",
			LastSeen: now.Add(-5 * time.Minute), // Should go offline
		},
		{
			ID:       3,
			UserID:   1,
			Name:     "machine-3",
			Status:   "offline",
			LastSeen: now.Add(-1 * time.Minute), // Should come online
		},
	}

	cfg := HeartbeatConfig{
		CheckInterval:    100 * time.Millisecond,
		OfflineThreshold: 2 * time.Minute,
	}

	monitor := NewHeartbeatMonitor(store, notifier, testLogger(), cfg)
	ctx := context.Background()

	// Run a single check
	monitor.checkAllMachines(ctx)

	// Verify machine 1 stayed online (no change)
	if _, changed := store.statusUpdates[1]; changed {
		t.Error("Machine 1 should not have changed status")
	}

	// Verify machine 2 went offline
	if store.statusUpdates[2] != "offline" {
		t.Errorf("Machine 2 should be offline, got: %s", store.statusUpdates[2])
	}
	if len(notifier.OfflineNotifications) != 1 || notifier.OfflineNotifications[0].ID != 2 {
		t.Error("Expected offline notification for machine 2")
	}

	// Verify machine 3 came online
	if store.statusUpdates[3] != "online" {
		t.Errorf("Machine 3 should be online, got: %s", store.statusUpdates[3])
	}
	if len(notifier.OnlineNotifications) != 1 || notifier.OnlineNotifications[0].ID != 3 {
		t.Error("Expected online notification for machine 3")
	}
}
