package storage

import (
	"context"
	"testing"
	"time"
)

func TestSQLiteStore_CreateUser(t *testing.T) {
	// Use in-memory SQLite database for testing
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	email := "test@example.com"
	passwordHash := "hashed_password_123"

	// Test creating a user
	user, err := store.CreateUser(ctx, email, passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify user data
	if user.ID == 0 {
		t.Error("Expected user ID to be set")
	}
	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}
	if user.PasswordHash != passwordHash {
		t.Errorf("Expected password hash %s, got %s", passwordHash, user.PasswordHash)
	}
	if user.CreatedAt.IsZero() {
		t.Error("Expected created_at to be set")
	}
	if time.Since(user.CreatedAt) > time.Second {
		t.Error("Expected created_at to be recent")
	}
}

func TestSQLiteStore_CreateUser_UniqueConstraint(t *testing.T) {
	// Use in-memory SQLite database for testing
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	email := "test@example.com"
	passwordHash := "hashed_password_123"

	// Create first user
	_, err = store.CreateUser(ctx, email, passwordHash)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	// Try to create second user with same email
	_, err = store.CreateUser(ctx, email, "different_hash")
	if err == nil {
		t.Error("Expected error when creating user with duplicate email")
	}
	if err.Error() != "user with email test@example.com already exists" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestSQLiteStore_GetUserByEmail(t *testing.T) {
	// Use in-memory SQLite database for testing
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	email := "test@example.com"
	passwordHash := "hashed_password_123"

	// Create a user first
	createdUser, err := store.CreateUser(ctx, email, passwordHash)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get the user by email
	retrievedUser, err := store.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	// Compare the users
	if retrievedUser.ID != createdUser.ID {
		t.Errorf("Expected ID %d, got %d", createdUser.ID, retrievedUser.ID)
	}
	if retrievedUser.Email != createdUser.Email {
		t.Errorf("Expected email %s, got %s", createdUser.Email, retrievedUser.Email)
	}
	if retrievedUser.PasswordHash != createdUser.PasswordHash {
		t.Errorf("Expected password hash %s, got %s", createdUser.PasswordHash, retrievedUser.PasswordHash)
	}
	if !retrievedUser.CreatedAt.Equal(createdUser.CreatedAt) {
		t.Errorf("Expected created_at %v, got %v", createdUser.CreatedAt, retrievedUser.CreatedAt)
	}
}

func TestSQLiteStore_GetUserByEmail_NotFound(t *testing.T) {
	// Use in-memory SQLite database for testing
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	email := "nonexistent@example.com"

	// Try to get a user that doesn't exist
	_, err = store.GetUserByEmail(ctx, email)
	if err == nil {
		t.Error("Expected error when getting non-existent user")
	}
	if err.Error() != "user with email nonexistent@example.com not found" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestSQLiteStore_MigrationIdempotency(t *testing.T) {
	// Use in-memory SQLite database for testing
	store1, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create first test store: %v", err)
	}
	defer store1.Close()

	// Create a user to verify the database is working
	ctx := context.Background()
	user1, err := store1.CreateUser(ctx, "test1@example.com", "hash1")
	if err != nil {
		t.Fatalf("Failed to create user in first store: %v", err)
	}

	// Create another store instance (should not re-run migrations)
	store2, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create second test store: %v", err)
	}
	defer store2.Close()

	// Verify we can still access the user created by the first store
	user2, err := store2.GetUserByEmail(ctx, "test1@example.com")
	if err != nil {
		t.Fatalf("Failed to get user from second store: %v", err)
	}

	if user1.ID != user2.ID {
		t.Errorf("Expected same user ID %d, got %d", user1.ID, user2.ID)
	}

	// Create another user with the second store
	_, err = store2.CreateUser(ctx, "test2@example.com", "hash2")
	if err != nil {
		t.Fatalf("Failed to create user in second store: %v", err)
	}
}

func TestSQLiteStore_MultipleUsers(t *testing.T) {
	// Use in-memory SQLite database for testing
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create multiple users
	users := []struct {
		email        string
		passwordHash string
	}{
		{"user1@example.com", "hash1"},
		{"user2@example.com", "hash2"},
		{"user3@example.com", "hash3"},
	}

	createdUsers := make([]*User, len(users))
	for i, u := range users {
		user, err := store.CreateUser(ctx, u.email, u.passwordHash)
		if err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
		createdUsers[i] = user
	}

	// Verify all users can be retrieved
	for i, u := range users {
		retrievedUser, err := store.GetUserByEmail(ctx, u.email)
		if err != nil {
			t.Fatalf("Failed to get user %d: %v", i, err)
		}

		if retrievedUser.Email != createdUsers[i].Email {
			t.Errorf("User %d: expected email %s, got %s", i, createdUsers[i].Email, retrievedUser.Email)
		}
		if retrievedUser.PasswordHash != createdUsers[i].PasswordHash {
			t.Errorf("User %d: expected password hash %s, got %s", i, createdUsers[i].PasswordHash, retrievedUser.PasswordHash)
		}
	}
}

func TestSQLiteStore_UpsertAdmin_CreateNew(t *testing.T) {
	// Use in-memory SQLite database for testing
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	email := "admin@example.com"
	passwordHash := "admin_hash_123"

	// Upsert admin user (should create new)
	user, err := store.UpsertAdmin(ctx, email, passwordHash)
	if err != nil {
		t.Fatalf("Failed to upsert admin: %v", err)
	}

	// Verify user was created correctly
	if user.ID == 0 {
		t.Error("Expected user ID to be set")
	}
	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}
	if user.PasswordHash != passwordHash {
		t.Errorf("Expected password hash %s, got %s", passwordHash, user.PasswordHash)
	}
	if user.CreatedAt.IsZero() {
		t.Error("Expected created_at to be set")
	}

	// Verify user can be retrieved
	retrievedUser, err := store.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("Failed to retrieve created admin: %v", err)
	}
	if retrievedUser.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, retrievedUser.ID)
	}
}

func TestSQLiteStore_UpsertAdmin_UpdateExisting(t *testing.T) {
	// Use in-memory SQLite database for testing
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	email := "admin@example.com"
	originalHash := "original_hash"
	newHash := "new_hash_456"

	// Create initial user
	originalUser, err := store.CreateUser(ctx, email, originalHash)
	if err != nil {
		t.Fatalf("Failed to create original user: %v", err)
	}

	// Upsert admin user (should update existing)
	updatedUser, err := store.UpsertAdmin(ctx, email, newHash)
	if err != nil {
		t.Fatalf("Failed to upsert existing admin: %v", err)
	}

	// Verify user was updated correctly
	if updatedUser.ID != originalUser.ID {
		t.Errorf("Expected same ID %d, got %d", originalUser.ID, updatedUser.ID)
	}
	if updatedUser.Email != email {
		t.Errorf("Expected email %s, got %s", email, updatedUser.Email)
	}
	if updatedUser.PasswordHash != newHash {
		t.Errorf("Expected new password hash %s, got %s", newHash, updatedUser.PasswordHash)
	}
	if updatedUser.CreatedAt != originalUser.CreatedAt {
		t.Errorf("Expected same created_at %v, got %v", originalUser.CreatedAt, updatedUser.CreatedAt)
	}

	// Verify updated user can be retrieved with new password
	retrievedUser, err := store.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("Failed to retrieve updated admin: %v", err)
	}
	if retrievedUser.PasswordHash != newHash {
		t.Errorf("Expected new password hash %s, got %s", newHash, retrievedUser.PasswordHash)
	}
}