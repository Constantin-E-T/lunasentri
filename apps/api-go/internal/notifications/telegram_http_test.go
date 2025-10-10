package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// mockTelegramStore implements storage.Store for testing
type mockTelegramStore struct {
	telegramRecipients map[int][]storage.TelegramRecipient // userID -> recipients
	nextTelegramID     int
	failCreate         bool
	failList           bool
	failUpdate         bool
	failDelete         bool
	failGet            bool
}

func newMockTelegramStore() *mockTelegramStore {
	return &mockTelegramStore{
		telegramRecipients: make(map[int][]storage.TelegramRecipient),
		nextTelegramID:     1,
	}
}

func (m *mockTelegramStore) ListTelegramRecipients(ctx context.Context, userID int) ([]storage.TelegramRecipient, error) {
	if m.failList {
		return nil, fmt.Errorf("mock list failure")
	}
	return m.telegramRecipients[userID], nil
}

func (m *mockTelegramStore) GetTelegramRecipient(ctx context.Context, id int, userID int) (*storage.TelegramRecipient, error) {
	if m.failGet {
		return nil, fmt.Errorf("mock get failure")
	}
	for _, recipient := range m.telegramRecipients[userID] {
		if recipient.ID == id {
			return &recipient, nil
		}
	}
	return nil, fmt.Errorf("telegram recipient with id %d not found for user %d", id, userID)
}

