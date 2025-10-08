package notifications

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// mockStore implements storage.Store for testing
type mockStore struct {
	users         []storage.User
	webhooks      map[int][]storage.Webhook // userID -> webhooks
	failureCounts map[int]int               // webhookID -> failure count
	successTimes  map[int]time.Time         // webhookID -> last success time
}

func newMockStore() *mockStore {
	return &mockStore{
		webhooks:      make(map[int][]storage.Webhook),
		failureCounts: make(map[int]int),
		successTimes:  make(map[int]time.Time),
	}
}

func (m *mockStore) ListUsers(ctx context.Context) ([]storage.User, error) {
	return m.users, nil
}

func (m *mockStore) ListWebhooks(ctx context.Context, userID int) ([]storage.Webhook, error) {
	return m.webhooks[userID], nil
}

func (m *mockStore) GetWebhook(ctx context.Context, id int, userID int) (*storage.Webhook, error) {
	for _, webhook := range m.webhooks[userID] {
		if webhook.ID == id {
			return &webhook, nil
		}
	}
	return nil, fmt.Errorf("webhook with id %d not found for user %d", id, userID)
}

func (m *mockStore) IncrementWebhookFailure(ctx context.Context, id int, lastErrorAt time.Time) error {
	m.failureCounts[id]++
	return nil
}

func (m *mockStore) MarkWebhookSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error {
	m.failureCounts[id] = 0
	m.successTimes[id] = lastSuccessAt
	return nil
}

// Implement other required methods with no-ops for this test
func (m *mockStore) CreateUser(ctx context.Context, email, passwordHash string) (*storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) GetUserByEmail(ctx context.Context, email string) (*storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) GetUserByID(ctx context.Context, id int) (*storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) UpdateUserPassword(ctx context.Context, userID int, passwordHash string) error {
	return fmt.Errorf("not implemented")
}
func (m *mockStore) UpsertAdmin(ctx context.Context, email, passwordHash string) (*storage.User, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) CreatePasswordReset(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) (*storage.PasswordReset, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) GetPasswordResetByHash(ctx context.Context, tokenHash string) (*storage.PasswordReset, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) MarkPasswordResetUsed(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockStore) DeleteUser(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockStore) CountUsers(ctx context.Context) (int, error) {
	return 0, fmt.Errorf("not implemented")
}
func (m *mockStore) PromoteToAdmin(ctx context.Context, userID int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockStore) DeletePasswordResetsForUser(ctx context.Context, userID int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockStore) ListAlertRules(ctx context.Context) ([]storage.AlertRule, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) CreateAlertRule(ctx context.Context, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*storage.AlertRule, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) UpdateAlertRule(ctx context.Context, id int, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*storage.AlertRule, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) DeleteAlertRule(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockStore) ListAlertEvents(ctx context.Context, limit int) ([]storage.AlertEvent, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) CreateAlertEvent(ctx context.Context, ruleID int, value float64) (*storage.AlertEvent, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) AckAlertEvent(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockStore) CreateWebhook(ctx context.Context, userID int, url, secretHash string) (*storage.Webhook, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) UpdateWebhook(ctx context.Context, id int, userID int, url string, secretHash *string, isActive *bool) (*storage.Webhook, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *mockStore) DeleteWebhook(ctx context.Context, id int, userID int) error {
	return fmt.Errorf("not implemented")
}
func (m *mockStore) Close() error {
	return nil
}

