package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/agent/internal/collector"
	"github.com/Constantin-E-T/lunasentri/apps/agent/internal/config"
	"github.com/Constantin-E-T/lunasentri/apps/agent/internal/transport"
)

const version = "1.0.0"

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create collector and transport client
	metricsCollector := collector.New()
	apiClient := transport.NewClient(cfg.ServerURL, cfg.APIKey)
	logger := apiClient.Logger()

	// Log startup
	logger.Info("LunaSentri agent starting", map[string]interface{}{
		"version":            version,
		"server_url":         cfg.ServerURL,
		"interval":           cfg.Interval.String(),
		"system_info_period": cfg.SystemInfoPeriod.String(),
		"max_retries":        cfg.MaxRetries,
		"retry_backoff":      cfg.RetryBackoff.String(),
		"config_file":        cfg.ConfigFile,
	})

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Collect and send initial system info
	systemInfo, err := metricsCollector.CollectSystemInfo(ctx)
	if err != nil {
		logger.Warn("Failed to collect initial system info", map[string]interface{}{
			"error": err.Error(),
		})
		systemInfo = nil
	} else {
		logger.Info("System info collected", map[string]interface{}{
			"hostname":  getStringPtr(systemInfo.Hostname),
			"platform":  getStringPtr(systemInfo.Platform),
			"cpu_cores": getIntPtr(systemInfo.CPUCores),
			"memory_mb": getInt64Ptr(systemInfo.MemoryTotalMB),
			"disk_gb":   getInt64Ptr(systemInfo.DiskTotalGB),
		})
	}

	// Start metrics collection loop
	metricsTicker := time.NewTicker(cfg.Interval)
	defer metricsTicker.Stop()

	// Start system info update loop
	sysInfoTicker := time.NewTicker(cfg.SystemInfoPeriod)
	defer sysInfoTicker.Stop()

	// Track whether to send system info with next metrics
	sendSystemInfo := true
	consecutiveFailures := 0

	logger.Info("Agent started, entering metrics loop", nil)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled, shutting down", nil)
			return

		case <-sigChan:
			logger.Info("Received shutdown signal", nil)
			cancel()
			return

		case <-metricsTicker.C:
			// Collect metrics
			metrics, err := metricsCollector.CollectMetrics(ctx)
			if err != nil {
				logger.Error("Failed to collect metrics", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}

			// Prepare system info if needed
			var sysInfoToSend *collector.SystemInfo
			if sendSystemInfo && systemInfo != nil {
				sysInfoToSend = systemInfo
				sendSystemInfo = false // Don't send again until next period
			}

			// Send metrics to API
			err = apiClient.SendMetrics(ctx, metrics, sysInfoToSend, cfg.MaxRetries, cfg.RetryBackoff)
			if err != nil {
				consecutiveFailures++
				logger.Error("Failed to send metrics", map[string]interface{}{
					"error":                err.Error(),
					"consecutive_failures": consecutiveFailures,
				})

				// Log warning if too many consecutive failures
				if consecutiveFailures >= 5 {
					logger.Warn("Multiple consecutive failures detected", map[string]interface{}{
						"consecutive_failures": consecutiveFailures,
						"suggestion":           "Check network connectivity and API key",
					})
				}
			} else {
				// Reset failure counter on success
				if consecutiveFailures > 0 {
					logger.Info("Metrics delivery recovered", map[string]interface{}{
						"previous_failures": consecutiveFailures,
					})
					consecutiveFailures = 0
				}
			}

		case <-sysInfoTicker.C:
			// Update system info periodically
			logger.Info("Refreshing system info", nil)
			newSystemInfo, err := metricsCollector.CollectSystemInfo(ctx)
			if err != nil {
				logger.Warn("Failed to refresh system info", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				systemInfo = newSystemInfo
				sendSystemInfo = true // Send with next metrics payload
				logger.Info("System info refreshed", nil)
			}
		}
	}
}

// Helper functions to safely dereference pointers for logging
func getStringPtr(s *string) string {
	if s == nil {
		return "unknown"
	}
	return *s
}

func getIntPtr(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func getInt64Ptr(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
