package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *SQLiteStore {
	// Create a temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "lunasentri_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Clean up on test completion
	t.Cleanup(func() {
		store.Close()
		os.RemoveAll(tmpDir)
	})

	return store
}

func TestAlertRules_CRUD(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	// Test creating alert rule
	rule, err := store.CreateAlertRule(ctx, "High CPU", "cpu_pct", "above", 80.0, 3)
	if err != nil {
		t.Fatalf("Failed to create alert rule: %v", err)
	}

	if rule.ID == 0 {
		t.Error("Expected rule ID to be set")
	}
	if rule.Name != "High CPU" {
		t.Errorf("Expected name 'High CPU', got '%s'", rule.Name)
	}
	if rule.Metric != "cpu_pct" {
		t.Errorf("Expected metric 'cpu_pct', got '%s'", rule.Metric)
	}
	if rule.ThresholdPct != 80.0 {
		t.Errorf("Expected threshold 80.0, got %f", rule.ThresholdPct)
	}
	if rule.Comparison != "above" {
		t.Errorf("Expected comparison 'above', got '%s'", rule.Comparison)
	}
	if rule.TriggerAfter != 3 {
		t.Errorf("Expected trigger_after 3, got %d", rule.TriggerAfter)
	}

	// Test listing alert rules
	rules, err := store.ListAlertRules(ctx)
	if err != nil {
		t.Fatalf("Failed to list alert rules: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}

	// Test updating alert rule
	updatedRule, err := store.UpdateAlertRule(ctx, rule.ID, "Very High CPU", "cpu_pct", "above", 90.0, 5)
	if err != nil {
		t.Fatalf("Failed to update alert rule: %v", err)
	}
	if updatedRule.Name != "Very High CPU" {
		t.Errorf("Expected updated name 'Very High CPU', got '%s'", updatedRule.Name)
	}
	if updatedRule.ThresholdPct != 90.0 {
		t.Errorf("Expected updated threshold 90.0, got %f", updatedRule.ThresholdPct)
	}
	if updatedRule.TriggerAfter != 5 {
		t.Errorf("Expected updated trigger_after 5, got %d", updatedRule.TriggerAfter)
	}
	if !updatedRule.UpdatedAt.After(updatedRule.CreatedAt) {
		t.Error("Expected updated_at to be after created_at")
	}

	// Test deleting alert rule
	err = store.DeleteAlertRule(ctx, rule.ID)
	if err != nil {
		t.Fatalf("Failed to delete alert rule: %v", err)
	}

	// Verify rule is deleted
	rules, err = store.ListAlertRules(ctx)
	if err != nil {
		t.Fatalf("Failed to list alert rules after deletion: %v", err)
	}
	if len(rules) != 0 {
		t.Errorf("Expected 0 rules after deletion, got %d", len(rules))
	}

	// Test deleting non-existent rule
	err = store.DeleteAlertRule(ctx, 999)
	if err == nil {
		t.Error("Expected error when deleting non-existent rule")
	}
}

func TestAlertRules_ValidationConstraints(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	// Test invalid metric
	_, err := store.CreateAlertRule(ctx, "Test", "invalid_metric", "above", 50.0, 1)
	if err == nil {
		t.Error("Expected error for invalid metric")
	}

	// Test invalid comparison
	_, err = store.CreateAlertRule(ctx, "Test", "cpu_pct", "invalid_comparison", 50.0, 1)
	if err == nil {
		t.Error("Expected error for invalid comparison")
	}

	// Test invalid threshold (negative)
	_, err = store.CreateAlertRule(ctx, "Test", "cpu_pct", "above", -1.0, 1)
	if err == nil {
		t.Error("Expected error for negative threshold")
	}

	// Test invalid threshold (> 100)
	_, err = store.CreateAlertRule(ctx, "Test", "cpu_pct", "above", 101.0, 1)
	if err == nil {
		t.Error("Expected error for threshold > 100")
	}

	// Test invalid trigger_after (< 1)
	_, err = store.CreateAlertRule(ctx, "Test", "cpu_pct", "above", 50.0, 0)
	if err == nil {
		t.Error("Expected error for trigger_after < 1")
	}
}

func TestAlertEvents_CRUD(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	// Create a rule first
	rule, err := store.CreateAlertRule(ctx, "High Memory", "mem_used_pct", "above", 85.0, 2)
	if err != nil {
		t.Fatalf("Failed to create alert rule: %v", err)
	}

	// Test creating alert event
	event, err := store.CreateAlertEvent(ctx, rule.ID, 87.5)
	if err != nil {
		t.Fatalf("Failed to create alert event: %v", err)
	}

	if event.ID == 0 {
		t.Error("Expected event ID to be set")
	}
	if event.RuleID != rule.ID {
		t.Errorf("Expected rule_id %d, got %d", rule.ID, event.RuleID)
	}
	if event.Value != 87.5 {
		t.Errorf("Expected value 87.5, got %f", event.Value)
	}
	if event.Acknowledged {
		t.Error("Expected event to be unacknowledged initially")
	}
	if event.AcknowledgedAt != nil {
		t.Error("Expected acknowledged_at to be nil initially")
	}

	// Test listing alert events
	events, err := store.ListAlertEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list alert events: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	// Test acknowledging alert event
	err = store.AckAlertEvent(ctx, event.ID)
	if err != nil {
		t.Fatalf("Failed to acknowledge alert event: %v", err)
	}

	// Verify event is acknowledged
	events, err = store.ListAlertEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list alert events after ack: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if !events[0].Acknowledged {
		t.Error("Expected event to be acknowledged")
	}
	if events[0].AcknowledgedAt == nil {
		t.Error("Expected acknowledged_at to be set")
	}

	// Test acknowledging already acknowledged event
	err = store.AckAlertEvent(ctx, event.ID)
	if err == nil {
		t.Error("Expected error when acknowledging already acknowledged event")
	}

	// Test acknowledging non-existent event
	err = store.AckAlertEvent(ctx, 999)
	if err == nil {
		t.Error("Expected error when acknowledging non-existent event")
	}
}

func TestAlertEvents_CascadeDelete(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	// Create a rule
	rule, err := store.CreateAlertRule(ctx, "Test Rule", "cpu_pct", "above", 50.0, 1)
	if err != nil {
		t.Fatalf("Failed to create alert rule: %v", err)
	}

	// Create multiple events for the rule
	_, err = store.CreateAlertEvent(ctx, rule.ID, 60.0)
	if err != nil {
		t.Fatalf("Failed to create alert event 1: %v", err)
	}
	_, err = store.CreateAlertEvent(ctx, rule.ID, 70.0)
	if err != nil {
		t.Fatalf("Failed to create alert event 2: %v", err)
	}

	// Verify events exist
	events, err := store.ListAlertEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list alert events: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	// Delete the rule
	err = store.DeleteAlertRule(ctx, rule.ID)
	if err != nil {
		t.Fatalf("Failed to delete alert rule: %v", err)
	}

	// Verify events are also deleted (cascade)
	events, err = store.ListAlertEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list alert events after rule deletion: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events after rule deletion, got %d", len(events))
	}
}

func TestAlertEvents_OrderingAndLimit(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	// Create rules
	rule1, err := store.CreateAlertRule(ctx, "Rule 1", "cpu_pct", "above", 50.0, 1)
	if err != nil {
		t.Fatalf("Failed to create rule 1: %v", err)
	}
	rule2, err := store.CreateAlertRule(ctx, "Rule 2", "mem_used_pct", "above", 80.0, 1)
	if err != nil {
		t.Fatalf("Failed to create rule 2: %v", err)
	}

	// Create events with different timestamps
	event1, err := store.CreateAlertEvent(ctx, rule1.ID, 60.0)
	if err != nil {
		t.Fatalf("Failed to create event 1: %v", err)
	}

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	event2, err := store.CreateAlertEvent(ctx, rule2.ID, 85.0)
	if err != nil {
		t.Fatalf("Failed to create event 2: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	event3, err := store.CreateAlertEvent(ctx, rule1.ID, 70.0)
	if err != nil {
		t.Fatalf("Failed to create event 3: %v", err)
	}

	// Acknowledge the middle event
	err = store.AckAlertEvent(ctx, event2.ID)
	if err != nil {
		t.Fatalf("Failed to acknowledge event 2: %v", err)
	}

	// List events - should prioritize unacknowledged first, then by time desc
	events, err := store.ListAlertEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list alert events: %v", err)
	}

	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	// First two should be unacknowledged (event3, event1), then acknowledged (event2)
	if events[0].ID != event3.ID {
		t.Errorf("Expected first event to be event3 (ID %d), got ID %d", event3.ID, events[0].ID)
	}
	if events[1].ID != event1.ID {
		t.Errorf("Expected second event to be event1 (ID %d), got ID %d", event1.ID, events[1].ID)
	}
	if events[2].ID != event2.ID {
		t.Errorf("Expected third event to be event2 (ID %d), got ID %d", event2.ID, events[2].ID)
	}

	// Test limit
	limitedEvents, err := store.ListAlertEvents(ctx, 2)
	if err != nil {
		t.Fatalf("Failed to list limited alert events: %v", err)
	}
	if len(limitedEvents) != 2 {
		t.Errorf("Expected 2 events with limit, got %d", len(limitedEvents))
	}
}