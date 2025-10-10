package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Machine represents a monitored machine in the system
type Machine struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	Name            string    `json:"name"`
	Hostname        string    `json:"hostname"`
	APIKey          string    `json:"api_key"` // Hashed API key
	Status          string    `json:"status"`  // "online", "offline"
	LastSeen        time.Time `json:"last_seen"`
	Platform        string    `json:"platform"`
	PlatformVersion string    `json:"platform_version"`
	KernelVersion   string    `json:"kernel_version"`
	CPUCores        int       `json:"cpu_cores"`
	MemoryTotalMB   int64     `json:"memory_total_mb"`
	DiskTotalGB     int64     `json:"disk_total_gb"`
	LastBootTime    time.Time `json:"last_boot_time"`
	CreatedAt       time.Time `json:"created_at"`
}

// MetricsHistory represents historical metrics for a machine
type MetricsHistory struct {
	ID            int       `json:"id"`
	MachineID     int       `json:"machine_id"`
	CPUPct        float64   `json:"cpu_pct"`
	MemUsedPct    float64   `json:"mem_used_pct"`
	DiskUsedPct   float64   `json:"disk_used_pct"`
	NetRxBytes    int64     `json:"net_rx_bytes"`
	NetTxBytes    int64     `json:"net_tx_bytes"`
	UptimeSeconds float64   `json:"uptime_seconds"`
	Timestamp     time.Time `json:"timestamp"`
}

// MachineSystemInfoUpdate represents optional system info updates for a machine.
type MachineSystemInfoUpdate struct {
	Hostname        *string
	Platform        *string
	PlatformVersion *string
	KernelVersion   *string
	CPUCores        *int
	MemoryTotalMB   *int64
	DiskTotalGB     *int64
	LastBootTime    *time.Time
}

// CreateMachine creates a new machine entry
func (s *SQLiteStore) CreateMachine(ctx context.Context, userID int, name, hostname, apiKeyHash string) (*Machine, error) {
	query := `
		INSERT INTO machines (user_id, name, hostname, api_key, status, created_at)
		VALUES (?, ?, ?, ?, 'offline', CURRENT_TIMESTAMP)
		RETURNING id, user_id, name, hostname, api_key, status, last_seen, platform, platform_version, kernel_version, cpu_cores, memory_total_mb, disk_total_gb, last_boot_time, created_at
	`

	var m Machine
	var lastSeen sql.NullTime
	var platform sql.NullString
	var platformVersion sql.NullString
	var kernelVersion sql.NullString
	var cpuCores sql.NullInt64
	var memoryTotal sql.NullInt64
	var diskTotal sql.NullInt64
	var lastBoot sql.NullTime

	err := s.db.QueryRowContext(ctx, query, userID, name, hostname, apiKeyHash).Scan(
		&m.ID, &m.UserID, &m.Name, &m.Hostname, &m.APIKey, &m.Status, &lastSeen,
		&platform, &platformVersion, &kernelVersion, &cpuCores, &memoryTotal, &diskTotal, &lastBoot, &m.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create machine: %w", err)
	}

	if lastSeen.Valid {
		m.LastSeen = lastSeen.Time
	}
	if platform.Valid {
		m.Platform = platform.String
	}
	if platformVersion.Valid {
		m.PlatformVersion = platformVersion.String
	}
	if kernelVersion.Valid {
		m.KernelVersion = kernelVersion.String
	}
	if cpuCores.Valid {
		m.CPUCores = int(cpuCores.Int64)
	}
	if memoryTotal.Valid {
		m.MemoryTotalMB = memoryTotal.Int64
	}
	if diskTotal.Valid {
		m.DiskTotalGB = diskTotal.Int64
	}
	if lastBoot.Valid {
		m.LastBootTime = lastBoot.Time
	}

	return &m, nil
}

