package machines

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

func TestMachineService(t *testing.T) {
	// Create a temporary database
	dbPath := "./test_service.db"
	defer os.Remove(dbPath)

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	service := NewService(store)
	ctx := context.Background()

	// Create test user
	user, err := store.CreateUser(ctx, "service@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("RegisterMachine", func(t *testing.T) {
		machine, apiKey, err := service.RegisterMachine(ctx, user.ID, "my-server", "server.example.com", "Primary test machine")
		if err != nil {
			t.Fatalf("Failed to register machine: %v", err)
		}

		if machine.ID == 0 {
			t.Error("Expected machine ID to be set")
		}
		if machine.Name != "my-server" {
			t.Errorf("Expected name 'my-server', got '%s'", machine.Name)
		}
		if apiKey == "" {
			t.Error("Expected API key to be returned")
		}
		if len(apiKey) < 20 {
			t.Error("API key seems too short")
		}

		// Verify we can authenticate with the API key
		authenticated, err := service.AuthenticateMachine(ctx, apiKey)
		if err != nil {
			t.Fatalf("Failed to authenticate with API key: %v", err)
		}
		if authenticated.ID != machine.ID {
			t.Error("Authenticated machine ID doesn't match")
		}
	})

	t.Run("GetMachine", func(t *testing.T) {
		machine, _, err := service.RegisterMachine(ctx, user.ID, "get-test", "get.com", "")
		if err != nil {
			t.Fatalf("Failed to register machine: %v", err)
		}

		retrieved, err := service.GetMachine(ctx, machine.ID, user.ID)
		if err != nil {
			t.Fatalf("Failed to get machine: %v", err)
		}

		if retrieved.ID != machine.ID {
			t.Error("Retrieved machine ID doesn't match")
		}
	})

	t.Run("GetMachineAccessControl", func(t *testing.T) {
		user2, err := store.CreateUser(ctx, "user2@example.com", "hash2")
		if err != nil {
			t.Fatalf("Failed to create user2: %v", err)
		}

		machine, _, err := service.RegisterMachine(ctx, user.ID, "protected", "protected.com", "")
		if err != nil {
			t.Fatalf("Failed to register machine: %v", err)
		}

		// Try to get machine as different user
		_, err = service.GetMachine(ctx, machine.ID, user2.ID)
		if err == nil {
			t.Error("Expected error when accessing machine owned by different user")
		}
	})

	t.Run("ListMachines", func(t *testing.T) {
		// Register a few machines
		service.RegisterMachine(ctx, user.ID, "server-1", "s1.com", "")
		service.RegisterMachine(ctx, user.ID, "server-2", "s2.com", "")

		machines, err := service.ListMachines(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to list machines: %v", err)
		}

		if len(machines) < 2 {
			t.Errorf("Expected at least 2 machines, got %d", len(machines))
		}
	})

	t.Run("AuthenticateMachine", func(t *testing.T) {
		_, apiKey, err := service.RegisterMachine(ctx, user.ID, "auth-test", "auth.com", "")
		if err != nil {
			t.Fatalf("Failed to register machine: %v", err)
		}

		// Authenticate with valid key
		machine, err := service.AuthenticateMachine(ctx, apiKey)
		if err != nil {
			t.Fatalf("Failed to authenticate: %v", err)
		}
		if machine.Name != "auth-test" {
			t.Error("Authenticated wrong machine")
		}

		// Try with invalid key
		_, err = service.AuthenticateMachine(ctx, "invalid-key-xyz")
		if err == nil {
			t.Error("Expected error with invalid API key")
		}
	})

	t.Run("RecordMetrics", func(t *testing.T) {
		machine, _, err := service.RegisterMachine(ctx, user.ID, "metrics-test", "metrics.com", "")
		if err != nil {
			t.Fatalf("Failed to register machine: %v", err)
		}

		// Record metrics
		err = service.RecordMetrics(ctx, machine.ID, 45.5, 67.8, 23.1, 12345, 67890, nil, nil)
		if err != nil {
			t.Fatalf("Failed to record metrics: %v", err)
		}

		// Verify machine last_seen was updated (status is managed by heartbeat monitor, not metrics endpoint)
		updated, err := store.GetMachineByID(ctx, machine.ID)
		if err != nil {
			t.Fatalf("Failed to get updated machine: %v", err)
		}

		// Verify last_seen was updated recently
		if time.Since(updated.LastSeen) > 5*time.Second {
			t.Error("Last seen should be very recent")
		}
	})

	t.Run("RecordMetricsWithSystemInfo", func(t *testing.T) {
		machine, _, err := service.RegisterMachine(ctx, user.ID, "system-info-test", "sys.example.com", "")
		if err != nil {
			t.Fatalf("Failed to register machine: %v", err)
		}

		hostname := "agent-host"
		platform := "linux"
		platformVersion := "debian 12"
		kernel := "6.1.0"
		cpuCores := 4
		mem := int64(16384)
		disk := int64(256)
		lastBoot := time.Now().Add(-12 * time.Hour).UTC()
		uptime := 43200.0

		info := &AgentSystemInfo{
			Hostname:        &hostname,
			Platform:        &platform,
			PlatformVersion: &platformVersion,
			KernelVersion:   &kernel,
			CPUCores:        &cpuCores,
			MemoryTotalMB:   &mem,
			DiskTotalGB:     &disk,
			LastBootTime:    &lastBoot,
		}

		if err := service.RecordMetrics(ctx, machine.ID, 25.0, 35.0, 45.0, 1000, 2000, &uptime, info); err != nil {
			t.Fatalf("Failed to record metrics with system info: %v", err)
		}

		updated, err := store.GetMachineByID(ctx, machine.ID)
		if err != nil {
			t.Fatalf("Failed to reload machine: %v", err)
		}

		if updated.Hostname != hostname {
			t.Errorf("expected hostname %s, got %s", hostname, updated.Hostname)
		}
		if updated.Platform != platform {
			t.Errorf("expected platform %s, got %s", platform, updated.Platform)
		}
		if updated.PlatformVersion != platformVersion {
			t.Errorf("expected platform version %s, got %s", platformVersion, updated.PlatformVersion)
		}
		if updated.KernelVersion != kernel {
			t.Errorf("expected kernel %s, got %s", kernel, updated.KernelVersion)
		}
		if updated.CPUCores != cpuCores {
			t.Errorf("expected cpu cores %d, got %d", cpuCores, updated.CPUCores)
		}
		if updated.MemoryTotalMB != mem {
			t.Errorf("expected memory %d, got %d", mem, updated.MemoryTotalMB)
		}
		if updated.DiskTotalGB != disk {
			t.Errorf("expected disk %d, got %d", disk, updated.DiskTotalGB)
		}
		if updated.LastBootTime.IsZero() {
			t.Errorf("expected last boot time to be set")
		}

		latest, err := service.GetLatestMetrics(ctx, machine.ID, user.ID)
		if err != nil {
			t.Fatalf("failed to fetch latest metrics: %v", err)
		}
		if latest.UptimeSeconds != uptime {
			t.Errorf("expected uptime %.0f, got %.0f", uptime, latest.UptimeSeconds)
		}
	})

	t.Run("GetLatestMetrics", func(t *testing.T) {
		machine, _, err := service.RegisterMachine(ctx, user.ID, "latest-test", "latest.com", "Latest metrics machine")
		if err != nil {
			t.Fatalf("Failed to register machine: %v", err)
		}

		// Record some metrics
		service.RecordMetrics(ctx, machine.ID, 10.0, 20.0, 30.0, 100, 200, nil, nil)
		time.Sleep(10 * time.Millisecond)
		service.RecordMetrics(ctx, machine.ID, 50.0, 60.0, 70.0, 500, 600, nil, nil)

		// Get latest
		metrics, err := service.GetLatestMetrics(ctx, machine.ID, user.ID)
		if err != nil {
			t.Fatalf("Failed to get latest metrics: %v", err)
		}

		if metrics.CPUPct != 50.0 {
			t.Errorf("Expected CPU 50.0, got %f", metrics.CPUPct)
		}
	})

	t.Run("GetMetricsHistory", func(t *testing.T) {
		machine, _, err := service.RegisterMachine(ctx, user.ID, "history-test", "history.com", "History metrics machine")
		if err != nil {
			t.Fatalf("Failed to register machine: %v", err)
		}

		// Record several metrics
		for i := 0; i < 3; i++ {
			service.RecordMetrics(ctx, machine.ID, float64(i*10), float64(i*15), float64(i*20), int64(i*100), int64(i*200), nil, nil)
			time.Sleep(10 * time.Millisecond)
		}

		from := time.Now().Add(-1 * time.Minute)
		to := time.Now()

		history, err := service.GetMetricsHistory(ctx, machine.ID, user.ID, from, to, 10)
		if err != nil {
			t.Fatalf("Failed to get metrics history: %v", err)
		}

		if len(history) < 3 {
			t.Errorf("Expected at least 3 metrics, got %d", len(history))
		}
	})

	t.Run("APIKeyHashing", func(t *testing.T) {
		key := "test-api-key-12345"
		hash1 := HashAPIKey(key)
		hash2 := HashAPIKey(key)

		// Same key should produce same hash
		if hash1 != hash2 {
			t.Error("Same key produced different hashes")
		}

		// Different key should produce different hash
		hash3 := HashAPIKey("different-key")
		if hash1 == hash3 {
			t.Error("Different keys produced same hash")
		}

		// Hash should be hex string (SHA-256 produces 64 hex chars)
		if len(hash1) != 64 {
			t.Errorf("Expected hash length 64, got %d", len(hash1))
		}
	})

	t.Run("DeleteMachine", func(t *testing.T) {
		machine, _, err := service.RegisterMachine(ctx, user.ID, "delete-test", "delete.com", "Delete test machine")
		if err != nil {
			t.Fatalf("Failed to register machine: %v", err)
		}

		err = service.DeleteMachine(ctx, machine.ID, user.ID)
		if err != nil {
			t.Fatalf("Failed to delete machine: %v", err)
		}

		// Try to get deleted machine
		_, err = service.GetMachine(ctx, machine.ID, user.ID)
		if err == nil {
			t.Error("Expected error when getting deleted machine")
		}
	})
}

