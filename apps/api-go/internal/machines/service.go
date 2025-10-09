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
func (s *Service) RegisterMachine(ctx context.Context, userID int, name, hostname string) (*storage.Machine, string, error) {
	// Generate API key
	apiKey, err := GenerateAPIKey()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash the API key for storage
	apiKeyHash := HashAPIKey(apiKey)

	// Create machine in database
	machine, err := s.store.CreateMachine(ctx, userID, name, hostname, apiKeyHash)
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

// RecordMetrics records metrics for a machine and updates its status
func (s *Service) RecordMetrics(ctx context.Context, machineID int, cpuPct, memUsedPct, diskUsedPct float64, netRxBytes, netTxBytes int64) error {
	// Insert metrics
	now := time.Now()
	if err := s.store.InsertMetrics(ctx, machineID, cpuPct, memUsedPct, diskUsedPct, netRxBytes, netTxBytes, now); err != nil {
		return fmt.Errorf("failed to insert metrics: %w", err)
	}

	// Update machine status to online
	if err := s.store.UpdateMachineStatus(ctx, machineID, "online", now); err != nil {
		return fmt.Errorf("failed to update machine status: %w", err)
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