// GetMachineByID retrieves a machine by ID
func (s *SQLiteStore) GetMachineByID(ctx context.Context, id int) (*Machine, error) {
	query := `
		SELECT id, user_id, name, hostname, api_key, status, last_seen,
		       platform, platform_version, kernel_version, cpu_cores,
		       memory_total_mb, disk_total_gb, last_boot_time, created_at
		FROM machines
		WHERE id = ?
	`

	var m Machine
	var lastSeen sql.NullTime
	var platform sql.NullString
	var platformVersion sql.NullString
	var kernelVersion sql.NullString
	var cpuCores sql.NullInt64
	var memoryTotal sql.NullInt64
	var diskTotal sql.NullInt64
	var lastBoot sql.NullTime

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.UserID, &m.Name, &m.Hostname, &m.APIKey, &m.Status, &lastSeen,
		&platform, &platformVersion, &kernelVersion, &cpuCores, &memoryTotal, &diskTotal, &lastBoot, &m.CreatedAt,
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
	if platform.Valid {
		m.Platform = platform.String
	}
	if platformVersion.Valid {
		m.PlatformVersion = platformVersion.String
	}
	if kernelVersion.Valid {
		m.KernelVersion = kernelVersion.String
	}
	if cpuCores.Valid {
		m.CPUCores = int(cpuCores.Int64)
	}
	if memoryTotal.Valid {
		m.MemoryTotalMB = memoryTotal.Int64
	}
	if diskTotal.Valid {
		m.DiskTotalGB = diskTotal.Int64
	}
	if lastBoot.Valid {
		m.LastBootTime = lastBoot.Time
	}

	return &m, nil
}

