package auth

import (
	"context"
	"log"
	"net/http"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserContextKey is the key used to store the user in the request context
	UserContextKey contextKey = "user"
)

// RequireAuth is a middleware that validates the session and loads the user into the context
func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		token, err := GetSessionCookie(r)
		if err != nil {
			log.Printf("Authentication failed: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate JWT and get user ID
		userID, err := s.ValidateSession(token)
		if err != nil {
			log.Printf("Session validation failed: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get user from database
		user, err := s.GetUser(r.Context(), userID)
		if err != nil {
			log.Printf("Failed to get user: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Store user in context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (*storage.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*storage.User)
	return user, ok
}