func (m *mockTelegramStore) CreateTelegramRecipient(ctx context.Context, userID int, chatID string) (*storage.TelegramRecipient, error) {
	if m.failCreate {
		return nil, fmt.Errorf("mock create failure")
	}

	// Check for duplicate chatID
	for _, recipient := range m.telegramRecipients[userID] {
		if recipient.ChatID == chatID {
			return nil, fmt.Errorf("telegram recipient with chat_id %s already exists", chatID)
		}
	}

	recipient := &storage.TelegramRecipient{
		ID:        m.nextTelegramID,
		UserID:    userID,
		ChatID:    chatID,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	m.nextTelegramID++

	if m.telegramRecipients[userID] == nil {
		m.telegramRecipients[userID] = make([]storage.TelegramRecipient, 0)
	}
	m.telegramRecipients[userID] = append(m.telegramRecipients[userID], *recipient)

	return recipient, nil
}

func (m *mockTelegramStore) UpdateTelegramRecipient(ctx context.Context, id int, userID int, chatID string, isActive *bool) (*storage.TelegramRecipient, error) {
	if m.failUpdate {
		return nil, fmt.Errorf("mock update failure")
	}

	// Find recipient
	for i, recipient := range m.telegramRecipients[userID] {
		if recipient.ID == id {
			// Check for duplicate chatID if changing
			if chatID != "" && chatID != recipient.ChatID {
				for _, otherRecipient := range m.telegramRecipients[userID] {
					if otherRecipient.ChatID == chatID && otherRecipient.ID != id {
						return nil, fmt.Errorf("telegram recipient with chat_id %s already exists", chatID)
					}
				}
				recipient.ChatID = chatID
			}

			if isActive != nil {
				recipient.IsActive = *isActive
			}

			m.telegramRecipients[userID][i] = recipient
			return &recipient, nil
		}
	}

	return nil, fmt.Errorf("telegram recipient with id %d not found for user %d", id, userID)
}

func (m *mockTelegramStore) DeleteTelegramRecipient(ctx context.Context, id int, userID int) error {
	if m.failDelete {
		return fmt.Errorf("mock delete failure")
	}

	// Find and remove recipient
	for i, recipient := range m.telegramRecipients[userID] {
		if recipient.ID == id {
			m.telegramRecipients[userID] = append(m.telegramRecipients[userID][:i], m.telegramRecipients[userID][i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("telegram recipient with id %d not found for user %d", id, userID)
}

// Implement other storage methods (not used in these tests)
func (m *mockTelegramStore) CreateUser(ctx context.Context, email, passwordHash string) (*storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) GetUserByEmail(ctx context.Context, email string) (*storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) GetUserByID(ctx context.Context, id int) (*storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) UpdateUserPassword(ctx context.Context, userID int, passwordHash string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) UpsertAdmin(ctx context.Context, email, passwordHash string) (*storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) CreatePasswordReset(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) (*storage.PasswordReset, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) GetPasswordResetByHash(ctx context.Context, tokenHash string) (*storage.PasswordReset, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) MarkPasswordResetUsed(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) ListUsers(ctx context.Context) ([]storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) DeleteUser(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) CountUsers(ctx context.Context) (int, error) {
	return 0, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) PromoteToAdmin(ctx context.Context, userID int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) DeletePasswordResetsForUser(ctx context.Context, userID int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) CreateAlertRule(ctx context.Context, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*storage.AlertRule, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) ListAlertRules(ctx context.Context) ([]storage.AlertRule, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) GetAlertRule(ctx context.Context, id int) (*storage.AlertRule, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) UpdateAlertRule(ctx context.Context, id int, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*storage.AlertRule, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) DeleteAlertRule(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) CreateAlertEvent(ctx context.Context, ruleID int, value float64) (*storage.AlertEvent, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) ListAlertEvents(ctx context.Context, limit int) ([]storage.AlertEvent, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) AckAlertEvent(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) ListActiveEvents(ctx context.Context, limit int) ([]storage.AlertEvent, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) ListWebhooks(ctx context.Context, userID int) ([]storage.Webhook, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) GetWebhook(ctx context.Context, id int, userID int) (*storage.Webhook, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) CreateWebhook(ctx context.Context, userID int, url, secretHash string) (*storage.Webhook, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) UpdateWebhook(ctx context.Context, id int, userID int, url string, secretHash *string, isActive *bool) (*storage.Webhook, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) DeleteWebhook(ctx context.Context, id int, userID int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) IncrementWebhookFailure(ctx context.Context, id int, lastErrorAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) MarkWebhookSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) UpdateWebhookDeliveryState(ctx context.Context, id int, lastAttemptAt time.Time, cooldownUntil *time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) ListEmailRecipients(ctx context.Context, userID int) ([]storage.EmailRecipient, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) GetEmailRecipient(ctx context.Context, id int, userID int) (*storage.EmailRecipient, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) CreateEmailRecipient(ctx context.Context, userID int, email string) (*storage.EmailRecipient, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) UpdateEmailRecipient(ctx context.Context, id int, userID int, email string, isActive *bool) (*storage.EmailRecipient, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) DeleteEmailRecipient(ctx context.Context, id int, userID int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) IncrementEmailFailure(ctx context.Context, id int, lastErrorAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) MarkEmailSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) UpdateEmailDeliveryState(ctx context.Context, id int, lastAttemptAt time.Time, cooldownUntil *time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) IncrementTelegramFailure(ctx context.Context, id int, lastErrorAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) MarkTelegramSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) UpdateTelegramDeliveryState(ctx context.Context, id int, lastAttemptAt time.Time, cooldownUntil *time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) CreateMachine(ctx context.Context, userID int, name, hostname, description, apiKeyHash string) (*storage.Machine, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) GetMachineByID(ctx context.Context, id int) (*storage.Machine, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) GetMachineByAPIKey(ctx context.Context, apiKeyHash string) (*storage.Machine, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) ListMachines(ctx context.Context, userID int) ([]storage.Machine, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) UpdateMachineStatus(ctx context.Context, id int, status string, lastSeen time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) UpdateMachineLastSeen(ctx context.Context, id int, lastSeen time.Time) error {
	return nil
}
func (m *mockTelegramStore) UpdateMachineSystemInfo(ctx context.Context, id int, info storage.MachineSystemInfoUpdate) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) UpdateMachineDetails(ctx context.Context, id int, updates map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) DeleteMachine(ctx context.Context, id int, userID int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) InsertMetrics(ctx context.Context, machineID int, cpuPct, memUsedPct, diskUsedPct float64, netRxBytes, netTxBytes int64, uptimeSeconds *float64, timestamp time.Time) error {
	return fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) GetLatestMetrics(ctx context.Context, machineID int) (*storage.MetricsHistory, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockTelegramStore) GetMetricsHistory(ctx context.Context, machineID int, from, to time.Time, limit int) ([]storage.MetricsHistory, error) {
	return nil, fmt.Errorf("not implemented")
}

// Heartbeat monitoring methods (stub implementations for testing)
func (m *mockTelegramStore) ListAllMachines(ctx context.Context) ([]storage.Machine, error) {
	return []storage.Machine{}, nil
}

func (m *mockTelegramStore) RecordMachineOfflineNotification(ctx context.Context, machineID int, notifiedAt time.Time) error {
	return nil
}

func (m *mockTelegramStore) GetMachineLastOfflineNotification(ctx context.Context, machineID int) (time.Time, error) {
	return time.Time{}, nil
}

func (m *mockTelegramStore) ClearMachineOfflineNotification(ctx context.Context, machineID int) error {
	return nil
}

func (m *mockTelegramStore) Close() error {
	return nil
}

// Helper to create a request with authenticated user context
func createTelegramAuthRequest(method, url string, body []byte, userID int) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	// Add user to context
	user := &storage.User{
		ID:    userID,
		Email: fmt.Sprintf("user%d@example.com", userID),
	}
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
	return req.WithContext(ctx)
}

