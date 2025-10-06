package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// Service provides authentication operations
type Service struct {
	store      storage.Store
	jwtSecret  []byte
	accessTTL  time.Duration
}

// NewService creates a new authentication service
func NewService(store storage.Store, jwtSecret string, accessTTL time.Duration) (*Service, error) {
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT secret cannot be empty")
	}
	if accessTTL <= 0 {
		return nil, fmt.Errorf("access token TTL must be positive")
	}

	return &Service{
		store:     store,
		jwtSecret: []byte(jwtSecret),
		accessTTL: accessTTL,
	}, nil
}

// Authenticate verifies user credentials and returns the user if valid
func (s *Service) Authenticate(ctx context.Context, email, password string) (*storage.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	// Get user by email
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

// GetUser retrieves a user by their ID
func (s *Service) GetUser(ctx context.Context, id int) (*storage.User, error) {
	user, err := s.store.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// CreateSession creates a new JWT token for the given user
func (s *Service) CreateSession(userID int) (string, error) {
	return CreateJWT(userID, s.jwtSecret, s.accessTTL)
}

// ValidateSession validates a JWT token and returns the user ID
func (s *Service) ValidateSession(token string) (int, error) {
	return ValidateJWT(token, s.jwtSecret)
}
