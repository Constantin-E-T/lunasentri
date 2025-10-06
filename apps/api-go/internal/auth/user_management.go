package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

const (
	// TempPasswordLength is the length of generated temporary passwords
	TempPasswordLength = 32
)

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"` // Optional - will be generated if not provided
}

// CreateUserResponse represents the response from creating a user
type CreateUserResponse struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	TempPassword string `json:"temp_password,omitempty"` // Only present if password was generated
}

// CreateUser creates a new user with email validation
func (s *Service) CreateUser(ctx context.Context, email, password string) (*storage.User, string, error) {
	// Validate email
	if email == "" {
		return nil, "", fmt.Errorf("email cannot be empty")
	}
	if !strings.Contains(email, "@") {
		return nil, "", fmt.Errorf("invalid email format")
	}

	// Generate temporary password if not provided
	tempPassword := ""
	if password == "" {
		var err error
		password, err = generateTempPassword()
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate temporary password: %w", err)
		}
		tempPassword = password
	}

	// Hash the password
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the user
	user, err := s.store.CreateUser(ctx, email, passwordHash)
	if err != nil {
		// Check for duplicate email error
		if strings.Contains(err.Error(), "already exists") {
			return nil, "", fmt.Errorf("user with email %s already exists", email)
		}
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	return user, tempPassword, nil
}

// ListUsers returns all users
func (s *Service) ListUsers(ctx context.Context) ([]storage.User, error) {
	users, err := s.store.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// DeleteUser deletes a user by ID with safety checks
func (s *Service) DeleteUser(ctx context.Context, userID, currentUserID int) error {
	// Prevent deleting self
	if userID == currentUserID {
		return fmt.Errorf("cannot delete your own account")
	}

	// Delete the user (store will prevent deleting last user)
	err := s.store.DeleteUser(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "cannot delete the last user") {
			return fmt.Errorf("cannot delete the last remaining user")
		}
		if err == storage.ErrUserNotFound {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// generateTempPassword generates a cryptographically secure temporary password
func generateTempPassword() (string, error) {
	bytes := make([]byte, TempPasswordLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as base64 URL-safe string (suitable for passwords)
	return base64.URLEncoding.EncodeToString(bytes), nil
}
