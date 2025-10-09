package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// mockStore is a mock implementation of storage.Store for testing
type mockStore struct {
	users map[string]*storage.User
}

func newMockStore() *mockStore {
	return &mockStore{
		users: make(map[string]*storage.User),
	}
}

func (m *mockStore) CreateUser(ctx context.Context, email, passwordHash string) (*storage.User, error) {
	user := &storage.User{
		ID:           len(m.users) + 1,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	m.users[email] = user
	return user, nil
}

func (m *mockStore) GetUserByEmail(ctx context.Context, email string) (*storage.User, error) {
	if user, ok := m.users[email]; ok {
		return user, nil
	}
	return nil, storage.ErrUserNotFound
}

func (m *mockStore) GetUserByID(ctx context.Context, id int) (*storage.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, storage.ErrUserNotFound
}

func (m *mockStore) UpsertAdmin(ctx context.Context, email, passwordHash string) (*storage.User, error) {
	return m.CreateUser(ctx, email, passwordHash)
}

func (m *mockStore) UpdateUserPassword(ctx context.Context, userID int, passwordHash string) error {
	for _, user := range m.users {
		if user.ID == userID {
			user.PasswordHash = passwordHash
			return nil
		}
	}
	return storage.ErrUserNotFound
}

func (m *mockStore) CreatePasswordReset(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) (*storage.PasswordReset, error) {
	return &storage.PasswordReset{
		ID:        1,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockStore) GetPasswordResetByHash(ctx context.Context, tokenHash string) (*storage.PasswordReset, error) {
	return nil, storage.ErrPasswordResetNotFound
}

func (m *mockStore) MarkPasswordResetUsed(ctx context.Context, id int) error {
	return nil
}

func (m *mockStore) ListUsers(ctx context.Context) ([]storage.User, error) {
	users := make([]storage.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

func (m *mockStore) DeleteUser(ctx context.Context, id int) error {
	// Find the user to delete
	var userToDelete *storage.User
	for _, user := range m.users {
		if user.ID == id {
			userToDelete = user
			break
		}
	}

	if userToDelete == nil {
		return storage.ErrUserNotFound
	}

	// If user is admin, check if they're the last admin
	if userToDelete.IsAdmin {
		adminCount := 0
		for _, user := range m.users {
			if user.IsAdmin {
				adminCount++
			}
		}
		if adminCount <= 1 {
			return fmt.Errorf("cannot delete the last admin")
		}
	}

	// Delete user
	for email, user := range m.users {
		if user.ID == id {
			delete(m.users, email)
			return nil
		}
	}
	return storage.ErrUserNotFound
}

func (m *mockStore) CountUsers(ctx context.Context) (int, error) {
	return len(m.users), nil
}

func (m *mockStore) PromoteToAdmin(ctx context.Context, userID int) error {
	for _, user := range m.users {
		if user.ID == userID {
			user.IsAdmin = true
			return nil
		}
	}
	return storage.ErrUserNotFound
}

func (m *mockStore) DeletePasswordResetsForUser(ctx context.Context, userID int) error {
	return nil
}

// Alert Rules methods (stub implementations for testing)
func (m *mockStore) ListAlertRules(ctx context.Context) ([]storage.AlertRule, error) {
	return []storage.AlertRule{}, nil
}

func (m *mockStore) CreateAlertRule(ctx context.Context, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*storage.AlertRule, error) {
	return &storage.AlertRule{
		ID:           1,
		Name:         name,
		Metric:       metric,
		ThresholdPct: thresholdPct,
		Comparison:   comparison,
		TriggerAfter: triggerAfter,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (m *mockStore) UpdateAlertRule(ctx context.Context, id int, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*storage.AlertRule, error) {
	return &storage.AlertRule{
		ID:           id,
		Name:         name,
		Metric:       metric,
		ThresholdPct: thresholdPct,
		Comparison:   comparison,
		TriggerAfter: triggerAfter,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (m *mockStore) DeleteAlertRule(ctx context.Context, id int) error {
	return nil
}

// Alert Events methods (stub implementations for testing)
func (m *mockStore) ListAlertEvents(ctx context.Context, limit int) ([]storage.AlertEvent, error) {
	return []storage.AlertEvent{}, nil
}

func (m *mockStore) CreateAlertEvent(ctx context.Context, ruleID int, value float64) (*storage.AlertEvent, error) {
	return &storage.AlertEvent{
		ID:           1,
		RuleID:       ruleID,
		TriggeredAt:  time.Now(),
		Value:        value,
		Acknowledged: false,
	}, nil
}

func (m *mockStore) AckAlertEvent(ctx context.Context, id int) error {
	return nil
}

// Webhook methods
func (m *mockStore) CreateWebhook(ctx context.Context, userID int, url string, secretHash string) (*storage.Webhook, error) {
	return &storage.Webhook{ID: 1, UserID: userID, URL: url, SecretHash: secretHash, IsActive: true}, nil
}

func (m *mockStore) ListWebhooks(ctx context.Context, userID int) ([]storage.Webhook, error) {
	return []storage.Webhook{}, nil
}

func (m *mockStore) GetWebhook(ctx context.Context, id int, userID int) (*storage.Webhook, error) {
	return &storage.Webhook{ID: id, UserID: userID, URL: "https://example.com/webhook", IsActive: true}, nil
}

func (m *mockStore) UpdateWebhook(ctx context.Context, id int, userID int, url string, secretHash *string, isActive *bool) (*storage.Webhook, error) {
	return &storage.Webhook{ID: id, UserID: userID, URL: url, IsActive: true}, nil
}

func (m *mockStore) DeleteWebhook(ctx context.Context, id int, userID int) error {
	return nil
}

func (m *mockStore) IncrementWebhookFailure(ctx context.Context, webhookID int, lastErrorAt time.Time) error {
	return nil
}

func (m *mockStore) MarkWebhookSuccess(ctx context.Context, webhookID int, lastSuccessAt time.Time) error {
	return nil
}

func (m *mockStore) UpdateWebhookDeliveryState(ctx context.Context, id int, lastAttemptAt time.Time, cooldownUntil *time.Time) error {
	return nil
}

func (m *mockStore) ListEmailRecipients(ctx context.Context, userID int) ([]storage.EmailRecipient, error) {
	return []storage.EmailRecipient{}, nil
}

func (m *mockStore) GetEmailRecipient(ctx context.Context, id int, userID int) (*storage.EmailRecipient, error) {
	return &storage.EmailRecipient{ID: id, UserID: userID, Email: "test@example.com", IsActive: true}, nil
}

func (m *mockStore) CreateEmailRecipient(ctx context.Context, userID int, email string) (*storage.EmailRecipient, error) {
	return &storage.EmailRecipient{ID: 1, UserID: userID, Email: email, IsActive: true}, nil
}

func (m *mockStore) UpdateEmailRecipient(ctx context.Context, id int, userID int, email string, isActive *bool) (*storage.EmailRecipient, error) {
	return &storage.EmailRecipient{ID: id, UserID: userID, Email: email, IsActive: true}, nil
}

func (m *mockStore) DeleteEmailRecipient(ctx context.Context, id int, userID int) error {
	return nil
}

func (m *mockStore) IncrementEmailFailure(ctx context.Context, id int, lastErrorAt time.Time) error {
	return nil
}

func (m *mockStore) MarkEmailSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error {
	return nil
}

func (m *mockStore) UpdateEmailDeliveryState(ctx context.Context, id int, lastAttemptAt time.Time, cooldownUntil *time.Time) error {
	return nil
}

func (m *mockStore) ListTelegramRecipients(ctx context.Context, userID int) ([]storage.TelegramRecipient, error) {
	return []storage.TelegramRecipient{}, nil
}

func (m *mockStore) GetTelegramRecipient(ctx context.Context, id int, userID int) (*storage.TelegramRecipient, error) {
	return &storage.TelegramRecipient{ID: id, UserID: userID, ChatID: "123456789", IsActive: true}, nil
}

func (m *mockStore) CreateTelegramRecipient(ctx context.Context, userID int, chatID string) (*storage.TelegramRecipient, error) {
	return &storage.TelegramRecipient{ID: 1, UserID: userID, ChatID: chatID, IsActive: true}, nil
}

func (m *mockStore) UpdateTelegramRecipient(ctx context.Context, id int, userID int, chatID string, isActive *bool) (*storage.TelegramRecipient, error) {
	return &storage.TelegramRecipient{ID: id, UserID: userID, ChatID: chatID, IsActive: true}, nil
}

func (m *mockStore) DeleteTelegramRecipient(ctx context.Context, id int, userID int) error {
	return nil
}

func (m *mockStore) IncrementTelegramFailure(ctx context.Context, id int, lastErrorAt time.Time) error {
	return nil
}

func (m *mockStore) MarkTelegramSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error {
	return nil
}

func (m *mockStore) UpdateTelegramDeliveryState(ctx context.Context, id int, lastAttemptAt time.Time, cooldownUntil *time.Time) error {
	return nil
}

func (m *mockStore) Close() error {
	return nil
}

func TestNewService(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := 15 * time.Minute

	service, err := NewService(store, secret, ttl)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	if service == nil {
		t.Fatal("NewService returned nil service")
	}
}

func TestNewService_EmptySecret(t *testing.T) {
	store := newMockStore()
	secret := ""
	ttl := 15 * time.Minute

	_, err := NewService(store, secret, ttl)
	if err == nil {
		t.Fatal("NewService should fail with empty secret")
	}
}

func TestNewService_InvalidTTL(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := -1 * time.Minute

	_, err := NewService(store, secret, ttl)
	if err == nil {
		t.Fatal("NewService should fail with negative TTL")
	}
}

func TestAuthenticate_Success(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := 15 * time.Minute

	service, err := NewService(store, secret, ttl)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	// Create a test user
	email := "test@example.com"
	password := "testpassword123"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	ctx := context.Background()
	_, err = store.CreateUser(ctx, email, hashedPassword)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Authenticate with correct credentials
	user, err := service.Authenticate(ctx, email, password)
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	if user.Email != email {
		t.Fatalf("Expected email %s, got %s", email, user.Email)
	}
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := 15 * time.Minute

	service, err := NewService(store, secret, ttl)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	// Create a test user
	email := "test@example.com"
	password := "testpassword123"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	ctx := context.Background()
	_, err = store.CreateUser(ctx, email, hashedPassword)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Authenticate with wrong password
	_, err = service.Authenticate(ctx, email, "wrongpassword")
	if err == nil {
		t.Fatal("Authenticate should fail with wrong password")
	}

	if err.Error() != "invalid credentials" {
		t.Fatalf("Expected 'invalid credentials' error, got: %v", err)
	}
}

func TestAuthenticate_NonexistentUser(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := 15 * time.Minute

	service, err := NewService(store, secret, ttl)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	ctx := context.Background()

	// Authenticate with nonexistent user
	_, err = service.Authenticate(ctx, "nonexistent@example.com", "password")
	if err == nil {
		t.Fatal("Authenticate should fail with nonexistent user")
	}

	if err.Error() != "invalid credentials" {
		t.Fatalf("Expected 'invalid credentials' error, got: %v", err)
	}
}

func TestCreateSession(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := 15 * time.Minute

	service, err := NewService(store, secret, ttl)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	userID := 123
	token, err := service.CreateSession(userID)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if token == "" {
		t.Fatal("CreateSession returned empty token")
	}
}

func TestValidateSession(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := 15 * time.Minute

	service, err := NewService(store, secret, ttl)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	userID := 456
	token, err := service.CreateSession(userID)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	validatedUserID, err := service.ValidateSession(token)
	if err != nil {
		t.Fatalf("ValidateSession failed: %v", err)
	}

	if validatedUserID != userID {
		t.Fatalf("Expected user ID %d, got %d", userID, validatedUserID)
	}
}

func TestGetUser(t *testing.T) {
	store := newMockStore()
	secret := "test-secret-key"
	ttl := 15 * time.Minute

	service, err := NewService(store, secret, ttl)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	// Create a test user
	email := "test@example.com"
	hashedPassword, _ := HashPassword("password")
	ctx := context.Background()
	createdUser, err := store.CreateUser(ctx, email, hashedPassword)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Get user by ID
	user, err := service.GetUser(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if user.Email != email {
		t.Fatalf("Expected email %s, got %s", email, user.Email)
	}
}
