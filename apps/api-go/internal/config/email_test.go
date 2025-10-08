package config

import (
	"os"
	"testing"
)

func TestLoadEmailConfig_NoProvider(t *testing.T) {
	// Clear environment
	os.Unsetenv("EMAIL_PROVIDER")
	os.Unsetenv("M365_TENANT_ID")
	os.Unsetenv("M365_CLIENT_ID")
	os.Unsetenv("M365_CLIENT_SECRET")
	os.Unsetenv("M365_SENDER")

	cfg, err := LoadEmailConfig()
	if err != nil {
		t.Fatalf("Expected no error when provider not set, got: %v", err)
	}

	if cfg.Provider != EmailProviderNone {
		t.Errorf("Expected provider to be empty, got: %s", cfg.Provider)
	}

	if cfg.IsEnabled() {
		t.Error("Expected email to be disabled when no provider set")
	}
}

func TestLoadEmailConfig_M365Valid(t *testing.T) {
	// Set valid M365 config
	os.Setenv("EMAIL_PROVIDER", "m365")
	os.Setenv("M365_TENANT_ID", "tenant-123")
	os.Setenv("M365_CLIENT_ID", "client-456")
	os.Setenv("M365_CLIENT_SECRET", "secret-789")
	os.Setenv("M365_SENDER", "alerts@example.com")

	defer func() {
		os.Unsetenv("EMAIL_PROVIDER")
		os.Unsetenv("M365_TENANT_ID")
		os.Unsetenv("M365_CLIENT_ID")
		os.Unsetenv("M365_CLIENT_SECRET")
		os.Unsetenv("M365_SENDER")
	}()

	cfg, err := LoadEmailConfig()
	if err != nil {
		t.Fatalf("Expected no error with valid M365 config, got: %v", err)
	}

	if cfg.Provider != EmailProviderM365 {
		t.Errorf("Expected provider to be m365, got: %s", cfg.Provider)
	}

	if !cfg.IsEnabled() {
		t.Error("Expected email to be enabled")
	}

	if cfg.M365TenantID != "tenant-123" {
		t.Errorf("Expected tenant ID tenant-123, got: %s", cfg.M365TenantID)
	}
	if cfg.M365ClientID != "client-456" {
		t.Errorf("Expected client ID client-456, got: %s", cfg.M365ClientID)
	}
	if cfg.M365Secret != "secret-789" {
		t.Errorf("Expected secret secret-789, got: %s", cfg.M365Secret)
	}
	if cfg.M365Sender != "alerts@example.com" {
		t.Errorf("Expected sender alerts@example.com, got: %s", cfg.M365Sender)
	}
}

func TestLoadEmailConfig_M365MissingTenantID(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "m365")
	os.Setenv("M365_CLIENT_ID", "client-456")
	os.Setenv("M365_CLIENT_SECRET", "secret-789")
	os.Setenv("M365_SENDER", "alerts@example.com")

	defer func() {
		os.Unsetenv("EMAIL_PROVIDER")
		os.Unsetenv("M365_CLIENT_ID")
		os.Unsetenv("M365_CLIENT_SECRET")
		os.Unsetenv("M365_SENDER")
	}()

	_, err := LoadEmailConfig()
	if err == nil {
		t.Fatal("Expected error when M365_TENANT_ID missing")
	}

	expected := "M365_TENANT_ID is required when EMAIL_PROVIDER=m365"
	if err.Error() != expected {
		t.Errorf("Expected error %q, got: %q", expected, err.Error())
	}
}

func TestLoadEmailConfig_M365MissingClientID(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "m365")
	os.Setenv("M365_TENANT_ID", "tenant-123")
	os.Setenv("M365_CLIENT_SECRET", "secret-789")
	os.Setenv("M365_SENDER", "alerts@example.com")

	defer func() {
		os.Unsetenv("EMAIL_PROVIDER")
		os.Unsetenv("M365_TENANT_ID")
		os.Unsetenv("M365_CLIENT_SECRET")
		os.Unsetenv("M365_SENDER")
	}()

	_, err := LoadEmailConfig()
	if err == nil {
		t.Fatal("Expected error when M365_CLIENT_ID missing")
	}

	expected := "M365_CLIENT_ID is required when EMAIL_PROVIDER=m365"
	if err.Error() != expected {
		t.Errorf("Expected error %q, got: %q", expected, err.Error())
	}
}

func TestLoadEmailConfig_M365MissingSecret(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "m365")
	os.Setenv("M365_TENANT_ID", "tenant-123")
	os.Setenv("M365_CLIENT_ID", "client-456")
	os.Setenv("M365_SENDER", "alerts@example.com")

	defer func() {
		os.Unsetenv("EMAIL_PROVIDER")
		os.Unsetenv("M365_TENANT_ID")
		os.Unsetenv("M365_CLIENT_ID")
		os.Unsetenv("M365_SENDER")
	}()

	_, err := LoadEmailConfig()
	if err == nil {
		t.Fatal("Expected error when M365_CLIENT_SECRET missing")
	}

	expected := "M365_CLIENT_SECRET is required when EMAIL_PROVIDER=m365"
	if err.Error() != expected {
		t.Errorf("Expected error %q, got: %q", expected, err.Error())
	}
}

func TestLoadEmailConfig_M365MissingSender(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "m365")
	os.Setenv("M365_TENANT_ID", "tenant-123")
	os.Setenv("M365_CLIENT_ID", "client-456")
	os.Setenv("M365_CLIENT_SECRET", "secret-789")

	defer func() {
		os.Unsetenv("EMAIL_PROVIDER")
		os.Unsetenv("M365_TENANT_ID")
		os.Unsetenv("M365_CLIENT_ID")
		os.Unsetenv("M365_CLIENT_SECRET")
	}()

	_, err := LoadEmailConfig()
	if err == nil {
		t.Fatal("Expected error when M365_SENDER missing")
	}

	expected := "M365_SENDER is required when EMAIL_PROVIDER=m365"
	if err.Error() != expected {
		t.Errorf("Expected error %q, got: %q", expected, err.Error())
	}
}

func TestLoadEmailConfig_UnsupportedProvider(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "sendgrid")
	defer os.Unsetenv("EMAIL_PROVIDER")

	_, err := LoadEmailConfig()
	if err == nil {
		t.Fatal("Expected error with unsupported provider")
	}

	expected := "unsupported EMAIL_PROVIDER: sendgrid"
	if err.Error() != expected {
		t.Errorf("Expected error %q, got: %q", expected, err.Error())
	}
}
