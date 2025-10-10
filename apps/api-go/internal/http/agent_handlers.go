package router

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/machines"
)

// RegisterMachineRequest represents the machine registration request
type RegisterMachineRequest struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname,omitempty"`
}

// RegisterMachineResponse represents the machine registration response
type RegisterMachineResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Hostname  string    `json:"hostname"`
	APIKey    string    `json:"api_key"` // Only returned once at registration
	CreatedAt time.Time `json:"created_at"`
}

// AgentMetricsRequest represents the metrics payload from an agent
type AgentMetricsRequest struct {
	Timestamp   *time.Time `json:"timestamp,omitempty"`
	CPUPct      float64    `json:"cpu_pct"`
	MemUsedPct  float64    `json:"mem_used_pct"`
	DiskUsedPct float64    `json:"disk_used_pct"`
	NetRxBytes  int64      `json:"net_rx_bytes,omitempty"`
	NetTxBytes  int64      `json:"net_tx_bytes,omitempty"`
}

// handleAgentRegister handles POST /agent/register (requires session auth)
func handleAgentRegister(machineService *machines.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get authenticated user from context (set by RequireAuth middleware)
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req RegisterMachineRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate name
		if req.Name == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "machine name is required"})
			return
		}

		// Register machine
		machine, apiKey, err := machineService.RegisterMachine(r.Context(), user.ID, req.Name, req.Hostname)
		if err != nil {
			log.Printf("Failed to register machine for user %d: %v", user.ID, err)
			http.Error(w, "Failed to register machine", http.StatusInternalServerError)
			return
		}

		log.Printf("Machine registered: id=%d, name=%s, user_id=%d", machine.ID, machine.Name, user.ID)

		// Return machine details with API key (only time it's visible)
		response := RegisterMachineResponse{
			ID:        machine.ID,
			Name:      machine.Name,
			Hostname:  machine.Hostname,
			APIKey:    apiKey,
			CreatedAt: machine.CreatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// handleAgentMetrics handles POST /agent/metrics (requires API key auth)
func handleAgentMetrics(machineService *machines.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get machine from context (set by RequireAPIKey middleware)
		machineID, ok := GetMachineIDFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req AgentMetricsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate metrics
		if req.CPUPct < 0 || req.CPUPct > 100 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "cpu_pct must be between 0 and 100"})
			return
		}
		if req.MemUsedPct < 0 || req.MemUsedPct > 100 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "mem_used_pct must be between 0 and 100"})
			return
		}
		if req.DiskUsedPct < 0 || req.DiskUsedPct > 100 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "disk_used_pct must be between 0 and 100"})
			return
		}

		// Record metrics (service handles status update and timestamp)
		if err := machineService.RecordMetrics(r.Context(), machineID, req.CPUPct, req.MemUsedPct, req.DiskUsedPct, req.NetRxBytes, req.NetTxBytes); err != nil {
			log.Printf("Failed to record metrics for machine %d: %v", machineID, err)
			http.Error(w, "Failed to record metrics", http.StatusInternalServerError)
			return
		}

		// Log structured info for monitoring
		userID, _ := GetUserIDFromContext(r.Context())
		log.Printf("Metrics recorded: machine_id=%d, user_id=%d, remote_ip=%s, cpu=%.1f%%, mem=%.1f%%, disk=%.1f%%",
			machineID, userID, getRemoteIP(r), req.CPUPct, req.MemUsedPct, req.DiskUsedPct)

		// Return 202 Accepted
		w.WriteHeader(http.StatusAccepted)
	}
}

// getRemoteIP extracts the remote IP from the request
func getRemoteIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for reverse proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// handleListMachines handles GET /machines (requires session auth)
func handleListMachines(machineService *machines.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get authenticated user from context (set by RequireAuth middleware)
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// List machines with computed statuses
		machinesList, err := machineService.ListMachinesWithComputedStatus(r.Context(), user.ID)
		if err != nil {
			log.Printf("Failed to list machines for user %d: %v", user.ID, err)
			http.Error(w, "Failed to list machines", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(machinesList)
	}
}
