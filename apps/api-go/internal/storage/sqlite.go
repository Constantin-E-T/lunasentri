package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore implements the Store interface using SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite-backed store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign key constraints
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	store := &SQLiteStore{db: db}

	// Run migrations
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

// migrate creates the necessary tables and schema
func (s *SQLiteStore) migrate() error {
	// Create migrations table to track applied migrations
	createMigrationsTable := `
	CREATE TABLE IF NOT EXISTS migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version TEXT UNIQUE NOT NULL,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := s.db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// List of migrations to apply
	migrations := []migration{
		{
			version: "001_create_users_table",
			sql: `
            CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                email TEXT UNIQUE NOT NULL,
                password_hash TEXT NOT NULL,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
            );`,
		},
		{
			version: "002_create_password_resets",
			sql: `
            CREATE TABLE IF NOT EXISTS password_resets (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                user_id INTEGER NOT NULL,
                token_hash TEXT UNIQUE NOT NULL,
                expires_at DATETIME NOT NULL,
                used_at DATETIME,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
            );
            CREATE INDEX IF NOT EXISTS idx_password_resets_token_hash ON password_resets(token_hash);
            `,
		},
		{
			version: "003_add_is_admin_to_users",
			sql: `
            ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT 0 NOT NULL;
            `,
		},
		{
			version: "004_alert_rules",
			sql: `
            CREATE TABLE IF NOT EXISTS alert_rules (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                metric TEXT NOT NULL CHECK (metric IN ('cpu_pct', 'mem_used_pct', 'disk_used_pct')),
                threshold_pct REAL NOT NULL CHECK (threshold_pct >= 0 AND threshold_pct <= 100),
                comparison TEXT NOT NULL CHECK (comparison IN ('above', 'below')),
                trigger_after INTEGER NOT NULL CHECK (trigger_after >= 1),
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
            );
            CREATE INDEX IF NOT EXISTS idx_alert_rules_metric ON alert_rules(metric);
            `,
		},
		{
			version: "005_alert_events",
			sql: `
            CREATE TABLE IF NOT EXISTS alert_events (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                rule_id INTEGER NOT NULL,
                triggered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                value REAL NOT NULL,
                acknowledged BOOLEAN DEFAULT 0 NOT NULL,
                acknowledged_at DATETIME,
                FOREIGN KEY(rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE
            );
            CREATE INDEX IF NOT EXISTS idx_alert_events_rule_id ON alert_events(rule_id);
            CREATE INDEX IF NOT EXISTS idx_alert_events_acknowledged ON alert_events(acknowledged);
            CREATE INDEX IF NOT EXISTS idx_alert_events_triggered_at ON alert_events(triggered_at);
            `,
		},
		{
			version: "006_webhooks",
			sql: `
            CREATE TABLE IF NOT EXISTS webhooks (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                user_id INTEGER NOT NULL,
                url TEXT NOT NULL,
                secret_hash TEXT NOT NULL,
                is_active BOOLEAN DEFAULT 1 NOT NULL,
                failure_count INTEGER DEFAULT 0 NOT NULL,
                last_success_at DATETIME,
                last_error_at DATETIME,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
                UNIQUE(user_id, url)
            );
            CREATE INDEX IF NOT EXISTS idx_webhooks_user_id ON webhooks(user_id);
            CREATE INDEX IF NOT EXISTS idx_webhooks_active ON webhooks(is_active);
            `,
		},
	}

	// Apply each migration if not already applied
	for _, m := range migrations {
		if err := s.applyMigration(m); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", m.version, err)
		}
	}

	return nil
}

// migration represents a database migration
type migration struct {
	version string
	sql     string
}

// applyMigration applies a single migration if it hasn't been applied already
func (s *SQLiteStore) applyMigration(m migration) error {
	// Check if migration has already been applied
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM migrations WHERE version = ?", m.version).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		// Migration already applied
		return nil
	}

	// Apply the migration
	if _, err := s.db.Exec(m.sql); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record that migration was applied
	_, err = s.db.Exec("INSERT INTO migrations (version) VALUES (?)", m.version)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

// CreateUser creates a new user with the given email and password hash
func (s *SQLiteStore) CreateUser(ctx context.Context, email, passwordHash string) (*User, error) {
	query := `
	INSERT INTO users (email, password_hash, is_admin, created_at)
	VALUES (?, ?, ?, ?)
	RETURNING id, email, password_hash, is_admin, created_at`

	now := time.Now()
	user := &User{}

	err := s.db.QueryRowContext(ctx, query, email, passwordHash, false, now).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			return nil, fmt.Errorf("user with email %s already exists", email)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email address
func (s *SQLiteStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, password_hash, is_admin, created_at FROM users WHERE email = ?`

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by their ID
func (s *SQLiteStore) GetUserByID(ctx context.Context, id int) (*User, error) {
	query := `SELECT id, email, password_hash, is_admin, created_at FROM users WHERE id = ?`

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpsertAdmin creates or updates an admin user with the given email and password hash
func (s *SQLiteStore) UpsertAdmin(ctx context.Context, email, passwordHash string) (*User, error) {
	// Try to get existing user first
	existingUser, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		// User doesn't exist, create new one
		if strings.Contains(err.Error(), "not found") {
			return s.CreateUser(ctx, email, passwordHash)
		}
		if errors.Is(err, ErrUserNotFound) {
			return s.CreateUser(ctx, email, passwordHash)
		}
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// User exists, update password hash
	query := `UPDATE users SET password_hash = ? WHERE email = ?`
	_, err = s.db.ExecContext(ctx, query, passwordHash, email)
	if err != nil {
		return nil, fmt.Errorf("failed to update user password: %w", err)
	}

	// Return updated user
	existingUser.PasswordHash = passwordHash
	return existingUser, nil
}

// UpdateUserPassword updates the password hash for a user
func (s *SQLiteStore) UpdateUserPassword(ctx context.Context, userID int, passwordHash string) error {
	query := `UPDATE users SET password_hash = ? WHERE id = ?`
	res, err := s.db.ExecContext(ctx, query, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify password update: %w", err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// CreatePasswordReset creates a password reset entry for a user
func (s *SQLiteStore) CreatePasswordReset(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) (*PasswordReset, error) {
	query := `
    INSERT INTO password_resets (user_id, token_hash, expires_at)
    VALUES (?, ?, ?)
    RETURNING id, user_id, token_hash, expires_at, used_at, created_at`

	var pr PasswordReset
	err := s.db.QueryRowContext(ctx, query, userID, tokenHash, expiresAt).Scan(
		&pr.ID,
		&pr.UserID,
		&pr.TokenHash,
		&pr.ExpiresAt,
		&pr.UsedAt,
		&pr.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create password reset: %w", err)
	}

	return &pr, nil
}

// GetPasswordResetByHash retrieves an active password reset entry by token hash
func (s *SQLiteStore) GetPasswordResetByHash(ctx context.Context, tokenHash string) (*PasswordReset, error) {
	query := `
    SELECT id, user_id, token_hash, expires_at, used_at, created_at
    FROM password_resets
    WHERE token_hash = ?`

	var pr PasswordReset
	err := s.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&pr.ID,
		&pr.UserID,
		&pr.TokenHash,
		&pr.ExpiresAt,
		&pr.UsedAt,
		&pr.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPasswordResetNotFound
		}
		return nil, fmt.Errorf("failed to get password reset: %w", err)
	}

	if pr.UsedAt != nil || time.Now().After(pr.ExpiresAt) {
		return nil, ErrPasswordResetNotFound
	}

	return &pr, nil
}

// MarkPasswordResetUsed marks a password reset entry as used
func (s *SQLiteStore) MarkPasswordResetUsed(ctx context.Context, id int) error {
	now := time.Now()
	query := `UPDATE password_resets SET used_at = ? WHERE id = ?`
	res, err := s.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark password reset used: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify password reset update: %w", err)
	}
	if rows == 0 {
		return ErrPasswordResetNotFound
	}
	return nil
}

// ListUsers retrieves all users ordered by email
func (s *SQLiteStore) ListUsers(ctx context.Context) ([]User, error) {
	query := `SELECT id, email, password_hash, is_admin, created_at FROM users ORDER BY email ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsAdmin, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// DeleteUser deletes a user by ID
func (s *SQLiteStore) DeleteUser(ctx context.Context, id int) error {
	// First, check if the user exists and if they are admin
	var exists bool
	var isAdmin bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = ?), COALESCE((SELECT is_admin FROM users WHERE id = ?), 0)", id, id).Scan(&exists, &isAdmin)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if !exists {
		return ErrUserNotFound
	}

	// Check if this is the last admin
	if isAdmin {
		var adminCount int
		err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_admin = 1").Scan(&adminCount)
		if err != nil {
			return fmt.Errorf("failed to count admins: %w", err)
		}

		if adminCount <= 1 {
			return fmt.Errorf("cannot delete the last admin")
		}
	}

	// Delete the user
	query := `DELETE FROM users WHERE id = ?`
	_, err = s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// CountUsers returns the total number of users
func (s *SQLiteStore) CountUsers(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// PromoteToAdmin promotes a user to admin status
func (s *SQLiteStore) PromoteToAdmin(ctx context.Context, userID int) error {
	query := `UPDATE users SET is_admin = 1 WHERE id = ?`
	res, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to promote user to admin: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify promotion: %w", err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DeletePasswordResetsForUser deletes all password reset tokens for a user
func (s *SQLiteStore) DeletePasswordResetsForUser(ctx context.Context, userID int) error {
	query := `DELETE FROM password_resets WHERE user_id = ?`
	_, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete password resets: %w", err)
	}
	return nil
}

// Alert Rules methods

// ListAlertRules retrieves all alert rules
func (s *SQLiteStore) ListAlertRules(ctx context.Context) ([]AlertRule, error) {
	query := `SELECT id, name, metric, threshold_pct, comparison, trigger_after, created_at, updated_at 
              FROM alert_rules ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert rules: %w", err)
	}
	defer rows.Close()

	var rules []AlertRule
	for rows.Next() {
		var rule AlertRule
		err := rows.Scan(&rule.ID, &rule.Name, &rule.Metric, &rule.ThresholdPct,
			&rule.Comparison, &rule.TriggerAfter, &rule.CreatedAt, &rule.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert rule: %w", err)
		}
		rules = append(rules, rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate alert rules: %w", err)
	}

	return rules, nil
}

// CreateAlertRule creates a new alert rule
func (s *SQLiteStore) CreateAlertRule(ctx context.Context, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*AlertRule, error) {
	now := time.Now()
	query := `INSERT INTO alert_rules (name, metric, threshold_pct, comparison, trigger_after, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?)
              RETURNING id, name, metric, threshold_pct, comparison, trigger_after, created_at, updated_at`

	rule := &AlertRule{}
	err := s.db.QueryRowContext(ctx, query, name, metric, thresholdPct, comparison, triggerAfter, now, now).Scan(
		&rule.ID, &rule.Name, &rule.Metric, &rule.ThresholdPct,
		&rule.Comparison, &rule.TriggerAfter, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert rule: %w", err)
	}

	return rule, nil
}

// UpdateAlertRule updates an existing alert rule
func (s *SQLiteStore) UpdateAlertRule(ctx context.Context, id int, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*AlertRule, error) {
	now := time.Now()
	query := `UPDATE alert_rules 
              SET name = ?, metric = ?, threshold_pct = ?, comparison = ?, trigger_after = ?, updated_at = ?
              WHERE id = ?
              RETURNING id, name, metric, threshold_pct, comparison, trigger_after, created_at, updated_at`

	rule := &AlertRule{}
	err := s.db.QueryRowContext(ctx, query, name, metric, thresholdPct, comparison, triggerAfter, now, id).Scan(
		&rule.ID, &rule.Name, &rule.Metric, &rule.ThresholdPct,
		&rule.Comparison, &rule.TriggerAfter, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alert rule with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to update alert rule: %w", err)
	}

	return rule, nil
}

// DeleteAlertRule deletes an alert rule (and cascades to delete related events)
func (s *SQLiteStore) DeleteAlertRule(ctx context.Context, id int) error {
	query := `DELETE FROM alert_rules WHERE id = ?`
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify deletion: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("alert rule with id %d not found", id)
	}

	return nil
}

