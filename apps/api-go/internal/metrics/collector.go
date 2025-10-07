package metrics

import (
	"context"
	"log"
	"sync"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

// Metrics represents the system metrics response
type Metrics struct {
	CPUPct      float64 `json:"cpu_pct"`
	MemUsedPct  float64 `json:"mem_used_pct"`
	DiskUsedPct float64 `json:"disk_used_pct"`
	UptimeS     float64 `json:"uptime_s"`
}

// Collector interface defines methods for collecting system metrics
type Collector interface {
	Snapshot(ctx context.Context) (Metrics, error)
}

// SystemCollector implements Collector using gopsutil for real system metrics
type SystemCollector struct {
	loggedErrors map[string]bool // Track logged errors to avoid spam
	mutex        sync.RWMutex    // Protect concurrent access to loggedErrors
}

// NewSystemCollector creates a new SystemCollector instance
func NewSystemCollector() *SystemCollector {
	return &SystemCollector{
		loggedErrors: make(map[string]bool),
	}
}

// hasLoggedError safely checks if an error type has been logged
func (sc *SystemCollector) hasLoggedError(errorType string) bool {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	return sc.loggedErrors[errorType]
}

// markErrorLogged safely marks an error type as logged
func (sc *SystemCollector) markErrorLogged(errorType string) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	sc.loggedErrors[errorType] = true
}

// Snapshot collects current system metrics
func (sc *SystemCollector) Snapshot(ctx context.Context) (Metrics, error) {
	result := Metrics{
		UptimeS: 0, // Will be set by caller with real uptime
	}

	// Collect CPU percentage
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		if !sc.hasLoggedError("cpu") {
			log.Printf("Failed to collect CPU metrics: %v", err)
			sc.markErrorLogged("cpu")
		}
		result.CPUPct = 0
	} else if len(cpuPercent) > 0 {
		result.CPUPct = cpuPercent[0]
	}

	// Collect memory usage
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		if !sc.hasLoggedError("memory") {
			log.Printf("Failed to collect memory metrics: %v", err)
			sc.markErrorLogged("memory")
		}
		result.MemUsedPct = 0
	} else {
		result.MemUsedPct = memInfo.UsedPercent
	}

	// Collect disk usage for root filesystem
	diskInfo, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		if !sc.hasLoggedError("disk") {
			log.Printf("Failed to collect disk metrics: %v", err)
			sc.markErrorLogged("disk")
		}
		result.DiskUsedPct = 0
	} else {
		result.DiskUsedPct = diskInfo.UsedPercent
	}

	return result, nil
}
