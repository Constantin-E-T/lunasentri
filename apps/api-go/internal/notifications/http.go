package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// WebhookRequest represents the request body for creating/updating webhooks
type WebhookRequest struct {
	URL      string `json:"url"`
	Secret   string `json:"secret"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// WebhookResponse represents the response body for webhook operations
type WebhookResponse struct {
	ID             int        `json:"id"`
	URL            string     `json:"url"`
	IsActive       bool       `json:"is_active"`
	FailureCount   int        `json:"failure_count"`
	LastSuccessAt  *time.Time `json:"last_success_at"`
	LastErrorAt    *time.Time `json:"last_error_at"`
	CooldownUntil  *time.Time `json:"cooldown_until"`
	LastAttemptAt  *time.Time `json:"last_attempt_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	SecretLastFour string     `json:"secret_last_four"`
}

// validateWebhookRequest validates webhook request data
func validateWebhookRequest(req *WebhookRequest) error {
	if req.URL == "" {
		return fmt.Errorf("url is required")
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return fmt.Errorf("invalid url format")
	}

	// Check if URL has a scheme
	if parsedURL.Scheme == "" {
		return fmt.Errorf("invalid url format")
	}

	// Require HTTPS
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("url must use HTTPS")
	}

	// Validate secret length (16-128 characters)
	if req.Secret != "" {
		if len(req.Secret) < 16 || len(req.Secret) > 128 {
			return fmt.Errorf("secret must be between 16 and 128 characters")
		}
	}

	return nil
}

// getSecretLastFour returns the last 4 characters of the secret for display
func getSecretLastFour(secret string) string {
	if len(secret) <= 4 {
		return secret
	}
	return secret[len(secret)-4:]
}

// webhookToResponse converts a storage.Webhook to WebhookResponse (hiding secret hash)
func webhookToResponse(webhook storage.Webhook, secretLastFour string) WebhookResponse {
	return WebhookResponse{
		ID:             webhook.ID,
		URL:            webhook.URL,
		IsActive:       webhook.IsActive,
		FailureCount:   webhook.FailureCount,
		LastSuccessAt:  webhook.LastSuccessAt,
		LastErrorAt:    webhook.LastErrorAt,
		CooldownUntil:  webhook.CooldownUntil,
		LastAttemptAt:  webhook.LastAttemptAt,
		CreatedAt:      webhook.CreatedAt,
		UpdatedAt:      webhook.UpdatedAt,
		SecretLastFour: secretLastFour,
	}
}

// HandleListWebhooks handles GET /notifications/webhooks
func HandleListWebhooks(store storage.Store) http.HandlerFunc {
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

		// Get webhooks for the user
		webhooks, err := store.ListWebhooks(r.Context(), user.ID)
		if err != nil {
			http.Error(w, `{"error":"Failed to list webhooks"}`, http.StatusInternalServerError)
			return
		}

		// Convert to response format (hide secret hashes)
		responses := make([]WebhookResponse, len(webhooks))
		for i, webhook := range webhooks {
			// We can't show the original secret, so just show "****" for last four
			responses[i] = webhookToResponse(webhook, "****")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responses)
	}
}

// HandleCreateWebhook handles POST /notifications/webhooks
func HandleCreateWebhook(store storage.Store) http.HandlerFunc {
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

		var req WebhookRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		// Set default value for is_active if not provided
		if req.IsActive == nil {
			defaultActive := true
			req.IsActive = &defaultActive
		}

		// Validate request
		if err := validateWebhookRequest(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Validate secret is provided for create
		if req.Secret == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "secret is required"})
			return
		}

		// Hash the secret
		secretHash := storage.HashSecret(req.Secret)
		secretLastFour := getSecretLastFour(req.Secret)

		// Create webhook
		webhook, err := store.CreateWebhook(r.Context(), user.ID, req.URL, secretHash)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{"error": "Webhook URL already exists for this user"})
				return
			}
			http.Error(w, `{"error":"Failed to create webhook"}`, http.StatusInternalServerError)
			return
		}

		// Update is_active if provided and different from default
		if req.IsActive != nil && !*req.IsActive {
			updatedWebhook, err := store.UpdateWebhook(r.Context(), webhook.ID, user.ID, "", nil, req.IsActive)
			if err != nil {
				http.Error(w, `{"error":"Failed to update webhook status"}`, http.StatusInternalServerError)
				return
			}
			webhook = updatedWebhook
		}

		// Return response
		response := webhookToResponse(*webhook, secretLastFour)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// HandleUpdateWebhook handles PUT /notifications/webhooks/{id}
