package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/alerts"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/metrics"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/notifications"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/system"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
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

// RouterConfig holds dependencies for the HTTP router
type RouterConfig struct {
	Collector        metrics.Collector
	ServerStartTime  time.Time
	AuthService      *auth.Service
	AlertService     *alerts.Service
	SystemService    system.Service
	Store            storage.Store
	WebhookNotifier  *notifications.Notifier
	TelegramNotifier *notifications.TelegramNotifier
	AccessTTL        time.Duration
	PasswordResetTTL time.Duration
	SecureCookie     bool
	LocalHostMetrics bool
}

// NewRouter creates a new HTTP router with all routes configured
func NewRouter(cfg *RouterConfig) *http.ServeMux {
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
	mux.HandleFunc("/system/info", handleSystemInfo(cfg.SystemService))

	// Register auth handlers (public endpoints)
	mux.HandleFunc("/auth/register", handleRegister(cfg.AuthService))
	mux.HandleFunc("/auth/login", handleLogin(cfg.AuthService, cfg.AccessTTL, cfg.SecureCookie))
	mux.HandleFunc("/auth/logout", handleLogout(cfg.SecureCookie))
	mux.HandleFunc("/auth/forgot-password", handleForgotPassword(cfg.AuthService, cfg.PasswordResetTTL))
	mux.HandleFunc("/auth/reset-password", handleResetPassword(cfg.AuthService))

	// Protected auth endpoints
	mux.Handle("/auth/me", cfg.AuthService.RequireAuth(handleMe()))
	mux.Handle("/auth/change-password", cfg.AuthService.RequireAuth(handleChangePassword(cfg.AuthService)))

	// User management endpoints (protected)
	mux.Handle("/auth/users", cfg.AuthService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleListUsers(cfg.AuthService)(w, r)
		} else if r.Method == http.MethodPost {
			handleCreateUser(cfg.AuthService)(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// User delete endpoint (protected) - handle /auth/users/{id}
	mux.Handle("/auth/users/", cfg.AuthService.RequireAuth(handleDeleteUser(cfg.AuthService)))

	// Register protected handlers (require authentication)
	mux.Handle("/metrics", cfg.AuthService.RequireAuth(handleMetrics(cfg.Collector, cfg.ServerStartTime, cfg.AlertService, cfg.LocalHostMetrics)))

	// WebSocket endpoint for streaming metrics (protected)
	mux.Handle("/ws", cfg.AuthService.RequireAuth(handleWebSocket(cfg.Collector, cfg.ServerStartTime, cfg.AlertService, cfg.LocalHostMetrics)))

	// Alert management endpoints (protected)
	mux.Handle("/alerts/rules", cfg.AuthService.RequireAuth(handleAlertRules(cfg.AlertService)))
	mux.Handle("/alerts/rules/", cfg.AuthService.RequireAuth(handleAlertRule(cfg.AlertService)))
	mux.Handle("/alerts/events", cfg.AuthService.RequireAuth(handleAlertEvents(cfg.AlertService)))
	mux.Handle("/alerts/events/", cfg.AuthService.RequireAuth(handleAlertEventAck(cfg.AlertService)))

	// Webhook notification endpoints (protected)
	mux.Handle("/notifications/webhooks", cfg.AuthService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			notifications.HandleListWebhooks(cfg.Store)(w, r)
		} else if r.Method == http.MethodPost {
			notifications.HandleCreateWebhook(cfg.Store)(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/notifications/webhooks/", cfg.AuthService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a test webhook request
		if strings.HasSuffix(r.URL.Path, "/test") && r.Method == http.MethodPost {
			notifications.HandleTestWebhook(cfg.WebhookNotifier, cfg.Store)(w, r)
			return
		}

		if r.Method == http.MethodPut {
			notifications.HandleUpdateWebhook(cfg.Store)(w, r)
		} else if r.Method == http.MethodDelete {
			notifications.HandleDeleteWebhook(cfg.Store)(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Telegram notification endpoints (protected)
	if cfg.TelegramNotifier != nil {
		mux.Handle("/notifications/telegram", cfg.AuthService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				notifications.HandleListTelegramRecipients(cfg.Store)(w, r)
			} else if r.Method == http.MethodPost {
				notifications.HandleCreateTelegramRecipient(cfg.Store)(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))
		mux.Handle("/notifications/telegram/", cfg.AuthService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/test") && r.Method == http.MethodPost {
				notifications.HandleTestTelegram(cfg.Store, cfg.TelegramNotifier)(w, r)
				return
			}
			if r.Method == http.MethodPut {
				notifications.HandleUpdateTelegramRecipient(cfg.Store)(w, r)
			} else if r.Method == http.MethodDelete {
				notifications.HandleDeleteTelegramRecipient(cfg.Store)(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})))
	}

	return mux
}

// CORSMiddleware adds CORS headers and handles preflight requests
func CORSMiddleware(next http.Handler) http.Handler {
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

// handleMetrics handles GET /metrics
func handleMetrics(collector metrics.Collector, startTime time.Time, alertService *alerts.Service, localHostMetrics bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// TODO: Phase 2 - Require machine_id query parameter for multi-machine support
		// For now, we check if local host metrics are enabled
		if !localHostMetrics {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Local host metrics disabled. Please register a machine and use machine_id parameter.",
				// TODO: Reference upcoming agent ingestion work (see docs/roadmap/MULTI_MACHINE_MONITORING.md)
			})
			return
		}

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
	}
}

// handleWebSocket handles WebSocket connections for streaming metrics
func handleWebSocket(collector metrics.Collector, startTime time.Time, alertService *alerts.Service, localHostMetrics bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Phase 2 - Require machine_id query parameter for multi-machine support
		if !localHostMetrics {
			http.Error(w, "Local host metrics disabled. Please register a machine first.", http.StatusUnprocessableEntity)
			return
		}

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
