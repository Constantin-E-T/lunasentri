package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// mockStore for HTTP handler tests
type mockHTTPStore struct {
	webhooks      map[int][]storage.Webhook // userID -> webhooks
	nextWebhookID int
	failCreate    bool
	failList      bool
	failUpdate    bool
	failDelete    bool
}

func newMockHTTPStore() *mockHTTPStore {
	return &mockHTTPStore{
		webhooks:      make(map[int][]storage.Webhook),
		nextWebhookID: 1,
	}
}

func (m *mockHTTPStore) ListWebhooks(ctx context.Context, userID int) ([]storage.Webhook, error) {
	if m.failList {
		return nil, fmt.Errorf("mock list failure")
	}
	return m.webhooks[userID], nil
}

func (m *mockHTTPStore) GetWebhook(ctx context.Context, id int, userID int) (*storage.Webhook, error) {
	for _, webhook := range m.webhooks[userID] {
		if webhook.ID == id {
			return &webhook, nil
		}
	}
	return nil, fmt.Errorf("webhook with id %d not found for user %d", id, userID)
}

func (m *mockHTTPStore) CreateWebhook(ctx context.Context, userID int, url, secretHash string) (*storage.Webhook, error) {
	if m.failCreate {
		return nil, fmt.Errorf("mock create failure")
	}

	// Check for duplicate URL
	for _, webhook := range m.webhooks[userID] {
		if webhook.URL == url {
			return nil, fmt.Errorf("UNIQUE constraint failed")
		}
	}

	webhook := &storage.Webhook{
		ID:         m.nextWebhookID,
		UserID:     userID,
		URL:        url,
		SecretHash: secretHash,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	m.nextWebhookID++

	if m.webhooks[userID] == nil {
		m.webhooks[userID] = make([]storage.Webhook, 0)
	}
	m.webhooks[userID] = append(m.webhooks[userID], *webhook)

	return webhook, nil
}

func (m *mockHTTPStore) UpdateWebhook(ctx context.Context, id int, userID int, url string, secretHash *string, isActive *bool) (*storage.Webhook, error) {
	if m.failUpdate {
		return nil, fmt.Errorf("mock update failure")
	}

	// Find webhook
	for i, webhook := range m.webhooks[userID] {
		if webhook.ID == id {
			// Check for duplicate URL if changing URL
			if url != "" && url != webhook.URL {
				for _, otherWebhook := range m.webhooks[userID] {
					if otherWebhook.URL == url && otherWebhook.ID != id {
						return nil, fmt.Errorf("UNIQUE constraint failed")
					}
				}
				webhook.URL = url
			}

			if secretHash != nil {
				webhook.SecretHash = *secretHash
			}

			if isActive != nil {
				webhook.IsActive = *isActive
			}

			webhook.UpdatedAt = time.Now()
			m.webhooks[userID][i] = webhook

			return &webhook, nil
		}
	}

	return nil, fmt.Errorf("webhook with id %d not found for user %d", id, userID)
}

func (m *mockHTTPStore) DeleteWebhook(ctx context.Context, id int, userID int) error {
	if m.failDelete {
		return fmt.Errorf("mock delete failure")
	}

	// Find and remove webhook
	for i, webhook := range m.webhooks[userID] {
		if webhook.ID == id {
			m.webhooks[userID] = append(m.webhooks[userID][:i], m.webhooks[userID][i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("webhook with id %d not found for user %d", id, userID)
}

// Mock implementations of other storage methods (not used in these tests)
func (m *mockHTTPStore) CreateUser(ctx context.Context, email, passwordHash string) (*storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) GetUserByEmail(ctx context.Context, email string) (*storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) GetUserByID(ctx context.Context, id int) (*storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) UpdateUserPassword(ctx context.Context, userID int, passwordHash string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) UpsertAdmin(ctx context.Context, email, passwordHash string) (*storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) CreatePasswordReset(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) (*storage.PasswordReset, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) GetPasswordResetByHash(ctx context.Context, tokenHash string) (*storage.PasswordReset, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) MarkPasswordResetUsed(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) ListUsers(ctx context.Context) ([]storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) DeleteUser(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) CountUsers(ctx context.Context) (int, error) {
	return 0, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) PromoteToAdmin(ctx context.Context, userID int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) DeletePasswordResetsForUser(ctx context.Context, userID int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) CreateAlertRule(ctx context.Context, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*storage.AlertRule, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) ListAlertRules(ctx context.Context) ([]storage.AlertRule, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) GetAlertRule(ctx context.Context, id int) (*storage.AlertRule, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) UpdateAlertRule(ctx context.Context, id int, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*storage.AlertRule, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) DeleteAlertRule(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) CreateAlertEvent(ctx context.Context, ruleID int, value float64) (*storage.AlertEvent, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) ListAlertEvents(ctx context.Context, limit int) ([]storage.AlertEvent, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) AckAlertEvent(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) ListActiveEvents(ctx context.Context, limit int) ([]storage.AlertEvent, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) IncrementWebhookFailure(ctx context.Context, id int, lastErrorAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) MarkWebhookSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) UpdateWebhookDeliveryState(ctx context.Context, id int, lastAttemptAt time.Time, cooldownUntil *time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) ListEmailRecipients(ctx context.Context, userID int) ([]storage.EmailRecipient, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) GetEmailRecipient(ctx context.Context, id int, userID int) (*storage.EmailRecipient, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) CreateEmailRecipient(ctx context.Context, userID int, email string) (*storage.EmailRecipient, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) UpdateEmailRecipient(ctx context.Context, id int, userID int, email string, isActive *bool) (*storage.EmailRecipient, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) DeleteEmailRecipient(ctx context.Context, id int, userID int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) IncrementEmailFailure(ctx context.Context, id int, lastErrorAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) MarkEmailSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) UpdateEmailDeliveryState(ctx context.Context, id int, lastAttemptAt time.Time, cooldownUntil *time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) ListTelegramRecipients(ctx context.Context, userID int) ([]storage.TelegramRecipient, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) GetTelegramRecipient(ctx context.Context, id int, userID int) (*storage.TelegramRecipient, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) CreateTelegramRecipient(ctx context.Context, userID int, chatID string) (*storage.TelegramRecipient, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) UpdateTelegramRecipient(ctx context.Context, id int, userID int, chatID string, isActive *bool) (*storage.TelegramRecipient, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) DeleteTelegramRecipient(ctx context.Context, id int, userID int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) IncrementTelegramFailure(ctx context.Context, id int, lastErrorAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) MarkTelegramSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) UpdateTelegramDeliveryState(ctx context.Context, id int, lastAttemptAt time.Time, cooldownUntil *time.Time) error {
	return fmt.Errorf("not implemented")
}

// Machine methods (not implemented for these tests)
func (m *mockHTTPStore) CreateMachine(ctx context.Context, userID int, name, hostname, description, apiKeyHash string) (*storage.Machine, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) GetMachineByID(ctx context.Context, id int) (*storage.Machine, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) GetMachineByAPIKey(ctx context.Context, apiKeyHash string) (*storage.Machine, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) ListMachines(ctx context.Context, userID int) ([]storage.Machine, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) UpdateMachineStatus(ctx context.Context, id int, status string, lastSeen time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) UpdateMachineLastSeen(ctx context.Context, id int, lastSeen time.Time) error {
	return nil
}
func (m *mockHTTPStore) UpdateMachineSystemInfo(ctx context.Context, id int, info storage.MachineSystemInfoUpdate) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) UpdateMachineDetails(ctx context.Context, id int, updates map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) DeleteMachine(ctx context.Context, id int, userID int) error {
	return fmt.Errorf("not implemented")
}

// Metrics history methods (not implemented for these tests)
func (m *mockHTTPStore) InsertMetrics(ctx context.Context, machineID int, cpuPct, memUsedPct, diskUsedPct float64, netRxBytes, netTxBytes int64, uptimeSeconds *float64, timestamp time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) GetLatestMetrics(ctx context.Context, machineID int) (*storage.MetricsHistory, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockHTTPStore) GetMetricsHistory(ctx context.Context, machineID int, from, to time.Time, limit int) ([]storage.MetricsHistory, error) {
	return nil, fmt.Errorf("not implemented")
}

// Heartbeat monitoring methods (stub implementations for testing)
func (m *mockHTTPStore) ListAllMachines(ctx context.Context) ([]storage.Machine, error) {
	return []storage.Machine{}, nil
}

func (m *mockHTTPStore) RecordMachineOfflineNotification(ctx context.Context, machineID int, notifiedAt time.Time) error {
	return nil
}

func (m *mockHTTPStore) GetMachineLastOfflineNotification(ctx context.Context, machineID int) (time.Time, error) {
	return time.Time{}, nil
}

func (m *mockHTTPStore) ClearMachineOfflineNotification(ctx context.Context, machineID int) error {
	return nil
}

// Machine credential management methods (not implemented for these tests)
func (m *mockHTTPStore) SetMachineEnabled(ctx context.Context, machineID int, enabled bool) error {
	return fmt.Errorf("not implemented")
}

func (m *mockHTTPStore) CreateMachineAPIKey(ctx context.Context, machineID int, apiKeyHash string) (*storage.MachineAPIKey, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockHTTPStore) RevokeMachineAPIKey(ctx context.Context, keyID int) error {
	return fmt.Errorf("not implemented")
}

func (m *mockHTTPStore) RevokeAllMachineAPIKeys(ctx context.Context, machineID int) error {
	return fmt.Errorf("not implemented")
}

func (m *mockHTTPStore) GetActiveAPIKeyForMachine(ctx context.Context, machineID int) (*storage.MachineAPIKey, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockHTTPStore) GetMachineAPIKeyByHash(ctx context.Context, apiKeyHash string) (*storage.MachineAPIKey, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockHTTPStore) ListMachineAPIKeys(ctx context.Context, machineID int) ([]storage.MachineAPIKey, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockHTTPStore) Close() error {
	return nil
}

// Helper to create a request with authenticated user context
func createAuthenticatedRequest(method, url string, body []byte, userID int) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	// Mock user in context using the same key as the auth middleware
	user := &storage.User{ID: userID, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
	return req.WithContext(ctx)
}

func TestHandleListWebhooks_Success(t *testing.T) {
	store := newMockHTTPStore()
	handler := HandleListWebhooks(store)

	// Add some test webhooks
	store.CreateWebhook(context.Background(), 1, "https://example.com/webhook1", "hash1")
	store.CreateWebhook(context.Background(), 1, "https://example.com/webhook2", "hash2")

	req := createAuthenticatedRequest("GET", "/notifications/webhooks", nil, 1)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []WebhookResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 webhooks, got %d", len(response))
	}

	// Check that secret hash is hidden
	for _, webhook := range response {
		if webhook.SecretLastFour != "****" {
			t.Errorf("Expected secret_last_four to be '****', got '%s'", webhook.SecretLastFour)
		}
	}
}

