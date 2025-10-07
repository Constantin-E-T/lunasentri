package auth

import (
	"context"
	"os"
	"testing"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// TestBootstrapIntegration tests the complete bootstrap flow with real storage
func TestBootstrapIntegration(t *testing.T) {
	// Use in-memory SQLite database for testing (real store)
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	// Save and restore original environment variables
	originalEmail := os.Getenv("ADMIN_EMAIL")
	originalPassword := os.Getenv("ADMIN_PASSWORD")
	defer func() {
		os.Setenv("ADMIN_EMAIL", originalEmail)
		os.Setenv("ADMIN_PASSWORD", originalPassword)
	}()

	ctx := context.Background()
	testEmail := "integration@test.com"
	testPassword := "integration_password_123"

	// Test 1: Bootstrap creates new admin user
	t.Run("CreateNewAdmin", func(t *testing.T) {
		os.Setenv("ADMIN_EMAIL", testEmail)
		os.Setenv("ADMIN_PASSWORD", testPassword)

		// Bootstrap should create new user
		err := BootstrapAdmin(ctx, store)
		if err != nil {
			t.Fatalf("BootstrapAdmin failed: %v", err)
		}

		// Verify user exists and password works
		user, err := store.GetUserByEmail(ctx, testEmail)
		if err != nil {
			t.Fatalf("Failed to retrieve admin user: %v", err)
		}

		if user.Email != testEmail {
			t.Errorf("Expected email %s, got %s", testEmail, user.Email)
		}

		// Verify password verification works
		err = VerifyPassword(user.PasswordHash, testPassword)
		if err != nil {
			t.Errorf("Password verification failed: %v", err)
		}
	})

	// Test 2: Bootstrap updates existing admin user
	t.Run("UpdateExistingAdmin", func(t *testing.T) {
		newPassword := "updated_password_456"
		os.Setenv("ADMIN_EMAIL", testEmail)
		os.Setenv("ADMIN_PASSWORD", newPassword)

		// Get original user ID
		originalUser, err := store.GetUserByEmail(ctx, testEmail)
		if err != nil {
			t.Fatalf("Failed to get original user: %v", err)
		}

		// Bootstrap should update existing user
		err = BootstrapAdmin(ctx, store)
		if err != nil {
			t.Fatalf("BootstrapAdmin update failed: %v", err)
		}

		// Verify user still exists with same ID but updated password
		updatedUser, err := store.GetUserByEmail(ctx, testEmail)
		if err != nil {
			t.Fatalf("Failed to retrieve updated admin user: %v", err)
		}

		if updatedUser.ID != originalUser.ID {
			t.Errorf("User ID should not change on update: original=%d, updated=%d",
				originalUser.ID, updatedUser.ID)
		}

		// Verify new password works
		err = VerifyPassword(updatedUser.PasswordHash, newPassword)
		if err != nil {
			t.Errorf("New password verification failed: %v", err)
		}

		// Verify old password no longer works
		err = VerifyPassword(updatedUser.PasswordHash, testPassword)
		if err == nil {
			t.Error("Old password should not work after update")
		}
	})

	// Test 3: Bootstrap is idempotent
	t.Run("IdempotentBootstrap", func(t *testing.T) {
		password := "idempotent_password_789"
		os.Setenv("ADMIN_EMAIL", testEmail)
		os.Setenv("ADMIN_PASSWORD", password)

		// Run bootstrap multiple times
		for i := 0; i < 3; i++ {
			err := BootstrapAdmin(ctx, store)
			if err != nil {
				t.Fatalf("BootstrapAdmin iteration %d failed: %v", i+1, err)
			}
		}

		// Verify user still exists and password works
		user, err := store.GetUserByEmail(ctx, testEmail)
		if err != nil {
			t.Fatalf("Failed to retrieve admin user after multiple bootstraps: %v", err)
		}

		err = VerifyPassword(user.PasswordHash, password)
		if err != nil {
			t.Errorf("Password verification failed after multiple bootstraps: %v", err)
		}
	})

	// Test 4: Bootstrap with different email creates separate user
	t.Run("DifferentEmailCreatesSeparateUser", func(t *testing.T) {
		secondEmail := "second@test.com"
		secondPassword := "second_password_123"
		os.Setenv("ADMIN_EMAIL", secondEmail)
		os.Setenv("ADMIN_PASSWORD", secondPassword)

		// Bootstrap second admin
		err := BootstrapAdmin(ctx, store)
		if err != nil {
			t.Fatalf("BootstrapAdmin for second user failed: %v", err)
		}

		// Verify both users exist
		firstUser, err := store.GetUserByEmail(ctx, testEmail)
		if err != nil {
			t.Fatalf("First user should still exist: %v", err)
		}

		secondUser, err := store.GetUserByEmail(ctx, secondEmail)
		if err != nil {
			t.Fatalf("Second user should exist: %v", err)
		}

		// Verify they have different IDs
		if firstUser.ID == secondUser.ID {
			t.Error("Different users should have different IDs")
		}

		// Verify both passwords work
		err = VerifyPassword(firstUser.PasswordHash, "idempotent_password_789")
		if err != nil {
			t.Errorf("First user password verification failed: %v", err)
		}

		err = VerifyPassword(secondUser.PasswordHash, secondPassword)
		if err != nil {
			t.Errorf("Second user password verification failed: %v", err)
		}
	})
}