// Test HandleListTelegramRecipients
func TestHandleListTelegramRecipients(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		userID         int
		setupStore     func(*mockTelegramStore)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful list with recipients",
			method: http.MethodGet,
			userID: 1,
			setupStore: func(m *mockTelegramStore) {
				m.CreateTelegramRecipient(context.Background(), 1, "123456789")
				m.CreateTelegramRecipient(context.Background(), 1, "987654321")
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var recipients []TelegramRecipientResponse
				if err := json.NewDecoder(rec.Body).Decode(&recipients); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if len(recipients) != 2 {
					t.Errorf("Expected 2 recipients, got %d", len(recipients))
				}
				if rec.Header().Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
				}
			},
		},
		{
			name:   "successful list with no recipients",
			method: http.MethodGet,
			userID: 1,
			setupStore: func(m *mockTelegramStore) {
				// No recipients
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var recipients []TelegramRecipientResponse
				if err := json.NewDecoder(rec.Body).Decode(&recipients); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if len(recipients) != 0 {
					t.Errorf("Expected 0 recipients, got %d", len(recipients))
				}
			},
		},
		{
			name:   "method not allowed",
			method: http.MethodPost,
			userID: 1,
			setupStore: func(m *mockTelegramStore) {
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("Expected error field in JSON response")
				}
			},
		},
		{
			name:   "storage failure",
			method: http.MethodGet,
			userID: 1,
			setupStore: func(m *mockTelegramStore) {
				m.failList = true
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("Expected error field in JSON response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockTelegramStore()
			tt.setupStore(store)

			handler := HandleListTelegramRecipients(store)
			req := createTelegramAuthRequest(tt.method, "/notifications/telegram", nil, tt.userID)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

// Test HandleCreateTelegramRecipient
func TestHandleCreateTelegramRecipient(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		requestBody    interface{}
		setupStore     func(*mockTelegramStore)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful create with notifier nil",
			userID: 1,
			requestBody: TelegramRecipientRequest{
				ChatID: "123456789",
			},
			setupStore:     func(m *mockTelegramStore) {},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var recipient TelegramRecipientResponse
				if err := json.NewDecoder(rec.Body).Decode(&recipient); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if recipient.ChatID != "123456789" {
					t.Errorf("Expected chat_id 123456789, got %s", recipient.ChatID)
				}
				if !recipient.IsActive {
					t.Error("Expected recipient to be active")
				}
			},
		},
		{
			name:   "invalid chat_id - not numeric",
			userID: 1,
			requestBody: TelegramRecipientRequest{
				ChatID: "invalid-chat-id",
			},
			setupStore:     func(m *mockTelegramStore) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("Expected error field in JSON response")
				}
			},
		},
		{
			name:   "missing chat_id",
			userID: 1,
			requestBody: TelegramRecipientRequest{
				ChatID: "",
			},
			setupStore:     func(m *mockTelegramStore) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("Expected error field in JSON response")
				}
			},
		},
		{
			name:   "duplicate chat_id",
			userID: 1,
			requestBody: TelegramRecipientRequest{
				ChatID: "123456789",
			},
			setupStore: func(m *mockTelegramStore) {
				m.CreateTelegramRecipient(context.Background(), 1, "123456789")
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("Expected error field in JSON response")
				}
			},
		},
		{
			name:           "invalid json",
			userID:         1,
			requestBody:    "invalid json",
			setupStore:     func(m *mockTelegramStore) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("Expected error field in JSON response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockTelegramStore()
			tt.setupStore(store)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			handler := HandleCreateTelegramRecipient(store)
			req := createTelegramAuthRequest(http.MethodPost, "/notifications/telegram", body, tt.userID)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

// Test HandleTestTelegram
func TestHandleTestTelegram(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		recipientID    string
		setupStore     func(*mockTelegramStore)
		notifierNil    bool
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "notifier not configured - returns 503",
			userID:      1,
			recipientID: "1",
			setupStore: func(m *mockTelegramStore) {
				m.CreateTelegramRecipient(context.Background(), 1, "123456789")
			},
			notifierNil:    true,
			expectedStatus: http.StatusServiceUnavailable,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if response["error"] != "Telegram notifier is not configured" {
					t.Errorf("Expected specific error message, got: %s", response["error"])
				}
				if rec.Header().Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
				}
			},
		},
		{
			name:        "recipient not found - still returns 503 when notifier nil",
			userID:      1,
			recipientID: "999",
			setupStore: func(m *mockTelegramStore) {
				// No recipients
			},
			notifierNil:    true,
			expectedStatus: http.StatusServiceUnavailable,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if response["error"] != "Telegram notifier is not configured" {
					t.Errorf("Expected notifier error, got: %s", response["error"])
				}
			},
		},
		{
			name:        "invalid recipient ID - still returns 503 when notifier nil",
			userID:      1,
			recipientID: "invalid",
			setupStore: func(m *mockTelegramStore) {
			},
			notifierNil:    true,
			expectedStatus: http.StatusServiceUnavailable,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if response["error"] != "Telegram notifier is not configured" {
					t.Errorf("Expected notifier error, got: %s", response["error"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockTelegramStore()
			tt.setupStore(store)

			var notifier *TelegramNotifier
			if !tt.notifierNil {
				// For now, we only test nil notifier scenarios
				// In the future, we can add a mock notifier
				notifier = nil
			}

			handler := HandleTestTelegram(store, notifier)
			url := fmt.Sprintf("/notifications/telegram/%s/test", tt.recipientID)
			req := createTelegramAuthRequest(http.MethodPost, url, nil, tt.userID)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

// Test HandleUpdateTelegramRecipient
func TestHandleUpdateTelegramRecipient(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		recipientID    string
		requestBody    interface{}
		setupStore     func(*mockTelegramStore)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successful update - toggle is_active",
			userID:      1,
			recipientID: "1",
			requestBody: TelegramRecipientRequest{
				IsActive: boolPtr(false),
			},
			setupStore: func(m *mockTelegramStore) {
				m.CreateTelegramRecipient(context.Background(), 1, "123456789")
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var recipient TelegramRecipientResponse
				if err := json.NewDecoder(rec.Body).Decode(&recipient); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if recipient.IsActive {
					t.Error("Expected recipient to be inactive")
				}
			},
		},
		{
			name:        "recipient not found",
			userID:      1,
			recipientID: "999",
			requestBody: TelegramRecipientRequest{
				IsActive: boolPtr(false),
			},
			setupStore:     func(m *mockTelegramStore) {},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("Expected error field in JSON response")
				}
			},
		},
		{
			name:        "invalid recipient ID",
			userID:      1,
			recipientID: "invalid",
			requestBody: TelegramRecipientRequest{
				IsActive: boolPtr(false),
			},
			setupStore:     func(m *mockTelegramStore) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("Expected error field in JSON response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockTelegramStore()
			tt.setupStore(store)

			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			handler := HandleUpdateTelegramRecipient(store)
			url := fmt.Sprintf("/notifications/telegram/%s", tt.recipientID)
			req := createTelegramAuthRequest(http.MethodPut, url, body, tt.userID)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

// Test HandleDeleteTelegramRecipient
func TestHandleDeleteTelegramRecipient(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		recipientID    string
		setupStore     func(*mockTelegramStore)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successful delete",
			userID:      1,
			recipientID: "1",
			setupStore: func(m *mockTelegramStore) {
				m.CreateTelegramRecipient(context.Background(), 1, "123456789")
			},
			expectedStatus: http.StatusNoContent,
			checkResponse:  nil,
		},
		{
			name:        "recipient not found",
			userID:      1,
			recipientID: "999",
			setupStore: func(m *mockTelegramStore) {
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("Expected error field in JSON response")
				}
			},
		},
		{
			name:        "invalid recipient ID",
			userID:      1,
			recipientID: "invalid",
			setupStore: func(m *mockTelegramStore) {
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("Expected error field in JSON response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockTelegramStore()
			tt.setupStore(store)

			handler := HandleDeleteTelegramRecipient(store)
			url := fmt.Sprintf("/notifications/telegram/%s", tt.recipientID)
			req := createTelegramAuthRequest(http.MethodDelete, url, nil, tt.userID)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}
