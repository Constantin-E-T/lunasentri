package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/alerts"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/config"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/metrics"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/notifications"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/system"
	"github.com/gorilla/websocket"
)

var (
	serverStartTime time.Time
	upgrader        = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Get allowed origin from environment variable, default to localhost:3000
			allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
			if allowedOrigin == "" {
				allowedOrigin = "http://localhost:3000"
			}
			origin := r.Header.Get("Origin")
			return origin == allowedOrigin
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// evaluateAlerts is a helper function to evaluate alerts with a dedicated background context
// This prevents context cancellation issues when request/websocket contexts are cancelled
func evaluateAlerts(alertService *alerts.Service, metricsData metrics.Metrics) {
	if alertService == nil {
		return
	}

	// Use a background context with sufficient timeout for email/webhook notifications
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := alertService.Evaluate(ctx, metricsData); err != nil {
		log.Printf("Failed to evaluate alerts: %v", err)
	}
}

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

// handleWebSocket handles WebSocket connections for streaming metrics
func handleWebSocket(collector metrics.Collector, startTime time.Time, alertService *alerts.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade HTTP to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}
		defer func() {
			conn.Close()
			log.Printf("WebSocket client disconnected: %s", r.RemoteAddr)
		}()

		log.Printf("WebSocket client connected: %s", r.RemoteAddr)

		// Set read/write deadlines
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

		// Set up ping/pong to detect disconnections
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		// Create ticker for sending metrics every 3 seconds
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		// Channel to signal when to stop the connection
		done := make(chan struct{})

		// Start goroutine to handle client disconnection detection
		go func() {
			defer close(done)
			for {
				// Read messages from client (mainly for detecting disconnection)
				_, _, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("WebSocket error: %v", err)
					}
					return
				}
			}
		}()

		// Main loop for sending metrics
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// Collect metrics
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				metricsData, err := collector.Snapshot(ctx)
				cancel()

				if err != nil {
					log.Printf("Failed to collect metrics for WebSocket: %v", err)
					metricsData = metrics.Metrics{}
				}

				// Set real uptime
				metricsData.UptimeS = time.Since(startTime).Seconds()

				// Evaluate alerts with the new metrics using background context
				evaluateAlerts(alertService, metricsData)

				// Set write deadline for this message
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

				// Send metrics as JSON
				if err := conn.WriteJSON(metricsData); err != nil {
					log.Printf("Failed to send WebSocket message: %v", err)
					return
				}

				// Send ping to keep connection alive
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("Failed to send ping: %v", err)
					return
				}
			}
		}
	}
}

