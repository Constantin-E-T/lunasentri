package mainpackage main



import (import (

	"context"	"context"

	"encoding/json"	"encoding/json"

	"errors"	"errors"

	"net/http"	"net/http"

	"net/http/httptest"	"net/http/httptest"

	"strings"	"net/url"

	"testing"	"strings"

	"time"	"testing"

	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/alerts"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/metrics"	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/alerts"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/metrics"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/system"	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"

)	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/system"

	"github.com/gorilla/websocket"

// fakeCollector implements metrics.Collector for testing)

type fakeCollector struct {

	metricsToReturn metrics.Metrics// fakeCollector implements metrics.Collector for testing

	errToReturn     errortype fakeCollector struct {

}	metricsToReturn metrics.Metrics

	errToReturn     error

func (f *fakeCollector) Snapshot(ctx context.Context) (metrics.Metrics, error) {}

	return f.metricsToReturn, f.errToReturn

}func (f *fakeCollector) Snapshot(ctx context.Context) (metrics.Metrics, error) {

	return f.metricsToReturn, f.errToReturn

// fakeSystemService implements system.Service for testing}

type fakeSystemService struct {

	systemInfoToReturn system.SystemInfo// fakeSystemService implements system.Service for testing

	errToReturn        errortype fakeSystemService struct {

}	systemInfoToReturn system.SystemInfo

	errToReturn        error

func (f *fakeSystemService) GetSystemInfo(ctx context.Context) (system.SystemInfo, error) {}

	return f.systemInfoToReturn, f.errToReturn

}func (f *fakeSystemService) GetSystemInfo(ctx context.Context) (system.SystemInfo, error) {

	return f.systemInfoToReturn, f.errToReturn

// createTestAuthService creates an auth service for testing}

func createTestAuthService(t *testing.T) *auth.Service {

	// Create in-memory SQLite store// createTestAuthService creates an auth service for testing (without actual auth checks for these tests)

	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")func createTestAuthService(t *testing.T) *auth.Service {

	if err != nil {	// Create in-memory SQLite store

		t.Fatalf("Failed to create test store: %v", err)	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")

	}	if err != nil {

		t.Fatalf("Failed to create test store: %v", err)

	// Create auth service with test secret	}

	authService, err := auth.NewService(store, "test-secret-key-for-testing", 15*time.Minute)

	if err != nil {	// Create auth service with test secret

		t.Fatalf("Failed to create test auth service: %v", err)	authService, err := auth.NewService(store, "test-secret-key-for-testing", 15*time.Minute)

	}	if err != nil {

		t.Fatalf("Failed to create test auth service: %v", err)

	return authService	}

}

	return authService

// createTestAlertService creates an alert service for testing}

func createTestAlertService(t *testing.T) *alerts.Service {

	// Create in-memory SQLite store// createTestAlertService creates an alert service for testing

	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")func createTestAlertService(t *testing.T) *alerts.Service {

	if err != nil {	// Create in-memory SQLite store

		t.Fatalf("Failed to create test store: %v", err)	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")

	}	if err != nil {

		t.Fatalf("Failed to create test store: %v", err)

	return alerts.NewService(store)	}

}

	return alerts.NewService(store)

func TestSystemInfoHandler(t *testing.T) {}

	// Setup fake system service with known values

	fakeSystemInfo := system.SystemInfo{func TestMetricsHandlerIntegration(t *testing.T) {

		Hostname:        "test-host",	// Setup fake collector with known values

		Platform:        "linux",	startTime := time.Now().Add(-10 * time.Second) // 10 seconds ago

		PlatformVersion: "ubuntu 22.04",	fakeMetrics := metrics.Metrics{

		KernelVersion:   "5.15.0-72-generic",		CPUPct:      45.5,

		UptimeS:         3600,		MemUsedPct:  67.2,

		CPUCores:        4,		DiskUsedPct: 23.8,

		MemoryTotalMB:   8192,		UptimeS:     0, // Will be overwritten by handler

		DiskTotalGB:     256,	}

		LastBootTime:    1640995200,

	}	collector := &fakeCollector{

		metricsToReturn: fakeMetrics,

	systemService := &fakeSystemService{		errToReturn:     nil,

		systemInfoToReturn: fakeSystemInfo,	}

		errToReturn:        nil,

	}	// Create auth service and server

	authService := createTestAuthService(t)

	// Create test services	alertService := createTestAlertService(t)

	authService := createTestAuthService(t)	systemService := &fakeSystemService{

	alertService := createTestAlertService(t)		systemInfoToReturn: system.SystemInfo{

	collector := &fakeCollector{			Hostname:        "test-host",

		metricsToReturn: metrics.Metrics{},			Platform:        "linux",

		errToReturn:     nil,			PlatformVersion: "ubuntu 22.04",

	}			KernelVersion:   "5.15.0",

			UptimeS:         3600,

	// Create server			CPUCores:        4,

	mux := newServer(collector, time.Now(), authService, alertService, systemService, 15*time.Minute, 15*time.Minute, false)			MealertService := createTestAlertService(t)

	server := httptest.NewServer(corsMiddleware(mux))	systemService := &fakeSystealertService := createTestAlertService(t)

	defer server.Close()	systemService := &fakeSystemService{

		systemInfoToReturn: system.SystemInfo{},

	// Make request to /system/info endpoint		errToReturn:        nil,

	resp, err := http.Get(server.URL + "/system/info")	}

	if err != nil {	alertService := createTestAlertService(t)

		t.Fatalf("Failed to make request: %v", err)	systemService := &fakeSystemService{

	}		systemInfoToReturn: system.SystemInfo{},

	defer resp.Body.Close()		errToReturn:        nil,

	}

	// Assert HTTP status	alertService := createTestAlertService(t)

	if resp.StatusCode != http.StatusOK {	systemService := &fakeSystemService{

		t.Errorf("Expected status 200, got %d", resp.StatusCode)		systemInfoToReturn: system.SystemInfo{},

	}		errToReturn:        nil,

	}

	// Assert Content-Type header	alertService := createTestAlertService(t)

	contentType := resp.Header.Get("Content-Type")	systemService := &fakeSystemService{

	if contentType != "application/json" {		systemInfoToReturn: system.SystemInfo{},

		t.Errorf("Expected Content-Type application/json, got %s", contentType)		errToReturn:        nil,

	}	}

	moryTotalMB:   8192,

	// Decode response			DiskTotalGB:     256,

	var receivedSystemInfo system.SystemInfo			LastBootTime:    1640995200,

	if err := json.NewDecoder(resp.Body).Decode(&receivedSystemInfo); err != nil {		},

		t.Fatalf("Failed to decode response: %v", err)		errToRetService{errToReturn: nil,

	}	}

	murn: errors.New("system error")}

	// Assert system info data	mux := newServer(collector, startTime, authService, alertService, systemService, alertService, systemService, 15*time.Minute, alertService, systemService, 15*time.Minute, alertService, systemService, 15*time.Minute, alertService, systemService, 15*time.Minute, alertService, systemService, 15*time.Minute, 15*time.Minute, false, 15*time.Minute, false, false, false, false, false)

	if receivedSystemInfo.Hostname != fakeSystemInfo.Hostname {	server := httptest.NewServer(corsMiddleware(mux))

		t.Errorf("Expected hostname %s, got %s", fakeSystemInfo.Hostname, receivedSystemInfo.Hostname)	defer server.Close()

	}

	if receivedSystemInfo.Platform != fakeSystemInfo.Platform {	// Make request to /metrics endpoint

		t.Errorf("Expected platform %s, got %s", fakeSystemInfo.Platform, receivedSystemInfo.Platform)	resp, err := http.Get(server.URL + "/metrics")

	}	if err != nil {

	if receivedSystemInfo.CPUCores != fakeSystemInfo.CPUCores {		t.Fatalf("Failed to make request: %v", err)

		t.Errorf("Expected CPU cores %d, got %d", fakeSystemInfo.CPUCores, receivedSystemInfo.CPUCores)	}

	}	defer resp.Body.Close()

	if receivedSystemInfo.MemoryTotalMB != fakeSystemInfo.MemoryTotalMB {

		t.Errorf("Expected memory %d MB, got %d MB", fakeSystemInfo.MemoryTotalMB, receivedSystemInfo.MemoryTotalMB)	// Assert HTTP status

	}	if resp.StatusCode != http.StatusOK {

		t.Errorf("Expected status 200, got %d", resp.StatusCode)

	t.Logf("System info test successful - received: %+v", receivedSystemInfo)	}

}

	// Assert Content-Type header

func TestSystemInfoHandlerError(t *testing.T) {	contentType := resp.Header.Get("Content-Type")

	// Setup fake system service that returns an error	if contentType != "application/json" {

	systemService := &fakeSystemService{		t.Errorf("Expected Content-Type application/json, got %s", contentType)

		systemInfoToReturn: system.SystemInfo{},	}

		errToReturn:        errors.New("failed to collect system info"),

	}	// Assert CORS headers are present

	corsOrigin := resp.Header.Get("Access-Control-Allow-Origin")

	// Create test services	if corsOrigin == "" {

	authService := createTestAuthService(t)		t.Error("Expected CORS Access-Control-Allow-Origin header to be present")

	alertService := createTestAlertService(t)	}

	collector := &fakeCollector{

		metricsToReturn: metrics.Metrics{},	corsMethods := resp.Header.Get("Access-Control-Allow-Methods")

		errToReturn:     nil,	if corsMethods == "" {

	}		t.Error("Expected CORS Access-Control-Allow-Methods header to be present")

	}

	// Create server

	mux := newServer(collector, time.Now(), authService, alertService, systemService, 15*time.Minute, 15*time.Minute, false)	corsHeaders := resp.Header.Get("Access-Control-Allow-Headers")

	server := httptest.NewServer(corsMiddleware(mux))	if corsHeaders == "" {

	defer server.Close()		t.Error("Expected CORS Access-Control-Allow-Headers header to be present")

	}

	// Make request to /system/info endpoint

	resp, err := http.Get(server.URL + "/system/info")	// Parse JSON response

	if err != nil {	var result metrics.Metrics

		t.Fatalf("Failed to make request: %v", err)	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {

	}		t.Fatalf("Failed to decode JSON response: %v", err)

	defer resp.Body.Close()	}



	// Assert HTTP status should be 500	// Assert collector values are preserved

	if resp.StatusCode != http.StatusInternalServerError {	if result.CPUPct != fakeMetrics.CPUPct {

		t.Errorf("Expected status 500, got %d", resp.StatusCode)		t.Errorf("Expected CPU %.2f, got %.2f", fakeMetrics.CPUPct, result.CPUPct)

	}	}

	if result.MemUsedPct != fakeMetrics.MemUsedPct {

	t.Logf("System info error test successful - got expected 500 status")		t.Errorf("Expected Memory %.2f, got %.2f", fakeMetrics.MemUsedPct, result.MemUsedPct)

}	}

	if result.DiskUsedPct != fakeMetrics.DiskUsedPct {

func TestSystemInfoHandlerMethodNotAllowed(t *testing.T) {		t.Errorf("Expected Disk %.2f, got %.2f", fakeMetrics.DiskUsedPct, result.DiskUsedPct)

	// Setup fake system service	}

	systemService := &fakeSystemService{

		systemInfoToReturn: system.SystemInfo{},	// Assert uptime is calculated correctly (should be ~10 seconds)

		errToReturn:        nil,	if result.UptimeS < 9 || result.UptimeS > 11 {

	}		t.Errorf("Expected uptime around 10 seconds, got %.2f", result.UptimeS)

	}

	// Create test services

	authService := createTestAuthService(t)	t.Logf("Integration test successful - CPU: %.2f%%, Memory: %.2f%%, Disk: %.2f%%, Uptime: %.2fs", 

	alertService := createTestAlertService(t)		result.CPUPct, result.MemUsedPct, result.DiskUsedPct, result.UptimeS)

	collector := &fakeCollector{}

		metricsToReturn: metrics.Metrics{},

		errToReturn:     nil,func TestMetricsHandlerWithCollectorError(t *testing.T) {

	}	// Setup fake collector that returns an error

	startTime := time.Now()

	// Create server	collector := &fakeCollector{

	mux := newServer(collector, time.Now(), authService, alertService, systemService, 15*time.Minute, 15*time.Minute, false)		metricsToReturn: metrics.Metrics{}, // Will be ignored due to error

	server := httptest.NewServer(corsMiddleware(mux))		errToReturn:     errors.New("simulated collector error"),

	defer server.Close()	}



	// Make POST request to /system/info endpoint (should fail)	// Create server with failing collector

	resp, err := http.Post(server.URL+"/system/info", "application/json", strings.NewReader("{}"))	authService := createTestAuthService(t)

	if err != nil {	mux := newServer(collector, startTime, authService, 15*time.Minute)

		t.Fatalf("Failed to make request: %v", err)	server := httptest.NewServer(corsMiddleware(mux))

	}	defer server.Close()

	defer resp.Body.Close()

	// Make request to /metrics endpoint

	// Assert HTTP status should be 405	resp, err := http.Get(server.URL + "/metrics")

	if resp.StatusCode != http.StatusMethodNotAllowed {	if err != nil {

		t.Errorf("Expected status 405, got %d", resp.StatusCode)		t.Fatalf("Failed to make request: %v", err)

	}	}

	defer resp.Body.Close()

	t.Logf("System info method not allowed test successful - got expected 405 status")

}	// Assert HTTP status is still 200 (we don't return 500 for collector errors)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 even with collector error, got %d", resp.StatusCode)
	}

	// Parse JSON response
	var result metrics.Metrics
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	// Assert collector error results in zeroed values
	if result.CPUPct != 0 {
		t.Errorf("Expected CPU 0 on collector error, got %.2f", result.CPUPct)
	}
	if result.MemUsedPct != 0 {
		t.Errorf("Expected Memory 0 on collector error, got %.2f", result.MemUsedPct)
	}
	if result.DiskUsedPct != 0 {
		t.Errorf("Expected Disk 0 on collector error, got %.2f", result.DiskUsedPct)
	}

	// Assert uptime is still calculated (should be ≥ 0)
	if result.UptimeS < 0 {
		t.Errorf("Expected uptime ≥ 0 even with collector error, got %.2f", result.UptimeS)
	}

	t.Logf("Error handling test successful - Zeroed metrics with uptime: %.2fs", result.UptimeS)
}

