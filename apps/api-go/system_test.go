package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/alerts"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/metrics"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/system"
)

// fakeCollector implements metrics.Collector for testing
type fakeCollector struct {
	metricsToReturn metrics.Metrics
	errToReturn     error
}

func (f *fakeCollector) Snapshot(ctx context.Context) (metrics.Metrics, error) {
	return f.metricsToReturn, f.errToReturn
}

// fakeSystemService implements system.Service for testing
type fakeSystemService struct {
	systemInfoToReturn system.SystemInfo
	errToReturn        error
}

func (f *fakeSystemService) GetSystemInfo(ctx context.Context) (system.SystemInfo, error) {
	return f.systemInfoToReturn, f.errToReturn
}

// createTestStore creates a store for testing
func createTestStore(t *testing.T) storage.Store {
	// Create in-memory SQLite store
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	return store
}

// createTestAuthService creates an auth service for testing
func createTestAuthService(t *testing.T) *auth.Service {
	// Create in-memory SQLite store
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	// Create auth service with test secret
	authService, err := auth.NewService(store, "test-secret-key-for-testing", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create test auth service: %v", err)
	}

	return authService
}

// createTestAlertService creates an alert service for testing
func createTestAlertService(t *testing.T) *alerts.Service {
	// Create in-memory SQLite store
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	// Create a no-op notifier for testing
	notifier := &noOpNotifier{}
	return alerts.NewService(store, notifier)
}

// noOpNotifier is a test implementation that does nothing
type noOpNotifier struct{}

func (n *noOpNotifier) Send(ctx context.Context, rule storage.AlertRule, event storage.AlertEvent) error {
	return nil
}

func (n *noOpNotifier) Notify(ctx context.Context, rule storage.AlertRule, event *storage.AlertEvent) error {
	return nil
}

func TestSystemInfoHandler(t *testing.T) {
	// Setup fake system service with known values
	fakeSystemInfo := system.SystemInfo{
		Hostname:        "test-host",
		Platform:        "linux",
		PlatformVersion: "ubuntu 22.04",
		KernelVersion:   "5.15.0-72-generic",
		UptimeS:         3600,
		CPUCores:        4,
		MemoryTotalMB:   8192,
		DiskTotalGB:     256,
		LastBootTime:    1640995200,
	}

	systemService := &fakeSystemService{
		systemInfoToReturn: fakeSystemInfo,
		errToReturn:        nil,
	}

	// Create test services
	store := createTestStore(t)
	authService := createTestAuthService(t)
	alertService := createTestAlertService(t)
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{},
		errToReturn:     nil,
	}

	// Create server
	mux := newServer(collector, time.Now(), authService, alertService, systemService, store, 15*time.Minute, 15*time.Minute, false)
	server := httptest.NewServer(corsMiddleware(mux))
	defer server.Close()

	// Make request to /system/info endpoint
	resp, err := http.Get(server.URL + "/system/info")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Assert Content-Type header
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Decode response
	var receivedSystemInfo system.SystemInfo
	if err := json.NewDecoder(resp.Body).Decode(&receivedSystemInfo); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Assert system info data
	if receivedSystemInfo.Hostname != fakeSystemInfo.Hostname {
		t.Errorf("Expected hostname %s, got %s", fakeSystemInfo.Hostname, receivedSystemInfo.Hostname)
	}
	if receivedSystemInfo.Platform != fakeSystemInfo.Platform {
		t.Errorf("Expected platform %s, got %s", fakeSystemInfo.Platform, receivedSystemInfo.Platform)
	}
	if receivedSystemInfo.CPUCores != fakeSystemInfo.CPUCores {
		t.Errorf("Expected CPU cores %d, got %d", fakeSystemInfo.CPUCores, receivedSystemInfo.CPUCores)
	}
	if receivedSystemInfo.MemoryTotalMB != fakeSystemInfo.MemoryTotalMB {
		t.Errorf("Expected memory %d MB, got %d MB", fakeSystemInfo.MemoryTotalMB, receivedSystemInfo.MemoryTotalMB)
	}

	t.Logf("System info test successful - received: %+v", receivedSystemInfo)
}

func TestSystemInfoHandlerError(t *testing.T) {
	// Setup fake system service that returns an error
	systemService := &fakeSystemService{
		systemInfoToReturn: system.SystemInfo{},
		errToReturn:        errors.New("failed to collect system info"),
	}

	// Create test services
	store := createTestStore(t)
	authService := createTestAuthService(t)
	alertService := createTestAlertService(t)
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{},
		errToReturn:     nil,
	}

	// Create server
	mux := newServer(collector, time.Now(), authService, alertService, systemService, store, 15*time.Minute, 15*time.Minute, false)
	server := httptest.NewServer(corsMiddleware(mux))
	defer server.Close()

	// Make request to /system/info endpoint
	resp, err := http.Get(server.URL + "/system/info")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert HTTP status should be 500
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	t.Logf("System info error test successful - got expected 500 status")
}

func TestSystemInfoHandlerMethodNotAllowed(t *testing.T) {
	// Setup fake system service
	systemService := &fakeSystemService{
		systemInfoToReturn: system.SystemInfo{},
		errToReturn:        nil,
	}

	// Create test services
	store := createTestStore(t)
	authService := createTestAuthService(t)
	alertService := createTestAlertService(t)
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{},
		errToReturn:     nil,
	}

	// Create server
	mux := newServer(collector, time.Now(), authService, alertService, systemService, store, 15*time.Minute, 15*time.Minute, false)
	server := httptest.NewServer(corsMiddleware(mux))
	defer server.Close()

	// Make POST request to /system/info endpoint (should fail)
	resp, err := http.Post(server.URL+"/system/info", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert HTTP status should be 405
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}

	t.Logf("System info method not allowed test successful - got expected 405 status")
}
