package system

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// SystemInfo represents the system information response
type SystemInfo struct {
	Hostname        string `json:"hostname"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	UptimeS         uint64 `json:"uptime_s"`
	CPUCores        int    `json:"cpu_cores"`
	MemoryTotalMB   uint64 `json:"memory_total_mb"`
	DiskTotalGB     uint64 `json:"disk_total_gb"`
	LastBootTime    uint64 `json:"last_boot_time"`
}

// Service interface defines methods for collecting system information
type Service interface {
	GetSystemInfo(ctx context.Context) (SystemInfo, error)
}

// SystemService implements Service using gopsutil for real system information
type SystemService struct {
	loggedErrors map[string]bool // Track logged errors to avoid spam
	mutex        sync.RWMutex    // Protect concurrent access to loggedErrors
	cache        *SystemInfo     // Cached system info
	cacheTime    time.Time       // When cache was last updated
	cacheTTL     time.Duration   // Cache time-to-live
}

// NewSystemService creates a new SystemService instance
func NewSystemService() *SystemService {
	return &SystemService{
		loggedErrors: make(map[string]bool),
		cacheTTL:     time.Minute, // Cache for 1 minute
	}
}

// hasLoggedError safely checks if an error type has been logged
func (ss *SystemService) hasLoggedError(errorType string) bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.loggedErrors[errorType]
}

// markErrorLogged safely marks an error type as logged
func (ss *SystemService) markErrorLogged(errorType string) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	ss.loggedErrors[errorType] = true
}

// getCachedInfo safely gets cached system info
func (ss *SystemService) getCachedInfo() *SystemInfo {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	if ss.cache != nil && time.Since(ss.cacheTime) < ss.cacheTTL {
		return ss.cache
	}
	return nil
}

// setCachedInfo safely sets cached system info
func (ss *SystemService) setCachedInfo(info SystemInfo) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	ss.cache = &info
	ss.cacheTime = time.Now()
}

// GetSystemInfo collects current system information
func (ss *SystemService) GetSystemInfo(ctx context.Context) (SystemInfo, error) {
	// Check cache first
	if cached := ss.getCachedInfo(); cached != nil {
		return *cached, nil
	}

	result := SystemInfo{}

	// Get host information
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		if !ss.hasLoggedError("host") {
			log.Printf("Failed to collect host info: %v", err)
			ss.markErrorLogged("host")
		}
		// Continue with partial data
	} else {
		result.Hostname = hostInfo.Hostname
		result.Platform = hostInfo.Platform
		result.PlatformVersion = hostInfo.PlatformVersion
		result.KernelVersion = hostInfo.KernelVersion
		result.UptimeS = hostInfo.Uptime
		result.LastBootTime = hostInfo.BootTime
	}

	// Get CPU core count
	cpuCounts, err := cpu.CountsWithContext(ctx, true) // logical cores
	if err != nil {
		if !ss.hasLoggedError("cpu") {
			log.Printf("Failed to collect CPU info: %v", err)
			ss.markErrorLogged("cpu")
		}
		result.CPUCores = 0
	} else {
		result.CPUCores = cpuCounts
	}

	// Get memory information
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		if !ss.hasLoggedError("memory") {
			log.Printf("Failed to collect memory info: %v", err)
			ss.markErrorLogged("memory")
		}
		result.MemoryTotalMB = 0
	} else {
		result.MemoryTotalMB = memInfo.Total / (1024 * 1024) // Convert bytes to MB
	}

	// Get disk information for root filesystem
	diskInfo, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		if !ss.hasLoggedError("disk") {
			log.Printf("Failed to collect disk info: %v", err)
			ss.markErrorLogged("disk")
		}
		result.DiskTotalGB = 0
	} else {
		result.DiskTotalGB = diskInfo.Total / (1024 * 1024 * 1024) // Convert bytes to GB
	}

	// Cache the result
	ss.setCachedInfo(result)

	return result, nil
}
