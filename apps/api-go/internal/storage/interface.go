package storage

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrPasswordResetNotFound is returned when a password reset token is not found or invalid
	ErrPasswordResetNotFound = errors.New("password reset token not found")
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
}

// Store defines the interface for user storage operations
type Store interface {
	// CreateUser creates a new user with the given email and password hash
	CreateUser(ctx context.Context, email, passwordHash string) (*User, error)

	// GetUserByEmail retrieves a user by their email address
	GetUserByEmail(ctx context.Context, email string) (*User, error)

	// GetUserByID retrieves a user by their ID
	GetUserByID(ctx context.Context, id int) (*User, error)

	// UpdateUserPassword updates the password hash for a user
	UpdateUserPassword(ctx context.Context, userID int, passwordHash string) error

	// UpsertAdmin creates or updates an admin user with the given email and password hash
	UpsertAdmin(ctx context.Context, email, passwordHash string) (*User, error)

	// CreatePasswordReset creates a password reset entry for a user
	CreatePasswordReset(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) (*PasswordReset, error)

	// GetPasswordResetByHash retrieves an active password reset entry by token hash
	GetPasswordResetByHash(ctx context.Context, tokenHash string) (*PasswordReset, error)

	// MarkPasswordResetUsed marks a password reset entry as used
	MarkPasswordResetUsed(ctx context.Context, id int) error

	// ListUsers retrieves all users ordered by email
	ListUsers(ctx context.Context) ([]User, error)

	// DeleteUser deletes a user by ID
	DeleteUser(ctx context.Context, id int) error

	// CountUsers returns the total number of users
	CountUsers(ctx context.Context) (int, error)

	// PromoteToAdmin promotes a user to admin status
	PromoteToAdmin(ctx context.Context, userID int) error

	// DeletePasswordResetsForUser deletes all password reset tokens for a user
	DeletePasswordResetsForUser(ctx context.Context, userID int) error

	// Alert Rules methods
	ListAlertRules(ctx context.Context) ([]AlertRule, error)
	CreateAlertRule(ctx context.Context, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*AlertRule, error)
	UpdateAlertRule(ctx context.Context, id int, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*AlertRule, error)
	DeleteAlertRule(ctx context.Context, id int) error

	// Alert Events methods
	ListAlertEvents(ctx context.Context, limit int) ([]AlertEvent, error)
	CreateAlertEvent(ctx context.Context, ruleID int, value float64) (*AlertEvent, error)
	AckAlertEvent(ctx context.Context, id int) error

	// Webhook methods
	ListWebhooks(ctx context.Context, userID int) ([]Webhook, error)
	GetWebhook(ctx context.Context, id int, userID int) (*Webhook, error)
	CreateWebhook(ctx context.Context, userID int, url, secretHash string) (*Webhook, error)
	UpdateWebhook(ctx context.Context, id int, userID int, url string, secretHash *string, isActive *bool) (*Webhook, error)
	DeleteWebhook(ctx context.Context, id int, userID int) error
	IncrementWebhookFailure(ctx context.Context, id int, lastErrorAt time.Time) error
	MarkWebhookSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error
	UpdateWebhookDeliveryState(ctx context.Context, id int, lastAttemptAt time.Time, cooldownUntil *time.Time) error

	// Email recipient methods
	ListEmailRecipients(ctx context.Context, userID int) ([]EmailRecipient, error)
	GetEmailRecipient(ctx context.Context, id int, userID int) (*EmailRecipient, error)
	CreateEmailRecipient(ctx context.Context, userID int, email string) (*EmailRecipient, error)
	UpdateEmailRecipient(ctx context.Context, id int, userID int, email string, isActive *bool) (*EmailRecipient, error)
	DeleteEmailRecipient(ctx context.Context, id int, userID int) error
	IncrementEmailFailure(ctx context.Context, id int, lastErrorAt time.Time) error
	MarkEmailSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error
	UpdateEmailDeliveryState(ctx context.Context, id int, lastAttemptAt time.Time, cooldownUntil *time.Time) error

	// Telegram recipient methods
	ListTelegramRecipients(ctx context.Context, userID int) ([]TelegramRecipient, error)
	GetTelegramRecipient(ctx context.Context, id int, userID int) (*TelegramRecipient, error)
	CreateTelegramRecipient(ctx context.Context, userID int, chatID string) (*TelegramRecipient, error)
	UpdateTelegramRecipient(ctx context.Context, id int, userID int, chatID string, isActive *bool) (*TelegramRecipient, error)
	DeleteTelegramRecipient(ctx context.Context, id int, userID int) error
	IncrementTelegramFailure(ctx context.Context, id int, lastErrorAt time.Time) error
	MarkTelegramSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error
	UpdateTelegramDeliveryState(ctx context.Context, id int, lastAttemptAt time.Time, cooldownUntil *time.Time) error

	// Machine methods
	CreateMachine(ctx context.Context, userID int, name, hostname, description, apiKeyHash string) (*Machine, error)
	GetMachineByID(ctx context.Context, id int) (*Machine, error)
	GetMachineByAPIKey(ctx context.Context, apiKeyHash string) (*Machine, error)
	ListMachines(ctx context.Context, userID int) ([]Machine, error)
	ListAllMachines(ctx context.Context) ([]Machine, error)
	UpdateMachineStatus(ctx context.Context, id int, status string, lastSeen time.Time) error
	UpdateMachineLastSeen(ctx context.Context, id int, lastSeen time.Time) error
	UpdateMachineSystemInfo(ctx context.Context, id int, info MachineSystemInfoUpdate) error
	UpdateMachineDetails(ctx context.Context, id int, updates map[string]interface{}) error
	DeleteMachine(ctx context.Context, id int, userID int) error

	// Machine heartbeat notification tracking
	RecordMachineOfflineNotification(ctx context.Context, machineID int, notifiedAt time.Time) error
	GetMachineLastOfflineNotification(ctx context.Context, machineID int) (time.Time, error)
	ClearMachineOfflineNotification(ctx context.Context, machineID int) error

	// Metrics history methods
	InsertMetrics(ctx context.Context, machineID int, cpuPct, memUsedPct, diskUsedPct float64, netRxBytes, netTxBytes int64, uptimeSeconds *float64, timestamp time.Time) error
	GetLatestMetrics(ctx context.Context, machineID int) (*MetricsHistory, error)
	GetMetricsHistory(ctx context.Context, machineID int, from, to time.Time, limit int) ([]MetricsHistory, error)

	// Close closes the storage connection
	Close() error
}

