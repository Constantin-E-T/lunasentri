package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/alerts"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/machines"
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
	MachineService   *machines.Service
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

	// System info endpoint (requires auth, machine-aware)
	mux.Handle("/system/info", cfg.AuthService.RequireAuth(handleSystemInfo(cfg.SystemService, cfg.MachineService, cfg.LocalHostMetrics)))

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
	mux.Handle("/metrics", cfg.AuthService.RequireAuth(handleMetrics(cfg.Collector, cfg.ServerStartTime, cfg.AlertService, cfg.MachineService, cfg.LocalHostMetrics)))

	// WebSocket endpoint for streaming metrics (protected)
	mux.Handle("/ws", cfg.AuthService.RequireAuth(handleWebSocket(cfg.Collector, cfg.ServerStartTime, cfg.AlertService, cfg.MachineService, cfg.LocalHostMetrics)))

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
	// Always register endpoints regardless of whether TelegramNotifier is configured
	mux.Handle("/notifications/telegram", cfg.AuthService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			notifications.HandleListTelegramRecipients(cfg.Store)(w, r)
		} else if r.Method == http.MethodPost {
			notifications.HandleCreateTelegramRecipient(cfg.Store)(w, r)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		}
	})))

	// Agent endpoints
	// POST /agent/register - Session authenticated (user registers a new machine)
	mux.Handle("/agent/register", cfg.AuthService.RequireAuth(handleAgentRegister(cfg.MachineService)))

	// POST /agent/metrics - API key authenticated (agent pushes metrics)
	mux.Handle("/agent/metrics", RequireAPIKey(cfg.MachineService)(http.HandlerFunc(handleAgentMetrics(cfg.MachineService))))

	// Machine management endpoints (session authenticated)
	mux.Handle("/machines", cfg.AuthService.RequireAuth(handleListMachines(cfg.MachineService)))
	mux.Handle("/machines/", cfg.AuthService.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle /machines/:id/disable
		if strings.HasSuffix(r.URL.Path, "/disable") && r.Method == http.MethodPost {
			handleDisableMachine(cfg.MachineService)(w, r)
			return
		}
		// Handle /machines/:id/enable
		if strings.HasSuffix(r.URL.Path, "/enable") && r.Method == http.MethodPost {
			handleEnableMachine(cfg.MachineService)(w, r)
			return
		}
		// Handle /machines/:id/rotate-key
		if strings.HasSuffix(r.URL.Path, "/rotate-key") && r.Method == http.MethodPost {
			handleRotateMachineAPIKey(cfg.MachineService)(w, r)
			return
		}

		// Handle basic CRUD operations
		switch r.Method {
		case http.MethodDelete:
			handleDeleteMachine(cfg.MachineService)(w, r)
		case http.MethodPatch:
			handleUpdateMachine(cfg.MachineService)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
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
func handleMetrics(collector metrics.Collector, startTime time.Time, alertService *alerts.Service, machineService *machines.Service, localHostMetrics bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Determine if machine_id was provided
		machineIDParam := r.URL.Query().Get("machine_id")

		if machineIDParam != "" {
			if machineService == nil {
				http.Error(w, "Machine metrics service unavailable", http.StatusInternalServerError)
				return
			}

			machineID, err := strconv.Atoi(machineIDParam)
			if err != nil {
				http.Error(w, "Invalid machine_id", http.StatusBadRequest)
				return
			}

			user, ok := auth.GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			latest, err := machineService.GetLatestMetrics(r.Context(), machineID, user.ID)
			if err != nil {
				if strings.Contains(err.Error(), "no metrics found") {
					empty := metrics.Metrics{}
					if err := json.NewEncoder(w).Encode(empty); err != nil {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					}
					return
				}
				log.Printf("Failed to fetch metrics for machine %d (user %d): %v", machineID, user.ID, err)
				http.Error(w, "Metrics not available for this machine", http.StatusInternalServerError)
				return
			}

			metricsData := metrics.Metrics{
				CPUPct:      latest.CPUPct,
				MemUsedPct:  latest.MemUsedPct,
				DiskUsedPct: latest.DiskUsedPct,
				UptimeS:     latest.UptimeSeconds,
			}

			evaluateAlerts(alertService, metricsData)

			if err := json.NewEncoder(w).Encode(metricsData); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			return
		}

		// No machine_id; fall back to local host metrics for development if enabled
		if !localHostMetrics {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Local host metrics disabled. Please select a machine.",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		metricsData, err := collector.Snapshot(ctx)
		if err != nil {
			log.Printf("Failed to collect metrics: %v", err)
			metricsData = metrics.Metrics{}
		}

		metricsData.UptimeS = time.Since(startTime).Seconds()

		evaluateAlerts(alertService, metricsData)

		if err := json.NewEncoder(w).Encode(metricsData); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// handleWebSocket handles WebSocket connections for streaming metrics
func handleWebSocket(collector metrics.Collector, startTime time.Time, alertService *alerts.Service, machineService *machines.Service, localHostMetrics bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		machineIDParam := r.URL.Query().Get("machine_id")
		var machineID int
		var err error
		var userID int
		streamMachineMetrics := machineIDParam != ""

		if streamMachineMetrics {
			if machineService == nil {
				http.Error(w, "Machine metrics service unavailable", http.StatusInternalServerError)
				return
			}

			machineID, err = strconv.Atoi(machineIDParam)
			if err != nil {
				http.Error(w, "Invalid machine_id", http.StatusBadRequest)
				return
			}

			user, ok := auth.GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			userID = user.ID

			// Ensure user owns machine
			if _, err := machineService.GetMachine(r.Context(), machineID, user.ID); err != nil {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}
		} else if !localHostMetrics {
			http.Error(w, "Local host metrics disabled. Please select a machine.", http.StatusUnprocessableEntity)
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
				var metricsData metrics.Metrics

				if streamMachineMetrics {
					latest, err := machineService.GetLatestMetrics(r.Context(), machineID, userID)
					if err != nil {
						log.Printf("Failed to fetch metrics for WebSocket stream machine_id=%d: %v", machineID, err)
						continue
					}
					metricsData = metrics.Metrics{
						CPUPct:      latest.CPUPct,
						MemUsedPct:  latest.MemUsedPct,
						DiskUsedPct: latest.DiskUsedPct,
						UptimeS:     latest.UptimeSeconds,
					}
				} else {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					tmp, err := collector.Snapshot(ctx)
					cancel()

					if err != nil {
						log.Printf("Failed to collect metrics for WebSocket: %v", err)
						tmp = metrics.Metrics{}
					}

					tmp.UptimeS = time.Since(startTime).Seconds()
					metricsData = tmp
				}

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
func handleSystemInfo(systemService system.Service, machineService *machines.Service, localHostMetrics bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		machineIDParam := r.URL.Query().Get("machine_id")

		if machineIDParam != "" {
			if machineService == nil {
				http.Error(w, "Machine service unavailable", http.StatusInternalServerError)
				return
			}

			machineID, err := strconv.Atoi(machineIDParam)
			if err != nil {
				http.Error(w, "Invalid machine_id", http.StatusBadRequest)
				return
			}

			user, ok := auth.GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			machine, err := machineService.GetMachineWithComputedStatus(r.Context(), machineID, user.ID)
			if err != nil {
				log.Printf("Failed to load machine %d for user %d: %v", machineID, user.ID, err)
				http.Error(w, "Machine not found", http.StatusNotFound)
				return
			}

			var latestMetrics *storage.MetricsHistory
			if m, err := machineService.GetLatestMetrics(r.Context(), machineID, user.ID); err == nil {
				latestMetrics = m
			}

			info := system.SystemInfo{
				Hostname:        firstNonEmpty(machine.Hostname, machine.Name),
				Platform:        machine.Platform,
				PlatformVersion: machine.PlatformVersion,
				KernelVersion:   machine.KernelVersion,
				CPUCores:        machine.CPUCores,
			}

			if machine.MemoryTotalMB > 0 {
				info.MemoryTotalMB = uint64(machine.MemoryTotalMB)
			}
			if machine.DiskTotalGB > 0 {
				info.DiskTotalGB = uint64(machine.DiskTotalGB)
			}
			if !machine.LastBootTime.IsZero() {
				info.LastBootTime = uint64(machine.LastBootTime.Unix())
			}
			if latestMetrics != nil && latestMetrics.UptimeSeconds > 0 {
				info.UptimeS = uint64(latestMetrics.UptimeSeconds)
			}

			if err := json.NewEncoder(w).Encode(info); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		if !localHostMetrics {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Local host system info disabled. Please select a machine.",
			})
			return
		}

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
