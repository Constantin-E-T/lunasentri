package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ForgotPasswordRequest represents the forgot password request body
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPasswordResponse represents the forgot password response
type ForgotPasswordResponse struct {
	ResetToken string `json:"reset_token"`
}

// ResetPasswordRequest represents the reset password request body
type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// ChangePasswordRequest represents the change password request body
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// UserProfile represents the user profile response
type UserProfile struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserRequest represents the create user request body
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

// CreateUserResponse represents the create user response
type CreateUserResponse struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
	TempPassword string    `json:"temp_password,omitempty"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// handleRegister handles POST /auth/register
func handleRegister(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate request
		if req.Email == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "email cannot be empty"})
			return
		}
		if !strings.Contains(req.Email, "@") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid email format"})
			return
		}
		if len(req.Password) < 8 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "password must be at least 8 characters"})
			return
		}

		// Create the user
		user, tempPassword, err := authService.CreateUser(r.Context(), req.Email, req.Password)
		if err != nil {
			log.Printf("Failed to register user: %v", err)

			// Return appropriate status codes
			if strings.Contains(err.Error(), "already exists") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{"error": "user with this email already exists"})
				return
			}

			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return user with optional temp password (shouldn't happen for registration)
		response := CreateUserResponse{
			ID:           user.ID,
			Email:        user.Email,
			IsAdmin:      user.IsAdmin,
			CreatedAt:    user.CreatedAt,
			TempPassword: tempPassword,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// handleLogin handles POST /auth/login
func handleLogin(authService *auth.Service, accessTTL time.Duration, secureCookie bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Authenticate user
		user, err := authService.Authenticate(r.Context(), req.Email, req.Password)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Create session token
		token, err := authService.CreateSession(user.ID)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		auth.SetSessionCookie(w, token, int(accessTTL.Seconds()), secureCookie)

		// Return user profile
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserProfile{
			ID:        user.ID,
			Email:     user.Email,
			IsAdmin:   user.IsAdmin,
			CreatedAt: user.CreatedAt,
		})
	}
}

// handleLogout handles POST /auth/logout
func handleLogout(secureCookie bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Clear session cookie
		auth.ClearSessionCookie(w, secureCookie)

		// Return 204 No Content
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleMe handles GET /auth/me
func handleMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get user from context (set by RequireAuth middleware)
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Return user profile
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserProfile{
			ID:        user.ID,
			Email:     user.Email,
			IsAdmin:   user.IsAdmin,
			CreatedAt: user.CreatedAt,
		})
	}
}

// handleForgotPassword handles POST /auth/forgot-password
func handleForgotPassword(authService *auth.Service, passwordResetTTL time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ForgotPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Generate password reset token
		token, err := authService.GeneratePasswordReset(r.Context(), req.Email, passwordResetTTL)
		if err != nil {
			log.Printf("Failed to generate password reset: %v", err)
			// Still return 202 to avoid leaking user existence
		}

		// Log the token to stdout for manual testing (dev mode)
		log.Printf("Password reset token for %s: %s", req.Email, token)

		// Return token in response (for development)
		// TODO: In production, send via email instead
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(ForgotPasswordResponse{
			ResetToken: token,
		})
	}
}

// handleResetPassword handles POST /auth/reset-password
func handleResetPassword(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ResetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Reset the password
		err := authService.ResetPassword(r.Context(), req.Token, req.Password)
		if err != nil {
			log.Printf("Password reset failed: %v", err)
			http.Error(w, "Invalid or expired reset token", http.StatusBadRequest)
			return
		}

		// Return 204 No Content on success
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleChangePassword handles POST /auth/change-password (requires authentication)
func handleChangePassword(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get user from context (set by RequireAuth middleware)
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Change the password
		err := authService.ChangePassword(r.Context(), user.ID, req.CurrentPassword, req.NewPassword)
		if err != nil {
			log.Printf("Password change failed for user %d: %v", user.ID, err)

			// Map errors to appropriate HTTP status codes
			errMsg := err.Error()
			if strings.Contains(errMsg, "current password is incorrect") {
				http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
				return
			}
			if strings.Contains(errMsg, "must be at least 8 characters") {
				http.Error(w, "New password must be at least 8 characters long", http.StatusBadRequest)
				return
			}
			if strings.Contains(errMsg, "user not found") {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}

			// Generic error for other cases
			http.Error(w, "Failed to change password", http.StatusBadRequest)
			return
		}

		log.Printf("Password successfully changed for user %d", user.ID)

		// Return 204 No Content on success
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleListUsers handles GET /auth/users
func handleListUsers(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// List all users
		users, err := authService.ListUsers(r.Context())
		if err != nil {
			log.Printf("Failed to list users: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Convert to UserProfile list
		profiles := make([]UserProfile, len(users))
		for i, user := range users {
			profiles[i] = UserProfile{
				ID:        user.ID,
				Email:     user.Email,
				IsAdmin:   user.IsAdmin,
				CreatedAt: user.CreatedAt,
			}
		}

		// Return user list
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profiles)
	}
}

// handleCreateUser handles POST /auth/users
func handleCreateUser(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Create the user
		user, tempPassword, err := authService.CreateUser(r.Context(), req.Email, req.Password)
		if err != nil {
			log.Printf("Failed to create user: %v", err)

			// Return appropriate status codes
			if strings.Contains(err.Error(), "already exists") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			if strings.Contains(err.Error(), "invalid email") || strings.Contains(err.Error(), "cannot be empty") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return user with optional temp password
		response := CreateUserResponse{
			ID:           user.ID,
			Email:        user.Email,
			IsAdmin:      user.IsAdmin,
			CreatedAt:    user.CreatedAt,
			TempPassword: tempPassword,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// handleDeleteUser handles DELETE /auth/users/{id}
func handleDeleteUser(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract user ID from URL path
		path := strings.TrimPrefix(r.URL.Path, "/auth/users/")
		if path == "" || path == r.URL.Path {
			http.Error(w, "User ID required", http.StatusBadRequest)
			return
		}

		// Parse user ID
		var userID int
		if _, err := fmt.Sscanf(path, "%d", &userID); err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Get current user from context
		currentUser, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Delete the user
		err := authService.DeleteUser(r.Context(), userID, currentUser.ID)
		if err != nil {
			log.Printf("Failed to delete user: %v", err)

			// Return appropriate status codes
			if strings.Contains(err.Error(), "cannot delete your own account") ||
				strings.Contains(err.Error(), "cannot delete the last admin") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			if strings.Contains(err.Error(), "user not found") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return 204 No Content on success
		w.WriteHeader(http.StatusNoContent)
	}
}
