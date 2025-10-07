package auth

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "test_password_123"

	// Test successful hashing
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Hash should not be empty
	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	// Hash should start with bcrypt prefix
	if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") && !strings.HasPrefix(hash, "$2y$") {
		t.Errorf("Hash doesn't appear to be bcrypt format: %s", hash[:10])
	}

	// Hash should be different from original password
	if hash == password {
		t.Error("Hash should be different from original password")
	}

	// Multiple hashes of same password should be different (due to salt)
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword failed: %v", err)
	}
	if hash == hash2 {
		t.Error("Multiple hashes of same password should be different")
	}
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	_, err := HashPassword("")
	if err == nil {
		t.Error("Expected error for empty password")
	}
	if !strings.Contains(err.Error(), "password cannot be empty") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "test_password_123"
	wrongPassword := "wrong_password"

	// Create hash
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test correct password verification
	err = VerifyPassword(hash, password)
	if err != nil {
		t.Errorf("VerifyPassword failed for correct password: %v", err)
	}

	// Test wrong password verification
	err = VerifyPassword(hash, wrongPassword)
	if err == nil {
		t.Error("Expected error for wrong password")
	}
	if !strings.Contains(err.Error(), "invalid password") {
		t.Errorf("Expected specific error message for wrong password, got: %v", err)
	}
}

func TestVerifyPassword_EmptyInputs(t *testing.T) {
	testCases := []struct {
		name     string
		hash     string
		password string
		wantErr  string
	}{
		{
			name:     "empty hash",
			hash:     "",
			password: "password",
			wantErr:  "hash cannot be empty",
		},
		{
			name:     "empty password",
			hash:     "$2a$12$test",
			password: "",
			wantErr:  "password cannot be empty",
		},
		{
			name:     "both empty",
			hash:     "",
			password: "",
			wantErr:  "hash cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := VerifyPassword(tc.hash, tc.password)
			if err == nil {
				t.Error("Expected error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Expected error containing '%s', got: %v", tc.wantErr, err)
			}
		})
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	err := VerifyPassword("invalid_hash", "password")
	if err == nil {
		t.Error("Expected error for invalid hash")
	}
	if !strings.Contains(err.Error(), "failed to verify password") {
		t.Errorf("Expected specific error message for invalid hash, got: %v", err)
	}
}

func TestPasswordHashAndVerifyIntegration(t *testing.T) {
	testPasswords := []string{
		"simple",
		"complex_P@ssw0rd!",
		"with spaces and symbols !@#$%^&*()",
		"very_long_password_with_many_characters_1234567890",
		"unicode_密码_🔐",
	}

	for _, password := range testPasswords {
		t.Run("password_"+password[:min(10, len(password))], func(t *testing.T) {
			// Hash the password
			hash, err := HashPassword(password)
			if err != nil {
				t.Fatalf("HashPassword failed: %v", err)
			}

			// Verify correct password
			err = VerifyPassword(hash, password)
			if err != nil {
				t.Errorf("VerifyPassword failed for correct password: %v", err)
			}

			// Verify wrong password fails
			err = VerifyPassword(hash, password+"_wrong")
			if err == nil {
				t.Error("Expected verification to fail for wrong password")
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
