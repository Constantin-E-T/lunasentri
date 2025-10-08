package storage

import (
	"context"
	"testing"
	"time"
)

// Webhook tests split out from sqlite_test.go to keep suites focused.

func TestSQLiteStore_CreateWebhook(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test user first
	user, err := store.CreateUser(ctx, "test@example.com", "password_hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test data
	url := "https://example.com/webhook"
	secretHash := HashSecret("my-secret")

	// Test creating a webhook
	webhook, err := store.CreateWebhook(ctx, user.ID, url, secretHash)
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	// Verify webhook data
	if webhook.ID == 0 {
		t.Error("Expected webhook ID to be set")
	}
	if webhook.UserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, webhook.UserID)
	}
	if webhook.URL != url {
		t.Errorf("Expected URL %s, got %s", url, webhook.URL)
	}
	if webhook.SecretHash != secretHash {
		t.Errorf("Expected secret hash %s, got %s", secretHash, webhook.SecretHash)
	}
	if !webhook.IsActive {
		t.Error("Expected webhook to be active by default")
	}
	if webhook.FailureCount != 0 {
		t.Errorf("Expected failure count 0, got %d", webhook.FailureCount)
	}
	if webhook.CreatedAt.IsZero() {
		t.Error("Expected created_at to be set")
	}
	if webhook.UpdatedAt.IsZero() {
		t.Error("Expected updated_at to be set")
	}
}