func TestMetricsHandlerConcurrentRequests(t *testing.T) {
	// Setup fake collector
	startTime := time.Now()
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{
			CPUPct:      50.0,
			MemUsedPct:  75.0,
			DiskUsedPct: 30.0,
		},
		errToReturn: nil,
	}

	// Create server
	authService := createTestAuthService(t)
	mux := newServer(collector, startTime, authService, 15*time.Minute)
	server := httptest.NewServer(corsMiddleware(mux))
	defer server.Close()

	// Make 10 concurrent requests
	numRequests := 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(server.URL + "/metrics")
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- errors.New("non-200 status code")
				return
			}

			var metrics metrics.Metrics
			if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
				results <- err
				return
			}

			results <- nil
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent request %d failed: %v", i, err)
		}
	}

	t.Logf("Concurrent requests test successful - %d requests completed", numRequests)
}

func TestMetricsHandlerJSONStructure(t *testing.T) {
	// Test that the JSON structure is exactly what we expect
	startTime := time.Now()
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{
			CPUPct:      12.34,
			MemUsedPct:  56.78,
			DiskUsedPct: 90.12,
		},
		errToReturn: nil,
	}

	authService := createTestAuthService(t)
	mux := newServer(collector, startTime, authService, 15*time.Minute)
	server := httptest.NewServer(corsMiddleware(mux))
	defer server.Close()

	resp, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Parse as generic map to check exact JSON structure
	var jsonMap map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonMap); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	// Check all required fields are present
	requiredFields := []string{"cpu_pct", "mem_used_pct", "disk_used_pct", "uptime_s"}
	for _, field := range requiredFields {
		if _, exists := jsonMap[field]; !exists {
			t.Errorf("Required field %s missing from JSON response", field)
		}
	}

	// Check no extra fields
	if len(jsonMap) != len(requiredFields) {
		t.Errorf("Expected exactly %d fields, got %d", len(requiredFields), len(jsonMap))
	}

	t.Logf("JSON structure test successful - all fields present: %v", requiredFields)
}

