package storage

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMachineOperations(t *testing.T) {
	// Create a temporary database
	dbPath := "./test_machines.db"
	defer os.Remove(dbPath)

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test user first
	user, err := store.CreateUser(ctx, "test@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("CreateMachine", func(t *testing.T) {
		machine, err := store.CreateMachine(ctx, user.ID, "test-machine", "test.example.com", "Primary test machine", "hashed-api-key-123")
		if err != nil {
			t.Fatalf("Failed to create machine: %v", err)
		}

		if machine.ID == 0 {
			t.Error("Expected machine ID to be set")
		}
		if machine.UserID != user.ID {
			t.Errorf("Expected user ID %d, got %d", user.ID, machine.UserID)
		}
		if machine.Name != "test-machine" {
			t.Errorf("Expected name 'test-machine', got '%s'", machine.Name)
		}
		if machine.Status != "offline" {
			t.Errorf("Expected status 'offline', got '%s'", machine.Status)
		}
	})

	t.Run("GetMachineByID", func(t *testing.T) {
		// Create a machine
		created, err := store.CreateMachine(ctx, user.ID, "machine-2", "host2.com", "Second test machine", "key-456")
		if err != nil {
			t.Fatalf("Failed to create machine: %v", err)
		}

		// Retrieve it
		machine, err := store.GetMachineByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("Failed to get machine: %v", err)
		}

		if machine.ID != created.ID {
			t.Errorf("Expected ID %d, got %d", created.ID, machine.ID)
		}
		if machine.Name != "machine-2" {
			t.Errorf("Expected name 'machine-2', got '%s'", machine.Name)
		}
	})

	t.Run("GetMachineByAPIKey", func(t *testing.T) {
		apiKeyHash := "unique-hash-789"
		created, err := store.CreateMachine(ctx, user.ID, "machine-3", "host3.com", "Third test machine", apiKeyHash)
		if err != nil {
			t.Fatalf("Failed to create machine: %v", err)
		}

		machine, err := store.GetMachineByAPIKey(ctx, apiKeyHash)
		if err != nil {
			t.Fatalf("Failed to get machine by API key: %v", err)
		}

		if machine.ID != created.ID {
			t.Errorf("Expected ID %d, got %d", created.ID, machine.ID)
		}
	})

	t.Run("ListMachines", func(t *testing.T) {
		// Create another user
		user2, err := store.CreateUser(ctx, "user2@example.com", "hashedpassword2")
		if err != nil {
			t.Fatalf("Failed to create user2: %v", err)
		}

		// Create machines for both users
		_, err = store.CreateMachine(ctx, user.ID, "user1-machine1", "host1.com", "User one machine", "key-user1-1")
		if err != nil {
			t.Fatalf("Failed to create machine: %v", err)
		}

		_, err = store.CreateMachine(ctx, user2.ID, "user2-machine1", "host2.com", "User two machine", "key-user2-1")
		if err != nil {
			t.Fatalf("Failed to create machine: %v", err)
		}

		// List machines for user1
		machines, err := store.ListMachines(ctx, user.ID)
		if err != nil {
			t.Fatalf("Failed to list machines: %v", err)
		}

		// Should have at least 1 machine (we created several in previous tests)
		if len(machines) == 0 {
			t.Error("Expected at least 1 machine for user1")
		}

		// All machines should belong to user1
		for _, m := range machines {
			if m.UserID != user.ID {
				t.Errorf("Expected all machines to belong to user %d, got machine with user %d", user.ID, m.UserID)
			}
		}
	})

	t.Run("UpdateMachineStatus", func(t *testing.T) {
		machine, err := store.CreateMachine(ctx, user.ID, "status-test", "status.com", "Status machine", "key-status")
		if err != nil {
			t.Fatalf("Failed to create machine: %v", err)
		}

		now := time.Now()
		err = store.UpdateMachineStatus(ctx, machine.ID, "online", now)
		if err != nil {
			t.Fatalf("Failed to update status: %v", err)
		}

		updated, err := store.GetMachineByID(ctx, machine.ID)
		if err != nil {
			t.Fatalf("Failed to get updated machine: %v", err)
		}

		if updated.Status != "online" {
			t.Errorf("Expected status 'online', got '%s'", updated.Status)
		}

		// Check last_seen was updated (allowing 1 second tolerance)
		if updated.LastSeen.Sub(now).Abs() > time.Second {
			t.Errorf("Expected last_seen to be around %v, got %v", now, updated.LastSeen)
		}
	})

	t.Run("DeleteMachine", func(t *testing.T) {
		machine, err := store.CreateMachine(ctx, user.ID, "delete-test", "delete.com", "Delete machine", "key-delete")
		if err != nil {
			t.Fatalf("Failed to create machine: %v", err)
		}

		err = store.DeleteMachine(ctx, machine.ID, user.ID)
		if err != nil {
			t.Fatalf("Failed to delete machine: %v", err)
		}

		// Try to get the deleted machine
		_, err = store.GetMachineByID(ctx, machine.ID)
		if err == nil {
			t.Error("Expected error when getting deleted machine")
		}
	})

	t.Run("DeleteMachineAccessControl", func(t *testing.T) {
		user3, err := store.CreateUser(ctx, "user3@example.com", "hash3")
		if err != nil {
			t.Fatalf("Failed to create user3: %v", err)
		}

		machine, err := store.CreateMachine(ctx, user.ID, "protected-machine", "protected.com", "Protected machine", "key-protected")
		if err != nil {
			t.Fatalf("Failed to create machine: %v", err)
		}

		// Try to delete machine as different user
		err = store.DeleteMachine(ctx, machine.ID, user3.ID)
		if err == nil {
			t.Error("Expected error when deleting machine as different user")
		}

		// Verify machine still exists
		retrieved, err := store.GetMachineByID(ctx, machine.ID)
		if err != nil {
			t.Errorf("Machine should still exist: %v", err)
		}
		if retrieved.ID != machine.ID {
			t.Error("Machine was incorrectly deleted")
		}
	})
}

