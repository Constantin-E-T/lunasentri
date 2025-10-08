package config

import (
	"errors"
	"fmt"
	"os"
)

// EmailProvider represents the type of email provider
type EmailProvider string

const (
	EmailProviderM365 EmailProvider = "m365"
	EmailProviderNone EmailProvider = ""
)

// EmailConfig holds email notification configuration
type EmailConfig struct {
	Provider     EmailProvider
	M365TenantID string
	M365ClientID string
	M365Secret   string
	M365Sender   string
}

// LoadEmailConfig loads email configuration from environment variables
func LoadEmailConfig() (*EmailConfig, error) {
	provider := EmailProvider(os.Getenv("EMAIL_PROVIDER"))

	cfg := &EmailConfig{
		Provider: provider,
	}

	// If no provider set, email is disabled
	if provider == EmailProviderNone {
		return cfg, nil
	}

	// Validate M365 configuration
	if provider == EmailProviderM365 {
		cfg.M365TenantID = os.Getenv("M365_TENANT_ID")
		cfg.M365ClientID = os.Getenv("M365_CLIENT_ID")
		cfg.M365Secret = os.Getenv("M365_CLIENT_SECRET")
		cfg.M365Sender = os.Getenv("M365_SENDER")

		if cfg.M365TenantID == "" {
			return nil, errors.New("M365_TENANT_ID is required when EMAIL_PROVIDER=m365")
		}
		if cfg.M365ClientID == "" {
			return nil, errors.New("M365_CLIENT_ID is required when EMAIL_PROVIDER=m365")
		}
		if cfg.M365Secret == "" {
			return nil, errors.New("M365_CLIENT_SECRET is required when EMAIL_PROVIDER=m365")
		}
		if cfg.M365Sender == "" {
			return nil, errors.New("M365_SENDER is required when EMAIL_PROVIDER=m365")
		}

		return cfg, nil
	}

	return nil, fmt.Errorf("unsupported EMAIL_PROVIDER: %s", provider)
}

// IsEnabled returns true if email notifications are enabled
func (c *EmailConfig) IsEnabled() bool {
	return c.Provider != EmailProviderNone
}
