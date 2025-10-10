package machines

import (
	"context"
	"strings"
	"testing"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

func TestDisableMachine(t *testing.T) {
	// Create in-memory test database
	store, err := storage.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	svc := NewService(store)
	ctx := context.Background()

	// Create test user
	user, _ := store.CreateUser(ctx, "test@example.com", "hashedpassword")

	// Create machine
	machine, apiKey, err := svc.RegisterMachine(ctx, user.ID, "test-machine", "test-host", "test description")
	if err != nil {
		t.Fatalf("Failed to register machine: %v", err)
	}

	// Machine should be enabled by default
	if !machine.IsEnabled {
		t.Error("New machine should be enabled")
	}

	// Should be able to authenticate with API key
	_, err = svc.AuthenticateMachine(ctx, apiKey)
	if err != nil {
		t.Errorf("Should be able to authenticate with valid API key: %v", err)
	}

	// Disable machine
	err = svc.DisableMachine(ctx, machine.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to disable machine: %v", err)
	}

	// Should not be able to authenticate after disabling
	_, err = svc.AuthenticateMachine(ctx, apiKey)
	if err == nil {
		t.Error("Should not be able to authenticate with disabled machine")
	}
	if !strings.Contains(err.Error(), "machine disabled") {
		t.Errorf("Expected 'machine disabled' error, got: %v", err)
	}
}

func TestEnableMachine(t *testing.T) {
	store, _ := storage.NewSQLiteStore(":memory:")
	defer store.Close()

	svc := NewService(store)
	ctx := context.Background()

	user, _ := store.CreateUser(ctx, "test@example.com", "hashedpassword")
	machine, apiKey, _ := svc.RegisterMachine(ctx, user.ID, "test-machine", "test-host", "test description")

	// Disable and then re-enable
	svc.DisableMachine(ctx, machine.ID, user.ID)
	err := svc.EnableMachine(ctx, machine.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to enable machine: %v", err)
	}

	// Should be able to authenticate after enabling
	_, err = svc.AuthenticateMachine(ctx, apiKey)
	if err != nil {
		t.Errorf("Should be able to authenticate after enabling: %v", err)
	}
}

func TestRotateMachineAPIKey(t *testing.T) {
	store, _ := storage.NewSQLiteStore(":memory:")
	defer store.Close()

	svc := NewService(store)
	ctx := context.Background()

	user, _ := store.CreateUser(ctx, "test@example.com", "hashedpassword")
	machine, oldAPIKey, _ := svc.RegisterMachine(ctx, user.ID, "test-machine", "test-host", "test description")

	// Rotate the API key
	newAPIKey, err := svc.RotateMachineAPIKey(ctx, machine.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to rotate API key: %v", err)
	}

	// New and old keys should be different
	if newAPIKey == oldAPIKey {
		t.Error("New API key should be different from old key")
	}

	// Old key should not work
	_, err = svc.AuthenticateMachine(ctx, oldAPIKey)
	if err == nil {
		t.Error("Old API key should not work after rotation")
	}

	// New key should work
	authenticatedMachine, err := svc.AuthenticateMachine(ctx, newAPIKey)
	if err != nil {
		t.Errorf("New API key should work: %v", err)
	}

	if authenticatedMachine.ID != machine.ID {
		t.Error("Authenticated machine should match original machine")
	}
}

func TestRotateMachineAPIKeyUnauthorized(t *testing.T) {
	store, _ := storage.NewSQLiteStore(":memory:")
	defer store.Close()

	svc := NewService(store)
	ctx := context.Background()

	user1, _ := store.CreateUser(ctx, "user1@example.com", "hashedpassword")
	user2, _ := store.CreateUser(ctx, "user2@example.com", "hashedpassword")

	machine, _, _ := svc.RegisterMachine(ctx, user1.ID, "test-machine", "test-host", "test description")

	// User2 should not be able to rotate user1's machine key
	_, err := svc.RotateMachineAPIKey(ctx, machine.ID, user2.ID)
	if err == nil {
		t.Error("User should not be able to rotate another user's machine API key")
	}
}

func TestAuthenticateMachineWithRevokedKey(t *testing.T) {
	store, _ := storage.NewSQLiteStore(":memory:")
	defer store.Close()

	svc := NewService(store)
	ctx := context.Background()

	user, _ := store.CreateUser(ctx, "test@example.com", "hashedpassword")
	machine, apiKey, _ := svc.RegisterMachine(ctx, user.ID, "test-machine", "test-host", "test description")

	// Revoke all keys
	err := store.RevokeAllMachineAPIKeys(ctx, machine.ID)
	if err != nil {
		t.Fatalf("Failed to revoke API keys: %v", err)
	}

	// Should not be able to authenticate with revoked key
	_, err = svc.AuthenticateMachine(ctx, apiKey)
	if err == nil {
		t.Error("Should not be able to authenticate with revoked API key")
	}
}

func TestGetMachineAPIKeyInfo(t *testing.T) {
	store, _ := storage.NewSQLiteStore(":memory:")
	defer store.Close()

	svc := NewService(store)
	ctx := context.Background()

	user, _ := store.CreateUser(ctx, "test@example.com", "hashedpassword")
	machine, _, _ := svc.RegisterMachine(ctx, user.ID, "test-machine", "test-host", "test description")

	// Get initial key info
	keys, err := svc.GetMachineAPIKeyInfo(ctx, machine.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get API key info: %v", err)
	}

	if len(keys) != 1 {
		t.Errorf("Expected 1 API key, got %d", len(keys))
	}

	// Rotate and check again
	svc.RotateMachineAPIKey(ctx, machine.ID, user.ID)
	keys, _ = svc.GetMachineAPIKeyInfo(ctx, machine.ID, user.ID)

	if len(keys) != 2 {
		t.Errorf("Expected 2 API keys after rotation, got %d", len(keys))
	}

	// Count active and revoked keys
	var activeCount, revokedCount int
	for _, key := range keys {
		if key.RevokedAt == nil {
			activeCount++
		} else {
			revokedCount++
		}
	}

	if activeCount != 1 {
		t.Errorf("Expected 1 active key, got %d", activeCount)
	}
	if revokedCount != 1 {
		t.Errorf("Expected 1 revoked key, got %d", revokedCount)
	}
}