// handleSystemInfo handles GET /system/info
func handleSystemInfo(systemService system.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// Create context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Get system information
		systemInfo, err := systemService.GetSystemInfo(ctx)
		if err != nil {
			log.Printf("Failed to collect system info: %v", err)
			http.Error(w, `{"error":"Failed to collect system information"}`, http.StatusInternalServerError)
			return
		}

		// Encode and send response
		if err := json.NewEncoder(w).Encode(systemInfo); err != nil {
			log.Printf("Failed to encode system info response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// newServer creates a new HTTP server with the given collector, auth service, alert service, and system service
func newServer(collector metrics.Collector, startTime time.Time, authService *auth.Service, alertService *alerts.Service, systemService system.Service, store storage.Store, webhookNotifier *notifications.Notifier, telegramNotifier *notifications.TelegramNotifier, accessTTL time.Duration, passwordResetTTL time.Duration, secureCookie bool) *http.ServeMux {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register public handlers
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "LunaSentri API - Coming Soon")
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy"}`)
	})

	// System info endpoint (public for monitoring)
	mux.HandleFunc("/system/info", handleSystemInfo(systemService))

	// Register auth handlers (public endpoints)
	mux.HandleFunc("/auth/register", handleRegister(authService))
	mux.HandleFunc("/auth/login", handleLogin(authService, accessTTL, secureCookie))
	mux.HandleFunc("/auth/logout", handleLogout(secureCookie))
	mux.HandleFunc("/auth/forgot-password", handleForgotPassword(authService, passwordResetTTL))
	mux.HandleFunc("/auth/reset-password", handleResetPassword(authService))

	// Protected auth endpoints
	mux.Handle("/auth/me", authService.RequireAuth(handleMe()))
	mux.Handle("/auth/change-password", authService.RequireAuth(handleChangePassword(authService)))

	// User management endpoints (protected)
	mux.Handle("/auth/users", authService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleListUsers(authService)(w, r)
		} else if r.Method == http.MethodPost {
			handleCreateUser(authService)(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// User delete endpoint (protected) - handle /auth/users/{id}
	mux.Handle("/auth/users/", authService.RequireAuth(handleDeleteUser(authService)))

	// Register protected handlers (require authentication)
	mux.Handle("/metrics", authService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Collect real system metrics
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		metricsData, err := collector.Snapshot(ctx)
		if err != nil {
			log.Printf("Failed to collect metrics: %v", err)
			// Return zeroed metrics on error
			metricsData = metrics.Metrics{}
		}

		// Set real uptime
		metricsData.UptimeS = time.Since(startTime).Seconds()

		// Evaluate alerts with the new metrics using background context
		evaluateAlerts(alertService, metricsData)

		if err := json.NewEncoder(w).Encode(metricsData); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})))

	// WebSocket endpoint for streaming metrics (protected)
	mux.Handle("/ws", authService.RequireAuth(handleWebSocket(collector, startTime, alertService)))

	// Alert management endpoints (protected)
	mux.Handle("/alerts/rules", authService.RequireAuth(handleAlertRules(alertService)))
	mux.Handle("/alerts/rules/", authService.RequireAuth(handleAlertRule(alertService)))
	mux.Handle("/alerts/events", authService.RequireAuth(handleAlertEvents(alertService)))
	mux.Handle("/alerts/events/", authService.RequireAuth(handleAlertEventAck(alertService)))

	// Webhook notification endpoints (protected)
	mux.Handle("/notifications/webhooks", authService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			notifications.HandleListWebhooks(store)(w, r)
		} else if r.Method == http.MethodPost {
			notifications.HandleCreateWebhook(store)(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/notifications/webhooks/", authService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a test webhook request
		if strings.HasSuffix(r.URL.Path, "/test") && r.Method == http.MethodPost {
			notifications.HandleTestWebhook(webhookNotifier, store)(w, r)
			return
		}

		if r.Method == http.MethodPut {
			notifications.HandleUpdateWebhook(store)(w, r)
		} else if r.Method == http.MethodDelete {
			notifications.HandleDeleteWebhook(store)(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Telegram notification endpoints (protected)
	if telegramNotifier != nil {
		mux.Handle("/notifications/telegram", authService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				notifications.HandleListTelegramRecipients(store)(w, r)
			} else if r.Method == http.MethodPost {
				notifications.HandleCreateTelegramRecipient(store)(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))
		mux.Handle("/notifications/telegram/", authService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/test") && r.Method == http.MethodPost {
				notifications.HandleTestTelegram(store, telegramNotifier)(w, r)
				return
			}
			if r.Method == http.MethodPut {
				notifications.HandleUpdateTelegramRecipient(store)(w, r)
			} else if r.Method == http.MethodDelete {
				notifications.HandleDeleteTelegramRecipient(store)(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))
	}

	return mux
}

// corsMiddleware adds CORS headers and handles preflight requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origin from environment variable, default to localhost:3000
		allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:3000"
		}

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// Alert rule request/response types
type AlertRuleRequest struct {
	Name         string  `json:"name"`
	Metric       string  `json:"metric"`
	ThresholdPct float64 `json:"threshold_pct"`
	Comparison   string  `json:"comparison"`
	TriggerAfter int     `json:"trigger_after"`
}

type AlertEventAckRequest struct {
	// No body needed, ID comes from URL
}

// validateAlertRule validates an alert rule request
func validateAlertRule(req *AlertRuleRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Metric != "cpu_pct" && req.Metric != "mem_used_pct" && req.Metric != "disk_used_pct" {
		return fmt.Errorf("metric must be one of: cpu_pct, mem_used_pct, disk_used_pct")
	}
	if req.ThresholdPct < 0 || req.ThresholdPct > 100 {
		return fmt.Errorf("threshold_pct must be between 0 and 100")
	}
	if req.Comparison != "above" && req.Comparison != "below" {
		return fmt.Errorf("comparison must be 'above' or 'below'")
	}
	if req.TriggerAfter < 1 {
		return fmt.Errorf("trigger_after must be >= 1")
	}
	return nil
}

// handleAlertRules handles GET /alerts/rules and POST /alerts/rules
func handleAlertRules(alertService *alerts.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case "GET":
			rules, err := alertService.ListRules(r.Context())
			if err != nil {
				log.Printf("Failed to list alert rules: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if err := json.NewEncoder(w).Encode(rules); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

		case "POST":
			var req AlertRuleRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if err := validateAlertRule(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			rule, err := alertService.UpsertRule(r.Context(), 0, req.Name, req.Metric, req.Comparison, req.ThresholdPct, req.TriggerAfter)
			if err != nil {
				log.Printf("Failed to create alert rule: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(rule); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleAlertRule handles PUT /alerts/rules/{id} and DELETE /alerts/rules/{id}
func handleAlertRule(alertService *alerts.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Extract ID from path
		path := strings.TrimPrefix(r.URL.Path, "/alerts/rules/")
		if path == "" {
			http.Error(w, "Alert rule ID required", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "Invalid alert rule ID", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "PUT":
			var req AlertRuleRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if err := validateAlertRule(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			rule, err := alertService.UpsertRule(r.Context(), id, req.Name, req.Metric, req.Comparison, req.ThresholdPct, req.TriggerAfter)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					http.Error(w, "Alert rule not found", http.StatusNotFound)
				} else {
					log.Printf("Failed to update alert rule: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}

			if err := json.NewEncoder(w).Encode(rule); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

		case "DELETE":
			if err := alertService.DeleteRule(r.Context(), id); err != nil {
				if strings.Contains(err.Error(), "not found") {
					http.Error(w, "Alert rule not found", http.StatusNotFound)
				} else {
					log.Printf("Failed to delete alert rule: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}

			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleAlertEvents handles GET /alerts/events
func handleAlertEvents(alertService *alerts.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// Default limit to 50, can be overridden by query param
		limit := 50
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		events, err := alertService.ListActiveEvents(r.Context(), limit)
		if err != nil {
			log.Printf("Failed to list alert events: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(events); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// handleAlertEventAck handles POST /alerts/events/{id}/ack
func handleAlertEventAck(alertService *alerts.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract ID from path - expecting /alerts/events/{id}/ack
		path := strings.TrimPrefix(r.URL.Path, "/alerts/events/")
		if !strings.HasSuffix(path, "/ack") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		idStr := strings.TrimSuffix(path, "/ack")
		if idStr == "" {
			http.Error(w, "Alert event ID required", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid alert event ID", http.StatusBadRequest)
			return
		}

		if err := alertService.AcknowledgeEvent(r.Context(), id); err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already acknowledged") {
				http.Error(w, "Alert event not found or already acknowledged", http.StatusNotFound)
			} else {
				log.Printf("Failed to acknowledge alert event: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func main() {
	// Record server start time for uptime calculation
	serverStartTime = time.Now()

	// Get database path from environment variable
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/lunasentri.db"
	}

	// Ensure data directory exists
	if err := storage.EnsureDataDir(dbPath); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize database
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	log.Printf("Database initialized at: %s", dbPath)

	// Bootstrap admin user if environment variables are set
	ctx := context.Background()
	if err := auth.BootstrapAdmin(ctx, store); err != nil {
		log.Fatalf("Failed to bootstrap admin user: %v", err)
	}

	// Get JWT secret from environment variable (required)
	jwtSecret := os.Getenv("AUTH_JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("AUTH_JWT_SECRET environment variable is required")
	}

	// Get access token TTL from environment variable, default to 15 minutes
	accessTTL := 15 * time.Minute
	if ttlStr := os.Getenv("ACCESS_TOKEN_TTL"); ttlStr != "" {
		if parsedTTL, err := time.ParseDuration(ttlStr); err == nil {
			accessTTL = parsedTTL
		} else {
			log.Printf("Warning: Invalid ACCESS_TOKEN_TTL value '%s', using default 15m", ttlStr)
		}
	}

	// Get password reset TTL from environment variable, default to 1 hour
	passwordResetTTL := 1 * time.Hour
	if ttlStr := os.Getenv("PASSWORD_RESET_TTL"); ttlStr != "" {
		if parsedTTL, err := time.ParseDuration(ttlStr); err == nil {
			passwordResetTTL = parsedTTL
		} else {
			log.Printf("Warning: Invalid PASSWORD_RESET_TTL value '%s', using default 1h", ttlStr)
		}
	}

	// Initialize auth service
	authService, err := auth.NewService(store, jwtSecret, accessTTL)
	if err != nil {
		log.Fatalf("Failed to initialize auth service: %v", err)
	}

	log.Printf("Auth service initialized (access token TTL: %v, password reset TTL: %v)", accessTTL, passwordResetTTL)

	// Get secure cookie setting from environment variable, default to true for production
	secureCookie := true
	if secureCookieEnv := os.Getenv("SECURE_COOKIE"); secureCookieEnv == "false" {
		secureCookie = false
		log.Println("Warning: Secure cookie flag disabled - only use in development")
	}

	// Initialize metrics collector
	metricsCollector := metrics.NewSystemCollector()

	// Initialize system service
	systemService := system.NewSystemService()

	// Load Telegram configuration
	telegramConfig, err := config.LoadTelegramConfig()
	if err != nil {
		log.Println("Telegram notifications disabled:", err)
		telegramConfig = nil
	}
	if telegramConfig != nil && telegramConfig.IsEnabled() {
		log.Println("Telegram notifications enabled")
	}

	// Initialize webhook notifier
	webhookNotifier := notifications.NewNotifier(store, log.Default())

	// Initialize Telegram notifier
	var telegramNotifier *notifications.TelegramNotifier
	if telegramConfig != nil && telegramConfig.IsEnabled() {
		telegramNotifier = notifications.NewTelegramNotifier(store, telegramConfig, log.Default())
	}

	// Create composite notifier that fans out to all channels
	compositeNotifier := notifications.NewCompositeNotifier(log.Default(), webhookNotifier, telegramNotifier)

	// Initialize alert service with composite notifier
	alertService := alerts.NewService(store, compositeNotifier)

	// Create server with real collector, auth service, alert service, and system service
	mux := newServer(metricsCollector, serverStartTime, authService, alertService, systemService, store, webhookNotifier, telegramNotifier, accessTTL, passwordResetTTL, secureCookie)

	// Create HTTP server with CORS middleware
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMiddleware(mux),
	}

	// Start server in a goroutine
	go func() {
		allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:3000"
		}
		log.Printf("LunaSentri API starting on port %s (endpoints: /, /health, /metrics, /ws) with CORS origin: %s", port, allowedOrigin)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("LunaSentri API shutting down...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("LunaSentri API stopped gracefully")
	}
}
