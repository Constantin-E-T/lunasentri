package router

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/auth"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/machines"
)

// RegisterMachineRequest represents the machine registration request
type RegisterMachineRequest struct {
	Name        string `json:"name"`
	Hostname    string `json:"hostname,omitempty"`
	Description string `json:"description,omitempty"`
}

// RegisterMachineResponse represents the machine registration response
type RegisterMachineResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Hostname    string    `json:"hostname"`
	Description string    `json:"description"`
	APIKey      string    `json:"api_key"` // Only returned once at registration
	CreatedAt   time.Time `json:"created_at"`
}

// AgentMetricsRequest represents the metrics payload from an agent
type AgentMetricsRequest struct {
	Timestamp   *time.Time              `json:"timestamp,omitempty"`
	CPUPct      float64                 `json:"cpu_pct"`
	MemUsedPct  float64                 `json:"mem_used_pct"`
	DiskUsedPct float64                 `json:"disk_used_pct"`
	NetRxBytes  int64                   `json:"net_rx_bytes,omitempty"`
	NetTxBytes  int64                   `json:"net_tx_bytes,omitempty"`
	UptimeS     *float64                `json:"uptime_s,omitempty"`
	SystemInfo  *AgentSystemInfoPayload `json:"system_info,omitempty"`
}

// AgentSystemInfoPayload represents optional system metadata supplied with metrics.
type AgentSystemInfoPayload struct {
	Hostname        *string    `json:"hostname,omitempty"`
	Platform        *string    `json:"platform,omitempty"`
	PlatformVersion *string    `json:"platform_version,omitempty"`
	KernelVersion   *string    `json:"kernel_version,omitempty"`
	CPUCores        *int       `json:"cpu_cores,omitempty"`
	MemoryTotalMB   *int64     `json:"memory_total_mb,omitempty"`
	DiskTotalGB     *int64     `json:"disk_total_gb,omitempty"`
	LastBootTime    *time.Time `json:"last_boot_time,omitempty"`
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
		machine, apiKey, err := machineService.RegisterMachine(r.Context(), user.ID, req.Name, req.Hostname, req.Description)
		if err != nil {
			log.Printf("Failed to register machine for user %d: %v", user.ID, err)
			http.Error(w, "Failed to register machine", http.StatusInternalServerError)
			return
		}

		log.Printf("Machine registered: id=%d, name=%s, user_id=%d", machine.ID, machine.Name, user.ID)

		// Return machine details with API key (only time it's visible)
		response := RegisterMachineResponse{
			ID:          machine.ID,
			Name:        machine.Name,
			Hostname:    machine.Hostname,
			Description: machine.Description,
			APIKey:      apiKey,
			CreatedAt:   machine.CreatedAt,
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

		var sysInfo *machines.AgentSystemInfo
		if req.SystemInfo != nil {
			sysInfo = &machines.AgentSystemInfo{
				Hostname:        req.SystemInfo.Hostname,
				Platform:        req.SystemInfo.Platform,
				PlatformVersion: req.SystemInfo.PlatformVersion,
				KernelVersion:   req.SystemInfo.KernelVersion,
				CPUCores:        req.SystemInfo.CPUCores,
				MemoryTotalMB:   req.SystemInfo.MemoryTotalMB,
				DiskTotalGB:     req.SystemInfo.DiskTotalGB,
				LastBootTime:    req.SystemInfo.LastBootTime,
			}
		}

		// Record metrics (service handles status update and timestamp)
		if err := machineService.RecordMetrics(r.Context(), machineID, req.CPUPct, req.MemUsedPct, req.DiskUsedPct, req.NetRxBytes, req.NetTxBytes, req.UptimeS, sysInfo); err != nil {
			log.Printf("Failed to record metrics for machine %d: %v", machineID, err)
			http.Error(w, "Failed to record metrics", http.StatusInternalServerError)
			return
		}

		// Log structured info for monitoring
		userID, _ := GetUserIDFromContext(r.Context())
		log.Printf("Metrics recorded: machine_id=%d, user_id=%d, remote_ip=%s, cpu=%.1f%%, mem=%.1f%%, disk=%.1f%%, uptime=%.0fs",
			machineID, userID, getRemoteIP(r), req.CPUPct, req.MemUsedPct, req.DiskUsedPct, func() float64 {
				if req.UptimeS != nil {
					return *req.UptimeS
				}
				return 0
			}())

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

// handleDeleteMachine handles DELETE /machines/:id (requires session auth)
func handleDeleteMachine(machineService *machines.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get authenticated user from context
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract machine ID from URL path
		// Expecting: DELETE /machines/{id}
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) != 2 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		machineID, err := strconv.Atoi(pathParts[1])
		if err != nil {
			http.Error(w, "Invalid machine ID", http.StatusBadRequest)
			return
		}

		// Delete the machine (service will verify ownership)
		err = machineService.DeleteMachine(r.Context(), machineID, user.ID)
		if err != nil {
			log.Printf("Failed to delete machine %d for user %d: %v", machineID, user.ID, err)
			http.Error(w, "Failed to delete machine", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Machine deleted successfully",
		})
	}
}

