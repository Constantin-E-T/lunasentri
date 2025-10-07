package alerts

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/metrics"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

// AlertNotifier defines the interface for sending alert notifications
type AlertNotifier interface {
	// Notify sends notifications for an alert event
	Notify(ctx context.Context, rule storage.AlertRule, event *storage.AlertEvent) error
}

// Service handles alert rule evaluation and event management
type Service struct {
	store        storage.Store
	notifier     AlertNotifier
	mu           sync.RWMutex
	ruleStates   map[int]*RuleState // rule ID -> current state
	rulesCache   []storage.AlertRule
	lastRefresh  time.Time
	refreshTTL   time.Duration
}

// RuleState tracks the consecutive breach count for a rule
type RuleState struct {
	ConsecutiveBreaches int
	LastValue           float64
	LastEvaluated       time.Time
}

// NewService creates a new alert service
func NewService(store storage.Store, notifier AlertNotifier) *Service {
	return &Service{
		store:       store,
		notifier:    notifier,
		ruleStates:  make(map[int]*RuleState),
		refreshTTL:  30 * time.Second, // Refresh rules every 30 seconds
	}
}

// refreshRulesIfNeeded refreshes the rules cache if it's stale
func (s *Service) refreshRulesIfNeeded(ctx context.Context) error {
	s.mu.RLock()
	needsRefresh := time.Since(s.lastRefresh) > s.refreshTTL
	s.mu.RUnlock()

	if !needsRefresh {
		return nil
	}

	rules, err := s.store.ListAlertRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh alert rules: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.rulesCache = rules
	s.lastRefresh = time.Now()

	// Clean up state for deleted rules
	activeRuleIDs := make(map[int]bool)
	for _, rule := range rules {
		activeRuleIDs[rule.ID] = true
	}

	for ruleID := range s.ruleStates {
		if !activeRuleIDs[ruleID] {
			delete(s.ruleStates, ruleID)
		}
	}

	return nil
}

// Evaluate evaluates all alert rules against the given metrics sample
func (s *Service) Evaluate(ctx context.Context, sample metrics.Metrics) error {
	if err := s.refreshRulesIfNeeded(ctx); err != nil {
		log.Printf("[ALERT] Failed to refresh rules: %v", err)
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for _, rule := range s.rulesCache {
		value := s.getMetricValue(sample, rule.Metric)
		
		// Get or create rule state
		state, exists := s.ruleStates[rule.ID]
		if !exists {
			state = &RuleState{}
			s.ruleStates[rule.ID] = state
		}

		state.LastValue = value
		state.LastEvaluated = now

		// Check if threshold is breached
		breached := s.isThresholdBreached(value, rule.ThresholdPct, rule.Comparison)

		if breached {
			state.ConsecutiveBreaches++
			
			// Fire alert if we've reached the trigger threshold
			if state.ConsecutiveBreaches == rule.TriggerAfter {
				if err := s.fireAlert(ctx, &rule, value); err != nil {
					log.Printf("[ALERT] Failed to fire alert for rule '%s': %v", rule.Name, err)
				}
			}
		} else {
			// Reset consecutive breaches when metric recovers
			if state.ConsecutiveBreaches > 0 {
				log.Printf("[ALERT] Rule '%s' recovered: %s=%.1f (was breached %d times)", 
					rule.Name, rule.Metric, value, state.ConsecutiveBreaches)
			}
			state.ConsecutiveBreaches = 0
		}
	}

	return nil
}

// getMetricValue extracts the specific metric value from the sample
func (s *Service) getMetricValue(sample metrics.Metrics, metricName string) float64 {
	switch metricName {
	case "cpu_pct":
		return sample.CPUPct
	case "mem_used_pct":
		return sample.MemUsedPct
	case "disk_used_pct":
		return sample.DiskUsedPct
	default:
		return 0
	}
}

// isThresholdBreached checks if the value breaches the threshold according to comparison
func (s *Service) isThresholdBreached(value, threshold float64, comparison string) bool {
	switch comparison {
	case "above":
		return value > threshold
	case "below":
		return value < threshold
	default:
		return false
	}
}

// fireAlert creates an alert event and logs it
func (s *Service) fireAlert(ctx context.Context, rule *storage.AlertRule, value float64) error {
	event, err := s.store.CreateAlertEvent(ctx, rule.ID, value)
	if err != nil {
		return fmt.Errorf("failed to create alert event: %w", err)
	}

	log.Printf("[ALERT] %s %s %.1f%% for %d samples (value=%.1f) - Event ID: %d", 
		rule.Name, rule.Comparison, rule.ThresholdPct, rule.TriggerAfter, value, event.ID)

	// Send notifications asynchronously if notifier is available
	if s.notifier != nil {
		go func() {
			// Create a timeout context for notification sending
			notifyCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := s.notifier.Notify(notifyCtx, *rule, event); err != nil {
				log.Printf("[ALERT] Failed to send notifications for event %d: %v", event.ID, err)
			}
		}()
	}

	return nil
}

// ListRules returns all alert rules
func (s *Service) ListRules(ctx context.Context) ([]storage.AlertRule, error) {
	return s.store.ListAlertRules(ctx)
}

// UpsertRule creates or updates an alert rule
func (s *Service) UpsertRule(ctx context.Context, id int, name, metric, comparison string, thresholdPct float64, triggerAfter int) (*storage.AlertRule, error) {
	if id == 0 {
		// Create new rule
		rule, err := s.store.CreateAlertRule(ctx, name, metric, comparison, thresholdPct, triggerAfter)
		if err != nil {
			return nil, err
		}
		
		// Reset rule state cache to pick up new rule
		s.mu.Lock()
		s.lastRefresh = time.Time{}
		s.mu.Unlock()
		
		return rule, nil
	} else {
		// Update existing rule
		rule, err := s.store.UpdateAlertRule(ctx, id, name, metric, comparison, thresholdPct, triggerAfter)
		if err != nil {
			return nil, err
		}
		
		// Reset state for updated rule
		s.mu.Lock()
		delete(s.ruleStates, id)
		s.lastRefresh = time.Time{}
		s.mu.Unlock()
		
		return rule, nil
	}
}

// DeleteRule deletes an alert rule
func (s *Service) DeleteRule(ctx context.Context, id int) error {
	err := s.store.DeleteAlertRule(ctx, id)
	if err != nil {
		return err
	}
	
	// Clean up rule state
	s.mu.Lock()
	delete(s.ruleStates, id)
	s.lastRefresh = time.Time{}
	s.mu.Unlock()
	
	return nil
}

// ListActiveEvents returns recent alert events (unacknowledged first)
func (s *Service) ListActiveEvents(ctx context.Context, limit int) ([]storage.AlertEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.store.ListAlertEvents(ctx, limit)
}

// AcknowledgeEvent acknowledges an alert event
func (s *Service) AcknowledgeEvent(ctx context.Context, eventID int) error {
	return s.store.AckAlertEvent(ctx, eventID)
}

// GetRuleStates returns the current state of all rules (for debugging/monitoring)
func (s *Service) GetRuleStates() map[int]RuleState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	states := make(map[int]RuleState)
	for id, state := range s.ruleStates {
		states[id] = *state
	}
	return states
}