func TestMetricsOperations(t *testing.T) {
	// Create a temporary database
	dbPath := "./test_metrics.db"
	defer os.Remove(dbPath)

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create test user and machine
	user, err := store.CreateUser(ctx, "metrics@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	machine, err := store.CreateMachine(ctx, user.ID, "metrics-machine", "metrics.com", "Metrics machine", "key-metrics")
	if err != nil {
		t.Fatalf("Failed to create machine: %v", err)
	}

	t.Run("InsertMetrics", func(t *testing.T) {
		now := time.Now()
		err := store.InsertMetrics(ctx, machine.ID, 45.5, 67.8, 23.1, 12345, 67890, nil, now)
		if err != nil {
			t.Fatalf("Failed to insert metrics: %v", err)
		}
	})

	t.Run("GetLatestMetrics", func(t *testing.T) {
		// Insert multiple metrics
		now := time.Now()
		store.InsertMetrics(ctx, machine.ID, 10.0, 20.0, 30.0, 100, 200, nil, now.Add(-2*time.Minute))
		store.InsertMetrics(ctx, machine.ID, 50.0, 60.0, 70.0, 500, 600, nil, now)

		metrics, err := store.GetLatestMetrics(ctx, machine.ID)
		if err != nil {
			t.Fatalf("Failed to get latest metrics: %v", err)
		}

		// Should get the most recent one
		if metrics.CPUPct != 50.0 {
			t.Errorf("Expected CPU 50.0, got %f", metrics.CPUPct)
		}
		if metrics.MemUsedPct != 60.0 {
			t.Errorf("Expected Mem 60.0, got %f", metrics.MemUsedPct)
		}
	})

	t.Run("GetMetricsHistory", func(t *testing.T) {
		// Create a new machine for this test to avoid interference
		machine2, err := store.CreateMachine(ctx, user.ID, "history-machine", "history.com", "History machine", "key-history")
		if err != nil {
			t.Fatalf("Failed to create machine: %v", err)
		}

		now := time.Now()
		// Insert 5 metrics over time
		for i := 0; i < 5; i++ {
			timestamp := now.Add(time.Duration(-i) * time.Minute)
			err := store.InsertMetrics(ctx, machine2.ID, float64(i*10), float64(i*15), float64(i*20), int64(i*100), int64(i*200), nil, timestamp)
			if err != nil {
				t.Fatalf("Failed to insert metric %d: %v", i, err)
			}
		}

		from := now.Add(-10 * time.Minute)
		to := now.Add(1 * time.Minute)

		history, err := store.GetMetricsHistory(ctx, machine2.ID, from, to, 10)
		if err != nil {
			t.Fatalf("Failed to get metrics history: %v", err)
		}

		if len(history) != 5 {
			t.Errorf("Expected 5 metrics, got %d", len(history))
		}

		// Verify they're in descending order (newest first)
		for i := 0; i < len(history)-1; i++ {
			if history[i].Timestamp.Before(history[i+1].Timestamp) {
				t.Error("Metrics should be in descending order by timestamp")
			}
		}
	})

	t.Run("CascadeDelete", func(t *testing.T) {
		// Create machine with metrics
		machine3, err := store.CreateMachine(ctx, user.ID, "cascade-test", "cascade.com", "Cascade machine", "key-cascade")
		if err != nil {
			t.Fatalf("Failed to create machine: %v", err)
		}

		// Insert metrics
		now := time.Now()
		store.InsertMetrics(ctx, machine3.ID, 11.1, 22.2, 33.3, 111, 222, nil, now)

		// Delete machine
		err = store.DeleteMachine(ctx, machine3.ID, user.ID)
		if err != nil {
			t.Fatalf("Failed to delete machine: %v", err)
		}

		// Try to get metrics - should fail or return nothing
		_, err = store.GetLatestMetrics(ctx, machine3.ID)
		if err == nil {
			t.Error("Expected error or no metrics after cascade delete")
		}
	})
}
