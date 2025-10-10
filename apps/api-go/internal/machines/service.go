package machines

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// Service provides machine management operations
type Service struct {
	store storage.Store
}

// AgentSystemInfo represents optional system metadata supplied by an agent.
type AgentSystemInfo struct {
	Hostname        *string
	Platform        *string
	PlatformVersion *string
	KernelVersion   *string
	CPUCores        *int
	MemoryTotalMB   *int64
	DiskTotalGB     *int64
	LastBootTime    *time.Time
}

// NewService creates a new machine service
func NewService(store storage.Store) *Service {
	return &Service{store: store}
}

// GenerateAPIKey generates a new random API key
func GenerateAPIKey() (string, error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as base64
	return base64.URLEncoding.EncodeToString(b), nil
}

// HashAPIKey hashes an API key for storage
func HashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return fmt.Sprintf("%x", hash)
}

// RegisterMachine registers a new machine for a user
func (s *Service) RegisterMachine(ctx context.Context, userID int, name, hostname, description string) (*storage.Machine, string, error) {
	// Generate API key
	apiKey, err := GenerateAPIKey()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash the API key for storage
	apiKeyHash := HashAPIKey(apiKey)

	// Create machine in database
	machine, err := s.store.CreateMachine(ctx, userID, name, hostname, description, apiKeyHash)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create machine: %w", err)
	}

	// Return machine and the plaintext API key (only time it's visible)
	return machine, apiKey, nil
}

