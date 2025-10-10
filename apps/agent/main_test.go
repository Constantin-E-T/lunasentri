package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/agent/internal/collector"
	"github.com/Constantin-E-T/lunasentri/apps/agent/internal/transport"
)

// TestMetricsPayloadShape verifies the metrics payload structure
func TestMetricsPayloadShape(t *testing.T) {
	// Create a mock server
	requestReceived := false
	var receivedPayload transport.MetricsPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/agent/metrics" {
			t.Errorf("Expected path /agent/metrics, got %s", r.URL.Path)
		}

		// Verify authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header 'Bearer test-api-key', got '%s'", authHeader)
		}

		// Decode payload
		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Errorf("Failed to decode payload: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		requestReceived = true
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	// Create client
	client := transport.NewClient(server.URL, "test-api-key")

	// Create sample metrics
	uptime := 12345.0
	metrics := &collector.Metrics{
		CPUPct:      45.5,
		MemUsedPct:  67.8,
		DiskUsedPct: 23.4,
		NetRxBytes:  1024000,
		NetTxBytes:  512000,
		UptimeS:     &uptime,
	}

	hostname := "test-host"
	platform := "linux"
	cpuCores := 4
	memoryMB := int64(8192)

	sysInfo := &collector.SystemInfo{
		Hostname:      &hostname,
		Platform:      &platform,
		CPUCores:      &cpuCores,
		MemoryTotalMB: &memoryMB,
	}

	// Send metrics
	ctx := context.Background()
	err := client.SendMetrics(ctx, metrics, sysInfo, 3, 1*time.Second)

	if err != nil {
		t.Fatalf("Failed to send metrics: %v", err)
	}

	if !requestReceived {
		t.Fatal("Request was not received by mock server")
	}

	// Verify payload structure
	if receivedPayload.CPUPct != 45.5 {
		t.Errorf("Expected CPU 45.5%%, got %.1f%%", receivedPayload.CPUPct)
	}

	if receivedPayload.MemUsedPct != 67.8 {
		t.Errorf("Expected Memory 67.8%%, got %.1f%%", receivedPayload.MemUsedPct)
	}

	if receivedPayload.DiskUsedPct != 23.4 {
		t.Errorf("Expected Disk 23.4%%, got %.1f%%", receivedPayload.DiskUsedPct)
	}

	if receivedPayload.NetRxBytes != 1024000 {
		t.Errorf("Expected NetRxBytes 1024000, got %d", receivedPayload.NetRxBytes)
	}

	if receivedPayload.NetTxBytes != 512000 {
		t.Errorf("Expected NetTxBytes 512000, got %d", receivedPayload.NetTxBytes)
	}

	if receivedPayload.UptimeS == nil || *receivedPayload.UptimeS != 12345.0 {
		t.Errorf("Expected UptimeS 12345.0, got %v", receivedPayload.UptimeS)
	}

	if receivedPayload.SystemInfo == nil {
		t.Fatal("Expected SystemInfo to be present")
	}

	if receivedPayload.SystemInfo.Hostname == nil || *receivedPayload.SystemInfo.Hostname != "test-host" {
		t.Errorf("Expected hostname 'test-host', got %v", receivedPayload.SystemInfo.Hostname)
	}

	if receivedPayload.SystemInfo.Platform == nil || *receivedPayload.SystemInfo.Platform != "linux" {
		t.Errorf("Expected platform 'linux', got %v", receivedPayload.SystemInfo.Platform)
	}
}

// TestRetryBehavior verifies retry logic
func TestRetryBehavior(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		// Fail first 2 attempts, succeed on 3rd
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := transport.NewClient(server.URL, "test-api-key")

	metrics := &collector.Metrics{
		CPUPct:      50.0,
		MemUsedPct:  60.0,
		DiskUsedPct: 70.0,
	}

	ctx := context.Background()
	err := client.SendMetrics(ctx, metrics, nil, 3, 100*time.Millisecond)

	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

// TestRetryExhaustion verifies behavior when all retries fail
func TestRetryExhaustion(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := transport.NewClient(server.URL, "test-api-key")

	metrics := &collector.Metrics{
		CPUPct:      50.0,
		MemUsedPct:  60.0,
		DiskUsedPct: 70.0,
	}

	ctx := context.Background()
	err := client.SendMetrics(ctx, metrics, nil, 2, 50*time.Millisecond)

	if err == nil {
		t.Fatal("Expected error after exhausting retries, got nil")
	}

	// Should be initial attempt + 2 retries = 3 total
	if attempts != 3 {
		t.Errorf("Expected 3 attempts (initial + 2 retries), got %d", attempts)
	}
}

// TestClientErrorNoRetry verifies 4xx errors don't retry (except 429)
func TestClientErrorNoRetry(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest) // 400
	}))
	defer server.Close()

	client := transport.NewClient(server.URL, "test-api-key")

	metrics := &collector.Metrics{
		CPUPct:      50.0,
		MemUsedPct:  60.0,
		DiskUsedPct: 70.0,
	}

	ctx := context.Background()
	err := client.SendMetrics(ctx, metrics, nil, 3, 50*time.Millisecond)

	if err == nil {
		t.Fatal("Expected error for 400 response, got nil")
	}

	// Should only attempt once (no retries for 4xx)
	if attempts != 1 {
		t.Errorf("Expected 1 attempt (no retries for 4xx), got %d", attempts)
	}
}
