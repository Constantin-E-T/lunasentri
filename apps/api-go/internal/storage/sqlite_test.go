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
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

func TestSQLiteStore_UpdateUserPassword(t *testing.T) {
    store, err := NewSQLiteStore("file::memory:?cache=shared")
    if err != nil {
        t.Fatalf("Failed to create test store: %v", err)
    }
    defer store.Close()

    ctx := context.Background()
    user, err := store.CreateUser(ctx, "update@example.com", "oldhash")
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }

    if err := store.UpdateUserPassword(ctx, user.ID, "newhash"); err != nil {
        t.Fatalf("UpdateUserPassword failed: %v", err)
    }

    updated, err := store.GetUserByEmail(ctx, "update@example.com")
    if err != nil {
        t.Fatalf("Failed to retrieve user: %v", err)
    }

    if updated.PasswordHash != "newhash" {
        t.Errorf("Expected password hash 'newhash', got %s", updated.PasswordHash)
    }
}

func TestSQLiteStore_PasswordResetLifecycle(t *testing.T) {
    store, err := NewSQLiteStore("file::memory:?cache=shared")
    if err != nil {
        t.Fatalf("Failed to create test store: %v", err)
    }
    defer store.Close()

    ctx := context.Background()
    user, err := store.CreateUser(ctx, "reset@example.com", "hash")
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }

    expiresAt := time.Now().Add(time.Hour)
    pr, err := store.CreatePasswordReset(ctx, user.ID, "tokenhash", expiresAt)
    if err != nil {
        t.Fatalf("CreatePasswordReset failed: %v", err)
    }

    fetched, err := store.GetPasswordResetByHash(ctx, "tokenhash")
    if err != nil {
        t.Fatalf("GetPasswordResetByHash failed: %v", err)
    }
    if fetched.ID != pr.ID {
        t.Errorf("Expected password reset ID %d, got %d", pr.ID, fetched.ID)
    }

    if err := store.MarkPasswordResetUsed(ctx, pr.ID); err != nil {
        t.Fatalf("MarkPasswordResetUsed failed: %v", err)
    }

    if _, err := store.GetPasswordResetByHash(ctx, "tokenhash"); err == nil {
        t.Fatal("Expected error when fetching used password reset token")
    }
}

func TestSQLiteStore_PasswordResetExpiry(t *testing.T) {
    store, err := NewSQLiteStore("file::memory:?cache=shared")
    if err != nil {
        t.Fatalf("Failed to create test store: %v", err)
    }
    defer store.Close()

    ctx := context.Background()
    user, err := store.CreateUser(ctx, "expired@example.com", "hash")
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }

    expiresAt := time.Now().Add(-time.Hour)
    _, err = store.CreatePasswordReset(ctx, user.ID, "expired-token", expiresAt)
    if err != nil {
        t.Fatalf("CreatePasswordReset failed: %v", err)
    }

    if _, err := store.GetPasswordResetByHash(ctx, "expired-token"); err == nil {
        t.Fatal("Expected error for expired token")
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

func TestSQLiteStore_ListUsers(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create multiple users
	_, err = store.CreateUser(ctx, "charlie@example.com", "hash3")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	_, err = store.CreateUser(ctx, "alice@example.com", "hash1")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	_, err = store.CreateUser(ctx, "bob@example.com", "hash2")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// List all users
	users, err := store.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	// Verify count
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Verify ordering by email (alphabetical)
	if users[0].Email != "alice@example.com" {
		t.Errorf("Expected first user to be alice@example.com, got %s", users[0].Email)
	}
	if users[1].Email != "bob@example.com" {
		t.Errorf("Expected second user to be bob@example.com, got %s", users[1].Email)
	}
	if users[2].Email != "charlie@example.com" {
		t.Errorf("Expected third user to be charlie@example.com, got %s", users[2].Email)
	}
}

func TestSQLiteStore_ListUsers_Empty(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// List users when none exist
	users, err := store.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	// Verify empty list
	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}
}

func TestSQLiteStore_DeleteUser(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create multiple users
	user1, err := store.CreateUser(ctx, "user1@example.com", "hash1")
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	user2, err := store.CreateUser(ctx, "user2@example.com", "hash2")
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Delete user1
	err = store.DeleteUser(ctx, user1.ID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	// Verify user1 is gone
	_, err = store.GetUserByEmail(ctx, "user1@example.com")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound for deleted user, got: %v", err)
	}

	// Verify user2 still exists
	retrievedUser, err := store.GetUserByEmail(ctx, "user2@example.com")
	if err != nil {
		t.Fatalf("Failed to get user2: %v", err)
	}
	if retrievedUser.ID != user2.ID {
		t.Errorf("Expected user2 ID %d, got %d", user2.ID, retrievedUser.ID)
	}
}

func TestSQLiteStore_DeleteUser_NotFound(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create a user so we have at least one
	_, err = store.CreateUser(ctx, "user@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Try to delete non-existent user
	err = store.DeleteUser(ctx, 9999)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

func TestSQLiteStore_DeleteUser_LastUser(t *testing.T) {
	store, err := NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create only one user
	user, err := store.CreateUser(ctx, "lastuser@example.com", "hash")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Try to delete the last user (should fail)
	err = store.DeleteUser(ctx, user.ID)
	if err == nil {
		t.Fatal("Expected error when deleting last user")
	}
	if err.Error() != "cannot delete the last user" {
		t.Errorf("Expected 'cannot delete the last user' error, got: %v", err)
	}

	// Verify user still exists
	users, err := store.ListUsers(ctx)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user to remain, got %d", len(users))
	}
}
