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

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
