package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

func TestCreateUser_Success(t *testing.T) {
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	service, err := NewService(store, "test-secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()

	// Create user with provided password
	user, tempPassword, err := service.CreateUser(ctx, "test@example.com", "mypassword")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Verify user was created
	if user.ID == 0 {
		t.Error("Expected user ID to be set")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}
	if tempPassword != "" {
		t.Error("Expected no temp password when password was provided")
	}

	// Verify password works
	if err := VerifyPassword(user.PasswordHash, "mypassword"); err != nil {
		t.Error("Password verification failed")
	}
}

func TestCreateUser_GeneratedPassword(t *testing.T) {
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	service, err := NewService(store, "test-secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()

	// Create user without password (should generate temp password)
	user, tempPassword, err := service.CreateUser(ctx, "test@example.com", "")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Verify temp password was generated
	if tempPassword == "" {
		t.Fatal("Expected temp password to be generated")
	}
	if len(tempPassword) < 40 {
		t.Errorf("Temp password too short: %d characters", len(tempPassword))
	}

	// Verify temp password works
	if err := VerifyPassword(user.PasswordHash, tempPassword); err != nil {
		t.Error("Temp password verification failed")
	}
}

func TestCreateUser_EmptyEmail(t *testing.T) {
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	service, err := NewService(store, "test-secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()

	// Try to create user with empty email
	_, _, err = service.CreateUser(ctx, "", "password")
	if err == nil {
		t.Fatal("Expected error for empty email")
	}
	if !strings.Contains(err.Error(), "email cannot be empty") {
		t.Errorf("Expected 'email cannot be empty' error, got: %v", err)
	}
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	service, err := NewService(store, "test-secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()

	// Try to create user with invalid email (no @)
	_, _, err = service.CreateUser(ctx, "notanemail", "password")
	if err == nil {
		t.Fatal("Expected error for invalid email")
	}
	if !strings.Contains(err.Error(), "invalid email format") {
		t.Errorf("Expected 'invalid email format' error, got: %v", err)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	service, err := NewService(store, "test-secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()

	// Create first user
	_, _, err = service.CreateUser(ctx, "test@example.com", "password1")
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	// Try to create user with same email
	_, _, err = service.CreateUser(ctx, "test@example.com", "password2")
	if err == nil {
		t.Fatal("Expected error for duplicate email")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %v", err)
	}
}

func TestListUsers(t *testing.T) {
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	service, err := NewService(store, "test-secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()

	// Create multiple users
	_, _, err = service.CreateUser(ctx, "user1@example.com", "password1")
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	_, _, err = service.CreateUser(ctx, "user2@example.com", "password2")
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// List users
	users, err := service.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	// Verify count
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	// Verify ordering (alphabetical by email)
	if users[0].Email != "user1@example.com" {
		t.Errorf("Expected first user to be user1@example.com, got %s", users[0].Email)
	}
	if users[1].Email != "user2@example.com" {
		t.Errorf("Expected second user to be user2@example.com, got %s", users[1].Email)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	service, err := NewService(store, "test-secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()

	// Create two users
	user1, _, err := service.CreateUser(ctx, "user1@example.com", "password1")
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	user2, _, err := service.CreateUser(ctx, "user2@example.com", "password2")
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Delete user1 as user2
	err = service.DeleteUser(ctx, user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	// Verify user1 is gone
	users, err := service.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user remaining, got %d", len(users))
	}
	if users[0].ID != user2.ID {
		t.Errorf("Expected user2 to remain, got user ID %d", users[0].ID)
	}
}

func TestDeleteUser_CannotDeleteSelf(t *testing.T) {
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	service, err := NewService(store, "test-secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()

	// Create user
	user, _, err := service.CreateUser(ctx, "user@example.com", "password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create second user so we're not the last user
	_, _, err = service.CreateUser(ctx, "other@example.com", "password")
	if err != nil {
		t.Fatalf("Failed to create other user: %v", err)
	}

	// Try to delete self
	err = service.DeleteUser(ctx, user.ID, user.ID)
	if err == nil {
		t.Fatal("Expected error when deleting self")
	}
	if !strings.Contains(err.Error(), "cannot delete your own account") {
		t.Errorf("Expected 'cannot delete your own account' error, got: %v", err)
	}
}

func TestDeleteUser_CannotDeleteLastUser(t *testing.T) {
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	service, err := NewService(store, "test-secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()

	// Create only one user
	user, _, err := service.CreateUser(ctx, "lastuser@example.com", "password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create second user to act as admin
	admin, _, err := service.CreateUser(ctx, "admin@example.com", "password")
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	// Delete the first user (now admin is last)
	err = service.DeleteUser(ctx, user.ID, admin.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Try to delete the last remaining user
	err = service.DeleteUser(ctx, admin.ID, 999) // Use different current user ID to bypass self-check
	if err == nil {
		t.Fatal("Expected error when deleting last user")
	}
	if !strings.Contains(err.Error(), "cannot delete the last remaining user") {
		t.Errorf("Expected 'cannot delete the last remaining user' error, got: %v", err)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	service, err := NewService(store, "test-secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()

	// Create a user so we have at least one
	currentUser, _, err := service.CreateUser(ctx, "current@example.com", "password")
	if err != nil {
		t.Fatalf("Failed to create current user: %v", err)
	}

	// Try to delete non-existent user
	err = service.DeleteUser(ctx, 9999, currentUser.ID)
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}
	if !strings.Contains(err.Error(), "user not found") {
		t.Errorf("Expected 'user not found' error, got: %v", err)
	}
}

func TestGenerateTempPassword(t *testing.T) {
	// Generate multiple passwords
	passwords := make(map[string]bool)
	for i := 0; i < 100; i++ {
		password, err := generateTempPassword()
		if err != nil {
			t.Fatalf("generateTempPassword failed: %v", err)
		}

		// Check password is not empty
		if password == "" {
			t.Error("Generated empty password")
		}

		// Check password has reasonable length
		if len(password) < 40 {
			t.Errorf("Password too short: %d characters", len(password))
		}

		// Check for uniqueness
		if passwords[password] {
			t.Errorf("Generated duplicate password: %s", password)
		}
		passwords[password] = true
	}
}
