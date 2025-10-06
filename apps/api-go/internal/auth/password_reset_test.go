package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

func TestGeneratePasswordReset_Success(t *testing.T) {
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

	// Create a test user
	passwordHash, _ := HashPassword("oldpassword")
	user, err := store.CreateUser(ctx, "test@example.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Generate password reset token
	token, err := service.GeneratePasswordReset(ctx, user.Email, 1*time.Hour)
	if err != nil {
		t.Fatalf("GeneratePasswordReset failed: %v", err)
	}

	// Token should not be empty
	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Token should be base64-encoded (at least 32 bytes worth)
	if len(token) < 40 {
		t.Errorf("Token too short: %d characters", len(token))
	}
}

func TestGeneratePasswordReset_NonExistentUser(t *testing.T) {
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

	// Try to generate reset token for non-existent user
	token, err := service.GeneratePasswordReset(ctx, "nonexistent@example.com", 1*time.Hour)

	// Should not return an error (to avoid leaking user existence)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should still return a token (fake one)
	if token == "" {
		t.Error("Expected non-empty token even for non-existent user")
	}
}

func TestGeneratePasswordReset_EmptyEmail(t *testing.T) {
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

	// Try with empty email
	_, err = service.GeneratePasswordReset(ctx, "", 1*time.Hour)
	if err == nil {
		t.Error("Expected error for empty email")
	}
	if !strings.Contains(err.Error(), "email cannot be empty") {
		t.Errorf("Expected 'email cannot be empty' error, got: %v", err)
	}
}

func TestResetPassword_Success(t *testing.T) {
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

	// Create a test user
	oldPasswordHash, _ := HashPassword("oldpassword")
	user, err := store.CreateUser(ctx, "reset@example.com", oldPasswordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Generate password reset token
	token, err := service.GeneratePasswordReset(ctx, user.Email, 1*time.Hour)
	if err != nil {
		t.Fatalf("GeneratePasswordReset failed: %v", err)
	}

	// Reset the password
	newPassword := "newpassword123"
	err = service.ResetPassword(ctx, token, newPassword)
	if err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}

	// Verify the password was updated
	updatedUser, err := store.GetUserByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	// Verify new password works
	if err := VerifyPassword(updatedUser.PasswordHash, newPassword); err != nil {
		t.Error("New password verification failed")
	}

	// Verify old password no longer works
	if err := VerifyPassword(updatedUser.PasswordHash, "oldpassword"); err == nil {
		t.Error("Old password should not work after reset")
	}
}

func TestResetPassword_TokenReuse(t *testing.T) {
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

	// Create a test user
	passwordHash, _ := HashPassword("password")
	user, err := store.CreateUser(ctx, "reuse@example.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Generate password reset token
	token, err := service.GeneratePasswordReset(ctx, user.Email, 1*time.Hour)
	if err != nil {
		t.Fatalf("GeneratePasswordReset failed: %v", err)
	}

	// Reset password first time
	err = service.ResetPassword(ctx, token, "newpassword1")
	if err != nil {
		t.Fatalf("First ResetPassword failed: %v", err)
	}

	// Try to reuse the same token
	err = service.ResetPassword(ctx, token, "newpassword2")
	if err == nil {
		t.Error("Expected error when reusing token")
	}
	if !strings.Contains(err.Error(), "invalid or expired") {
		t.Errorf("Expected 'invalid or expired' error, got: %v", err)
	}
}

func TestResetPassword_ExpiredToken(t *testing.T) {
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

	// Create a test user
	passwordHash, _ := HashPassword("password")
	user, err := store.CreateUser(ctx, "expired@example.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Generate password reset token with negative TTL (already expired)
	token, err := service.GeneratePasswordReset(ctx, user.Email, -1*time.Hour)
	if err != nil {
		t.Fatalf("GeneratePasswordReset failed: %v", err)
	}

	// Try to use the expired token
	err = service.ResetPassword(ctx, token, "newpassword")
	if err == nil {
		t.Error("Expected error for expired token")
	}
	if !strings.Contains(err.Error(), "invalid or expired") {
		t.Errorf("Expected 'invalid or expired' error, got: %v", err)
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
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

	// Try to reset password with a completely invalid token
	err = service.ResetPassword(ctx, "invalid-token-12345", "newpassword")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestResetPassword_WeakPassword(t *testing.T) {
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

	// Create a test user
	passwordHash, _ := HashPassword("password")
	user, err := store.CreateUser(ctx, "weak@example.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Generate password reset token
	token, err := service.GeneratePasswordReset(ctx, user.Email, 1*time.Hour)
	if err != nil {
		t.Fatalf("GeneratePasswordReset failed: %v", err)
	}

	// Try to reset with weak password (less than 8 characters)
	err = service.ResetPassword(ctx, token, "weak")
	if err == nil {
		t.Error("Expected error for weak password")
	}
	if !strings.Contains(err.Error(), "at least 8 characters") {
		t.Errorf("Expected password length error, got: %v", err)
	}
}

func TestResetPassword_EmptyToken(t *testing.T) {
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

	// Try with empty token
	err = service.ResetPassword(ctx, "", "newpassword")
	if err == nil {
		t.Error("Expected error for empty token")
	}
	if !strings.Contains(err.Error(), "token cannot be empty") {
		t.Errorf("Expected 'token cannot be empty' error, got: %v", err)
	}
}

func TestResetPassword_EmptyPassword(t *testing.T) {
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

	// Create a test user and generate token
	passwordHash, _ := HashPassword("password")
	user, err := store.CreateUser(ctx, "empty@example.com", passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	token, err := service.GeneratePasswordReset(ctx, user.Email, 1*time.Hour)
	if err != nil {
		t.Fatalf("GeneratePasswordReset failed: %v", err)
	}

	// Try with empty password
	err = service.ResetPassword(ctx, token, "")
	if err == nil {
		t.Error("Expected error for empty password")
	}
	if !strings.Contains(err.Error(), "password cannot be empty") {
		t.Errorf("Expected 'password cannot be empty' error, got: %v", err)
	}
}

func TestGenerateSecureToken(t *testing.T) {
	// Generate multiple tokens
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateSecureToken()
		if err != nil {
			t.Fatalf("generateSecureToken failed: %v", err)
		}

		// Check token is not empty
		if token == "" {
			t.Error("Generated empty token")
		}

		// Check token has reasonable length (base64 of 32 bytes)
		if len(token) < 40 {
			t.Errorf("Token too short: %d characters", len(token))
		}

		// Check for uniqueness
		if tokens[token] {
			t.Errorf("Generated duplicate token: %s", token)
		}
		tokens[token] = true
	}
}

func TestHashToken(t *testing.T) {
	token := "test-token-12345"

	// Hash the token
	hash, err := hashToken(token)
	if err != nil {
		t.Fatalf("hashToken failed: %v", err)
	}

	// Hash should not be empty
	if hash == "" {
		t.Error("Generated empty hash")
	}

	// Hash should be different from the original token
	if hash == token {
		t.Error("Hash should differ from token")
	}

	// Hashing the same token again should produce the same hash (SHA256 is deterministic)
	hash2, err := hashToken(token)
	if err != nil {
		t.Fatalf("hashToken failed on second call: %v", err)
	}

	if hash != hash2 {
		t.Error("Expected same hash for deterministic SHA256")
	}

	// Different tokens should produce different hashes
	hash3, err := hashToken("different-token")
	if err != nil {
		t.Fatalf("hashToken failed for different token: %v", err)
	}

	if hash == hash3 {
		t.Error("Different tokens should produce different hashes")
	}
}

func TestHashToken_EmptyToken(t *testing.T) {
	_, err := hashToken("")
	if err == nil {
		t.Error("Expected error for empty token")
	}
	if !strings.Contains(err.Error(), "token cannot be empty") {
		t.Errorf("Expected 'token cannot be empty' error, got: %v", err)
	}
}