// Alert Events methods

// ListAlertEvents retrieves recent alert events (unacknowledged first, limited)
func (s *SQLiteStore) ListAlertEvents(ctx context.Context, limit int) ([]AlertEvent, error) {
	query := `SELECT id, rule_id, triggered_at, value, acknowledged, acknowledged_at 
              FROM alert_events 
              ORDER BY acknowledged ASC, triggered_at DESC 
              LIMIT ?`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert events: %w", err)
	}
	defer rows.Close()

	var events []AlertEvent
	for rows.Next() {
		var event AlertEvent
		err := rows.Scan(&event.ID, &event.RuleID, &event.TriggeredAt,
			&event.Value, &event.Acknowledged, &event.AcknowledgedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert event: %w", err)
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate alert events: %w", err)
	}

	return events, nil
}

// CreateAlertEvent creates a new alert event
func (s *SQLiteStore) CreateAlertEvent(ctx context.Context, ruleID int, value float64) (*AlertEvent, error) {
	now := time.Now()
	query := `INSERT INTO alert_events (rule_id, triggered_at, value, acknowledged)
              VALUES (?, ?, ?, ?)
              RETURNING id, rule_id, triggered_at, value, acknowledged, acknowledged_at`

	event := &AlertEvent{}
	err := s.db.QueryRowContext(ctx, query, ruleID, now, value, false).Scan(
		&event.ID, &event.RuleID, &event.TriggeredAt,
		&event.Value, &event.Acknowledged, &event.AcknowledgedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert event: %w", err)
	}

	return event, nil
}

