package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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

// newServer creates a new HTTP server with the given collector
func newServer(collector metrics.Collector, startTime time.Time) *http.ServeMux {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "LunaSentri API - Coming Soon")
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy"}`)
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// WebSocket endpoint for streaming metrics
	mux.HandleFunc("/ws", handleWebSocket(collector, startTime))

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

	// Initialize metrics collector
	metricsCollector := metrics.NewSystemCollector()

	// Create server with real collector
	mux := newServer(metricsCollector, serverStartTime)

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
