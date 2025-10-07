package system

import (
	"context"
	"testing"
	"time"
)

func TestNewSystemService(t *testing.T) {
	service := NewSystemService()
	if service == nil {
		t.Fatal("NewSystemService() returned nil")
	}
	if service.loggedErrors == nil {
		t.Error("loggedErrors map not initialized")
	}
	if service.cacheTTL != time.Minute {
		t.Errorf("Expected cacheTTL to be 1 minute, got %v", service.cacheTTL)
	}
}

func TestGetSystemInfo(t *testing.T) {
	service := NewSystemService()
	ctx := context.Background()

	info, err := service.GetSystemInfo(ctx)
	if err != nil {
		t.Fatalf("GetSystemInfo() failed: %v", err)
	}

	// Check that we got some data (even if some fields might be empty due to errors)
	// We can't guarantee specific values since this runs on different systems
	t.Logf("System Info: %+v", info)

	// Test that hostname is likely populated (most systems should have this)
	if info.Hostname == "" {
		t.Log("Warning: Hostname is empty - this might be expected in some test environments")
	}

	// Test that CPU cores is reasonable (should be > 0 on real systems)
	if info.CPUCores == 0 {
		t.Log("Warning: CPU cores is 0 - this might be expected in some test environments")
	}

	// Test that memory total is reasonable (should be > 0 on real systems)
	if info.MemoryTotalMB == 0 {
		t.Log("Warning: Memory total is 0 - this might be expected in some test environments")
	}

	// Test that disk total is reasonable (should be > 0 on real systems)
	if info.DiskTotalGB == 0 {
		t.Log("Warning: Disk total is 0 - this might be expected in some test environments")
	}
}

func TestGetSystemInfoCaching(t *testing.T) {
	service := NewSystemService()
	ctx := context.Background()

	// First call should populate cache
	info1, err := service.GetSystemInfo(ctx)
	if err != nil {
		t.Fatalf("First GetSystemInfo() failed: %v", err)
	}

	// Second call should use cache (should be fast)
	start := time.Now()
	info2, err := service.GetSystemInfo(ctx)
	if err != nil {
		t.Fatalf("Second GetSystemInfo() failed: %v", err)
	}
	duration := time.Since(start)

	// Results should be identical (cached)
	if info1 != info2 {
		t.Error("Cached results don't match original results")
	}

	// Should be very fast (cached call)
	if duration > 10*time.Millisecond {
		t.Logf("Warning: Cached call took %v, expected < 10ms", duration)
	}
}

func TestGetSystemInfoCacheExpiry(t *testing.T) {
	service := NewSystemService()
	service.cacheTTL = 100 * time.Millisecond // Short TTL for testing
	ctx := context.Background()

	// First call
	info1, err := service.GetSystemInfo(ctx)
	if err != nil {
		t.Fatalf("First GetSystemInfo() failed: %v", err)
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Second call should refresh cache
	info2, err := service.GetSystemInfo(ctx)
	if err != nil {
		t.Fatalf("Second GetSystemInfo() failed: %v", err)
	}

	// Results should still be the same (system info doesn't change quickly)
	// but the uptime might be slightly different
	if info1.Hostname != info2.Hostname {
		t.Error("Hostname changed between calls")
	}
	if info1.Platform != info2.Platform {
		t.Error("Platform changed between calls")
	}
}

func TestErrorLogging(t *testing.T) {
	service := NewSystemService()

	// Test error logging functionality
	if service.hasLoggedError("test") {
		t.Error("hasLoggedError returned true for unlogged error")
	}

	service.markErrorLogged("test")
	if !service.hasLoggedError("test") {
		t.Error("hasLoggedError returned false for logged error")
	}
}