func TestNotifier_Send_Success(t *testing.T) {
	// Setup mock server
	var receivedPayload WebhookPayload
	var receivedSignature string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		receivedSignature = r.Header.Get("X-LunaSentri-Signature")
		if receivedSignature == "" {
			t.Error("Expected X-LunaSentri-Signature header")
		}

		// Read and parse payload
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		if err := json.Unmarshal(body, &receivedPayload); err != nil {
			t.Fatalf("Failed to unmarshal payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Setup mock store
	store := newMockStore()
	secret := "test-secret"
	secretHash := storage.HashSecret(secret)

	store.users = []storage.User{{ID: 1, Email: "test@example.com"}}
	store.webhooks[1] = []storage.Webhook{
		{
			ID:         1,
			UserID:     1,
			URL:        server.URL,
			SecretHash: secretHash,
			IsActive:   true,
		},
	}

	// Create notifier
	notifier := NewNotifier(store, log.Default())

	// Test data
	rule := storage.AlertRule{
		ID:           1,
		Name:         "Test Rule",
		Metric:       "cpu_pct",
		Comparison:   "above",
		ThresholdPct: 80.0,
		TriggerAfter: 3,
	}

	event := storage.AlertEvent{
		ID:          1,
		RuleID:      1,
		Value:       85.5,
		TriggeredAt: time.Now(),
	}

	// Send notification
	ctx := context.Background()
	err := notifier.Send(ctx, rule, event)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}

	// Wait a bit for goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Verify payload
	if receivedPayload.RuleID != rule.ID {
		t.Errorf("Expected rule ID %d, got %d", rule.ID, receivedPayload.RuleID)
	}
	if receivedPayload.RuleName != rule.Name {
		t.Errorf("Expected rule name %s, got %s", rule.Name, receivedPayload.RuleName)
	}
	if receivedPayload.Value != event.Value {
		t.Errorf("Expected value %f, got %f", event.Value, receivedPayload.Value)
	}

	// Verify signature
	payloadBytes, _ := json.Marshal(receivedPayload)
	expectedSignature, _ := notifier.createSignature(payloadBytes, secretHash)
	if receivedSignature != expectedSignature {
		t.Errorf("Expected signature %s, got %s", expectedSignature, receivedSignature)
	}

	// Verify success was marked
	if store.failureCounts[1] != 0 {
		t.Errorf("Expected failure count 0, got %d", store.failureCounts[1])
	}
	if store.successTimes[1].IsZero() {
		t.Error("Expected success time to be set")
	}
}

func TestNotifier_Send_Failure_And_Retry(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	// Don't close server until after test completes

	// Setup mock store
	store := newMockStore()
	secret := "test-secret"
	secretHash := storage.HashSecret(secret)

	store.users = []storage.User{{ID: 1, Email: "test@example.com"}}
	store.webhooks[1] = []storage.Webhook{
		{
			ID:         1,
			UserID:     1,
			URL:        server.URL,
			SecretHash: secretHash,
			IsActive:   true,
		},
	}

	// Create notifier
	notifier := NewNotifier(store, log.Default())

	// Test data
	rule := storage.AlertRule{ID: 1, Name: "Test Rule"}
	event := storage.AlertEvent{ID: 1, RuleID: 1, Value: 85.5, TriggeredAt: time.Now()}

	// Send notification
	ctx := context.Background()
	err := notifier.Send(ctx, rule, event)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}

	// Wait for retries to complete (1s + 2s + processing time)
	time.Sleep(4 * time.Second)

	// Close server now
	server.Close()

	// Verify it attempted 3 times (2 failures + 1 success)
	if attempt != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempt)
	}

	// Verify success was marked on final attempt
	if store.failureCounts[1] != 0 {
		t.Errorf("Expected failure count 0 after success, got %d", store.failureCounts[1])
	}
}

func TestNotifier_Send_MaxRetries_Exceeded(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Setup mock store
	store := newMockStore()
	secret := "test-secret"
	secretHash := storage.HashSecret(secret)

	store.users = []storage.User{{ID: 1, Email: "test@example.com"}}
	store.webhooks[1] = []storage.Webhook{
		{
			ID:         1,
			UserID:     1,
			URL:        server.URL,
			SecretHash: secretHash,
			IsActive:   true,
		},
	}

	// Create notifier
	notifier := NewNotifier(store, log.Default())

	// Test data
	rule := storage.AlertRule{ID: 1, Name: "Test Rule"}
	event := storage.AlertEvent{ID: 1, RuleID: 1, Value: 85.5, TriggeredAt: time.Now()}

	// Send notification
	ctx := context.Background()
	err := notifier.Send(ctx, rule, event)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}

	// Wait for retries to complete
	time.Sleep(4 * time.Second)

	// Verify it attempted 3 times (all failures)
	if attempt != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempt)
	}

	// Verify failure was marked
	if store.failureCounts[1] != 1 {
		t.Errorf("Expected failure count 1, got %d", store.failureCounts[1])
	}
}

func TestNotifier_Send_Context_Cancellation(t *testing.T) {
	// Server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second) // Simulate slow response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Setup mock store
	store := newMockStore()
	secret := "test-secret"
	secretHash := storage.HashSecret(secret)

	store.users = []storage.User{{ID: 1, Email: "test@example.com"}}
	store.webhooks[1] = []storage.Webhook{
		{
			ID:         1,
			UserID:     1,
			URL:        server.URL,
			SecretHash: secretHash,
			IsActive:   true,
		},
	}

	// Create notifier
	notifier := NewNotifier(store, log.Default())

	// Test data
	rule := storage.AlertRule{ID: 1, Name: "Test Rule"}
	event := storage.AlertEvent{ID: 1, RuleID: 1, Value: 85.5, TriggeredAt: time.Now()}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Send notification - should timeout
	err := notifier.Send(ctx, rule, event)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}

	// Wait a bit for the goroutine to start and potentially timeout
	time.Sleep(200 * time.Millisecond)

	// The test passes if no panic occurs and the context cancellation is handled
}

