package auth

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

func TestBootstrapAdmin_Success(t *testing.T) {
	// Use in-memory SQLite database for testing
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	// Set environment variables
	originalEmail := os.Getenv("ADMIN_EMAIL")
	originalPassword := os.Getenv("ADMIN_PASSWORD")
	defer func() {
		os.Setenv("ADMIN_EMAIL", originalEmail)
		os.Setenv("ADMIN_PASSWORD", originalPassword)
	}()

	testEmail := "admin@test.com"
	testPassword := "secure_password_123"
	os.Setenv("ADMIN_EMAIL", testEmail)
	os.Setenv("ADMIN_PASSWORD", testPassword)

	ctx := context.Background()

	// Bootstrap admin user
	err = BootstrapAdmin(ctx, store)
	if err != nil {
		t.Fatalf("BootstrapAdmin failed: %v", err)
	}

	// Verify user was created
	user, err := store.GetUserByEmail(ctx, testEmail)
	if err != nil {
		t.Fatalf("Failed to retrieve admin user: %v", err)
	}

	if user.Email != testEmail {
		t.Errorf("Expected email %s, got %s", testEmail, user.Email)
	}

	// Verify password hash works
	err = VerifyPassword(user.PasswordHash, testPassword)
	if err != nil {
		t.Errorf("Password verification failed: %v", err)
	}
}

func TestBootstrapAdmin_UpdateExisting(t *testing.T) {
	// Use in-memory SQLite database for testing
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	testEmail := "admin@test.com"
	originalPassword := "original_password"
	newPassword := "new_password_456"

	// Create initial user with original password
	originalHash, err := HashPassword(originalPassword)
	if err != nil {
		t.Fatalf("Failed to hash original password: %v", err)
	}

	_, err = store.CreateUser(ctx, testEmail, originalHash)
	if err != nil {
		t.Fatalf("Failed to create initial user: %v", err)
	}

	// Set environment variables for bootstrap
	originalEmail := os.Getenv("ADMIN_EMAIL")
	originalEnvPassword := os.Getenv("ADMIN_PASSWORD")
	defer func() {
		os.Setenv("ADMIN_EMAIL", originalEmail)
		os.Setenv("ADMIN_PASSWORD", originalEnvPassword)
	}()

	os.Setenv("ADMIN_EMAIL", testEmail)
	os.Setenv("ADMIN_PASSWORD", newPassword)

	// Bootstrap admin user (should update existing)
	err = BootstrapAdmin(ctx, store)
	if err != nil {
		t.Fatalf("BootstrapAdmin failed: %v", err)
	}

	// Verify user was updated
	user, err := store.GetUserByEmail(ctx, testEmail)
	if err != nil {
		t.Fatalf("Failed to retrieve updated admin user: %v", err)
	}

	// Verify new password works
	err = VerifyPassword(user.PasswordHash, newPassword)
	if err != nil {
		t.Errorf("New password verification failed: %v", err)
	}

	// Verify old password no longer works
	err = VerifyPassword(user.PasswordHash, originalPassword)
	if err == nil {
		t.Error("Old password should not work after update")
	}
}

func TestBootstrapAdmin_SkippedWhenEnvVarsNotSet(t *testing.T) {
	// Use in-memory SQLite database for testing
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	// Save original env vars
	originalEmail := os.Getenv("ADMIN_EMAIL")
	originalPassword := os.Getenv("ADMIN_PASSWORD")
	defer func() {
		os.Setenv("ADMIN_EMAIL", originalEmail)
		os.Setenv("ADMIN_PASSWORD", originalPassword)
	}()

	testCases := []struct {
		name     string
		email    string
		password string
	}{
		{"both unset", "", ""},
		{"email unset", "", "password"},
		{"password unset", "admin@test.com", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("ADMIN_EMAIL", tc.email)
			os.Setenv("ADMIN_PASSWORD", tc.password)

			ctx := context.Background()

			// Bootstrap should not fail but should skip
			err = BootstrapAdmin(ctx, store)
			if err != nil {
				t.Errorf("BootstrapAdmin should not fail when env vars not set: %v", err)
			}

			// Verify no user was created (if email was provided)
			if tc.email != "" {
				_, err = store.GetUserByEmail(ctx, tc.email)
				if err == nil {
					t.Error("User should not have been created when bootstrap is skipped")
				}
				if !strings.Contains(err.Error(), "not found") {
					t.Errorf("Expected 'not found' error, got: %v", err)
				}
			}
		})
	}
}

func TestBootstrapAdmin_ErrorHandling(t *testing.T) {
	// Use in-memory SQLite database for testing
	store, err := storage.NewSQLiteStore("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	defer store.Close()

	// Save original env vars
	originalEmail := os.Getenv("ADMIN_EMAIL")
	originalPassword := os.Getenv("ADMIN_PASSWORD")
	defer func() {
		os.Setenv("ADMIN_EMAIL", originalEmail)
		os.Setenv("ADMIN_PASSWORD", originalPassword)
	}()

	ctx := context.Background()

	// Test with empty password (edge case that shouldn't happen but worth testing)
	os.Setenv("ADMIN_EMAIL", "admin@test.com")
	os.Setenv("ADMIN_PASSWORD", "")

	err = BootstrapAdmin(ctx, store)
	// This should actually skip since password is empty, not error
	if err != nil {
		t.Errorf("Bootstrap should skip when password is empty, got error: %v", err)
	}
}