// PasswordReset represents a password reset token entry
type PasswordReset struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	TokenHash string     `json:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// AlertRule represents an alert rule for monitoring metrics
type AlertRule struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Metric       string    `json:"metric"` // "cpu_pct", "mem_used_pct", "disk_used_pct"
	ThresholdPct float64   `json:"threshold_pct"`
	Comparison   string    `json:"comparison"`    // "above" | "below"
	TriggerAfter int       `json:"trigger_after"` // number of consecutive samples before firing
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AlertEvent represents an alert event triggered by a rule
type AlertEvent struct {
	ID             int        `json:"id"`
	RuleID         int        `json:"rule_id"`
	TriggeredAt    time.Time  `json:"triggered_at"`
	Value          float64    `json:"value"`
	Acknowledged   bool       `json:"acknowledged"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
}

// Webhook represents a user webhook configuration for alert notifications
type Webhook struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	URL           string     `json:"url"`
	SecretHash    string     `json:"secret_hash"` // SHA-256 hex hash of user-provided secret
	IsActive      bool       `json:"is_active"`
	FailureCount  int        `json:"failure_count"`
	LastSuccessAt *time.Time `json:"last_success_at"`
	LastErrorAt   *time.Time `json:"last_error_at"`
	CooldownUntil *time.Time `json:"cooldown_until"`  // Circuit breaker cooldown end time
	LastAttemptAt *time.Time `json:"last_attempt_at"` // Last delivery attempt for rate limiting
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// EmailRecipient represents an email notification recipient for alert notifications
type EmailRecipient struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	Email         string     `json:"email"`
	IsActive      bool       `json:"is_active"`
	FailureCount  int        `json:"failure_count"`
	LastSuccessAt *time.Time `json:"last_success_at"`
	LastErrorAt   *time.Time `json:"last_error_at"`
	CooldownUntil *time.Time `json:"cooldown_until"`  // Circuit breaker cooldown end time
	LastAttemptAt *time.Time `json:"last_attempt_at"` // Last delivery attempt for rate limiting
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// TelegramRecipient represents a Telegram chat that receives alert notifications
type TelegramRecipient struct {
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
