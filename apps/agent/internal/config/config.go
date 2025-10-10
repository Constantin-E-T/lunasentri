package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the agent configuration
type Config struct {
	ServerURL        string        `yaml:"server_url"`
	APIKey           string        `yaml:"api_key"`
	Interval         time.Duration `yaml:"interval"`
	SystemInfoPeriod time.Duration `yaml:"system_info_period"`
	MaxRetries       int           `yaml:"max_retries"`
	RetryBackoff     time.Duration `yaml:"retry_backoff"`
	ConfigFile       string        `yaml:"-"` // Not from file
}

// FileConfig represents the YAML configuration file structure
type FileConfig struct {
	ServerURL        string `yaml:"server_url"`
	APIKey           string `yaml:"api_key"`
	Interval         string `yaml:"interval"`           // Duration as string in YAML
	SystemInfoPeriod string `yaml:"system_info_period"` // Duration as string in YAML
	MaxRetries       int    `yaml:"max_retries"`
	RetryBackoff     string `yaml:"retry_backoff"` // Duration as string in YAML
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		ServerURL:        "https://api.example.com",
		APIKey:           "",
		Interval:         10 * time.Second,
		SystemInfoPeriod: 1 * time.Hour,
		MaxRetries:       3,
		RetryBackoff:     5 * time.Second,
	}
}

// Load loads configuration with the following precedence:
// 1. Command-line flags
// 2. Environment variables
// 3. Config file
// 4. Default values
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Define flags
	var (
		serverURL        = flag.String("server-url", "", "LunaSentri server URL")
		apiKey           = flag.String("api-key", "", "Machine API key")
		interval         = flag.Duration("interval", 0, "Metrics collection interval")
		systemInfoPeriod = flag.Duration("system-info-period", 0, "System info update period")
		configFile       = flag.String("config", "", "Path to configuration file")
		maxRetries       = flag.Int("max-retries", 0, "Maximum retry attempts")
		retryBackoff     = flag.Duration("retry-backoff", 0, "Retry backoff duration")
	)

	flag.Parse()

	// Try to load config file (check flag, then default locations)
	configPath := *configFile
	if configPath == "" {
		// Try default locations
		configPath = findConfigFile()
	}

	if configPath != "" {
		if err := loadConfigFile(configPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
		cfg.ConfigFile = configPath
	}

	// Override with environment variables
	if url := os.Getenv("LUNASENTRI_SERVER_URL"); url != "" {
		cfg.ServerURL = url
	}
	if key := os.Getenv("LUNASENTRI_API_KEY"); key != "" {
		cfg.APIKey = key
	}
	if intervalStr := os.Getenv("LUNASENTRI_INTERVAL"); intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			cfg.Interval = d
		}
	}
	if periodStr := os.Getenv("LUNASENTRI_SYSTEM_INFO_PERIOD"); periodStr != "" {
		if d, err := time.ParseDuration(periodStr); err == nil {
			cfg.SystemInfoPeriod = d
		}
	}
	if retriesStr := os.Getenv("LUNASENTRI_MAX_RETRIES"); retriesStr != "" {
		var retries int
		if _, err := fmt.Sscanf(retriesStr, "%d", &retries); err == nil && retries > 0 {
			cfg.MaxRetries = retries
		}
	}
	if backoffStr := os.Getenv("LUNASENTRI_RETRY_BACKOFF"); backoffStr != "" {
		if d, err := time.ParseDuration(backoffStr); err == nil {
			cfg.RetryBackoff = d
		}
	}

	// Override with command-line flags (highest precedence)
	if *serverURL != "" {
		cfg.ServerURL = *serverURL
	}
	if *apiKey != "" {
		cfg.APIKey = *apiKey
	}
	if *interval != 0 {
		cfg.Interval = *interval
	}
	if *systemInfoPeriod != 0 {
		cfg.SystemInfoPeriod = *systemInfoPeriod
	}
	if *maxRetries != 0 {
		cfg.MaxRetries = *maxRetries
	}
	if *retryBackoff != 0 {
		cfg.RetryBackoff = *retryBackoff
	}

	// Validate required fields
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required (set via --api-key flag, LUNASENTRI_API_KEY env var, or config file)")
	}

	return cfg, nil
}

// findConfigFile searches for config file in default locations
func findConfigFile() string {
	// Check /etc/lunasentri/agent.yaml (system-wide)
	if fileExists("/etc/lunasentri/agent.yaml") {
		return "/etc/lunasentri/agent.yaml"
	}

	// Check ~/.config/lunasentri/agent.yaml (user-specific)
	if homeDir, err := os.UserHomeDir(); err == nil {
		userConfig := filepath.Join(homeDir, ".config", "lunasentri", "agent.yaml")
		if fileExists(userConfig) {
			return userConfig
		}
	}

	return ""
}

// loadConfigFile loads configuration from a YAML file
func loadConfigFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fileCfg FileConfig
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return err
	}

	// Apply values from file
	if fileCfg.ServerURL != "" {
		cfg.ServerURL = fileCfg.ServerURL
	}
	if fileCfg.APIKey != "" {
		cfg.APIKey = fileCfg.APIKey
	}
	if fileCfg.Interval != "" {
		if d, err := time.ParseDuration(fileCfg.Interval); err == nil {
			cfg.Interval = d
		}
	}
	if fileCfg.SystemInfoPeriod != "" {
		if d, err := time.ParseDuration(fileCfg.SystemInfoPeriod); err == nil {
			cfg.SystemInfoPeriod = d
		}
	}
	if fileCfg.MaxRetries > 0 {
		cfg.MaxRetries = fileCfg.MaxRetries
	}
	if fileCfg.RetryBackoff != "" {
		if d, err := time.ParseDuration(fileCfg.RetryBackoff); err == nil {
			cfg.RetryBackoff = d
		}
	}

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
