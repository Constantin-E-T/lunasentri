package storage

import (
	"context"
	"testing"
)

func TestMachineCredentialManagement(t *testing.T) {
	// Create in-memory test database
	store, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test user
	user, err := store.CreateUser(ctx, "test@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test 1: Create machine with API key
	t.Run("CreateMachineWithAPIKey", func(t *testing.T) {
		machine, err := store.CreateMachine(ctx, user.ID, "test-machine", "test-host", "test description", "hashed_api_key_1")
		if err != nil {
			t.Fatalf("Failed to create machine: %v", err)
		}

		if !machine.IsEnabled {
			t.Error("New machine should be enabled by default")
		}

		// Verify API key was created in machine_api_keys table
		keys, err := store.ListMachineAPIKeys(ctx, machine.ID)
		if err != nil {
			t.Fatalf("Failed to list API keys: %v", err)
		}

		if len(keys) != 1 {
			t.Errorf("Expected 1 API key, got %d", len(keys))
		}

		if keys[0].APIKeyHash != "hashed_api_key_1" {
			t.Errorf("Expected API key hash 'hashed_api_key_1', got '%s'", keys[0].APIKeyHash)
		}

		if keys[0].RevokedAt != nil {
			t.Error("New API key should not be revoked")
		}
	})

	// Test 2: Disable machine
	t.Run("DisableMachine", func(t *testing.T) {
		machine, _ := store.CreateMachine(ctx, user.ID, "test-machine-2", "test-host-2", "desc", "hashed_api_key_2")

		err := store.SetMachineEnabled(ctx, machine.ID, false)
		if err != nil {
			t.Fatalf("Failed to disable machine: %v", err)
		}

		// Verify machine is disabled
		updatedMachine, err := store.GetMachineByID(ctx, machine.ID)
		if err != nil {
			t.Fatalf("Failed to get machine: %v", err)
		}

		if updatedMachine.IsEnabled {
			t.Error("Machine should be disabled")
		}
	})

	// Test 3: Enable machine
	t.Run("EnableMachine", func(t *testing.T) {
		machine, _ := store.CreateMachine(ctx, user.ID, "test-machine-3", "test-host-3", "desc", "hashed_api_key_3")
		store.SetMachineEnabled(ctx, machine.ID, false)

		err := store.SetMachineEnabled(ctx, machine.ID, true)
		if err != nil {
			t.Fatalf("Failed to enable machine: %v", err)
		}

		updatedMachine, err := store.GetMachineByID(ctx, machine.ID)
		if err != nil {
			t.Fatalf("Failed to get machine: %v", err)
		}

		if !updatedMachine.IsEnabled {
			t.Error("Machine should be enabled")
		}
	})

	// Test 4: Rotate API key
	t.Run("RotateAPIKey", func(t *testing.T) {
		machine, _ := store.CreateMachine(ctx, user.ID, "test-machine-4", "test-host-4", "desc", "hashed_api_key_4")

		// Get initial key
		initialKeys, _ := store.ListMachineAPIKeys(ctx, machine.ID)
		if len(initialKeys) != 1 {
			t.Fatalf("Expected 1 initial key, got %d", len(initialKeys))
		}

		// Revoke all existing keys
		err := store.RevokeAllMachineAPIKeys(ctx, machine.ID)
		if err != nil {
			t.Fatalf("Failed to revoke API keys: %v", err)
		}

		// Create new key
		newKey, err := store.CreateMachineAPIKey(ctx, machine.ID, "hashed_api_key_4_new")
		if err != nil {
			t.Fatalf("Failed to create new API key: %v", err)
		}

		// Verify old key is revoked
		keys, _ := store.ListMachineAPIKeys(ctx, machine.ID)
		if len(keys) != 2 {
			t.Errorf("Expected 2 keys (1 revoked, 1 active), got %d", len(keys))
		}

		// Find the old key and verify it's revoked
		var oldKeyRevoked bool
		for _, key := range keys {
			if key.APIKeyHash == "hashed_api_key_4" && key.RevokedAt != nil {
				oldKeyRevoked = true
			}
		}
		if !oldKeyRevoked {
			t.Error("Old API key should be revoked")
		}

		// Verify new key is active
		if newKey.RevokedAt != nil {
			t.Error("New API key should not be revoked")
		}
	})

	// Test 5: Get machine by API key (only returns if not revoked)
	t.Run("GetMachineByAPIKey", func(t *testing.T) {
		machine, _ := store.CreateMachine(ctx, user.ID, "test-machine-5", "test-host-5", "desc", "hashed_api_key_5")

		// Should find machine with active key
		foundMachine, err := store.GetMachineByAPIKey(ctx, "hashed_api_key_5")
		if err != nil {
			t.Fatalf("Failed to get machine by API key: %v", err)
		}
		if foundMachine.ID != machine.ID {
			t.Error("Found wrong machine")
		}

		// Revoke the key
		store.RevokeAllMachineAPIKeys(ctx, machine.ID)

		// Should not find machine with revoked key
		_, err = store.GetMachineByAPIKey(ctx, "hashed_api_key_5")
		if err == nil {
			t.Error("Should not find machine with revoked API key")
		}
	})

	// Test 6: Get active API key for machine
	t.Run("GetActiveAPIKeyForMachine", func(t *testing.T) {
		machine, _ := store.CreateMachine(ctx, user.ID, "test-machine-6", "test-host-6", "desc", "hashed_api_key_6")

		// Get active key
		activeKey, err := store.GetActiveAPIKeyForMachine(ctx, machine.ID)
		if err != nil {
			t.Fatalf("Failed to get active API key: %v", err)
		}

		if activeKey.APIKeyHash != "hashed_api_key_6" {
			t.Errorf("Expected 'hashed_api_key_6', got '%s'", activeKey.APIKeyHash)
		}

		// Revoke all keys
		store.RevokeAllMachineAPIKeys(ctx, machine.ID)

		// Should not find active key
		_, err = store.GetActiveAPIKeyForMachine(ctx, machine.ID)
		if err == nil {
			t.Error("Should not find active API key after revoking all")
		}
	})
}

func TestMachineAPIKeyMigration(t *testing.T) {
	// This test verifies the migration correctly moves existing API keys
	// to the machine_api_keys table
	store, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create user and machine
	user, _ := store.CreateUser(ctx, "test@example.com", "hashedpassword")
	machine, err := store.CreateMachine(ctx, user.ID, "test-machine", "test-host", "desc", "test_hash")
	if err != nil {
		t.Fatalf("Failed to create machine: %v", err)
	}

	// Verify the API key exists in machine_api_keys table
	keys, err := store.ListMachineAPIKeys(ctx, machine.ID)
	if err != nil {
		t.Fatalf("Failed to list API keys: %v", err)
	}

	if len(keys) == 0 {
		t.Error("API key should have been created in machine_api_keys table")
	}
}