// AckAlertEvent acknowledges an alert event
func (s *SQLiteStore) AckAlertEvent(ctx context.Context, id int) error {
	now := time.Now()
	query := `UPDATE alert_events 
              SET acknowledged = 1, acknowledged_at = ?
              WHERE id = ? AND acknowledged = 0`

	res, err := s.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to acknowledge alert event: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify acknowledgment: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("alert event with id %d not found or already acknowledged", id)
	}

	return nil
}

// Webhook methods

// ListWebhooks returns all webhooks for a specific user
func (s *SQLiteStore) ListWebhooks(ctx context.Context, userID int) ([]Webhook, error) {
	query := `SELECT id, user_id, url, secret_hash, is_active, failure_count, 
              last_success_at, last_error_at, created_at, updated_at
              FROM webhooks WHERE user_id = ?
              ORDER BY created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []Webhook
	for rows.Next() {
		var w Webhook
		err := rows.Scan(&w.ID, &w.UserID, &w.URL, &w.SecretHash, &w.IsActive,
			&w.FailureCount, &w.LastSuccessAt, &w.LastErrorAt,
			&w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}
		webhooks = append(webhooks, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate webhooks: %w", err)
	}

	return webhooks, nil
}

// CreateWebhook creates a new webhook for a user
func (s *SQLiteStore) CreateWebhook(ctx context.Context, userID int, url, secretHash string) (*Webhook, error) {
	now := time.Now()
	query := `INSERT INTO webhooks (user_id, url, secret_hash, is_active, failure_count, 
              created_at, updated_at)
              VALUES (?, ?, ?, 1, 0, ?, ?)
              RETURNING id, user_id, url, secret_hash, is_active, failure_count,
              last_success_at, last_error_at, created_at, updated_at`

	webhook := &Webhook{}
	err := s.db.QueryRowContext(ctx, query, userID, url, secretHash, now, now).Scan(
		&webhook.ID, &webhook.UserID, &webhook.URL, &webhook.SecretHash,
		&webhook.IsActive, &webhook.FailureCount, &webhook.LastSuccessAt,
		&webhook.LastErrorAt, &webhook.CreatedAt, &webhook.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	return webhook, nil
}

// UpdateWebhook updates an existing webhook
func (s *SQLiteStore) UpdateWebhook(ctx context.Context, id int, userID int, url string, secretHash *string, isActive *bool) (*Webhook, error) {
	now := time.Now()

	// Build dynamic query based on what fields are being updated
	setParts := []string{"updated_at = ?"}
	args := []interface{}{now}

	if url != "" {
		setParts = append(setParts, "url = ?")
		args = append(args, url)
	}

	if secretHash != nil {
		setParts = append(setParts, "secret_hash = ?")
		args = append(args, *secretHash)
	}

	if isActive != nil {
		setParts = append(setParts, "is_active = ?")
		args = append(args, *isActive)
	}

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`UPDATE webhooks SET %s WHERE id = ? AND user_id = ?`, setClause)
	args = append(args, id, userID)

	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to verify update: %w", err)
	}
	if rows == 0 {
		return nil, fmt.Errorf("webhook with id %d not found for user %d", id, userID)
	}

	// Return updated webhook
	selectQuery := `SELECT id, user_id, url, secret_hash, is_active, failure_count,
                    last_success_at, last_error_at, created_at, updated_at
                    FROM webhooks WHERE id = ? AND user_id = ?`

	webhook := &Webhook{}
	err = s.db.QueryRowContext(ctx, selectQuery, id, userID).Scan(
		&webhook.ID, &webhook.UserID, &webhook.URL, &webhook.SecretHash,
		&webhook.IsActive, &webhook.FailureCount, &webhook.LastSuccessAt,
		&webhook.LastErrorAt, &webhook.CreatedAt, &webhook.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated webhook: %w", err)
	}

	return webhook, nil
}

// DeleteWebhook deletes a webhook for a user
func (s *SQLiteStore) DeleteWebhook(ctx context.Context, id int, userID int) error {
	query := `DELETE FROM webhooks WHERE id = ? AND user_id = ?`

	res, err := s.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify deletion: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("webhook with id %d not found for user %d", id, userID)
	}

	return nil
}

// IncrementWebhookFailure increments the failure count and updates last error time
func (s *SQLiteStore) IncrementWebhookFailure(ctx context.Context, id int, lastErrorAt time.Time) error {
	query := `UPDATE webhooks 
              SET failure_count = failure_count + 1, last_error_at = ?, updated_at = ?
              WHERE id = ?`

	res, err := s.db.ExecContext(ctx, query, lastErrorAt, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to increment webhook failure: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify failure increment: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("webhook with id %d not found", id)
	}

	return nil
}

// MarkWebhookSuccess resets failure count and updates last success time
func (s *SQLiteStore) MarkWebhookSuccess(ctx context.Context, id int, lastSuccessAt time.Time) error {
	query := `UPDATE webhooks 
              SET failure_count = 0, last_success_at = ?, updated_at = ?
              WHERE id = ?`

	res, err := s.db.ExecContext(ctx, query, lastSuccessAt, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to mark webhook success: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify success update: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("webhook with id %d not found", id)
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
