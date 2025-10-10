package router

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/machines"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// Helper: create test store
func createTestStoreForAgentTests(t *testing.T) storage.Store {
	t.Helper()

	// Use unique in-memory database for each test to avoid conflicts
	store, err := storage.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Failed to close test store: %v", err)
		}
	})

	return store
}

// Helper: create test auth service
func createTestAuthServiceForAgent(t *testing.T, store storage.Store) *auth.Service {
	t.Helper()

	authService, err := auth.NewService(store, "test-secret-key", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	return authService
}

// Helper: create test user and session token
func createTestUserWithSession(t *testing.T, authService *auth.Service) (int, string) {
	t.Helper()

	ctx := context.Background()

	// Create user
	user, _, err := authService.CreateUser(ctx, "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create session token
	token, err := authService.CreateSession(user.ID)
	if err != nil {
		t.Fatalf("Failed to create session token: %v", err)
	}

	return user.ID, token
}

func TestAgentRegister(t *testing.T) {
	store := createTestStoreForAgentTests(t)
	authService := createTestAuthServiceForAgent(t, store)
	machineService := machines.NewService(store)

	// Create test user with session
	userID, sessionToken := createTestUserWithSession(t, authService)

	t.Run("successful registration", func(t *testing.T) {
		req := RegisterMachineRequest{
			Name:     "test-machine",
			Hostname: "test.example.com",
		}
		body, _ := json.Marshal(req)

		httpReq := httptest.NewRequest(http.MethodPost, "/agent/register", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		// Set session cookie with correct name
		httpReq.AddCookie(&http.Cookie{
			Name:  "lunasentri_session",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()

		// Wrap with auth middleware
		handler := authService.RequireAuth(handleAgentRegister(machineService))
		handler.ServeHTTP(w, httpReq)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		var resp RegisterMachineResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if resp.Name != "test-machine" {
			t.Errorf("Expected name 'test-machine', got '%s'", resp.Name)
		}
		if resp.APIKey == "" {
			t.Error("Expected API key to be returned")
		}
		if resp.ID == 0 {
			t.Error("Expected machine ID to be set")
		}

		// Verify machine was created in database
		machine, err := store.GetMachineByID(context.Background(), resp.ID)
		if err != nil {
			t.Fatalf("Failed to get machine from database: %v", err)
		}
		if machine.UserID != userID {
			t.Errorf("Expected machine user_id %d, got %d", userID, machine.UserID)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		req := RegisterMachineRequest{
			Name: "",
		}
		body, _ := json.Marshal(req)

		httpReq := httptest.NewRequest(http.MethodPost, "/agent/register", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.AddCookie(&http.Cookie{
			Name:  "lunasentri_session",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		handler := authService.RequireAuth(handleAgentRegister(machineService))
		handler.ServeHTTP(w, httpReq)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		req := RegisterMachineRequest{
			Name: "test-machine",
		}
		body, _ := json.Marshal(req)

		httpReq := httptest.NewRequest(http.MethodPost, "/agent/register", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler := authService.RequireAuth(handleAgentRegister(machineService))
		handler.ServeHTTP(w, httpReq)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})
}

func TestAgentMetrics(t *testing.T) {
	store := createTestStoreForAgentTests(t)
	authService := createTestAuthServiceForAgent(t, store)
	machineService := machines.NewService(store)

	ctx := context.Background()

	// Create a test user and machine
	user, _, err := authService.CreateUser(ctx, "agent@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	machine, apiKey, err := machineService.RegisterMachine(ctx, user.ID, "test-machine", "test.local", "Agent test machine")
	if err != nil {
		t.Fatalf("Failed to register machine: %v", err)
	}

	t.Run("successful metrics ingestion", func(t *testing.T) {
		hostname := "agent-host"
		platform := "linux"
		platformVersion := "ubuntu 22.04"
		kernel := "6.2.0"
		cpuCores := 8
		memory := int64(32768)
		disk := int64(512)
		lastBoot := time.Now().Add(-2 * time.Hour).UTC()
		uptime := 7200.0

		req := AgentMetricsRequest{
			CPUPct:      45.5,
			MemUsedPct:  67.8,
			DiskUsedPct: 23.1,
			NetRxBytes:  1024,
			NetTxBytes:  2048,
			UptimeS:     &uptime,
			SystemInfo: &AgentSystemInfoPayload{
				Hostname:        &hostname,
				Platform:        &platform,
				PlatformVersion: &platformVersion,
				KernelVersion:   &kernel,
				CPUCores:        &cpuCores,
				MemoryTotalMB:   &memory,
				DiskTotalGB:     &disk,
				LastBootTime:    &lastBoot,
			},
		}
		body, _ := json.Marshal(req)

		httpReq := httptest.NewRequest(http.MethodPost, "/agent/metrics", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("X-API-Key", apiKey)

		w := httptest.NewRecorder()
		handler := RequireAPIKey(machineService)(http.HandlerFunc(handleAgentMetrics(machineService)))
		handler.ServeHTTP(w, httpReq)

		if w.Code != http.StatusAccepted {
			t.Errorf("Expected status 202, got %d: %s", w.Code, w.Body.String())
		}

		// Verify metrics were stored
		metrics, err := store.GetLatestMetrics(context.Background(), machine.ID)
		if err != nil {
			t.Fatalf("Failed to get metrics: %v", err)
		}

		if metrics.CPUPct != 45.5 {
			t.Errorf("Expected CPU 45.5, got %.1f", metrics.CPUPct)
		}
		if metrics.MemUsedPct != 67.8 {
			t.Errorf("Expected memory 67.8, got %.1f", metrics.MemUsedPct)
		}
		if metrics.UptimeSeconds != uptime {
			t.Errorf("Expected uptime %.0f, got %.0f", uptime, metrics.UptimeSeconds)
		}

		// Verify machine status was updated to online
		updatedMachine, err := store.GetMachineByID(context.Background(), machine.ID)
		if err != nil {
			t.Fatalf("Failed to get machine: %v", err)
		}
		if updatedMachine.Status != "online" {
			t.Errorf("Expected status 'online', got '%s'", updatedMachine.Status)
		}
		if updatedMachine.Hostname != hostname {
			t.Errorf("Expected hostname %s, got %s", hostname, updatedMachine.Hostname)
		}
		if updatedMachine.Platform != platform {
			t.Errorf("Expected platform %s, got %s", platform, updatedMachine.Platform)
		}
		if updatedMachine.PlatformVersion != platformVersion {
			t.Errorf("Expected platform version %s, got %s", platformVersion, updatedMachine.PlatformVersion)
		}
		if updatedMachine.KernelVersion != kernel {
			t.Errorf("Expected kernel %s, got %s", kernel, updatedMachine.KernelVersion)
		}
		if updatedMachine.CPUCores != cpuCores {
			t.Errorf("Expected cpu cores %d, got %d", cpuCores, updatedMachine.CPUCores)
		}
		if updatedMachine.MemoryTotalMB != memory {
			t.Errorf("Expected memory total %d, got %d", memory, updatedMachine.MemoryTotalMB)
		}
		if updatedMachine.DiskTotalGB != disk {
			t.Errorf("Expected disk total %d, got %d", disk, updatedMachine.DiskTotalGB)
		}
		if updatedMachine.LastBootTime.IsZero() {
			t.Errorf("Expected last boot time to be set")
		}
	})

	t.Run("invalid CPU percentage", func(t *testing.T) {
		req := AgentMetricsRequest{
			CPUPct:      150.0, // Invalid
			MemUsedPct:  50.0,
			DiskUsedPct: 30.0,
		}
		body, _ := json.Marshal(req)

		httpReq := httptest.NewRequest(http.MethodPost, "/agent/metrics", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("X-API-Key", apiKey)

		w := httptest.NewRecorder()
		handler := RequireAPIKey(machineService)(http.HandlerFunc(handleAgentMetrics(machineService)))
		handler.ServeHTTP(w, httpReq)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid API key", func(t *testing.T) {
		req := AgentMetricsRequest{
			CPUPct:      45.5,
			MemUsedPct:  67.8,
			DiskUsedPct: 23.1,
		}
		body, _ := json.Marshal(req)

		httpReq := httptest.NewRequest(http.MethodPost, "/agent/metrics", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("X-API-Key", "invalid-key")

		w := httptest.NewRecorder()
		handler := RequireAPIKey(machineService)(http.HandlerFunc(handleAgentMetrics(machineService)))
		handler.ServeHTTP(w, httpReq)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("missing API key", func(t *testing.T) {
		req := AgentMetricsRequest{
			CPUPct:      45.5,
			MemUsedPct:  67.8,
			DiskUsedPct: 23.1,
		}
		body, _ := json.Marshal(req)

		httpReq := httptest.NewRequest(http.MethodPost, "/agent/metrics", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler := RequireAPIKey(machineService)(http.HandlerFunc(handleAgentMetrics(machineService)))
		handler.ServeHTTP(w, httpReq)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})
}

func TestAPIKeyMiddleware(t *testing.T) {
	store := createTestStoreForAgentTests(t)
	authService := createTestAuthServiceForAgent(t, store)
	machineService := machines.NewService(store)

	ctx := context.Background()

	// Create test user and machine
	user, _, err := authService.CreateUser(ctx, "middleware@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	machine, apiKey, err := machineService.RegisterMachine(ctx, user.ID, "middleware-test", "test.local", "Middleware test machine")
	if err != nil {
		t.Fatalf("Failed to register machine: %v", err)
	}

	t.Run("valid API key in X-API-Key header", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/test", nil)
		httpReq.Header.Set("X-API-Key", apiKey)

		w := httptest.NewRecorder()

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			machineID, ok := GetMachineIDFromContext(r.Context())
			if !ok {
				t.Error("Expected machine ID in context")
			}
			if machineID != machine.ID {
				t.Errorf("Expected machine ID %d, got %d", machine.ID, machineID)
			}

			userID, ok := GetUserIDFromContext(r.Context())
			if !ok {
				t.Error("Expected user ID in context")
			}
			if userID != user.ID {
				t.Errorf("Expected user ID %d, got %d", user.ID, userID)
			}

			w.WriteHeader(http.StatusOK)
		})

		handler := RequireAPIKey(machineService)(testHandler)
		handler.ServeHTTP(w, httpReq)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("valid API key in Authorization header", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/test", nil)
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)

		w := httptest.NewRecorder()

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := RequireAPIKey(machineService)(testHandler)
		handler.ServeHTTP(w, httpReq)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestListMachines(t *testing.T) {
	store := createTestStoreForAgentTests(t)
	authService := createTestAuthServiceForAgent(t, store)
	machineService := machines.NewService(store)

	// Create test user with session
	userID, sessionToken := createTestUserWithSession(t, authService)

	// Create another user to test isolation
	ctx := context.Background()
	otherUser, _, err := authService.CreateUser(ctx, "other@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create other user: %v", err)
	}

	// Register machines for first user
	machine1, _, err := machineService.RegisterMachine(ctx, userID, "machine-1", "host-1.example.com", "First user machine")
	if err != nil {
		t.Fatalf("Failed to register machine 1: %v", err)
	}

	machine2, _, err := machineService.RegisterMachine(ctx, userID, "machine-2", "host-2.example.com", "Second user machine")
	if err != nil {
		t.Fatalf("Failed to register machine 2: %v", err)
	}

	// Register machine for other user (should not appear in response)
	_, _, err = machineService.RegisterMachine(ctx, otherUser.ID, "other-machine", "other.example.com", "Other user machine")
	if err != nil {
		t.Fatalf("Failed to register other machine: %v", err)
	}

	t.Run("successful list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/machines", nil)
		req.AddCookie(&http.Cookie{
			Name:  auth.CookieName,
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		handler := authService.RequireAuth(handleListMachines(machineService))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var machinesList []storage.Machine
		if err := json.Unmarshal(w.Body.Bytes(), &machinesList); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Should return 2 machines for the authenticated user
		if len(machinesList) != 2 {
			t.Errorf("Expected 2 machines, got %d", len(machinesList))
		}

		// Verify machines belong to the user
		foundMachine1 := false
		foundMachine2 := false
		for _, m := range machinesList {
			if m.UserID != userID {
				t.Errorf("Machine %d belongs to user %d, expected %d", m.ID, m.UserID, userID)
			}
			if m.ID == machine1.ID {
				foundMachine1 = true
			}
			if m.ID == machine2.ID {
				foundMachine2 = true
			}
		}

		if !foundMachine1 || !foundMachine2 {
			t.Error("Expected to find both registered machines")
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/machines", nil)
		// No session cookie

		w := httptest.NewRecorder()
		handler := authService.RequireAuth(handleListMachines(machineService))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/machines", nil)
		req.AddCookie(&http.Cookie{
			Name:  auth.CookieName,
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		handler := authService.RequireAuth(handleListMachines(machineService))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}