func TestWebSocketHandler(t *testing.T) {
	// Setup fake collector with known values
	startTime := time.Now().Add(-5 * time.Second) // 5 seconds ago
	fakeMetrics := metrics.Metrics{
		CPUPct:      25.5,
		MemUsedPct:  55.0,
		DiskUsedPct: 33.3,
		UptimeS:     0, // Will be overwritten by handler
	}
	
	collector := &fakeCollector{
		metricsToReturn: fakeMetrics,
		errToReturn:     nil,
	}

	// Create server with fake collector
	authService := createTestAuthService(t)
	alertService := createTestAlertService(t)
	systemService := &fakeSystemService{
		systemInfoToReturn: system.SystemInfo{},
		errToReturn:        nil,
	}
	mux := newServer(collector, startTime, authService, alertService, systemService, 15*time.Minute, 15*time.Minute, false)
	server := httptest.NewServer(corsMiddleware(mux))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	u, err := url.Parse(wsURL)
	if err != nil {
		t.Fatalf("Failed to parse WebSocket URL: %v", err)
	}

	// Set proper Origin header for CORS
	headers := http.Header{}
	headers.Set("Origin", "http://localhost:3000")

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Read first message
	var receivedMetrics metrics.Metrics
	err = conn.ReadJSON(&receivedMetrics)
	if err != nil {
		t.Fatalf("Failed to read WebSocket message: %v", err)
	}

	// Verify metrics data
	if receivedMetrics.CPUPct != fakeMetrics.CPUPct {
		t.Errorf("Expected CPU %f, got %f", fakeMetrics.CPUPct, receivedMetrics.CPUPct)
	}
	if receivedMetrics.MemUsedPct != fakeMetrics.MemUsedPct {
		t.Errorf("Expected Memory %f, got %f", fakeMetrics.MemUsedPct, receivedMetrics.MemUsedPct)
	}
	if receivedMetrics.DiskUsedPct != fakeMetrics.DiskUsedPct {
		t.Errorf("Expected Disk %f, got %f", fakeMetrics.DiskUsedPct, receivedMetrics.DiskUsedPct)
	}
	if receivedMetrics.UptimeS <= 0 {
		t.Errorf("Expected positive uptime, got %f", receivedMetrics.UptimeS)
	}

	t.Logf("WebSocket test successful - received metrics: %+v", receivedMetrics)
}

