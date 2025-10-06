package storage

import (
	"context"
	"time"
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

	// UpsertAdmin creates or updates an admin user with the given email and password hash
	UpsertAdmin(ctx context.Context, email, passwordHash string) (*User, error)

	// Close closes the storage connection
	Close() error
}