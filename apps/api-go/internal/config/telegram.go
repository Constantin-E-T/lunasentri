package config

import (
	"fmt"
	"os"
)

// TelegramConfig holds configuration for Telegram bot notifications
type TelegramConfig struct {
	BotToken string
}

// LoadTelegramConfig loads Telegram configuration from environment variables
func LoadTelegramConfig() (*TelegramConfig, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	if botToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	return &TelegramConfig{
		BotToken: botToken,
	}, nil
}

// IsEnabled returns true if Telegram notifications are configured
func (c *TelegramConfig) IsEnabled() bool {
	return c != nil && c.BotToken != ""
}
