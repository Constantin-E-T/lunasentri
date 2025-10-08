package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

func TestHandleTestWebhook_RateLimited(t *testing.T) {
	store := newMockHTTPStore()

	// Create a webhook
	webhook := storage.Webhook{
		ID:       1,
		UserID:   1,
		URL:      "https://example.com/webhook",
		IsActive: true,
	}
	store.webhooks[1] = []storage.Webhook{webhook}

	// Create notifier that returns rate limit error
	rateLimitErr := &RateLimitError{
		Type:    "cooldown",
		Message: "Webhook in cooldown until 2025-01-01T15:00:00Z",
	}
	notifier := &mockNotifier{
		sendTestError: rateLimitErr,
	}

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/notifications/webhooks/1/test", nil)
	ctx := context.WithValue(req.Context(), auth.UserContextKey, &storage.User{ID: 1})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler := HandleTestWebhook(notifier, store)
	handler(rec, req)

	// Check response
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Check response body contains rate limit message
	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	expectedError := "Webhook in cooldown until 2025-01-01T15:00:00Z"
	if response["error"] != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, response["error"])
	}

	// Verify notifier was called
	if len(notifier.sendTestCalls) != 1 {
		t.Errorf("Expected 1 SendTest call, got %d", len(notifier.sendTestCalls))
	}
}

func TestHandleTestWebhook_RateLimitWithInterval(t *testing.T) {
	store := newMockHTTPStore()

	// Create a webhook
	webhook := storage.Webhook{
		ID:       1,
		UserID:   1,
		URL:      "https://example.com/webhook",
		IsActive: true,
	}
	store.webhooks[1] = []storage.Webhook{webhook}

	// Create notifier that returns rate limit error for minimum interval
	rateLimitErr := &RateLimitError{
		Type:    "rate_limit",
		Message: "Rate limit active, can retry in 25s",
	}
	notifier := &mockNotifier{
		sendTestError: rateLimitErr,
	}

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/notifications/webhooks/1/test", nil)
	ctx := context.WithValue(req.Context(), auth.UserContextKey, &storage.User{ID: 1})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler := HandleTestWebhook(notifier, store)
	handler(rec, req)

	// Check response
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Check response body contains rate limit message
	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	expectedError := "Rate limit active, can retry in 25s"
	if response["error"] != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, response["error"])
	}
}
