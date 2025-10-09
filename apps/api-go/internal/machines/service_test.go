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
		machine, apiKey, err := service.RegisterMachine(ctx, user.ID, "my-server", "server.example.com")
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
		machine, _, err := service.RegisterMachine(ctx, user.ID, "get-test", "get.com")
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

		machine, _, err := service.RegisterMachine(ctx, user.ID, "protected", "protected.com")
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
		service.RegisterMachine(ctx, user.ID, "server-1", "s1.com")
		service.RegisterMachine(ctx, user.ID, "server-2", "s2.com")

		machines, err := service.ListMachines(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to list machines: %v", err)
		}

		if len(machines) < 2 {
			t.Errorf("Expected at least 2 machines, got %d", len(machines))
		}
	})

	t.Run("AuthenticateMachine", func(t *testing.T) {
		_, apiKey, err := service.RegisterMachine(ctx, user.ID, "auth-test", "auth.com")
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
		machine, _, err := service.RegisterMachine(ctx, user.ID, "metrics-test", "metrics.com")
		if err != nil {
			t.Fatalf("Failed to register machine: %v", err)
		}

		// Record metrics
		err = service.RecordMetrics(ctx, machine.ID, 45.5, 67.8, 23.1, 12345, 67890)
		if err != nil {
			t.Fatalf("Failed to record metrics: %v", err)
		}

		// Verify machine status updated to online
		updated, err := store.GetMachineByID(ctx, machine.ID)
		if err != nil {
			t.Fatalf("Failed to get updated machine: %v", err)
		}
		if updated.Status != "online" {
			t.Errorf("Expected status 'online', got '%s'", updated.Status)
		}

		// Verify last_seen was updated recently
		if time.Since(updated.LastSeen) > 5*time.Second {
			t.Error("Last seen should be very recent")
		}
	})

	t.Run("GetLatestMetrics", func(t *testing.T) {
		machine, _, err := service.RegisterMachine(ctx, user.ID, "latest-test", "latest.com")
		if err != nil {
			t.Fatalf("Failed to register machine: %v", err)
		}

		// Record some metrics
		service.RecordMetrics(ctx, machine.ID, 10.0, 20.0, 30.0, 100, 200)
		time.Sleep(10 * time.Millisecond)
		service.RecordMetrics(ctx, machine.ID, 50.0, 60.0, 70.0, 500, 600)

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
		machine, _, err := service.RegisterMachine(ctx, user.ID, "history-test", "history.com")
		if err != nil {
			t.Fatalf("Failed to register machine: %v", err)
		}

		// Record several metrics
		for i := 0; i < 3; i++ {
			service.RecordMetrics(ctx, machine.ID, float64(i*10), float64(i*15), float64(i*20), int64(i*100), int64(i*200))
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
		machine, _, err := service.RegisterMachine(ctx, user.ID, "delete-test", "delete.com")
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