func TestGenerateAPIKey(t *testing.T) {
	// Generate several keys
	keys := make(map[string]bool)
	for i := 0; i < 10; i++ {
		key, err := GenerateAPIKey()
		if err != nil {
			t.Fatalf("Failed to generate API key: %v", err)
		}

		// Should not be empty
		if key == "" {
			t.Error("Generated empty API key")
		}

		// Should be reasonably long
		if len(key) < 20 {
			t.Errorf("API key seems too short: %s", key)
		}

		// Should be unique
		if keys[key] {
			t.Error("Generated duplicate API key")
		}
		keys[key] = true
	}
}

func TestStatusHelpers(t *testing.T) {
	t.Run("IsOnline", func(t *testing.T) {
		// Recent timestamp should be online
		recentTime := time.Now().Add(-30 * time.Second)
		if !IsOnline(recentTime) {
			t.Error("Expected recent timestamp to be online")
		}

		// Old timestamp should be offline
		oldTime := time.Now().Add(-5 * time.Minute)
		if IsOnline(oldTime) {
			t.Error("Expected old timestamp to be offline")
		}

		// Zero time should be offline
		if IsOnline(time.Time{}) {
			t.Error("Expected zero time to be offline")
		}

		// Boundary test - just under threshold
		justOnline := time.Now().Add(-OfflineThreshold + 5*time.Second)
		if !IsOnline(justOnline) {
			t.Error("Expected timestamp just under threshold to be online")
		}

		// Boundary test - just over threshold
		justOffline := time.Now().Add(-OfflineThreshold - 5*time.Second)
		if IsOnline(justOffline) {
			t.Error("Expected timestamp just over threshold to be offline")
		}
	})

	t.Run("ComputeStatus", func(t *testing.T) {
		// Recent timestamp should return "online"
		recentTime := time.Now().Add(-30 * time.Second)
		if status := ComputeStatus(recentTime); status != "online" {
			t.Errorf("Expected status 'online', got '%s'", status)
		}

		// Old timestamp should return "offline"
		oldTime := time.Now().Add(-5 * time.Minute)
		if status := ComputeStatus(oldTime); status != "offline" {
			t.Errorf("Expected status 'offline', got '%s'", status)
		}

		// Zero time should return "offline"
		if status := ComputeStatus(time.Time{}); status != "offline" {
			t.Errorf("Expected status 'offline', got '%s'", status)
		}
	})
}

