package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/metrics"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
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

// UserProfile represents the user profile response
type UserProfile struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
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
	CreatedAt    time.Time `json:"created_at"`
	TempPassword string    `json:"temp_password,omitempty"`
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
				strings.Contains(err.Error(), "cannot delete the last remaining user") {
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
func handleWebSocket(collector metrics.Collector, startTime time.Time) http.HandlerFunc {
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

// newServer creates a new HTTP server with the given collector and auth service
func newServer(collector metrics.Collector, startTime time.Time, authService *auth.Service, accessTTL time.Duration, passwordResetTTL time.Duration, secureCookie bool) *http.ServeMux {
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

	// Register auth handlers (public endpoints)
	mux.HandleFunc("/auth/login", handleLogin(authService, accessTTL, secureCookie))
	mux.HandleFunc("/auth/logout", handleLogout(secureCookie))
	mux.HandleFunc("/auth/forgot-password", handleForgotPassword(authService, passwordResetTTL))
	mux.HandleFunc("/auth/reset-password", handleResetPassword(authService))

	// Protected auth endpoints
	mux.Handle("/auth/me", authService.RequireAuth(handleMe()))

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

		if err := json.NewEncoder(w).Encode(metricsData); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})))

	// WebSocket endpoint for streaming metrics (protected)
	mux.Handle("/ws", authService.RequireAuth(handleWebSocket(collector, startTime)))

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

	// Create server with real collector and auth service
	mux := newServer(metricsCollector, serverStartTime, authService, accessTTL, passwordResetTTL, secureCookie)

	// Create HTTP server with CORS middleware
	port := "8080"
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
