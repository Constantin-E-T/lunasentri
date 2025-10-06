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
