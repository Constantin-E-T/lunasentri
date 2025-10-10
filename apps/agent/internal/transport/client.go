package transport

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/agent/internal/collector"
)

// MetricsPayload represents the metrics payload sent to the API
type MetricsPayload struct {
	Timestamp   *time.Time  `json:"timestamp,omitempty"`
	CPUPct      float64     `json:"cpu_pct"`
	MemUsedPct  float64     `json:"mem_used_pct"`
	DiskUsedPct float64     `json:"disk_used_pct"`
	NetRxBytes  int64       `json:"net_rx_bytes,omitempty"`
	NetTxBytes  int64       `json:"net_tx_bytes,omitempty"`
	UptimeS     *float64    `json:"uptime_s,omitempty"`
	SystemInfo  *SystemInfo `json:"system_info,omitempty"`
}

// SystemInfo represents system metadata in the payload
type SystemInfo struct {
	Hostname        *string    `json:"hostname,omitempty"`
	Platform        *string    `json:"platform,omitempty"`
	PlatformVersion *string    `json:"platform_version,omitempty"`
	KernelVersion   *string    `json:"kernel_version,omitempty"`
	CPUCores        *int       `json:"cpu_cores,omitempty"`
	MemoryTotalMB   *int64     `json:"memory_total_mb,omitempty"`
	DiskTotalGB     *int64     `json:"disk_total_gb,omitempty"`
	LastBootTime    *time.Time `json:"last_boot_time,omitempty"`
}

// Client handles communication with the LunaSentri API
type Client struct {
	serverURL  string
	apiKey     string
	httpClient *http.Client
	logger     *Logger
}

// Logger provides structured JSON logging
type Logger struct {
	apiKeyHash string // First 8 chars of API key hash for debugging
}

// NewLogger creates a new logger instance
func NewLogger(apiKey string) *Logger {
	// Hash the API key and take first 8 chars for logging
	hash := sha256.Sum256([]byte(apiKey))
	hashStr := hex.EncodeToString(hash[:])
	return &Logger{
		apiKeyHash: hashStr[:8],
	}
}

// Log writes a structured JSON log entry
func (l *Logger) Log(level, msg string, fields map[string]interface{}) {
	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     level,
		"msg":       msg,
	}

	// Add API key hash for debugging
	entry["api_key_hash"] = l.apiKeyHash

	// Merge additional fields
	for k, v := range fields {
		entry[k] = v
	}

	jsonBytes, _ := json.Marshal(entry)
	fmt.Println(string(jsonBytes))
}

// Info logs an info message
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.Log("info", msg, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.Log("warn", msg, fields)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.Log("error", msg, fields)
}

// NewClient creates a new API client
func NewClient(serverURL, apiKey string) *Client {
	return &Client{
		serverURL: serverURL,
		apiKey:    apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: NewLogger(apiKey),
	}
}

// SendMetrics sends metrics to the API with retry logic
func (c *Client) SendMetrics(ctx context.Context, metrics *collector.Metrics, sysInfo *collector.SystemInfo, maxRetries int, retryBackoff time.Duration) error {
	// Build payload
	payload := MetricsPayload{
		CPUPct:      metrics.CPUPct,
		MemUsedPct:  metrics.MemUsedPct,
		DiskUsedPct: metrics.DiskUsedPct,
		NetRxBytes:  metrics.NetRxBytes,
		NetTxBytes:  metrics.NetTxBytes,
		UptimeS:     metrics.UptimeS,
	}

	// Add system info if provided
	if sysInfo != nil {
		payload.SystemInfo = &SystemInfo{
			Hostname:        sysInfo.Hostname,
			Platform:        sysInfo.Platform,
			PlatformVersion: sysInfo.PlatformVersion,
			KernelVersion:   sysInfo.KernelVersion,
			CPUCores:        sysInfo.CPUCores,
			MemoryTotalMB:   sysInfo.MemoryTotalMB,
			DiskTotalGB:     sysInfo.DiskTotalGB,
			LastBootTime:    sysInfo.LastBootTime,
		}
	}

	// Marshal payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Retry logic
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := retryBackoff * time.Duration(1<<uint(attempt-1))
			c.logger.Warn("Retrying metrics send", map[string]interface{}{
				"attempt":      attempt,
				"max_retries":  maxRetries,
				"backoff_secs": backoff.Seconds(),
			})

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL+"/agent/metrics", bytes.NewReader(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		// Send request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to send request: %w", err)
			c.logger.Error("HTTP request failed", map[string]interface{}{
				"error":   err.Error(),
				"attempt": attempt,
			})
			continue
		}

		// Read response body
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Check status code
		if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK {
			c.logger.Info("Metrics sent successfully", map[string]interface{}{
				"status_code": resp.StatusCode,
				"cpu_pct":     fmt.Sprintf("%.1f", metrics.CPUPct),
				"mem_pct":     fmt.Sprintf("%.1f", metrics.MemUsedPct),
				"disk_pct":    fmt.Sprintf("%.1f", metrics.DiskUsedPct),
			})
			return nil
		}

		// Non-200 response
		lastErr = fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		c.logger.Error("API error", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(body),
			"attempt":     attempt,
		})

		// Don't retry on client errors (4xx except 429)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return lastErr
		}
	}

	// All retries exhausted
	c.logger.Error("All retry attempts failed", map[string]interface{}{
		"max_retries": maxRetries,
		"last_error":  lastErr.Error(),
	})
	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// Logger returns the client's logger
func (c *Client) Logger() *Logger {
	return c.logger
}