func TestGetMachineWithComputedStatus(t *testing.T) {
	// Create a temporary database
	dbPath := "./test_computed_status.db"
	defer os.Remove(dbPath)

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	service := NewService(store)
	ctx := context.Background()

	// Create test user
	user, err := store.CreateUser(ctx, "status@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Register machine
	machine, _, err := service.RegisterMachine(ctx, user.ID, "status-test", "status.local", "Status computation machine")
	if err != nil {
		t.Fatalf("Failed to register machine: %v", err)
	}

	// Initially should be offline (no metrics yet)
	retrieved, err := service.GetMachineWithComputedStatus(ctx, machine.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get machine: %v", err)
	}
	if retrieved.Status != "offline" {
		t.Errorf("Expected status 'offline', got '%s'", retrieved.Status)
	}

	// Record metrics (should make it online)
	err = service.RecordMetrics(ctx, machine.ID, 50.0, 60.0, 70.0, 1024, 2048, nil, nil)
	if err != nil {
		t.Fatalf("Failed to record metrics: %v", err)
	}

	// Now should be online
	retrieved, err = service.GetMachineWithComputedStatus(ctx, machine.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get machine: %v", err)
	}
	if retrieved.Status != "online" {
		t.Errorf("Expected status 'online', got '%s'", retrieved.Status)
	}
}

func TestListMachinesWithComputedStatus(t *testing.T) {
	// Create a temporary database
	dbPath := "./test_list_computed_status.db"
	defer os.Remove(dbPath)

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	service := NewService(store)
	ctx := context.Background()

	// Create test user
	user, err := store.CreateUser(ctx, "list@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Register multiple machines
	machine1, _, err := service.RegisterMachine(ctx, user.ID, "machine1", "m1.local", "First machine")
	if err != nil {
		t.Fatalf("Failed to register machine1: %v", err)
	}

	machine2, _, err := service.RegisterMachine(ctx, user.ID, "machine2", "m2.local", "Second machine")
	if err != nil {
		t.Fatalf("Failed to register machine2: %v", err)
	}

	// Record metrics only for machine1
	err = service.RecordMetrics(ctx, machine1.ID, 50.0, 60.0, 70.0, 1024, 2048, nil, nil)
	if err != nil {
		t.Fatalf("Failed to record metrics: %v", err)
	}

	// List machines with computed status
	machines, err := service.ListMachinesWithComputedStatus(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to list machines: %v", err)
	}

	if len(machines) != 2 {
		t.Fatalf("Expected 2 machines, got %d", len(machines))
	}

	// Find machines in the list
	var m1Status, m2Status string
	for _, m := range machines {
		if m.ID == machine1.ID {
			m1Status = m.Status
		} else if m.ID == machine2.ID {
			m2Status = m.Status
		}
	}

	// machine1 should be online, machine2 should be offline
	if m1Status != "online" {
		t.Errorf("Expected machine1 status 'online', got '%s'", m1Status)
	}
	if m2Status != "offline" {
		t.Errorf("Expected machine2 status 'offline', got '%s'", m2Status)
	}
}
