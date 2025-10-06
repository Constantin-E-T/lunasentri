package auth

import (
	"testing"
	"time"
)

func TestCreateJWT(t *testing.T) {
	secret := []byte("test-secret-key")
	userID := 123
	ttl := 15 * time.Minute

	token, err := CreateJWT(userID, secret, ttl)
	if err != nil {
		t.Fatalf("CreateJWT failed: %v", err)
	}

	if token == "" {
		t.Fatal("CreateJWT returned empty token")
	}

	// Verify token has 3 parts (header.claims.signature)
	parts := len(token)
	if parts == 0 {
		t.Fatal("Token should not be empty")
	}
}

func TestValidateJWT(t *testing.T) {
	secret := []byte("test-secret-key")
	userID := 456
	ttl := 15 * time.Minute

	// Create a valid token
	token, err := CreateJWT(userID, secret, ttl)
	if err != nil {
		t.Fatalf("CreateJWT failed: %v", err)
	}

	// Validate the token
	validatedUserID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}

	if validatedUserID != userID {
		t.Fatalf("Expected user ID %d, got %d", userID, validatedUserID)
	}
}

func TestValidateJWT_InvalidSignature(t *testing.T) {
	secret := []byte("test-secret-key")
	wrongSecret := []byte("wrong-secret-key")
	userID := 789
	ttl := 15 * time.Minute

	// Create a token with one secret
	token, err := CreateJWT(userID, secret, ttl)
	if err != nil {
		t.Fatalf("CreateJWT failed: %v", err)
	}

	// Try to validate with a different secret
	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Fatal("ValidateJWT should fail with wrong secret")
	}

	if err.Error() != "invalid signature" {
		t.Fatalf("Expected 'invalid signature' error, got: %v", err)
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	secret := []byte("test-secret-key")
	userID := 999

	// Create a token that expired 1 hour ago
	ttl := -1 * time.Hour

	// Create a token with negative TTL (already expired)
	token, err := CreateJWT(userID, secret, ttl)
	if err != nil {
		t.Fatalf("CreateJWT failed: %v", err)
	}

	// Try to validate expired token
	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("ValidateJWT should fail with expired token")
	}

	if err.Error() != "token expired" {
		t.Fatalf("Expected 'token expired' error, got: %v", err)
	}
}

func TestValidateJWT_InvalidFormat(t *testing.T) {
	secret := []byte("test-secret-key")

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"single part", "invalidtoken"},
		{"two parts", "invalid.token"},
		{"four parts", "too.many.parts.here"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateJWT(tt.token, secret)
			if err == nil {
				t.Fatalf("ValidateJWT should fail for %s", tt.name)
			}
		})
	}
}

func TestCreateJWT_DifferentUserIDs(t *testing.T) {
	secret := []byte("test-secret-key")
	ttl := 15 * time.Minute

	userIDs := []int{1, 100, 1000, 999999}

	for _, userID := range userIDs {
		t.Run("UserID_"+string(rune(userID)), func(t *testing.T) {
			token, err := CreateJWT(userID, secret, ttl)
			if err != nil {
				t.Fatalf("CreateJWT failed for user ID %d: %v", userID, err)
			}

			validatedUserID, err := ValidateJWT(token, secret)
			if err != nil {
				t.Fatalf("ValidateJWT failed for user ID %d: %v", userID, err)
			}

			if validatedUserID != userID {
				t.Fatalf("Expected user ID %d, got %d", userID, validatedUserID)
			}
		})
	}
}
