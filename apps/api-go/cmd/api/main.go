package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/alerts"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/config"
	router "github.com/Constantin-E-T/lunasentri/apps/api-go/internal/http"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/metrics"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/notifications"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/system"
)

func main() {
	// Record server start time for uptime calculation
	serverStartTime := time.Now()

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

	// Get LOCAL_HOST_METRICS flag from environment variable, default to false
	localHostMetrics := false
	if localHostMetricsEnv := os.Getenv("LOCAL_HOST_METRICS"); localHostMetricsEnv == "true" {
		localHostMetrics = true
		log.Println("Warning: LOCAL_HOST_METRICS enabled - local system metrics collection active")
		log.Println("         This is for development/testing only. For production, use machine agents.")
	} else {
		log.Println("LOCAL_HOST_METRICS disabled - metrics require registered machines (see docs/AGENT_GUIDELINES.md)")
	}

	// Initialize metrics collector (only if local host metrics enabled)
	var metricsCollector metrics.Collector
	if localHostMetrics {
		metricsCollector = metrics.NewSystemCollector()
	} else {
		// Use a no-op collector that returns empty metrics
		metricsCollector = metrics.NewNoOpCollector()
	}

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

	// Create HTTP router with all dependencies
	routerCfg := &router.RouterConfig{
		Collector:        metricsCollector,
		ServerStartTime:  serverStartTime,
		AuthService:      authService,
		AlertService:     alertService,
		SystemService:    systemService,
		Store:            store,
		WebhookNotifier:  webhookNotifier,
		TelegramNotifier: telegramNotifier,
		AccessTTL:        accessTTL,
		PasswordResetTTL: passwordResetTTL,
		SecureCookie:     secureCookie,
		LocalHostMetrics: localHostMetrics,
	}
	mux := router.NewRouter(routerCfg)

	// Create HTTP server with CORS middleware
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router.CORSMiddleware(mux),
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