func TestSQLiteStore_CreateWebhook_UniqueConstraint(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test user
	user, err := store.CreateUser(ctx, "test@example.com", "password_hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	url := "https://example.com/webhook"
	secretHash := HashSecret("my-secret")

	// Create first webhook
	_, err = store.CreateWebhook(ctx, user.ID, url, secretHash)
	if err != nil {
		t.Fatalf("Failed to create first webhook: %v", err)
	}

	// Try to create duplicate webhook (same user_id, url)
	_, err = store.CreateWebhook(ctx, user.ID, url, secretHash)
	if err == nil {
		t.Error("Expected error when creating duplicate webhook")
	}
}

func TestSQLiteStore_ListWebhooks(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create test users
	user1, err := store.CreateUser(ctx, "user1@example.com", "password_hash")
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2, err := store.CreateUser(ctx, "user2@example.com", "password_hash")
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create webhooks for user1
	_, err = store.CreateWebhook(ctx, user1.ID, "https://example.com/webhook1", HashSecret("secret1"))
	if err != nil {
		t.Fatalf("Failed to create webhook1: %v", err)
	}

	_, err = store.CreateWebhook(ctx, user1.ID, "https://example.com/webhook2", HashSecret("secret2"))
	if err != nil {
		t.Fatalf("Failed to create webhook2: %v", err)
	}

	// Create webhook for user2
	_, err = store.CreateWebhook(ctx, user2.ID, "https://example.com/webhook3", HashSecret("secret3"))
	if err != nil {
		t.Fatalf("Failed to create webhook3: %v", err)
	}

	// Fetch webhooks for user1
	webhooks, err := store.ListWebhooks(ctx, user1.ID)
	if err != nil {
		t.Fatalf("Failed to list webhooks for user1: %v", err)
	}
	if len(webhooks) != 2 {
		t.Errorf("Expected 2 webhooks for user1, got %d", len(webhooks))
	}

	// Fetch webhooks for user2
	webhooks, err = store.ListWebhooks(ctx, user2.ID)
	if err != nil {
		t.Fatalf("Failed to list webhooks for user2: %v", err)
	}
	if len(webhooks) != 1 {
		t.Errorf("Expected 1 webhook for user2, got %d", len(webhooks))
	}
}

func TestSQLiteStore_GetWebhook(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create test users
	user1, err := store.CreateUser(ctx, "user1@example.com", "password_hash")
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2, err := store.CreateUser(ctx, "user2@example.com", "password_hash")
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create a webhook for user1
	secret := "my-secret"
	webhook, err := store.CreateWebhook(ctx, user1.ID, "https://example.com/webhook", HashSecret(secret))
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	// Fetch the webhook with the correct user ID
	fetchedWebhook, err := store.GetWebhook(ctx, webhook.ID, user1.ID)
	if err != nil {
		t.Fatalf("Failed to fetch webhook: %v", err)
	}
	if fetchedWebhook.ID != webhook.ID {
		t.Errorf("Expected webhook ID %d, got %d", webhook.ID, fetchedWebhook.ID)
	}
	if fetchedWebhook.UserID != user1.ID {
		t.Errorf("Expected user ID %d, got %d", user1.ID, fetchedWebhook.UserID)
	}

	// Try to fetch the webhook with a different user ID
	_, err = store.GetWebhook(ctx, webhook.ID, user2.ID)
	if err == nil {
		t.Error("Expected error when fetching webhook with incorrect user ID")
	}

	// Try to fetch a non-existent webhook
	_, err = store.GetWebhook(ctx, 99999, user1.ID)
	if err == nil {
		t.Error("Expected error when fetching non-existent webhook")
	}
}

func TestSQLiteStore_UpdateWebhook(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test user
	user, err := store.CreateUser(ctx, "test@example.com", "password_hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a webhook
	webhook, err := store.CreateWebhook(ctx, user.ID, "https://example.com/old", HashSecret("old-secret"))
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	// Update URL
	newURL := "https://example.com/new"
	updatedWebhook, err := store.UpdateWebhook(ctx, webhook.ID, user.ID, newURL, nil, nil)
	if err != nil {
		t.Fatalf("Failed to update webhook: %v", err)
	}
	if updatedWebhook.URL != newURL {
		t.Errorf("Expected URL %s, got %s", newURL, updatedWebhook.URL)
	}

	// Update secret hash
	newSecretHash := HashSecret("new-secret")
	updatedWebhook, err = store.UpdateWebhook(ctx, webhook.ID, user.ID, "", &newSecretHash, nil)
	if err != nil {
		t.Fatalf("Failed to update webhook secret: %v", err)
	}
	if updatedWebhook.SecretHash != newSecretHash {
		t.Errorf("Expected secret hash %s, got %s", newSecretHash, updatedWebhook.SecretHash)
	}

	// Update is_active flag
	isActive := false
	updatedWebhook, err = store.UpdateWebhook(ctx, webhook.ID, user.ID, "", nil, &isActive)
	if err != nil {
		t.Fatalf("Failed to update webhook is_active: %v", err)
	}
	if updatedWebhook.IsActive != isActive {
		t.Errorf("Expected is_active %v, got %v", isActive, updatedWebhook.IsActive)
	}

	// Try to update non-existent webhook
	_, err = store.UpdateWebhook(ctx, 99999, user.ID, "https://example.com/not-found", nil, nil)
	if err == nil {
		t.Error("Expected error when updating non-existent webhook")
	}
}

func TestSQLiteStore_DeleteWebhook(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test user
	user, err := store.CreateUser(ctx, "test@example.com", "password_hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a webhook
	webhook, err := store.CreateWebhook(ctx, user.ID, "https://example.com/webhook", HashSecret("secret"))
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	// Delete the webhook
	err = store.DeleteWebhook(ctx, webhook.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to delete webhook: %v", err)
	}

	// Verify deletion
	webhooks, err := store.ListWebhooks(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to list webhooks: %v", err)
	}
	if len(webhooks) != 0 {
		t.Errorf("Expected 0 webhooks after deletion, got %d", len(webhooks))
	}

	// Try deleting non-existent webhook
	err = store.DeleteWebhook(ctx, 99999, user.ID)
	if err == nil {
		t.Error("Expected error when deleting non-existent webhook")
	}
}

func TestSQLiteStore_IncrementWebhookFailure(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test user
	user, err := store.CreateUser(ctx, "user@example.com", "password_hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a webhook
	webhook, err := store.CreateWebhook(ctx, user.ID, "https://example.com/webhook", HashSecret("secret"))
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	// Increment failure count
	errorTime := time.Now()
	err = store.IncrementWebhookFailure(ctx, webhook.ID, errorTime)
	if err != nil {
		t.Fatalf("Failed to increment failure count: %v", err)
	}

	// Verify failure count and last error time
	updatedWebhook, err := store.GetWebhook(ctx, webhook.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get webhook: %v", err)
	}

	if updatedWebhook.FailureCount != 1 {
		t.Errorf("Expected failure count 1, got %d", updatedWebhook.FailureCount)
	}
	if updatedWebhook.LastErrorAt == nil {
		t.Error("Expected LastErrorAt to be set")
	}

	// Increment failure count again
	err = store.IncrementWebhookFailure(ctx, webhook.ID, time.Now())
	if err != nil {
		t.Fatalf("Failed to increment failure count: %v", err)
	}

	// Verify failure count increased
	updatedWebhook, err = store.GetWebhook(ctx, webhook.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get webhook: %v", err)
	}

	if updatedWebhook.FailureCount != 2 {
		t.Errorf("Expected failure count 2, got %d", updatedWebhook.FailureCount)
	}
}

func TestSQLiteStore_MarkWebhookSuccess(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test user
	user, err := store.CreateUser(ctx, "user@example.com", "password_hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a webhook
	webhook, err := store.CreateWebhook(ctx, user.ID, "https://example.com/webhook", HashSecret("secret"))
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	// Increment failure count to simulate previous failures
	err = store.IncrementWebhookFailure(ctx, webhook.ID, time.Now())
	if err != nil {
		t.Fatalf("Failed to increment failure count: %v", err)
	}

	// Mark webhook success
	successTime := time.Now().Add(1 * time.Hour)
	err = store.MarkWebhookSuccess(ctx, webhook.ID, successTime)
	if err != nil {
		t.Fatalf("Failed to mark webhook success: %v", err)
	}

	// Verify failure count reset and last success time set
	updatedWebhook, err := store.GetWebhook(ctx, webhook.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get webhook: %v", err)
	}

	if updatedWebhook.FailureCount != 0 {
		t.Errorf("Expected failure count 0 after success, got %d", updatedWebhook.FailureCount)
	}
	if updatedWebhook.LastSuccessAt == nil {
		t.Error("Expected LastSuccessAt to be set")
	}
}

func TestSQLiteStore_WebhookCascadeDelete(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test user
	user, err := store.CreateUser(ctx, "test@example.com", "password_hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a webhook
	_, err = store.CreateWebhook(ctx, user.ID, "https://example.com/webhook", HashSecret("secret"))
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	// Verify webhook exists
	webhooks, err := store.ListWebhooks(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to list webhooks: %v", err)
	}
	if len(webhooks) != 1 {
		t.Errorf("Expected 1 webhook before user deletion, got %d", len(webhooks))
	}

	// Delete the user
	err = store.DeleteUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify webhooks were cascade deleted
	webhooks, err = store.ListWebhooks(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to list webhooks after user deletion: %v", err)
	}
	if len(webhooks) != 0 {
		t.Errorf("Expected 0 webhooks after user deletion, got %d", len(webhooks))
	}
}

func TestSQLiteStore_UpdateWebhookDeliveryState(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a test user
	user, err := store.CreateUser(ctx, "test@example.com", "password_hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a webhook
	webhook, err := store.CreateWebhook(ctx, user.ID, "https://example.com/webhook", HashSecret("secret123"))
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}

	// Verify initial state
	if webhook.LastAttemptAt != nil {
		t.Error("Expected LastAttemptAt to be nil initially")
	}
	if webhook.CooldownUntil != nil {
		t.Error("Expected CooldownUntil to be nil initially")
	}

	// Test updating with last attempt time only
	lastAttemptTime := time.Now()
	err = store.UpdateWebhookDeliveryState(ctx, webhook.ID, lastAttemptTime, nil)
	if err != nil {
		t.Fatalf("Failed to update webhook delivery state: %v", err)
	}

	// Verify last attempt time was set
	updatedWebhook, err := store.GetWebhook(ctx, webhook.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated webhook: %v", err)
	}

	if updatedWebhook.LastAttemptAt == nil {
		t.Error("Expected LastAttemptAt to be set")
	} else if !updatedWebhook.LastAttemptAt.Truncate(time.Second).Equal(lastAttemptTime.Truncate(time.Second)) {
		t.Errorf("Expected LastAttemptAt %v, got %v", lastAttemptTime.Truncate(time.Second), updatedWebhook.LastAttemptAt.Truncate(time.Second))
	}

	if updatedWebhook.CooldownUntil != nil {
		t.Error("Expected CooldownUntil to remain nil")
	}

	// Test updating with cooldown time
	cooldownTime := time.Now().Add(15 * time.Minute)
	newAttemptTime := time.Now().Add(1 * time.Minute)
	err = store.UpdateWebhookDeliveryState(ctx, webhook.ID, newAttemptTime, &cooldownTime)
	if err != nil {
		t.Fatalf("Failed to update webhook delivery state with cooldown: %v", err)
	}

	// Verify cooldown time was set
	updatedWebhook, err = store.GetWebhook(ctx, webhook.ID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated webhook: %v", err)
	}

	if updatedWebhook.CooldownUntil == nil {
		t.Error("Expected CooldownUntil to be set")
	} else if !updatedWebhook.CooldownUntil.Truncate(time.Second).Equal(cooldownTime.Truncate(time.Second)) {
		t.Errorf("Expected CooldownUntil %v, got %v", cooldownTime.Truncate(time.Second), updatedWebhook.CooldownUntil.Truncate(time.Second))
	}

	// Test updating non-existent webhook
	err = store.UpdateWebhookDeliveryState(ctx, 99999, time.Now(), nil)
	if err == nil {
		t.Error("Expected error when updating non-existent webhook")
	}
	if err.Error() != "webhook with id 99999 not found" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}
