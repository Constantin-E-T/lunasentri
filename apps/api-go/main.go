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
)

// Metrics represents the system metrics response
type Metrics struct {
	CPUPct      float64 `json:"cpu_pct"`
	MemUsedPct  float64 `json:"mem_used_pct"`
	DiskUsedPct float64 `json:"disk_used_pct"`
	UptimeS     float64 `json:"uptime_s"`
}

var serverStartTime time.Time

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
		
		// Calculate real uptime in seconds
		uptime := time.Since(serverStartTime).Seconds()
		
		// TODO: Replace placeholder values with real system metrics
		metrics := Metrics{
			CPUPct:      45.2,  // Placeholder value
			MemUsedPct:  67.8,  // Placeholder value
			DiskUsedPct: 23.5,  // Placeholder value
			UptimeS:     uptime, // Real uptime since server start
		}
		
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

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
		log.Printf("LunaSentri API starting on port %s (endpoints: /, /health, /metrics) with CORS origin: %s", port, allowedOrigin)
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
