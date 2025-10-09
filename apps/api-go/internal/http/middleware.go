package router

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/machines"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// Context keys for machine and user data
type contextKey string

const (
	machineIDKey contextKey = "machine_id"
	userIDKey    contextKey = "user_id"
	machineKey   contextKey = "machine"
)

// RequireAPIKey is middleware that validates API keys for agent requests
func RequireAPIKey(machineService *machines.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract API key from X-API-Key header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				// Try Authorization header as fallback (Bearer token format)
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					apiKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if apiKey == "" {
				log.Printf("Agent request rejected: missing API key from %s", getRemoteIP(r))
				http.Error(w, "Unauthorized: API key required", http.StatusUnauthorized)
				return
			}

			// Authenticate machine using API key
			machine, err := machineService.AuthenticateMachine(r.Context(), apiKey)
			if err != nil {
				log.Printf("Agent request rejected: invalid API key from %s", getRemoteIP(r))
				http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
				return
			}

			// Check if machine is revoked (we could add a "revoked" status in the future)
			// For now, if the machine exists in the DB, it's valid

			// Add machine ID and user ID to context
			ctx := context.WithValue(r.Context(), machineIDKey, machine.ID)
			ctx = context.WithValue(ctx, userIDKey, machine.UserID)
			ctx = context.WithValue(ctx, machineKey, machine)

			log.Printf("Agent authenticated: machine_id=%d, user_id=%d, machine_name=%s, remote_ip=%s",
				machine.ID, machine.UserID, machine.Name, getRemoteIP(r))

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetMachineIDFromContext retrieves the machine ID from the request context
func GetMachineIDFromContext(ctx context.Context) (int, bool) {
	machineID, ok := ctx.Value(machineIDKey).(int)
	return machineID, ok
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDKey).(int)
	return userID, ok
}

// GetMachineFromContext retrieves the full machine object from the request context
func GetMachineFromContext(ctx context.Context) (*storage.Machine, bool) {
	machine, ok := ctx.Value(machineKey).(*storage.Machine)
	return machine, ok
}
