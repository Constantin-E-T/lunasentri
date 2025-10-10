package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/machines"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

func TestHandleDisableMachine(t *testing.T) {
	store, _ := storage.NewSQLiteStore(":memory:")
	defer store.Close()

	svc := machines.NewService(store)
	ctx := context.Background()

	// Create test user and machine
	user, _ := store.CreateUser(ctx, "test@example.com", "hashedpassword")
	machine, _, _ := svc.RegisterMachine(ctx, user.ID, "test-machine", "test-host", "test description")

	// Create request
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/machines/%d/disable", machine.ID), nil)
	
	// Add user to context
	reqCtx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(reqCtx)

	w := httptest.NewRecorder()

	// Call handler
	handler := handleDisableMachine(svc)
	handler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)

	if response["message"] != "Machine disabled successfully" {
		t.Errorf("Unexpected response message: %s", response["message"])
	}

	// Verify machine is disabled in database
	updatedMachine, _ := store.GetMachineByID(ctx, machine.ID)
	if updatedMachine.IsEnabled {
		t.Error("Machine should be disabled")
	}
}

func TestHandleEnableMachine(t *testing.T) {
	store, _ := storage.NewSQLiteStore(":memory:")
	defer store.Close()

	svc := machines.NewService(store)
	ctx := context.Background()

	user, _ := store.CreateUser(ctx, "test@example.com", "hashedpassword")
	machine, _, _ := svc.RegisterMachine(ctx, user.ID, "test-machine", "test-host", "test description")

	// Disable first
	svc.DisableMachine(ctx, machine.ID, user.ID)

	// Create request to enable
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/machines/%d/enable", machine.ID), nil)
	reqCtx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(reqCtx)

	w := httptest.NewRecorder()

	handler := handleEnableMachine(svc)
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify machine is enabled
	updatedMachine, _ := store.GetMachineByID(ctx, machine.ID)
	if !updatedMachine.IsEnabled {
		t.Error("Machine should be enabled")
	}
}

func TestHandleRotateMachineAPIKey(t *testing.T) {
	store, _ := storage.NewSQLiteStore(":memory:")
	defer store.Close()

	svc := machines.NewService(store)
	ctx := context.Background()

	user, _ := store.CreateUser(ctx, "test@example.com", "hashedpassword")
	machine, oldAPIKey, _ := svc.RegisterMachine(ctx, user.ID, "test-machine", "test-host", "test description")

	// Create request
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/machines/%d/rotate-key", machine.ID), nil)
	reqCtx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(reqCtx)

	w := httptest.NewRecorder()

	handler := handleRotateMachineAPIKey(svc)
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	// Should return new API key
	newAPIKey, ok := response["api_key"].(string)
	if !ok || newAPIKey == "" {
		t.Error("Response should contain new API key")
	}

	if newAPIKey == oldAPIKey {
		t.Error("New API key should be different from old key")
	}

	// Old key should not work
	_, err := svc.AuthenticateMachine(ctx, oldAPIKey)
	if err == nil {
		t.Error("Old API key should not work after rotation")
	}

	// New key should work
	_, err = svc.AuthenticateMachine(ctx, newAPIKey)
	if err != nil {
		t.Errorf("New API key should work: %v", err)
	}
}

func TestHandleDisableMachineUnauthorized(t *testing.T) {
	store, _ := storage.NewSQLiteStore(":memory:")
	defer store.Close()

	svc := machines.NewService(store)
	ctx := context.Background()

	user1, _ := store.CreateUser(ctx, "user1@example.com", "hashedpassword")
	user2, _ := store.CreateUser(ctx, "user2@example.com", "hashedpassword")

	machine, _, _ := svc.RegisterMachine(ctx, user1.ID, "test-machine", "test-host", "test description")

	// User2 tries to disable user1's machine
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/machines/%d/disable", machine.ID), nil)
	reqCtx := context.WithValue(req.Context(), auth.UserContextKey, user2)
	req = req.WithContext(reqCtx)

	w := httptest.NewRecorder()

	handler := handleDisableMachine(svc)
	handler(w, req)

	// Should fail
	if w.Code == http.StatusOK {
		t.Error("Should not allow user to disable another user's machine")
	}
}

func TestAgentAuthenticationWithDisabledMachine(t *testing.T) {
	store, _ := storage.NewSQLiteStore(":memory:")
	defer store.Close()

	svc := machines.NewService(store)
	ctx := context.Background()

	user, _ := store.CreateUser(ctx, "test@example.com", "hashedpassword")
	machine, apiKey, _ := svc.RegisterMachine(ctx, user.ID, "test-machine", "test-host", "test description")

	// Should work initially
	_, err := svc.AuthenticateMachine(ctx, apiKey)
	if err != nil {
		t.Fatalf("Initial authentication should work: %v", err)
	}

	// Disable machine
	svc.DisableMachine(ctx, machine.ID, user.ID)

	// Should fail after disabling
	_, err = svc.AuthenticateMachine(ctx, apiKey)
	if err == nil {
		t.Error("Authentication should fail for disabled machine")
	}

	// Re-enable machine
	svc.EnableMachine(ctx, machine.ID, user.ID)

	// Should work again
	_, err = svc.AuthenticateMachine(ctx, apiKey)
	if err != nil {
		t.Errorf("Authentication should work after re-enabling: %v", err)
	}
}
