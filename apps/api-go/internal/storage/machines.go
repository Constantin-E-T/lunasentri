package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Machine represents a monitored machine in the system
type Machine struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Hostname  string    `json:"hostname"`
	APIKey    string    `json:"api_key"` // Hashed API key
	Status    string    `json:"status"`  // "online", "offline"
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
}

// MetricsHistory represents historical metrics for a machine
type MetricsHistory struct {
	ID          int       `json:"id"`
	MachineID   int       `json:"machine_id"`
	CPUPct      float64   `json:"cpu_pct"`
	MemUsedPct  float64   `json:"mem_used_pct"`
	DiskUsedPct float64   `json:"disk_used_pct"`
	NetRxBytes  int64     `json:"net_rx_bytes"`
	NetTxBytes  int64     `json:"net_tx_bytes"`
	Timestamp   time.Time `json:"timestamp"`
}

// CreateMachine creates a new machine entry
func (s *SQLiteStore) CreateMachine(ctx context.Context, userID int, name, hostname, apiKeyHash string) (*Machine, error) {
	query := `
		INSERT INTO machines (user_id, name, hostname, api_key, status, created_at)
		VALUES (?, ?, ?, ?, 'offline', CURRENT_TIMESTAMP)
		RETURNING id, user_id, name, hostname, api_key, status, last_seen, created_at
	`

	var m Machine
	var lastSeen sql.NullTime
	err := s.db.QueryRowContext(ctx, query, userID, name, hostname, apiKeyHash).Scan(
		&m.ID, &m.UserID, &m.Name, &m.Hostname, &m.APIKey, &m.Status, &lastSeen, &m.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create machine: %w", err)
	}

	if lastSeen.Valid {
		m.LastSeen = lastSeen.Time
	}

	return &m, nil
}

// GetMachineByID retrieves a machine by ID
func (s *SQLiteStore) GetMachineByID(ctx context.Context, id int) (*Machine, error) {
	query := `
		SELECT id, user_id, name, hostname, api_key, status, last_seen, created_at
		FROM machines
		WHERE id = ?
	`

	var m Machine
	var lastSeen sql.NullTime
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.UserID, &m.Name, &m.Hostname, &m.APIKey, &m.Status, &lastSeen, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("machine not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get machine: %w", err)
	}

	if lastSeen.Valid {
		m.LastSeen = lastSeen.Time
	}

	return &m, nil
}

// GetMachineByAPIKey retrieves a machine by API key hash
func (s *SQLiteStore) GetMachineByAPIKey(ctx context.Context, apiKeyHash string) (*Machine, error) {
	query := `
		SELECT id, user_id, name, hostname, api_key, status, last_seen, created_at
		FROM machines
		WHERE api_key = ?
	`

	var m Machine
	var lastSeen sql.NullTime
	err := s.db.QueryRowContext(ctx, query, apiKeyHash).Scan(
		&m.ID, &m.UserID, &m.Name, &m.Hostname, &m.APIKey, &m.Status, &lastSeen, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("machine not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get machine: %w", err)
	}

	if lastSeen.Valid {
		m.LastSeen = lastSeen.Time
	}

	return &m, nil
}

// ListMachines retrieves all machines for a user
func (s *SQLiteStore) ListMachines(ctx context.Context, userID int) ([]Machine, error) {
	query := `
		SELECT id, user_id, name, hostname, api_key, status, last_seen, created_at
		FROM machines
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list machines: %w", err)
	}
	defer rows.Close()

	var machines []Machine
	for rows.Next() {
		var m Machine
		var lastSeen sql.NullTime
		if err := rows.Scan(&m.ID, &m.UserID, &m.Name, &m.Hostname, &m.APIKey, &m.Status, &lastSeen, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan machine: %w", err)
		}
		if lastSeen.Valid {
			m.LastSeen = lastSeen.Time
		}
		machines = append(machines, m)
	}

	return machines, nil
}

// UpdateMachineStatus updates the status and last_seen timestamp of a machine
func (s *SQLiteStore) UpdateMachineStatus(ctx context.Context, id int, status string, lastSeen time.Time) error {
	query := `
		UPDATE machines
		SET status = ?, last_seen = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query, status, lastSeen, id)
	if err != nil {
		return fmt.Errorf("failed to update machine status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("machine not found")
	}

	return nil
}

// DeleteMachine deletes a machine and all its metrics history
func (s *SQLiteStore) DeleteMachine(ctx context.Context, id int, userID int) error {
	// First verify the machine belongs to the user
	query := `DELETE FROM machines WHERE id = ? AND user_id = ?`

	result, err := s.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete machine: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("machine not found or access denied")
	}

	return nil
}

// InsertMetrics inserts a metrics record into the history table
func (s *SQLiteStore) InsertMetrics(ctx context.Context, machineID int, cpuPct, memUsedPct, diskUsedPct float64, netRxBytes, netTxBytes int64, timestamp time.Time) error {
	query := `
		INSERT INTO metrics_history (machine_id, cpu_pct, mem_used_pct, disk_used_pct, net_rx_bytes, net_tx_bytes, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query, machineID, cpuPct, memUsedPct, diskUsedPct, netRxBytes, netTxBytes, timestamp)
	if err != nil {
		return fmt.Errorf("failed to insert metrics: %w", err)
	}

	return nil
}

// GetLatestMetrics retrieves the most recent metrics for a machine
func (s *SQLiteStore) GetLatestMetrics(ctx context.Context, machineID int) (*MetricsHistory, error) {
	query := `
		SELECT id, machine_id, cpu_pct, mem_used_pct, disk_used_pct, net_rx_bytes, net_tx_bytes, timestamp
		FROM metrics_history
		WHERE machine_id = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var m MetricsHistory
	err := s.db.QueryRowContext(ctx, query, machineID).Scan(
		&m.ID, &m.MachineID, &m.CPUPct, &m.MemUsedPct, &m.DiskUsedPct, &m.NetRxBytes, &m.NetTxBytes, &m.Timestamp,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no metrics found for machine")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest metrics: %w", err)
	}

	return &m, nil
}

// GetMetricsHistory retrieves metrics history for a machine within a time range
func (s *SQLiteStore) GetMetricsHistory(ctx context.Context, machineID int, from, to time.Time, limit int) ([]MetricsHistory, error) {
	query := `
		SELECT id, machine_id, cpu_pct, mem_used_pct, disk_used_pct, net_rx_bytes, net_tx_bytes, timestamp
		FROM metrics_history
		WHERE machine_id = ? AND timestamp >= ? AND timestamp <= ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, machineID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics history: %w", err)
	}
	defer rows.Close()

	var metrics []MetricsHistory
	for rows.Next() {
		var m MetricsHistory
		if err := rows.Scan(&m.ID, &m.MachineID, &m.CPUPct, &m.MemUsedPct, &m.DiskUsedPct, &m.NetRxBytes, &m.NetTxBytes, &m.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan metrics: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}
