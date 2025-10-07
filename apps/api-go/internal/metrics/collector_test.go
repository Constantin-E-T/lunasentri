package metrics

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSystemCollector_Snapshot(t *testing.T) {
	collector := NewSystemCollector()

	// Create context with timeout to avoid hanging tests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call Snapshot and ensure it doesn't panic
	metrics, err := collector.Snapshot(ctx)

	// Should not return an error (errors are handled internally)
	if err != nil {
		t.Errorf("Snapshot() returned unexpected error: %v", err)
	}

	// Test CPU percentage is within valid range (0-100) or 0 on unsupported platforms
	if metrics.CPUPct < 0 || metrics.CPUPct > 100 {
		t.Errorf("CPU percentage out of range: got %f, want 0-100", metrics.CPUPct)
	}

	// Test memory percentage is within valid range (0-100) or 0 on unsupported platforms
	if metrics.MemUsedPct < 0 || metrics.MemUsedPct > 100 {
		t.Errorf("Memory percentage out of range: got %f, want 0-100", metrics.MemUsedPct)
	}

	// Test disk percentage is within valid range (0-100) or 0 on unsupported platforms
	if metrics.DiskUsedPct < 0 || metrics.DiskUsedPct > 100 {
		t.Errorf("Disk percentage out of range: got %f, want 0-100", metrics.DiskUsedPct)
	}

	// UptimeS should be 0 as it's set by the caller
	if metrics.UptimeS != 0 {
		t.Errorf("UptimeS should be 0 (set by caller): got %f, want 0", metrics.UptimeS)
	}

	t.Logf("Collected metrics - CPU: %.2f%%, Memory: %.2f%%, Disk: %.2f%%",
		metrics.CPUPct, metrics.MemUsedPct, metrics.DiskUsedPct)
}

func TestNewSystemCollector(t *testing.T) {
	collector := NewSystemCollector()

	if collector == nil {
		t.Error("NewSystemCollector() returned nil")
	}

	if collector.loggedErrors == nil {
		t.Error("NewSystemCollector() did not initialize loggedErrors map")
	}
}

func TestSystemCollector_MultipleSnapshots(t *testing.T) {
	collector := NewSystemCollector()
	ctx := context.Background()

	// Take multiple snapshots to ensure no issues with concurrent access
	for i := 0; i < 3; i++ {
		metrics, err := collector.Snapshot(ctx)
		if err != nil {
			t.Errorf("Snapshot %d returned error: %v", i, err)
		}

		// Basic range checks
		if metrics.CPUPct < 0 || metrics.CPUPct > 100 {
			t.Errorf("Snapshot %d: CPU percentage out of range: %f", i, metrics.CPUPct)
		}
		if metrics.MemUsedPct < 0 || metrics.MemUsedPct > 100 {
			t.Errorf("Snapshot %d: Memory percentage out of range: %f", i, metrics.MemUsedPct)
		}
		if metrics.DiskUsedPct < 0 || metrics.DiskUsedPct > 100 {
			t.Errorf("Snapshot %d: Disk percentage out of range: %f", i, metrics.DiskUsedPct)
		}
	}
}

func TestSystemCollector_ConcurrentAccess(t *testing.T) {
	collector := NewSystemCollector()
	ctx := context.Background()

	// Number of concurrent goroutines
	numGoroutines := 10
	numCallsPerGoroutine := 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Channel to collect any panics or errors
	errors := make(chan error, numGoroutines*numCallsPerGoroutine)

	// Run multiple goroutines calling Snapshot concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numCallsPerGoroutine; j++ {
				_, err := collector.Snapshot(ctx)
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent Snapshot() call failed: %v", err)
	}

	t.Logf("Successfully completed %d concurrent calls across %d goroutines",
		numGoroutines*numCallsPerGoroutine, numGoroutines)
}

func TestSystemCollector_ThreadSafety(t *testing.T) {
	collector := NewSystemCollector()

	// Test that hasLoggedError and markErrorLogged are thread-safe
	var wg sync.WaitGroup
	numGoroutines := 20

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Try to access and modify the logged errors map concurrently
			errorType := "test_error"

			if !collector.hasLoggedError(errorType) {
				collector.markErrorLogged(errorType)
			}

			// Verify it's now marked as logged
			if !collector.hasLoggedError(errorType) {
				t.Errorf("Goroutine %d: Error should be marked as logged", id)
			}
		}(i)
	}

	wg.Wait()

	// Final verification
	if !collector.hasLoggedError("test_error") {
		t.Error("Final check: test_error should be marked as logged")
	}
}
