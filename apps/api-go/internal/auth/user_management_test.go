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

	// Create two users (user1 will be admin as first user)
	user1, _, err := service.CreateUser(ctx, "user1@example.com", "password1")
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	user2, _, err := service.CreateUser(ctx, "user2@example.com", "password2")
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Delete user2 (non-admin) as user1 (admin)
	err = service.DeleteUser(ctx, user2.ID, user1.ID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	// Verify user2 is gone
	users, err := service.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user remaining, got %d", len(users))
	}
	if users[0].ID != user1.ID {
		t.Errorf("Expected user1 to remain, got user ID %d", users[0].ID)
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

// Note: The TestDeleteUser_CannotDeleteLastAdmin test is defined below
// and covers the scenario of preventing last admin deletion

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

func TestCreateUser_FirstUserIsAdmin(t *testing.T) {
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
	firstUser, _, err := service.CreateUser(ctx, "first@example.com", "password")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Verify first user is admin
	if !firstUser.IsAdmin {
		t.Error("Expected first user to be admin")
	}

	// Create second user
	secondUser, _, err := service.CreateUser(ctx, "second@example.com", "password")
	if err != nil {
		t.Fatalf("CreateUser failed for second user: %v", err)
	}

	// Verify second user is not admin
	if secondUser.IsAdmin {
		t.Error("Expected second user to not be admin")
	}
}

func TestDeleteUser_CannotDeleteLastAdmin(t *testing.T) {
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

	// Create admin user (first user)
	admin, _, err := service.CreateUser(ctx, "admin@example.com", "password")
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	// Verify admin is actually admin
	if !admin.IsAdmin {
		t.Fatal("Expected first user to be admin")
	}

	// Create regular user (second user)
	regularUser, _, err := service.CreateUser(ctx, "user@example.com", "password")
	if err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	// Try to delete the only admin as regular user
	err = service.DeleteUser(ctx, admin.ID, regularUser.ID)
	if err == nil {
		t.Fatal("Expected error when deleting last admin")
	}
	if !strings.Contains(err.Error(), "cannot delete the last admin") {
		t.Errorf("Expected 'cannot delete the last admin' error, got: %v", err)
	}

	// Verify admin still exists
	users, err := service.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestDeleteUser_CanDeleteAdminWithMultipleAdmins(t *testing.T) {
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

	// Create first admin
	admin1, _, err := service.CreateUser(ctx, "admin1@example.com", "password")
	if err != nil {
		t.Fatalf("Failed to create admin1: %v", err)
	}

	// Create second user and promote to admin
	admin2, _, err := service.CreateUser(ctx, "admin2@example.com", "password")
	if err != nil {
		t.Fatalf("Failed to create admin2: %v", err)
	}
	err = store.PromoteToAdmin(ctx, admin2.ID)
	if err != nil {
		t.Fatalf("Failed to promote admin2: %v", err)
	}

	// Delete first admin (should succeed since there's another admin)
	err = service.DeleteUser(ctx, admin1.ID, admin2.ID)
	if err != nil {
		t.Fatalf("Should be able to delete admin when another admin exists: %v", err)
	}

	// Verify admin1 is gone
	users, err := service.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user remaining, got %d", len(users))
	}
	if users[0].ID != admin2.ID {
		t.Error("Expected admin2 to remain")
	}
}

func TestChangePassword_Success(t *testing.T) {
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

	// Create a user
	currentPassword := "oldpassword123"
	user, _, err := service.CreateUser(ctx, "user@example.com", currentPassword)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Change password
	newPassword := "newpassword456"
	err = service.ChangePassword(ctx, user.ID, currentPassword, newPassword)
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	// Verify old password no longer works
	updatedUser, err := service.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if err := VerifyPassword(updatedUser.PasswordHash, currentPassword); err == nil {
		t.Error("Old password should not work after change")
	}

	// Verify new password works
	if err := VerifyPassword(updatedUser.PasswordHash, newPassword); err != nil {
		t.Error("New password should work after change")
	}
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
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

	// Create a user
	currentPassword := "oldpassword123"
	user, _, err := service.CreateUser(ctx, "user@example.com", currentPassword)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Try to change password with wrong current password
	err = service.ChangePassword(ctx, user.ID, "wrongpassword", "newpassword456")
	if err == nil {
		t.Fatal("Expected error for wrong current password")
	}
	if !strings.Contains(err.Error(), "current password is incorrect") {
		t.Errorf("Expected 'current password is incorrect' error, got: %v", err)
	}

	// Verify password hasn't changed
	updatedUser, err := service.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if err := VerifyPassword(updatedUser.PasswordHash, currentPassword); err != nil {
		t.Error("Original password should still work")
	}
}

func TestChangePassword_WeakNewPassword(t *testing.T) {
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

	// Create a user
	currentPassword := "oldpassword123"
	user, _, err := service.CreateUser(ctx, "user@example.com", currentPassword)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Try to change password to a weak password (< 8 characters)
	err = service.ChangePassword(ctx, user.ID, currentPassword, "weak")
	if err == nil {
		t.Fatal("Expected error for weak password")
	}
	if !strings.Contains(err.Error(), "must be at least 8 characters") {
		t.Errorf("Expected 'must be at least 8 characters' error, got: %v", err)
	}

	// Verify password hasn't changed
	updatedUser, err := service.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if err := VerifyPassword(updatedUser.PasswordHash, currentPassword); err != nil {
		t.Error("Original password should still work")
	}
}

func TestChangePassword_NonexistentUser(t *testing.T) {
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

	// Try to change password for non-existent user
	err = service.ChangePassword(ctx, 9999, "oldpassword", "newpassword123")
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}
	if !strings.Contains(err.Error(), "user not found") {
		t.Errorf("Expected 'user not found' error, got: %v", err)
	}
}

func TestChangePassword_EmptyCurrentPassword(t *testing.T) {
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

	// Create a user
	user, _, err := service.CreateUser(ctx, "user@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Try to change password with empty current password
	err = service.ChangePassword(ctx, user.ID, "", "newpassword123")
	if err == nil {
		t.Fatal("Expected error for empty current password")
	}
	if !strings.Contains(err.Error(), "current password cannot be empty") {
		t.Errorf("Expected 'current password cannot be empty' error, got: %v", err)
	}
}

func TestChangePassword_EmptyNewPassword(t *testing.T) {
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

	// Create a user
	currentPassword := "password123"
	user, _, err := service.CreateUser(ctx, "user@example.com", currentPassword)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Try to change password with empty new password
	err = service.ChangePassword(ctx, user.ID, currentPassword, "")
	if err == nil {
		t.Fatal("Expected error for empty new password")
	}
	if !strings.Contains(err.Error(), "new password cannot be empty") {
		t.Errorf("Expected 'new password cannot be empty' error, got: %v", err)
	}
}