func TestNotifier_CreateSignature(t *testing.T) {
	notifier := NewNotifier(newMockStore(), log.Default())

	payload := []byte(`{"test": "data"}`)
	secret := "test-secret"
	secretHash := storage.HashSecret(secret)

	signature, err := notifier.createSignature(payload, secretHash)
	if err != nil {
		t.Fatalf("Failed to create signature: %v", err)
	}

	// Verify signature format
	if !strings.HasPrefix(signature, "sha256=") {
		t.Errorf("Expected signature to start with 'sha256=', got %s", signature)
	}

	// Verify signature is deterministic
	signature2, err := notifier.createSignature(payload, secretHash)
	if err != nil {
		t.Fatalf("Failed to create signature again: %v", err)
	}

	if signature != signature2 {
		t.Error("Signature should be deterministic")
	}

	// Verify signature manually
	secretBytes, _ := hex.DecodeString(secretHash)
	h := hmac.New(sha256.New, secretBytes)
	h.Write(payload)
	expectedSig := "sha256=" + hex.EncodeToString(h.Sum(nil))

	if signature != expectedSig {
		t.Errorf("Expected signature %s, got %s", expectedSig, signature)
	}
}

func TestNotifier_Send_NoActiveWebhooks(t *testing.T) {
	// Setup mock store with no webhooks
	store := newMockStore()
	store.users = []storage.User{{ID: 1, Email: "test@example.com"}}

	// Create notifier
	notifier := NewNotifier(store, log.Default())

	// Test data
	rule := storage.AlertRule{ID: 1, Name: "Test Rule"}
	event := storage.AlertEvent{ID: 1, RuleID: 1, Value: 85.5, TriggeredAt: time.Now()}

	// Send notification
	ctx := context.Background()
	err := notifier.Send(ctx, rule, event)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}

	// Should succeed without error even with no webhooks
}

func TestNotifier_Send_InactiveWebhooks(t *testing.T) {
	// Setup mock store with inactive webhooks only
	store := newMockStore()
	store.users = []storage.User{{ID: 1, Email: "test@example.com"}}
	store.webhooks[1] = []storage.Webhook{
		{
			ID:         1,
			UserID:     1,
			URL:        "http://example.com/webhook",
			SecretHash: storage.HashSecret("secret"),
			IsActive:   false, // Inactive
		},
	}

	// Create notifier
	notifier := NewNotifier(store, log.Default())

	// Test data
	rule := storage.AlertRule{ID: 1, Name: "Test Rule"}
	event := storage.AlertEvent{ID: 1, RuleID: 1, Value: 85.5, TriggeredAt: time.Now()}

	// Send notification
	ctx := context.Background()
	err := notifier.Send(ctx, rule, event)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}

	// Should succeed without error even with only inactive webhooks
}

func TestNotifier_SendTest_Success(t *testing.T) {
	// Setup mock HTTP server
	var receivedPayload WebhookPayload
	var receivedSignature string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		receivedSignature = r.Header.Get("X-LunaSentri-Signature")
		if receivedSignature == "" {
			t.Error("Expected X-LunaSentri-Signature header")
		}

		// Read and parse payload
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		if err := json.Unmarshal(body, &receivedPayload); err != nil {
			t.Fatalf("Failed to unmarshal payload: %v", err)
		}

		// Return success
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Setup mock store
	store := newMockStore()

	// Create notifier
	notifier := NewNotifier(store, log.Default())

	// Create test webhook
	secret := "test-secret-key-12345"
	webhook := storage.Webhook{
		ID:         1,
		UserID:     1,
		URL:        server.URL,
		SecretHash: storage.HashSecret(secret),
		IsActive:   true,
	}

	// Send test notification
	ctx := context.Background()
	err := notifier.SendTest(ctx, webhook)
	if err != nil {
		t.Fatalf("SendTest failed: %v", err)
	}

	// Verify the payload was marked as a test
	if receivedPayload.RuleName != "Test Webhook" {
		t.Errorf("Expected test rule name 'Test Webhook', got %s", receivedPayload.RuleName)
	}
	if receivedPayload.RuleID != 0 {
		t.Errorf("Expected test rule ID 0, got %d", receivedPayload.RuleID)
	}
	if receivedPayload.EventID != 0 {
		t.Errorf("Expected test event ID 0, got %d", receivedPayload.EventID)
	}

	// Verify signature was received
	if receivedSignature == "" {
		t.Error("No signature received")
	}

	// Verify success was marked
	if store.failureCounts[webhook.ID] != 0 {
		t.Errorf("Expected failure count 0, got %d", store.failureCounts[webhook.ID])
	}
	if store.successTimes[webhook.ID].IsZero() {
		t.Error("Expected success time to be set")
	}
}

func TestNotifier_SendTest_Failure(t *testing.T) {
	// Setup mock HTTP server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Setup mock store
	store := newMockStore()

	// Create notifier
	notifier := NewNotifier(store, log.Default())

	// Create test webhook
	secret := "test-secret-key-12345"
	webhook := storage.Webhook{
		ID:         1,
		UserID:     1,
		URL:        server.URL,
		SecretHash: storage.HashSecret(secret),
		IsActive:   true,
	}

	// Send test notification (should fail after retries)
	ctx := context.Background()
	err := notifier.SendTest(ctx, webhook)
	if err == nil {
		t.Fatal("Expected SendTest to fail, but it succeeded")
	}

	// Verify failure was tracked
	if store.failureCounts[webhook.ID] != 1 {
		t.Errorf("Expected failure count 1, got %d", store.failureCounts[webhook.ID])
	}
}