func TestWebSocketCORSValidation(t *testing.T) {
	// Setup fake collector
	startTime := time.Now()
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{},
		errToReturn:     nil,
	}

	// Create server
	authService := createTestAuthService(t)
	alertService := createTestAlertService(t)
	systemService := &fakeSystemService{
		systemInfoToReturn: system.SystemInfo{},
		errToReturn:        nil,
	}
	mux := newServer(collector, startTime, authService, alertService, systemService, 15*time.Minute, 15*time.Minute, false)
	server := httptest.NewServer(corsMiddleware(mux))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	u, err := url.Parse(wsURL)
	if err != nil {
		t.Fatalf("Failed to parse WebSocket URL: %v", err)
	}

	// Test with wrong Origin header (should fail)
	headers := http.Header{}
	headers.Set("Origin", "http://malicious-site.com")

	_, _, err = websocket.DefaultDialer.Dial(u.String(), headers)
	if err == nil {
		t.Fatal("Expected WebSocket connection to fail with wrong Origin, but it succeeded")
	}

	t.Logf("CORS validation test successful - connection rejected for wrong origin")
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

	// Create auth service and server
	authService := createTestAuthService(t)
	alertService := createTestAlertService(t)
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{},
		errToReturn:     nil,
	}
	mux := newServer(collector, time.Now(), authService, alertService, systemService, 15*time.Minute, 15*time.Minute, false)
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

	// Create auth service and server
	authService := createTestAuthService(t)
	alertService := createTestAlertService(t)
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{},
		errToReturn:     nil,
	}
	mux := newServer(collector, time.Now(), authService, alertService, systemService, 15*time.Minute, 15*time.Minute, false)
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

	// Create auth service and server
	authService := createTestAuthService(t)
	alertService := createTestAlertService(t)
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{},
		errToReturn:     nil,
	}
	mux := newServer(collector, time.Now(), authService, alertService, systemService, 15*time.Minute, 15*time.Minute, false)
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