func TestHandleListWebhooks_WrongMethod(t *testing.T) {
	store := newMockHTTPStore()
	handler := HandleListWebhooks(store)

	req := createAuthenticatedRequest("POST", "/notifications/webhooks", nil, 1)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleCreateWebhook_Success(t *testing.T) {
	store := newMockHTTPStore()
	handler := HandleCreateWebhook(store)

	requestBody := WebhookRequest{
		URL:    "https://example.com/webhook",
		Secret: "mysecretkey12345",
	}
	body, _ := json.Marshal(requestBody)

	req := createAuthenticatedRequest("POST", "/notifications/webhooks", body, 1)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response WebhookResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.URL != requestBody.URL {
		t.Errorf("Expected URL %s, got %s", requestBody.URL, response.URL)
	}

	if response.SecretLastFour != "2345" {
		t.Errorf("Expected secret_last_four '2345', got '%s'", response.SecretLastFour)
	}

	if !response.IsActive {
		t.Error("Expected webhook to be active by default")
	}
}

func TestHandleCreateWebhook_ValidationErrors(t *testing.T) {
	store := newMockHTTPStore()
	handler := HandleCreateWebhook(store)

	testCases := []struct {
		name           string
		requestBody    WebhookRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing URL",
			requestBody:    WebhookRequest{Secret: "mysecretkey12345"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "url is required",
		},
		{
			name:           "non-HTTPS URL",
			requestBody:    WebhookRequest{URL: "http://example.com/webhook", Secret: "mysecretkey12345"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "url must use HTTPS",
		},
		{
			name:           "missing secret",
			requestBody:    WebhookRequest{URL: "https://example.com/webhook"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "secret is required",
		},
		{
			name:           "secret too short",
			requestBody:    WebhookRequest{URL: "https://example.com/webhook", Secret: "short"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "secret must be between 16 and 128 characters",
		},
		{
			name:           "secret too long",
			requestBody:    WebhookRequest{URL: "https://example.com/webhook", Secret: strings.Repeat("a", 129)},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "secret must be between 16 and 128 characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.requestBody)
			req := createAuthenticatedRequest("POST", "/notifications/webhooks", body, 1)
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			var errorResponse map[string]string
			if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			if !strings.Contains(errorResponse["error"], tc.expectedError) {
				t.Errorf("Expected error to contain '%s', got '%s'", tc.expectedError, errorResponse["error"])
			}
		})
	}
}

func TestHandleCreateWebhook_DuplicateURL(t *testing.T) {
	store := newMockHTTPStore()
	handler := HandleCreateWebhook(store)

	// Create first webhook
	store.CreateWebhook(context.Background(), 1, "https://example.com/webhook", "hash")

	// Try to create duplicate
	requestBody := WebhookRequest{
		URL:    "https://example.com/webhook",
		Secret: "mysecretkey12345",
	}
	body, _ := json.Marshal(requestBody)

	req := createAuthenticatedRequest("POST", "/notifications/webhooks", body, 1)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}

	var errorResponse map[string]string
	if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if !strings.Contains(errorResponse["error"], "already exists") {
		t.Errorf("Expected error to contain 'already exists', got '%s'", errorResponse["error"])
	}
}

func TestHandleUpdateWebhook_Success(t *testing.T) {
	store := newMockHTTPStore()
	handler := HandleUpdateWebhook(store)

	// Create a webhook first
	webhook, _ := store.CreateWebhook(context.Background(), 1, "https://example.com/webhook", "oldhash")

	requestBody := WebhookRequest{
		URL:    "https://example.com/updated",
		Secret: "newsecretkey12345",
	}
	isActive := false
	requestBody.IsActive = &isActive

	body, _ := json.Marshal(requestBody)

	req := createAuthenticatedRequest("PUT", fmt.Sprintf("/notifications/webhooks/%d", webhook.ID), body, 1)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response WebhookResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.URL != requestBody.URL {
		t.Errorf("Expected URL %s, got %s", requestBody.URL, response.URL)
	}

	if response.IsActive != false {
		t.Error("Expected webhook to be inactive")
	}

	if response.SecretLastFour != "2345" {
		t.Errorf("Expected secret_last_four '2345', got '%s'", response.SecretLastFour)
	}
}

func TestHandleUpdateWebhook_NotFound(t *testing.T) {
	store := newMockHTTPStore()
	handler := HandleUpdateWebhook(store)

	requestBody := WebhookRequest{URL: "https://example.com/updated"}
	body, _ := json.Marshal(requestBody)

	req := createAuthenticatedRequest("PUT", "/notifications/webhooks/999", body, 1)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleDeleteWebhook_Success(t *testing.T) {
	store := newMockHTTPStore()
	handler := HandleDeleteWebhook(store)

	// Create a webhook first
	webhook, _ := store.CreateWebhook(context.Background(), 1, "https://example.com/webhook", "hash")

	req := createAuthenticatedRequest("DELETE", fmt.Sprintf("/notifications/webhooks/%d", webhook.ID), nil, 1)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Verify webhook was deleted
	webhooks, _ := store.ListWebhooks(context.Background(), 1)
	if len(webhooks) != 0 {
		t.Error("Expected webhook to be deleted")
	}
}

func TestHandleDeleteWebhook_NotFound(t *testing.T) {
	store := newMockHTTPStore()
	handler := HandleDeleteWebhook(store)

	req := createAuthenticatedRequest("DELETE", "/notifications/webhooks/999", nil, 1)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleDeleteWebhook_InvalidID(t *testing.T) {
	store := newMockHTTPStore()
	handler := HandleDeleteWebhook(store)

	req := createAuthenticatedRequest("DELETE", "/notifications/webhooks/invalid", nil, 1)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestWebhookValidation(t *testing.T) {
	testCases := []struct {
		name        string
		request     WebhookRequest
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid request",
			request:     WebhookRequest{URL: "https://example.com/webhook", Secret: "validSecret123456"},
			expectError: false,
		},
		{
			name:        "empty URL",
			request:     WebhookRequest{Secret: "validSecret123456"},
			expectError: true,
			errorMsg:    "url is required",
		},
		{
			name:        "invalid URL format",
			request:     WebhookRequest{URL: "not-a-url", Secret: "validSecret123456"},
			expectError: true,
			errorMsg:    "invalid url format",
		},
		{
			name:        "HTTP URL",
			request:     WebhookRequest{URL: "http://example.com/webhook", Secret: "validSecret123456"},
			expectError: true,
			errorMsg:    "url must use HTTPS",
		},
		{
			name:        "secret too short",
			request:     WebhookRequest{URL: "https://example.com/webhook", Secret: "short"},
			expectError: true,
			errorMsg:    "secret must be between 16 and 128 characters",
		},
		{
			name:        "secret too long",
			request:     WebhookRequest{URL: "https://example.com/webhook", Secret: strings.Repeat("a", 129)},
			expectError: true,
			errorMsg:    "secret must be between 16 and 128 characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateWebhookRequest(&tc.request)
			if tc.expectError {
				if err == nil {
					t.Error("Expected validation error")
				} else if !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestGetSecretLastFour(t *testing.T) {
	testCases := []struct {
		secret   string
		expected string
	}{
		{"mysecretkey1234", "1234"},
		{"short", "hort"}, // 5 chars -> last 4
		{"test", "test"},  // 4 chars -> full string
		{"", ""},          // empty -> empty
		{"abc", "abc"},    // 3 chars -> full string
		{"abcd", "abcd"},  // 4 chars -> full string
	}

	for _, tc := range testCases {
		result := getSecretLastFour(tc.secret)
		if result != tc.expected {
			t.Errorf("For secret '%s' (len=%d), expected '%s', got '%s'", tc.secret, len(tc.secret), tc.expected, result)
		}
	}
}

// mockNotifier is a test implementation that records calls
type mockNotifier struct {
	sendTestCalls []storage.Webhook
	sendTestError error
}

func (m *mockNotifier) Notify(ctx context.Context, rule storage.AlertRule, event *storage.AlertEvent) error {
	return nil
}

func (m *mockNotifier) SendTest(ctx context.Context, webhook storage.Webhook) error {
	m.sendTestCalls = append(m.sendTestCalls, webhook)
	return m.sendTestError
}

func TestHandleTestWebhook_Success(t *testing.T) {
	store := newMockHTTPStore()
	notifier := &mockNotifier{}

	// Create a test user and webhook
	user := &storage.User{ID: 1, Email: "test@example.com"}
	webhook, _ := store.CreateWebhook(context.Background(), user.ID, "https://example.com/webhook", storage.HashSecret("secret"))

	// Create request with user context
	req := httptest.NewRequest(http.MethodPost, "/notifications/webhooks/1/test", nil)
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Handle request
	handler := HandleTestWebhook(notifier, store)
	handler(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Verify notifier was called
	if len(notifier.sendTestCalls) != 1 {
		t.Fatalf("Expected 1 SendTest call, got %d", len(notifier.sendTestCalls))
	}

	if notifier.sendTestCalls[0].ID != webhook.ID {
		t.Errorf("Expected webhook ID %d, got %d", webhook.ID, notifier.sendTestCalls[0].ID)
	}

	// Verify response body
	var response TestWebhookResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "sent" {
		t.Errorf("Expected status 'sent', got '%s'", response.Status)
	}
	if response.WebhookID != webhook.ID {
		t.Errorf("Expected webhook_id %d, got %d", webhook.ID, response.WebhookID)
	}
	if response.TriggeredAt == "" {
		t.Error("Expected triggered_at to be set")
	}
}

func TestHandleTestWebhook_Unauthorized(t *testing.T) {
	store := newMockHTTPStore()
	notifier := &mockNotifier{}

	// Create request without user context
	req := httptest.NewRequest(http.MethodPost, "/notifications/webhooks/1/test", nil)
	rec := httptest.NewRecorder()

	// Handle request
	handler := HandleTestWebhook(notifier, store)
	handler(rec, req)

	// Check response
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}

	// Verify notifier was not called
	if len(notifier.sendTestCalls) != 0 {
		t.Errorf("Expected 0 SendTest calls, got %d", len(notifier.sendTestCalls))
	}
}

func TestHandleTestWebhook_NotFound(t *testing.T) {
	store := newMockHTTPStore()
	notifier := &mockNotifier{}

	// Create a test user (but no webhook)
	user := &storage.User{ID: 1, Email: "test@example.com"}

	// Create request for non-existent webhook
	req := httptest.NewRequest(http.MethodPost, "/notifications/webhooks/999/test", nil)
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Handle request
	handler := HandleTestWebhook(notifier, store)
	handler(rec, req)

	// Check response
	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Verify notifier was not called
	if len(notifier.sendTestCalls) != 0 {
		t.Errorf("Expected 0 SendTest calls, got %d", len(notifier.sendTestCalls))
	}
}

func TestHandleTestWebhook_InactiveWebhook(t *testing.T) {
	store := newMockHTTPStore()
	notifier := &mockNotifier{}

	// Create a test user and inactive webhook
	user := &storage.User{ID: 1, Email: "test@example.com"}
	webhook, _ := store.CreateWebhook(context.Background(), user.ID, "https://example.com/webhook", storage.HashSecret("secret"))

	// Make webhook inactive
	isActive := false
	store.UpdateWebhook(context.Background(), webhook.ID, user.ID, "", nil, &isActive)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/notifications/webhooks/1/test", nil)
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Handle request
	handler := HandleTestWebhook(notifier, store)
	handler(rec, req)

	// Check response
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Verify error message
	var errorResp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&errorResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	expectedMsg := "Webhook must be active to send a test payload"
	if errorResp["error"] != expectedMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedMsg, errorResp["error"])
	}

	// Verify notifier was not called
	if len(notifier.sendTestCalls) != 0 {
		t.Errorf("Expected 0 SendTest calls, got %d", len(notifier.sendTestCalls))
	}
}

func TestHandleTestWebhook_NotifierFailure(t *testing.T) {
	store := newMockHTTPStore()
	notifier := &mockNotifier{
		sendTestError: fmt.Errorf("webhook delivery failed"),
	}

	// Create a test user and webhook
	user := &storage.User{ID: 1, Email: "test@example.com"}
	_, _ = store.CreateWebhook(context.Background(), user.ID, "https://example.com/webhook", storage.HashSecret("secret"))

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/notifications/webhooks/1/test", nil)
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Handle request
	handler := HandleTestWebhook(notifier, store)
	handler(rec, req)

	// Check response
	if rec.Code != http.StatusBadGateway {
		t.Errorf("Expected status 502, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Verify notifier was called
	if len(notifier.sendTestCalls) != 1 {
		t.Errorf("Expected 1 SendTest call, got %d", len(notifier.sendTestCalls))
	}

	// Verify error message contains the failure
	var errorResp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&errorResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if !strings.Contains(errorResp["error"], "Failed to send test webhook") {
		t.Errorf("Expected error to contain 'Failed to send test webhook', got '%s'", errorResp["error"])
	}
}

func TestHandleTestWebhook_InvalidID(t *testing.T) {
	store := newMockHTTPStore()
	notifier := &mockNotifier{}

	// Create a test user
	user := &storage.User{ID: 1, Email: "test@example.com"}

	// Create request with invalid ID
	req := httptest.NewRequest(http.MethodPost, "/notifications/webhooks/invalid/test", nil)
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Handle request
	handler := HandleTestWebhook(notifier, store)
	handler(rec, req)

	// Check response
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}

	// Verify notifier was not called
	if len(notifier.sendTestCalls) != 0 {
		t.Errorf("Expected 0 SendTest calls, got %d", len(notifier.sendTestCalls))
	}
}

func TestHandleTestWebhook_WrongUser(t *testing.T) {
	store := newMockHTTPStore()
	notifier := &mockNotifier{}

	// Create two users
	user1 := &storage.User{ID: 1, Email: "user1@example.com"}
	user2 := &storage.User{ID: 2, Email: "user2@example.com"}

	// Create webhook for user1
	_, _ = store.CreateWebhook(context.Background(), user1.ID, "https://example.com/webhook", storage.HashSecret("secret"))

	// Try to access user1's webhook as user2
	req := httptest.NewRequest(http.MethodPost, "/notifications/webhooks/1/test", nil)
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user2)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Handle request
	handler := HandleTestWebhook(notifier, store)
	handler(rec, req)

	// Check response
	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Verify notifier was not called
	if len(notifier.sendTestCalls) != 0 {
		t.Errorf("Expected 0 SendTest calls, got %d", len(notifier.sendTestCalls))
	}
}
