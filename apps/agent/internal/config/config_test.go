package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ServerURL != "https://api.example.com" {
		t.Errorf("Expected default server URL to be 'https://api.example.com', got '%s'", cfg.ServerURL)
	}

	if cfg.Interval != 10*time.Second {
		t.Errorf("Expected default interval to be 10s, got %v", cfg.Interval)
	}

	if cfg.SystemInfoPeriod != 1*time.Hour {
		t.Errorf("Expected default system info period to be 1h, got %v", cfg.SystemInfoPeriod)
	}

	if cfg.MaxRetries != 3 {
		t.Errorf("Expected default max retries to be 3, got %d", cfg.MaxRetries)
	}
}

func TestLoadFromConfigFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "agent.yaml")

	configContent := `
server_url: "https://file.example.com"
api_key: "file-key-456"
interval: "20s"
system_info_period: "30m"
max_retries: 7
retry_backoff: "15s"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	cfg := DefaultConfig()
	if err := loadConfigFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to load config file: %v", err)
	}

	if cfg.ServerURL != "https://file.example.com" {
		t.Errorf("Expected server URL from file, got '%s'", cfg.ServerURL)
	}

	if cfg.APIKey != "file-key-456" {
		t.Errorf("Expected API key from file, got '%s'", cfg.APIKey)
	}

	if cfg.Interval != 20*time.Second {
		t.Errorf("Expected interval 20s from file, got %v", cfg.Interval)
	}

	if cfg.SystemInfoPeriod != 30*time.Minute {
		t.Errorf("Expected system info period 30m from file, got %v", cfg.SystemInfoPeriod)
	}

	if cfg.MaxRetries != 7 {
		t.Errorf("Expected max retries 7 from file, got %d", cfg.MaxRetries)
	}

	if cfg.RetryBackoff != 15*time.Second {
		t.Errorf("Expected retry backoff 15s from file, got %v", cfg.RetryBackoff)
	}
}

func TestConfigPrecedence(t *testing.T) {
	// Create config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "agent.yaml")

	configContent := `
server_url: "https://file.example.com"
api_key: "file-key"
interval: "20s"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set env var (should override file)
	os.Setenv("LUNASENTRI_SERVER_URL", "https://env.example.com")
	os.Setenv("LUNASENTRI_API_KEY", "env-key")
	defer func() {
		os.Unsetenv("LUNASENTRI_SERVER_URL")
		os.Unsetenv("LUNASENTRI_API_KEY")
	}()

	cfg := DefaultConfig()
	if err := loadConfigFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to load config file: %v", err)
	}

	// Apply env vars
	if url := os.Getenv("LUNASENTRI_SERVER_URL"); url != "" {
		cfg.ServerURL = url
	}
	if key := os.Getenv("LUNASENTRI_API_KEY"); key != "" {
		cfg.APIKey = key
	}

	// Env var should win
	if cfg.ServerURL != "https://env.example.com" {
		t.Errorf("Expected env var to override file, got '%s'", cfg.ServerURL)
	}

	if cfg.APIKey != "env-key" {
		t.Errorf("Expected env var to override file, got '%s'", cfg.APIKey)
	}

	// File value should be used for interval (no env override)
	if cfg.Interval != 20*time.Second {
		t.Errorf("Expected interval from file, got %v", cfg.Interval)
	}
}
