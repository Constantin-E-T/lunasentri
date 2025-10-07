package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost (12 provides good security vs performance balance)
	DefaultCost = 12
)

// HashPassword hashes a plaintext password using bcrypt with the default cost
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword verifies that a plaintext password matches the given hash
func VerifyPassword(hash, password string) error {
	if hash == "" {
		return fmt.Errorf("hash cannot be empty")
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return fmt.Errorf("invalid password")
		}
		return fmt.Errorf("failed to verify password: %w", err)
	}

	return nil
}