// handleUpdateMachine handles PATCH /machines/:id (requires session auth)
func handleUpdateMachine(machineService *machines.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get authenticated user from context
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract machine ID from URL path
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) != 2 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		machineID, err := strconv.Atoi(pathParts[1])
		if err != nil {
			http.Error(w, "Invalid machine ID", http.StatusBadRequest)
			return
		}

		// Parse update request
		var req struct {
			Name        *string `json:"name"`
			Hostname    *string `json:"hostname"`
			Description *string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Update the machine (service will verify ownership)
		err = machineService.UpdateMachine(r.Context(), machineID, user.ID, req.Name, req.Hostname, req.Description)
		if err != nil {
			log.Printf("Failed to update machine %d for user %d: %v", machineID, user.ID, err)
			http.Error(w, "Failed to update machine", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Machine updated successfully",
		})
	}
}

// handleDisableMachine handles POST /machines/:id/disable (requires session auth)
func handleDisableMachine(machineService *machines.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get authenticated user from context
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}

		// Extract machine ID from URL path
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 2 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid path"})
			return
		}

		machineID, err := strconv.Atoi(pathParts[1])
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid machine ID"})
			return
		}

		// Disable the machine
		if err := machineService.DisableMachine(r.Context(), machineID, user.ID); err != nil {
			log.Printf("Failed to disable machine %d for user %d: %v", machineID, user.ID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to disable machine"})
			return
		}

		log.Printf("Machine disabled: id=%d, user_id=%d", machineID, user.ID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Machine disabled successfully",
		})
	}
}

// handleEnableMachine handles POST /machines/:id/enable (requires session auth)
func handleEnableMachine(machineService *machines.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get authenticated user from context
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}

		// Extract machine ID from URL path
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 2 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid path"})
			return
		}

		machineID, err := strconv.Atoi(pathParts[1])
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid machine ID"})
			return
		}

		// Enable the machine
		if err := machineService.EnableMachine(r.Context(), machineID, user.ID); err != nil {
			log.Printf("Failed to enable machine %d for user %d: %v", machineID, user.ID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to enable machine"})
			return
		}

		log.Printf("Machine enabled: id=%d, user_id=%d", machineID, user.ID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Machine enabled successfully",
		})
	}
}

// handleRotateMachineAPIKey handles POST /machines/:id/rotate-key (requires session auth)
func handleRotateMachineAPIKey(machineService *machines.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get authenticated user from context
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}

		// Extract machine ID from URL path
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 2 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid path"})
			return
		}

		machineID, err := strconv.Atoi(pathParts[1])
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid machine ID"})
			return
		}

		// Rotate the API key
		newAPIKey, err := machineService.RotateMachineAPIKey(r.Context(), machineID, user.ID)
		if err != nil {
			log.Printf("Failed to rotate API key for machine %d, user %d: %v", machineID, user.ID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to rotate API key"})
			return
		}

		log.Printf("API key rotated for machine: id=%d, user_id=%d", machineID, user.ID)

		// Return the new API key (only time it's visible)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "API key rotated successfully",
			"api_key": newAPIKey,
		})
	}
}