// GetMachineByAPIKey retrieves a machine by API key hash
func (s *SQLiteStore) GetMachineByAPIKey(ctx context.Context, apiKeyHash string) (*Machine, error) {
	query := `
		SELECT id, user_id, name, hostname, api_key, status, last_seen,
		       platform, platform_version, kernel_version, cpu_cores,
		       memory_total_mb, disk_total_gb, last_boot_time, created_at
		FROM machines
		WHERE api_key = ?
	`

	var m Machine
	var lastSeen sql.NullTime
	var platform sql.NullString
	var platformVersion sql.NullString
	var kernelVersion sql.NullString
	var cpuCores sql.NullInt64
	var memoryTotal sql.NullInt64
	var diskTotal sql.NullInt64
	var lastBoot sql.NullTime

	err := s.db.QueryRowContext(ctx, query, apiKeyHash).Scan(
		&m.ID, &m.UserID, &m.Name, &m.Hostname, &m.APIKey, &m.Status, &lastSeen,
		&platform, &platformVersion, &kernelVersion, &cpuCores, &memoryTotal, &diskTotal, &lastBoot, &m.CreatedAt,
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
	if platform.Valid {
		m.Platform = platform.String
	}
	if platformVersion.Valid {
		m.PlatformVersion = platformVersion.String
	}
	if kernelVersion.Valid {
		m.KernelVersion = kernelVersion.String
	}
	if cpuCores.Valid {
		m.CPUCores = int(cpuCores.Int64)
	}
	if memoryTotal.Valid {
		m.MemoryTotalMB = memoryTotal.Int64
	}
	if diskTotal.Valid {
		m.DiskTotalGB = diskTotal.Int64
	}
	if lastBoot.Valid {
		m.LastBootTime = lastBoot.Time
	}

	return &m, nil
}

// ListMachines retrieves all machines for a user
func (s *SQLiteStore) ListMachines(ctx context.Context, userID int) ([]Machine, error) {
	query := `
		SELECT id, user_id, name, hostname, api_key, status, last_seen,
		       platform, platform_version, kernel_version, cpu_cores,
		       memory_total_mb, disk_total_gb, last_boot_time, created_at
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
		var platform sql.NullString
		var platformVersion sql.NullString
		var kernelVersion sql.NullString
		var cpuCores sql.NullInt64
		var memoryTotal sql.NullInt64
		var diskTotal sql.NullInt64
		var lastBoot sql.NullTime
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.Name, &m.Hostname, &m.APIKey, &m.Status, &lastSeen,
			&platform, &platformVersion, &kernelVersion, &cpuCores, &memoryTotal, &diskTotal, &lastBoot, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan machine: %w", err)
		}
		if lastSeen.Valid {
			m.LastSeen = lastSeen.Time
		}
		if platform.Valid {
			m.Platform = platform.String
		}
		if platformVersion.Valid {
			m.PlatformVersion = platformVersion.String
		}
		if kernelVersion.Valid {
			m.KernelVersion = kernelVersion.String
		}
		if cpuCores.Valid {
			m.CPUCores = int(cpuCores.Int64)
		}
		if memoryTotal.Valid {
			m.MemoryTotalMB = memoryTotal.Int64
		}
		if diskTotal.Valid {
			m.DiskTotalGB = diskTotal.Int64
		}
		if lastBoot.Valid {
			m.LastBootTime = lastBoot.Time
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

// UpdateMachineSystemInfo updates optional system metadata for a machine.
func (s *SQLiteStore) UpdateMachineSystemInfo(ctx context.Context, id int, info MachineSystemInfoUpdate) error {
	setClauses := []string{}
	args := []interface{}{}

	if info.Hostname != nil {
		setClauses = append(setClauses, "hostname = ?")
		args = append(args, *info.Hostname)
	}
	if info.Platform != nil {
		setClauses = append(setClauses, "platform = ?")
		args = append(args, *info.Platform)
	}
	if info.PlatformVersion != nil {
		setClauses = append(setClauses, "platform_version = ?")
		args = append(args, *info.PlatformVersion)
	}
	if info.KernelVersion != nil {
		setClauses = append(setClauses, "kernel_version = ?")
		args = append(args, *info.KernelVersion)
	}
	if info.CPUCores != nil {
		setClauses = append(setClauses, "cpu_cores = ?")
		args = append(args, *info.CPUCores)
	}
	if info.MemoryTotalMB != nil {
		setClauses = append(setClauses, "memory_total_mb = ?")
		args = append(args, *info.MemoryTotalMB)
	}
	if info.DiskTotalGB != nil {
		setClauses = append(setClauses, "disk_total_gb = ?")
		args = append(args, *info.DiskTotalGB)
	}
	if info.LastBootTime != nil {
		setClauses = append(setClauses, "last_boot_time = ?")
		args = append(args, info.LastBootTime.UTC())
	}

	if len(setClauses) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE machines SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, id)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update machine system info: %w", err)
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
func (s *SQLiteStore) InsertMetrics(ctx context.Context, machineID int, cpuPct, memUsedPct, diskUsedPct float64, netRxBytes, netTxBytes int64, uptimeSeconds *float64, timestamp time.Time) error {
	query := `
		INSERT INTO metrics_history (machine_id, cpu_pct, mem_used_pct, disk_used_pct, net_rx_bytes, net_tx_bytes, uptime_seconds, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	var uptime interface{}
	if uptimeSeconds != nil {
		uptime = *uptimeSeconds
	} else {
		uptime = nil
	}

	_, err := s.db.ExecContext(ctx, query, machineID, cpuPct, memUsedPct, diskUsedPct, netRxBytes, netTxBytes, uptime, timestamp)
	if err != nil {
		return fmt.Errorf("failed to insert metrics: %w", err)
	}

	return nil
}

// GetLatestMetrics retrieves the most recent metrics for a machine
func (s *SQLiteStore) GetLatestMetrics(ctx context.Context, machineID int) (*MetricsHistory, error) {
	query := `
		SELECT id, machine_id, cpu_pct, mem_used_pct, disk_used_pct, net_rx_bytes, net_tx_bytes, uptime_seconds, timestamp
		FROM metrics_history
		WHERE machine_id = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var m MetricsHistory
	var uptime sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, machineID).Scan(
		&m.ID, &m.MachineID, &m.CPUPct, &m.MemUsedPct, &m.DiskUsedPct, &m.NetRxBytes, &m.NetTxBytes, &uptime, &m.Timestamp,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no metrics found for machine")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest metrics: %w", err)
	}

	if uptime.Valid {
		m.UptimeSeconds = uptime.Float64
	}

	return &m, nil
}

// GetMetricsHistory retrieves metrics history for a machine within a time range
func (s *SQLiteStore) GetMetricsHistory(ctx context.Context, machineID int, from, to time.Time, limit int) ([]MetricsHistory, error) {
	query := `
		SELECT id, machine_id, cpu_pct, mem_used_pct, disk_used_pct, net_rx_bytes, net_tx_bytes, uptime_seconds, timestamp
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
		var uptime sql.NullFloat64
		if err := rows.Scan(&m.ID, &m.MachineID, &m.CPUPct, &m.MemUsedPct, &m.DiskUsedPct, &m.NetRxBytes, &m.NetTxBytes, &uptime, &m.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan metrics: %w", err)
		}
		if uptime.Valid {
			m.UptimeSeconds = uptime.Float64
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}
