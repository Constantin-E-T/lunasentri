package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/alerts"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	router "github.com/Constantin-E-T/lunasentri/apps/api-go/internal/http"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/metrics"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/notifications"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/system"
)

// fakeCollector implements metrics.Collector for testing.
type fakeCollector struct {
	metricsToReturn metrics.Metrics
	errToReturn     error
}

func (f *fakeCollector) Snapshot(ctx context.Context) (metrics.Metrics, error) {
	return f.metricsToReturn, f.errToReturn
}

// fakeSystemService implements system.Service for testing.
type fakeSystemService struct {
	systemInfoToReturn system.SystemInfo
	errToReturn        error
}

func (f *fakeSystemService) GetSystemInfo(ctx context.Context) (system.SystemInfo, error) {
	return f.systemInfoToReturn, f.errToReturn
}

// createTestStore creates a store for testing.
func createTestStore(t *testing.T) storage.Store {
	t.Helper()

	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
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

// createTestAuthService creates an auth service for testing.
func createTestAuthService(t *testing.T) *auth.Service {
	t.Helper()

	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Failed to close auth test store: %v", err)
		}
	})

	authService, err := auth.NewService(store, "test-secret-key-for-testing", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create test auth service: %v", err)
	}

	return authService
}

// createTestAlertService creates an alert service for testing.
func createTestAlertService(t *testing.T) *alerts.Service {
	t.Helper()

	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Failed to close alert test store: %v", err)
		}
	})

	notifier := notifications.NewNotifier(store, log.Default())
	return alerts.NewService(store, notifier)
}

// newTestServer constructs an httptest.Server backed by the router.
func newTestServer(t *testing.T, systemService system.Service, collector metrics.Collector) *httptest.Server {
	t.Helper()

	store := createTestStore(t)
	authService := createTestAuthService(t)
	alertService := createTestAlertService(t)
	notifier := notifications.NewNotifier(store, log.Default())

	routerCfg := &router.RouterConfig{
		Collector:        collector,
		ServerStartTime:  time.Now(),
		AuthService:      authService,
		AlertService:     alertService,
		SystemService:    systemService,
		Store:            store,
		WebhookNotifier:  notifier,
		TelegramNotifier: nil,
		AccessTTL:        15 * time.Minute,
		PasswordResetTTL: 15 * time.Minute,
		SecureCookie:     false,
		LocalHostMetrics: true,
	}

	handler := router.CORSMiddleware(router.NewRouter(routerCfg))
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func TestSystemInfoHandler(t *testing.T) {
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
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{},
		errToReturn:     nil,
	}

	server := newTestServer(t, systemService, collector)

	resp, err := http.Get(server.URL + "/system/info")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var received system.SystemInfo
	if err := json.NewDecoder(resp.Body).Decode(&received); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if received.Hostname != fakeSystemInfo.Hostname {
		t.Errorf("Expected hostname %s, got %s", fakeSystemInfo.Hostname, received.Hostname)
	}
	if received.Platform != fakeSystemInfo.Platform {
		t.Errorf("Expected platform %s, got %s", fakeSystemInfo.Platform, received.Platform)
	}
	if received.CPUCores != fakeSystemInfo.CPUCores {
		t.Errorf("Expected CPU cores %d, got %d", fakeSystemInfo.CPUCores, received.CPUCores)
	}
	if received.MemoryTotalMB != fakeSystemInfo.MemoryTotalMB {
		t.Errorf("Expected memory %d MB, got %d MB", fakeSystemInfo.MemoryTotalMB, received.MemoryTotalMB)
	}
}

func TestSystemInfoHandlerError(t *testing.T) {
	systemService := &fakeSystemService{
		systemInfoToReturn: system.SystemInfo{},
		errToReturn:        errors.New("failed to collect system info"),
	}
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{},
		errToReturn:     nil,
	}

	server := newTestServer(t, systemService, collector)

	resp, err := http.Get(server.URL + "/system/info")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}

func TestSystemInfoHandlerMethodNotAllowed(t *testing.T) {
	systemService := &fakeSystemService{
		systemInfoToReturn: system.SystemInfo{},
		errToReturn:        nil,
	}
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{},
		errToReturn:     nil,
	}

	server := newTestServer(t, systemService, collector)

	resp, err := http.Post(server.URL+"/system/info", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}
