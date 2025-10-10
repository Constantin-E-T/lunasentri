package collector

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// Metrics represents the collected system metrics
type Metrics struct {
	CPUPct      float64
	MemUsedPct  float64
	DiskUsedPct float64
	NetRxBytes  int64
	NetTxBytes  int64
	UptimeS     *float64
}

// SystemInfo represents system metadata
type SystemInfo struct {
	Hostname        *string
	Platform        *string
	PlatformVersion *string
	KernelVersion   *string
	CPUCores        *int
	MemoryTotalMB   *int64
	DiskTotalGB     *int64
	LastBootTime    *time.Time
}

// Collector collects system metrics
type Collector struct {
	// Track previous network stats for delta calculation
	prevNetStats *net.IOCountersStat
	// Track errors to avoid log spam
	loggedErrors map[string]bool
}

// New creates a new Collector instance
func New() *Collector {
	return &Collector{
		loggedErrors: make(map[string]bool),
	}
}

// CollectMetrics collects current system metrics
func (c *Collector) CollectMetrics(ctx context.Context) (*Metrics, error) {
	metrics := &Metrics{}

	// Collect CPU percentage
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		c.logOnce("cpu", "Failed to collect CPU metrics: %v", err)
		metrics.CPUPct = 0
	} else if len(cpuPercent) > 0 {
		metrics.CPUPct = cpuPercent[0]
	}

	// Collect memory usage
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		c.logOnce("memory", "Failed to collect memory metrics: %v", err)
		metrics.MemUsedPct = 0
	} else {
		metrics.MemUsedPct = memInfo.UsedPercent
	}

	// Collect disk usage for root filesystem
	diskInfo, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		c.logOnce("disk", "Failed to collect disk metrics: %v", err)
		metrics.DiskUsedPct = 0
	} else {
		metrics.DiskUsedPct = diskInfo.UsedPercent
	}

	// Collect network stats (total across all interfaces)
	netStats, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		c.logOnce("network", "Failed to collect network metrics: %v", err)
		metrics.NetRxBytes = 0
		metrics.NetTxBytes = 0
	} else if len(netStats) > 0 {
		// Use cumulative totals
		metrics.NetRxBytes = int64(netStats[0].BytesRecv)
		metrics.NetTxBytes = int64(netStats[0].BytesSent)
	}

	// Collect uptime (optional)
	uptime, err := host.UptimeWithContext(ctx)
	if err != nil {
		c.logOnce("uptime", "Failed to collect uptime: %v", err)
	} else {
		uptimeFloat := float64(uptime)
		metrics.UptimeS = &uptimeFloat
	}

	return metrics, nil
}

// CollectSystemInfo collects system metadata
func (c *Collector) CollectSystemInfo(ctx context.Context) (*SystemInfo, error) {
	info := &SystemInfo{}

	// Get host information
	hostInfo, err := host.InfoWithContext(ctx)
	if err == nil {
		info.Hostname = &hostInfo.Hostname

		if hostInfo.Platform != "" {
			info.Platform = &hostInfo.Platform
		}

		if hostInfo.PlatformVersion != "" {
			info.PlatformVersion = &hostInfo.PlatformVersion
		}

		if hostInfo.KernelVersion != "" {
			info.KernelVersion = &hostInfo.KernelVersion
		}

		// Get boot time
		if hostInfo.BootTime != 0 {
			bootTime := time.Unix(int64(hostInfo.BootTime), 0)
			info.LastBootTime = &bootTime
		}
	}

	// Get CPU cores
	cpuCores := runtime.NumCPU()
	info.CPUCores = &cpuCores

	// Get total memory (in MB)
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		totalMB := int64(memInfo.Total / (1024 * 1024))
		info.MemoryTotalMB = &totalMB
	}

	// Get total disk space (in GB)
	diskInfo, err := disk.UsageWithContext(ctx, "/")
	if err == nil {
		totalGB := int64(diskInfo.Total / (1024 * 1024 * 1024))
		info.DiskTotalGB = &totalGB
	}

	return info, nil
}

// logOnce logs an error message only once per error type
func (c *Collector) logOnce(errorType, format string, args ...interface{}) {
	if !c.loggedErrors[errorType] {
		log.Printf(format, args...)
		c.loggedErrors[errorType] = true
	}
}