// GetMachine retrieves a machine by ID, ensuring the user owns it
func (s *Service) GetMachine(ctx context.Context, machineID, userID int) (*storage.Machine, error) {
	machine, err := s.store.GetMachineByID(ctx, machineID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if machine.UserID != userID {
		return nil, fmt.Errorf("access denied: machine does not belong to user")
	}

	return machine, nil
}

// ListMachines lists all machines for a user
func (s *Service) ListMachines(ctx context.Context, userID int) ([]storage.Machine, error) {
	return s.store.ListMachines(ctx, userID)
}

// DeleteMachine deletes a machine, ensuring the user owns it
func (s *Service) DeleteMachine(ctx context.Context, machineID, userID int) error {
	return s.store.DeleteMachine(ctx, machineID, userID)
}

// AuthenticateMachine validates an API key and returns the associated machine
func (s *Service) AuthenticateMachine(ctx context.Context, apiKey string) (*storage.Machine, error) {
	// Hash the provided API key
	apiKeyHash := HashAPIKey(apiKey)

	// Look up machine by API key hash
	machine, err := s.store.GetMachineByAPIKey(ctx, apiKeyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	return machine, nil
}

// RecordMetrics records metrics for a machine and updates its status.
// Optionally accepts uptimeSeconds and system information for enrichment.
func (s *Service) RecordMetrics(ctx context.Context, machineID int, cpuPct, memUsedPct, diskUsedPct float64, netRxBytes, netTxBytes int64, uptimeSeconds *float64, sysInfo *AgentSystemInfo) error {
	// Insert metrics
	now := time.Now()
	if err := s.store.InsertMetrics(ctx, machineID, cpuPct, memUsedPct, diskUsedPct, netRxBytes, netTxBytes, uptimeSeconds, now); err != nil {
		return fmt.Errorf("failed to insert metrics: %w", err)
	}

	// Update machine status to online
	if err := s.store.UpdateMachineStatus(ctx, machineID, "online", now); err != nil {
		return fmt.Errorf("failed to update machine status: %w", err)
	}

	// Optionally enrich system info
	if sysInfo != nil {
		update := storage.MachineSystemInfoUpdate{
			Hostname:        sysInfo.Hostname,
			Platform:        sysInfo.Platform,
			PlatformVersion: sysInfo.PlatformVersion,
			KernelVersion:   sysInfo.KernelVersion,
			CPUCores:        sysInfo.CPUCores,
			MemoryTotalMB:   sysInfo.MemoryTotalMB,
			DiskTotalGB:     sysInfo.DiskTotalGB,
			LastBootTime:    sysInfo.LastBootTime,
		}
		if err := s.store.UpdateMachineSystemInfo(ctx, machineID, update); err != nil {
			return fmt.Errorf("failed to update machine system info: %w", err)
		}
	}

	return nil
}

// GetLatestMetrics retrieves the latest metrics for a machine
func (s *Service) GetLatestMetrics(ctx context.Context, machineID, userID int) (*storage.MetricsHistory, error) {
	// Verify ownership
	machine, err := s.GetMachine(ctx, machineID, userID)
	if err != nil {
		return nil, err
	}

	// Get latest metrics
	return s.store.GetLatestMetrics(ctx, machine.ID)
}

// GetMetricsHistory retrieves historical metrics for a machine
func (s *Service) GetMetricsHistory(ctx context.Context, machineID, userID int, from, to time.Time, limit int) ([]storage.MetricsHistory, error) {
	// Verify ownership
	machine, err := s.GetMachine(ctx, machineID, userID)
	if err != nil {
		return nil, err
	}

	// Get metrics history
	return s.store.GetMetricsHistory(ctx, machine.ID, from, to, limit)
}

// OfflineThreshold is the duration after which a machine is considered offline
// if it hasn't reported metrics (default: 2 minutes = 4 missed 30-second intervals)
const OfflineThreshold = 2 * time.Minute

// IsOnline determines if a machine is online based on its last_seen timestamp
func IsOnline(lastSeen time.Time) bool {
	if lastSeen.IsZero() {
		return false
	}
	return time.Since(lastSeen) <= OfflineThreshold
}

// ComputeStatus determines the status string for a machine based on last_seen
func ComputeStatus(lastSeen time.Time) string {
	if IsOnline(lastSeen) {
		return "online"
	}
	return "offline"
}

// UpdateMachineStatuses checks all machines and updates their status based on last_seen
// This should be called periodically (e.g., every minute) by a background job
func (s *Service) UpdateMachineStatuses(ctx context.Context) error {
	// Note: This would require a new storage method to get all machines
	// For now, we rely on status being updated when metrics are recorded
	// Future enhancement: Add a background job to check all machines
	return nil
}

// GetMachineWithComputedStatus retrieves a machine and computes its real-time status
func (s *Service) GetMachineWithComputedStatus(ctx context.Context, machineID, userID int) (*storage.Machine, error) {
	machine, err := s.GetMachine(ctx, machineID, userID)
	if err != nil {
		return nil, err
	}

	// Compute real-time status
	machine.Status = ComputeStatus(machine.LastSeen)

	return machine, nil
}

// ListMachinesWithComputedStatus lists all machines for a user with computed statuses
func (s *Service) ListMachinesWithComputedStatus(ctx context.Context, userID int) ([]storage.Machine, error) {
	machines, err := s.ListMachines(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Compute real-time status for each machine
	for i := range machines {
		machines[i].Status = ComputeStatus(machines[i].LastSeen)
	}

	return machines, nil
}

// UpdateMachine updates machine details, ensuring the user owns it
func (s *Service) UpdateMachine(ctx context.Context, machineID, userID int, name, hostname, description *string) error {
	// First verify ownership
	machine, err := s.store.GetMachineByID(ctx, machineID)
	if err != nil {
		return err
	}

	if machine.UserID != userID {
		return fmt.Errorf("access denied: machine does not belong to user")
	}

	// Build update query dynamically based on what's provided
	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if hostname != nil {
		updates["hostname"] = *hostname
	}
	if description != nil {
		updates["description"] = *description
	}

	// If nothing to update, return early
	if len(updates) == 0 {
		return nil
	}

	// Perform update in storage layer
	return s.store.UpdateMachineDetails(ctx, machineID, updates)
}
