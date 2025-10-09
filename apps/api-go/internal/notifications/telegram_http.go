package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// TelegramRecipientRequest represents the request body for creating/updating Telegram recipients
type TelegramRecipientRequest struct {
	ChatID   string `json:"chat_id"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// TelegramRecipientResponse represents the response body for Telegram recipient operations
type TelegramRecipientResponse struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	ChatID        string     `json:"chat_id"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`
	LastSuccessAt *time.Time `json:"last_success_at,omitempty"`
	LastErrorAt   *time.Time `json:"last_error_at,omitempty"`
	FailureCount  int        `json:"failure_count"`
	CooldownUntil *time.Time `json:"cooldown_until,omitempty"`
}

// validateTelegramRecipientRequest validates Telegram recipient request data
func validateTelegramRecipientRequest(req *TelegramRecipientRequest) error {
	if req.ChatID == "" {
		return fmt.Errorf("chat_id is required")
	}

	// Validate chat_id format (should be numeric, can be negative for groups)
	if _, err := strconv.ParseInt(req.ChatID, 10, 64); err != nil {
		return fmt.Errorf("chat_id must be a valid numeric string")
	}

	return nil
}

// telegramRecipientToResponse converts a storage.TelegramRecipient to TelegramRecipientResponse
func telegramRecipientToResponse(recipient storage.TelegramRecipient) TelegramRecipientResponse {
	return TelegramRecipientResponse{
		ID:            recipient.ID,
		UserID:        recipient.UserID,
		ChatID:        recipient.ChatID,
		IsActive:      recipient.IsActive,
		CreatedAt:     recipient.CreatedAt,
		LastAttemptAt: recipient.LastAttemptAt,
		LastSuccessAt: recipient.LastSuccessAt,
		LastErrorAt:   recipient.LastErrorAt,
		FailureCount:  recipient.FailureCount,
		CooldownUntil: recipient.CooldownUntil,
	}
}

// HandleListTelegramRecipients handles GET /notifications/telegram
func HandleListTelegramRecipients(store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, ok := r.Context().Value(auth.UserContextKey).(*storage.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		recipients, err := store.ListTelegramRecipients(r.Context(), user.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list telegram recipients: %v", err), http.StatusInternalServerError)
			return
		}

		response := make([]TelegramRecipientResponse, len(recipients))
		for i, recipient := range recipients {
			response[i] = telegramRecipientToResponse(recipient)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// HandleCreateTelegramRecipient handles POST /notifications/telegram
func HandleCreateTelegramRecipient(store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, ok := r.Context().Value(auth.UserContextKey).(*storage.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req TelegramRecipientRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if err := validateTelegramRecipientRequest(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		recipient, err := store.CreateTelegramRecipient(r.Context(), user.ID, req.ChatID)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to create telegram recipient: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(telegramRecipientToResponse(*recipient))
	}
}

// HandleUpdateTelegramRecipient handles PUT /notifications/telegram/{id}
func HandleUpdateTelegramRecipient(store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, ok := r.Context().Value(auth.UserContextKey).(*storage.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract ID from URL path
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 3 {
			http.Error(w, "Invalid URL path", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(pathParts[2])
		if err != nil {
			http.Error(w, "Invalid telegram recipient ID", http.StatusBadRequest)
			return
		}

		var req TelegramRecipientRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		// Validate chat_id if provided
		if req.ChatID != "" {
			if err := validateTelegramRecipientRequest(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		recipient, err := store.UpdateTelegramRecipient(r.Context(), id, user.ID, req.ChatID, req.IsActive)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to update telegram recipient: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(telegramRecipientToResponse(*recipient))
	}
}

// HandleDeleteTelegramRecipient handles DELETE /notifications/telegram/{id}
func HandleDeleteTelegramRecipient(store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, ok := r.Context().Value(auth.UserContextKey).(*storage.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract ID from URL path
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 3 {
			http.Error(w, "Invalid URL path", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(pathParts[2])
		if err != nil {
			http.Error(w, "Invalid telegram recipient ID", http.StatusBadRequest)
			return
		}

		if err := store.DeleteTelegramRecipient(r.Context(), id, user.ID); err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unauthorized") {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to delete telegram recipient: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// HandleTestTelegram handles POST /notifications/telegram/{id}/test
func HandleTestTelegram(store storage.Store, telegramNotifier *TelegramNotifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, ok := r.Context().Value(auth.UserContextKey).(*storage.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract ID from URL path (format: /notifications/telegram/{id}/test)
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 4 {
			http.Error(w, "Invalid URL path", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(pathParts[2])
		if err != nil {
			http.Error(w, "Invalid telegram recipient ID", http.StatusBadRequest)
			return
		}

		// Verify ownership
		recipient, err := store.GetTelegramRecipient(r.Context(), id, user.ID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to get telegram recipient: %v", err), http.StatusInternalServerError)
			return
		}

		// Send test message
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		if err := telegramNotifier.SendTest(ctx, *recipient); err != nil {
			http.Error(w, fmt.Sprintf("Failed to send test message: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Test message sent successfully",
		})
	}
}