func HandleUpdateWebhook(store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get user from context (set by RequireAuth middleware)
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract webhook ID from URL path
		path := strings.TrimPrefix(r.URL.Path, "/notifications/webhooks/")
		if path == "" || path == r.URL.Path {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Webhook ID required"})
			return
		}

		// Parse webhook ID
		webhookID, err := strconv.Atoi(path)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid webhook ID"})
			return
		}

		var req WebhookRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		// Validate if URL is provided
		if req.URL != "" {
			if err := validateWebhookRequest(&req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
		}

		// Prepare update parameters
		var secretHash *string
		var secretLastFour string

		if req.Secret != "" {
			// Validate secret length
			if len(req.Secret) < 16 || len(req.Secret) > 128 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "secret must be between 16 and 128 characters"})
				return
			}
			hash := storage.HashSecret(req.Secret)
			secretHash = &hash
			secretLastFour = getSecretLastFour(req.Secret)
		}

		// Update webhook
		webhook, err := store.UpdateWebhook(r.Context(), webhookID, user.ID, req.URL, secretHash, req.IsActive)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Webhook not found"})
				return
			}
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{"error": "Webhook URL already exists for this user"})
				return
			}
			http.Error(w, `{"error":"Failed to update webhook"}`, http.StatusInternalServerError)
			return
		}

		// If no secret was updated, show "****" for last four
		if secretLastFour == "" {
			secretLastFour = "****"
		}

		// Return response
		response := webhookToResponse(*webhook, secretLastFour)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// HandleDeleteWebhook handles DELETE /notifications/webhooks/{id}
func HandleDeleteWebhook(store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get user from context (set by RequireAuth middleware)
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract webhook ID from URL path
		path := strings.TrimPrefix(r.URL.Path, "/notifications/webhooks/")
		if path == "" || path == r.URL.Path {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Webhook ID required"})
			return
		}

		// Parse webhook ID
		webhookID, err := strconv.Atoi(path)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid webhook ID"})
			return
		}

		// Delete webhook
		err = store.DeleteWebhook(r.Context(), webhookID, user.ID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Webhook not found"})
				return
			}
			http.Error(w, `{"error":"Failed to delete webhook"}`, http.StatusInternalServerError)
			return
		}

		// Return 204 No Content on success
		w.WriteHeader(http.StatusNoContent)
	}
}

// TestWebhookResponse represents the response body for test webhook operation
type TestWebhookResponse struct {
	Status      string `json:"status"`
	WebhookID   int    `json:"webhook_id"`
	TriggeredAt string `json:"triggered_at"`
}

// HandleTestWebhook handles POST /notifications/webhooks/{id}/test
func HandleTestWebhook(notifier AlertNotifier, store storage.Store) http.HandlerFunc {
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

		// Extract webhook ID from URL path
		// Path format: /notifications/webhooks/{id}/test
		path := strings.TrimPrefix(r.URL.Path, "/notifications/webhooks/")
		path = strings.TrimSuffix(path, "/test")

		if path == "" || path == r.URL.Path {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Webhook ID required"})
			return
		}

		// Parse webhook ID
		webhookID, err := strconv.Atoi(path)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid webhook ID"})
			return
		}

		// Fetch webhook with ownership verification
		webhook, err := store.GetWebhook(r.Context(), webhookID, user.ID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Webhook not found"})
				return
			}
			http.Error(w, `{"error":"Failed to fetch webhook"}`, http.StatusInternalServerError)
			return
		}

		// Require webhook to be active
		if !webhook.IsActive {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Webhook must be active to send a test payload"})
			return
		}

		// Create a context with timeout for the test request
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Send test notification
		if err := notifier.SendTest(ctx, *webhook); err != nil {
			// Check if it's a rate limit error
			var rateLimitErr *RateLimitError
			if errors.As(err, &rateLimitErr) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{"error": rateLimitErr.Message})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to send test webhook: %v", err)})
			return
		}

		// Return success response
		response := TestWebhookResponse{
			Status:      "sent",
			WebhookID:   webhookID,
			TriggeredAt: time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
