package notifications

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

func TestNotifier_checkDeliveryPreconditions(t *testing.T) {
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	notifier := NewNotifier(store, logger)

	// Test webhook with no rate limiting constraints
	webhook := storage.Webhook{
		ID:            1,
		UserID:        1,
		URL:           "https://example.com/webhook",
		SecretHash:    "hashedsecret",
		IsActive:      true,
		FailureCount:  0,
		LastAttemptAt: nil,
		CooldownUntil: nil,
	}

	err = notifier.checkDeliveryPreconditions(webhook)
	if err != nil {
		t.Errorf("Expected no error for webhook with no constraints, got: %v", err)
	}

	// Test webhook in cooldown
	futureTime := time.Now().Add(15 * time.Minute)
	webhook.CooldownUntil = &futureTime

	err = notifier.checkDeliveryPreconditions(webhook)
	if err == nil {
		t.Error("Expected error for webhook in cooldown")
	}

	var rateLimitErr *RateLimitError
	if !isRateLimitError(err, &rateLimitErr) {
		t.Errorf("Expected RateLimitError, got: %T", err)
	} else if rateLimitErr.Type != "cooldown" {
		t.Errorf("Expected cooldown error type, got: %s", rateLimitErr.Type)
	}

	// Test webhook with past cooldown (should be allowed)
	pastTime := time.Now().Add(-1 * time.Minute)
	webhook.CooldownUntil = &pastTime

	err = notifier.checkDeliveryPreconditions(webhook)
	if err != nil {
		t.Errorf("Expected no error for webhook with past cooldown, got: %v", err)
	}

	// Test webhook with recent attempt (rate limited)
	webhook.CooldownUntil = nil
	recentTime := time.Now().Add(-10 * time.Second) // Less than MinAttemptInterval (30s)
	webhook.LastAttemptAt = &recentTime

	err = notifier.checkDeliveryPreconditions(webhook)
	if err == nil {
		t.Error("Expected error for rate limited webhook")
	}

	if !isRateLimitError(err, &rateLimitErr) {
		t.Errorf("Expected RateLimitError, got: %T", err)
	} else if rateLimitErr.Type != "rate_limit" {
		t.Errorf("Expected rate_limit error type, got: %s", rateLimitErr.Type)
	}

	// Test webhook with old attempt (should be allowed)
	oldTime := time.Now().Add(-60 * time.Second) // More than MinAttemptInterval (30s)
	webhook.LastAttemptAt = &oldTime

	err = notifier.checkDeliveryPreconditions(webhook)
	if err != nil {
		t.Errorf("Expected no error for webhook with old attempt, got: %v", err)
	}
}

func TestNotifier_shouldEnterCooldown(t *testing.T) {
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	notifier := NewNotifier(store, logger)

	// Test webhook with low failure count
	webhook := storage.Webhook{
		FailureCount: FailureThreshold - 1, // 2 failures (threshold is 3)
		LastErrorAt:  nil,
	}

	if notifier.shouldEnterCooldown(webhook) {
		t.Error("Expected no cooldown for webhook below failure threshold")
	}

	// Test webhook meeting failure threshold with recent errors
	recentError := time.Now().Add(-5 * time.Minute) // Within FailureWindow (10 minutes)
	webhook.FailureCount = FailureThreshold
	webhook.LastErrorAt = &recentError

	if !notifier.shouldEnterCooldown(webhook) {
		t.Error("Expected cooldown for webhook meeting failure threshold with recent errors")
	}

	// Test webhook meeting failure threshold with old errors
	oldError := time.Now().Add(-15 * time.Minute) // Outside FailureWindow (10 minutes)
	webhook.LastErrorAt = &oldError

	if notifier.shouldEnterCooldown(webhook) {
		t.Error("Expected no cooldown for webhook with old errors outside failure window")
	}

	// Test webhook exceeding failure threshold with no LastErrorAt (safety measure)
	webhook.FailureCount = FailureThreshold + 1
	webhook.LastErrorAt = nil

	if !notifier.shouldEnterCooldown(webhook) {
		t.Error("Expected cooldown for webhook exceeding threshold with no LastErrorAt (safety)")
	}
}

func TestRateLimitError(t *testing.T) {
	now := time.Now()
	retryTime := now.Add(30 * time.Second)

	err := &RateLimitError{
		Type:    "rate_limit",
		Message: "Rate limit active",
		RetryAt: &retryTime,
	}

	if err.Error() != "Rate limit active" {
		t.Errorf("Expected error message 'Rate limit active', got: %s", err.Error())
	}

	if err.Type != "rate_limit" {
		t.Errorf("Expected type 'rate_limit', got: %s", err.Type)
	}

	if err.RetryAt == nil || !err.RetryAt.Equal(retryTime) {
		t.Errorf("Expected RetryAt %v, got %v", retryTime, err.RetryAt)
	}
}

// Helper function to check if an error is a RateLimitError
func isRateLimitError(err error, target **RateLimitError) bool {
	if rateLimitErr, ok := err.(*RateLimitError); ok {
		*target = rateLimitErr
		return true
	}
	return false
}
