package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/metrics"
)

// fakeCollector implements metrics.Collector for testing
type fakeCollector struct {
	metricsToReturn metrics.Metrics
	errToReturn     error
}

func (f *fakeCollector) Snapshot(ctx context.Context) (metrics.Metrics, error) {
	return f.metricsToReturn, f.errToReturn
}

func TestMetricsHandlerIntegration(t *testing.T) {
	// Setup fake collector with known values
	startTime := time.Now().Add(-10 * time.Second) // 10 seconds ago
	fakeMetrics := metrics.Metrics{
		CPUPct:      45.5,
		MemUsedPct:  67.2,
		DiskUsedPct: 23.8,
		UptimeS:     0, // Will be overwritten by handler
	}
	
	collector := &fakeCollector{
		metricsToReturn: fakeMetrics,
		errToReturn:     nil,
	}

	// Create server with fake collector
	mux := newServer(collector, startTime)
	server := httptest.NewServer(corsMiddleware(mux))
	defer server.Close()

	// Make request to /metrics endpoint
	resp, err := http.Get(server.URL + "/metrics")
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

	// Assert CORS headers are present
	corsOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if corsOrigin == "" {
		t.Error("Expected CORS Access-Control-Allow-Origin header to be present")
	}

	corsMethods := resp.Header.Get("Access-Control-Allow-Methods")
	if corsMethods == "" {
		t.Error("Expected CORS Access-Control-Allow-Methods header to be present")
	}

	corsHeaders := resp.Header.Get("Access-Control-Allow-Headers")
	if corsHeaders == "" {
		t.Error("Expected CORS Access-Control-Allow-Headers header to be present")
	}

	// Parse JSON response
	var result metrics.Metrics
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	// Assert collector values are preserved
	if result.CPUPct != fakeMetrics.CPUPct {
		t.Errorf("Expected CPU %.2f, got %.2f", fakeMetrics.CPUPct, result.CPUPct)
	}
	if result.MemUsedPct != fakeMetrics.MemUsedPct {
		t.Errorf("Expected Memory %.2f, got %.2f", fakeMetrics.MemUsedPct, result.MemUsedPct)
	}
	if result.DiskUsedPct != fakeMetrics.DiskUsedPct {
		t.Errorf("Expected Disk %.2f, got %.2f", fakeMetrics.DiskUsedPct, result.DiskUsedPct)
	}

	// Assert uptime is calculated correctly (should be ~10 seconds)
	if result.UptimeS < 9 || result.UptimeS > 11 {
		t.Errorf("Expected uptime around 10 seconds, got %.2f", result.UptimeS)
	}

	t.Logf("Integration test successful - CPU: %.2f%%, Memory: %.2f%%, Disk: %.2f%%, Uptime: %.2fs", 
		result.CPUPct, result.MemUsedPct, result.DiskUsedPct, result.UptimeS)
}

func TestMetricsHandlerWithCollectorError(t *testing.T) {
	// Setup fake collector that returns an error
	startTime := time.Now()
	collector := &fakeCollector{
		metricsToReturn: metrics.Metrics{}, // Will be ignored due to error
		errToReturn:     errors.New("simulated collector error"),
	}

	// Create server with failing collector
	mux := newServer(collector, startTime)
	server := httptest.NewServer(corsMiddleware(mux))
	defer server.Close()

	// Make request to /metrics endpoint
	resp, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert HTTP status is still 200 (we don't return 500 for collector errors)
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
	mux := newServer(collector, startTime)
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

	mux := newServer(collector, startTime)
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