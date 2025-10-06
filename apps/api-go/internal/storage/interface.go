package storage

import (
	"context"
	"errors"
	"time"
)

var (
    // ErrUserNotFound is returned when a user is not found
    ErrUserNotFound = errors.New("user not found")
    // ErrPasswordResetNotFound is returned when a password reset token is not found or invalid
    ErrPasswordResetNotFound = errors.New("password reset token not found")
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

// Store defines the interface for user storage operations
type Store interface {
	// CreateUser creates a new user with the given email and password hash
	CreateUser(ctx context.Context, email, passwordHash string) (*User, error)

	// GetUserByEmail retrieves a user by their email address
	GetUserByEmail(ctx context.Context, email string) (*User, error)

    // GetUserByID retrieves a user by their ID
    GetUserByID(ctx context.Context, id int) (*User, error)

    // UpdateUserPassword updates the password hash for a user
    UpdateUserPassword(ctx context.Context, userID int, passwordHash string) error

    // UpsertAdmin creates or updates an admin user with the given email and password hash
    UpsertAdmin(ctx context.Context, email, passwordHash string) (*User, error)

    // CreatePasswordReset creates a password reset entry for a user
    CreatePasswordReset(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) (*PasswordReset, error)

    // GetPasswordResetByHash retrieves an active password reset entry by token hash
    GetPasswordResetByHash(ctx context.Context, tokenHash string) (*PasswordReset, error)

    // MarkPasswordResetUsed marks a password reset entry as used
    MarkPasswordResetUsed(ctx context.Context, id int) error

	// Close closes the storage connection
	Close() error
}

// PasswordReset represents a password reset token entry
type PasswordReset struct {
    ID        int        `json:"id"`
    UserID    int        `json:"user_id"`
    TokenHash string     `json:"token_hash"`
    ExpiresAt time.Time  `json:"expires_at"`
    UsedAt    *time.Time `json:"used_at"`
    CreatedAt time.Time  `json:"created_at"`
}
