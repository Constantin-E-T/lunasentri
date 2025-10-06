package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

const (
	// DefaultPasswordResetTTL is the default time-to-live for password reset tokens
	DefaultPasswordResetTTL = 1 * time.Hour

	// PasswordResetTokenLength is the number of random bytes for the reset token
	PasswordResetTokenLength = 32
)

// GeneratePasswordResetRequest represents the request to generate a password reset token
type GeneratePasswordResetRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest represents the request to reset a password
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"password"`
}

// GeneratePasswordResetResponse represents the response from generating a password reset token
type GeneratePasswordResetResponse struct {
	ResetToken string `json:"reset_token"`
}

// GeneratePasswordReset generates a password reset token for the given email
func (s *Service) GeneratePasswordReset(ctx context.Context, email string, passwordResetTTL time.Duration) (string, error) {
	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}

	// Get user by email
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal whether the user exists or not - log the attempt and return success
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Printf("Password reset requested for non-existent email: %s", email)
			// Return a fake token to avoid leaking user existence
			return generateSecureToken()
		}
		return "", fmt.Errorf("failed to lookup user: %w", err)
	}

	// Generate a secure random token
	token, err := generateSecureToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash the token before storing
	tokenHash, err := hashToken(token)
	if err != nil {
		return "", fmt.Errorf("failed to hash token: %w", err)
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(passwordResetTTL)

	// Store the password reset entry
	_, err = s.store.CreatePasswordReset(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create password reset: %w", err)
	}

	log.Printf("Password reset token generated for user %s (expires at %v)", user.Email, expiresAt)

	return token, nil
}

// ResetPassword resets a user's password using a valid reset token
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	if newPassword == "" {
		return fmt.Errorf("new password cannot be empty")
	}

	// Validate password strength (minimum 8 characters)
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Hash the token to look it up
	tokenHash, err := hashToken(token)
	if err != nil {
		return fmt.Errorf("invalid token format")
	}

	// Get the password reset entry
	passwordReset, err := s.store.GetPasswordResetByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, storage.ErrPasswordResetNotFound) {
			return fmt.Errorf("invalid or expired reset token")
		}
		return fmt.Errorf("failed to validate token: %w", err)
	}

	// Hash the new password
	newPasswordHash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update the user's password
	err = s.store.UpdateUserPassword(ctx, passwordReset.UserID, newPasswordHash)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark the token as used
	err = s.store.MarkPasswordResetUsed(ctx, passwordReset.ID)
	if err != nil {
		log.Printf("Warning: failed to mark password reset as used: %v", err)
		// Don't fail the request since the password was already updated
	}

	log.Printf("Password successfully reset for user ID %d", passwordReset.UserID)

	return nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, PasswordResetTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as base64 URL-safe string
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// hashToken hashes a token using SHA256 for secure storage
// Note: SHA256 is deterministic (unlike bcrypt) which allows for token lookup
func hashToken(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}

	// Use SHA256 to hash the token (deterministic for lookup)
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:]), nil
}
