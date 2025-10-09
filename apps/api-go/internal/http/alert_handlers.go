package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/alerts"
)

// AlertRuleRequest represents an alert rule request/response
type AlertRuleRequest struct {
	Name         string  `json:"name"`
	Metric       string  `json:"metric"`
	ThresholdPct float64 `json:"threshold_pct"`
	Comparison   string  `json:"comparison"`
	TriggerAfter int     `json:"trigger_after"`
}

// AlertEventAckRequest represents an alert event acknowledgment request
type AlertEventAckRequest struct {
	// No body needed, ID comes from URL
}

// validateAlertRule validates an alert rule request
func validateAlertRule(req *AlertRuleRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Metric != "cpu_pct" && req.Metric != "mem_used_pct" && req.Metric != "disk_used_pct" {
		return fmt.Errorf("metric must be one of: cpu_pct, mem_used_pct, disk_used_pct")
	}
	if req.ThresholdPct < 0 || req.ThresholdPct > 100 {
		return fmt.Errorf("threshold_pct must be between 0 and 100")
	}
	if req.Comparison != "above" && req.Comparison != "below" {
		return fmt.Errorf("comparison must be 'above' or 'below'")
	}
	if req.TriggerAfter < 1 {
		return fmt.Errorf("trigger_after must be >= 1")
	}
	return nil
}

// handleAlertRules handles GET /alerts/rules and POST /alerts/rules
func handleAlertRules(alertService *alerts.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case "GET":
			rules, err := alertService.ListRules(r.Context())
			if err != nil {
				log.Printf("Failed to list alert rules: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if err := json.NewEncoder(w).Encode(rules); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

		case "POST":
			var req AlertRuleRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if err := validateAlertRule(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			rule, err := alertService.UpsertRule(r.Context(), 0, req.Name, req.Metric, req.Comparison, req.ThresholdPct, req.TriggerAfter)
			if err != nil {
				log.Printf("Failed to create alert rule: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(rule); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleAlertRule handles PUT /alerts/rules/{id} and DELETE /alerts/rules/{id}
func handleAlertRule(alertService *alerts.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Extract ID from path
		path := strings.TrimPrefix(r.URL.Path, "/alerts/rules/")
		if path == "" {
			http.Error(w, "Alert rule ID required", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "Invalid alert rule ID", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "PUT":
			var req AlertRuleRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if err := validateAlertRule(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			rule, err := alertService.UpsertRule(r.Context(), id, req.Name, req.Metric, req.Comparison, req.ThresholdPct, req.TriggerAfter)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					http.Error(w, "Alert rule not found", http.StatusNotFound)
				} else {
					log.Printf("Failed to update alert rule: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}

			if err := json.NewEncoder(w).Encode(rule); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

		case "DELETE":
			if err := alertService.DeleteRule(r.Context(), id); err != nil {
				if strings.Contains(err.Error(), "not found") {
					http.Error(w, "Alert rule not found", http.StatusNotFound)
				} else {
					log.Printf("Failed to delete alert rule: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}

			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleAlertEvents handles GET /alerts/events
func handleAlertEvents(alertService *alerts.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// Default limit to 50, can be overridden by query param
		limit := 50
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		events, err := alertService.ListActiveEvents(r.Context(), limit)
		if err != nil {
			log.Printf("Failed to list alert events: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(events); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// handleAlertEventAck handles POST /alerts/events/{id}/ack
func handleAlertEventAck(alertService *alerts.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract ID from path - expecting /alerts/events/{id}/ack
		path := strings.TrimPrefix(r.URL.Path, "/alerts/events/")
		if !strings.HasSuffix(path, "/ack") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		idStr := strings.TrimSuffix(path, "/ack")
		if idStr == "" {
			http.Error(w, "Alert event ID required", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid alert event ID", http.StatusBadRequest)
			return
		}

		if err := alertService.AcknowledgeEvent(r.Context(), id); err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already acknowledged") {
				http.Error(w, "Alert event not found or already acknowledged", http.StatusNotFound)
			} else {
				log.Printf("Failed to acknowledge alert event: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
