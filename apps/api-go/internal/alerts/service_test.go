package alerts

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/metrics"
	"github.com/Constantin-E-T/lunasentri/apps/api-go/internal/storage"
)

func setupTestAlertService(t *testing.T) (*Service, *storage.SQLiteStore) {
	// Create a temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "lunasentri_alert_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Clean up on test completion
	t.Cleanup(func() {
		store.Close()
		os.RemoveAll(tmpDir)
	})

	service := NewService(store)
	return service, store
}

func TestAlertService_Evaluate_AboveThreshold(t *testing.T) {
	service, store := setupTestAlertService(t)
	ctx := context.Background()

	// Create a rule: CPU above 80% for 3 consecutive samples
	_, err := store.CreateAlertRule(ctx, "High CPU", "cpu_pct", "above", 80.0, 3)
	if err != nil {
		t.Fatalf("Failed to create alert rule: %v", err)
	}

	// Test samples that breach the threshold
	samples := []metrics.Metrics{
		{CPUPct: 85.0, MemUsedPct: 50.0, DiskUsedPct: 30.0}, // 1st breach
		{CPUPct: 87.0, MemUsedPct: 50.0, DiskUsedPct: 30.0}, // 2nd breach
		{CPUPct: 82.0, MemUsedPct: 50.0, DiskUsedPct: 30.0}, // 3rd breach - should fire alert
	}

	for i, sample := range samples {
		err := service.Evaluate(ctx, sample)
		if err != nil {
			t.Fatalf("Failed to evaluate sample %d: %v", i+1, err)
		}
	}

	// Check that an alert event was created
	events, err := store.ListAlertEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list alert events: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 alert event, got %d", len(events))
	}

	event := events[0]
	if event.RuleID == 0 {
		t.Error("Expected event rule_id to be set")
	}
	if event.Value != 82.0 {
		t.Errorf("Expected event value 82.0, got %f", event.Value)
	}

	// Check rule state
	states := service.GetRuleStates()
	if len(states) == 0 {
		t.Error("Expected rule state to exist")
	} else {
		// Check first rule state
		for _, state := range states {
			if state.ConsecutiveBreaches != 3 {
				t.Errorf("Expected 3 consecutive breaches, got %d", state.ConsecutiveBreaches)
			}
			break
		}
	}
}

func TestAlertService_Evaluate_BelowThreshold(t *testing.T) {
	service, store := setupTestAlertService(t)
	ctx := context.Background()

	// Create a rule: Memory below 20% for 2 consecutive samples
	_, err := store.CreateAlertRule(ctx, "Low Memory", "mem_used_pct", "below", 20.0, 2)
	if err != nil {
		t.Fatalf("Failed to create alert rule: %v", err)
	}

	// Test samples that breach the threshold
	samples := []metrics.Metrics{
		{CPUPct: 50.0, MemUsedPct: 15.0, DiskUsedPct: 30.0}, // 1st breach
		{CPUPct: 50.0, MemUsedPct: 18.0, DiskUsedPct: 30.0}, // 2nd breach - should fire alert
	}

	for i, sample := range samples {
		err := service.Evaluate(ctx, sample)
		if err != nil {
			t.Fatalf("Failed to evaluate sample %d: %v", i+1, err)
		}
	}

	// Check that an alert event was created
	events, err := store.ListAlertEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list alert events: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 alert event, got %d", len(events))
	}

	event := events[0]
	if event.Value != 18.0 {
		t.Errorf("Expected event value 18.0, got %f", event.Value)
	}
}

func TestAlertService_Evaluate_Recovery(t *testing.T) {
	service, store := setupTestAlertService(t)
	ctx := context.Background()

	// Create a rule: CPU above 70% for 2 consecutive samples
	_, err := store.CreateAlertRule(ctx, "CPU Alert", "cpu_pct", "above", 70.0, 2)
	if err != nil {
		t.Fatalf("Failed to create alert rule: %v", err)
	}

	// Test samples: breach, breach, recover, breach
	samples := []metrics.Metrics{
		{CPUPct: 75.0, MemUsedPct: 50.0, DiskUsedPct: 30.0}, // 1st breach
		{CPUPct: 80.0, MemUsedPct: 50.0, DiskUsedPct: 30.0}, // 2nd breach - should fire alert
		{CPUPct: 60.0, MemUsedPct: 50.0, DiskUsedPct: 30.0}, // Recovery - resets counter
		{CPUPct: 85.0, MemUsedPct: 50.0, DiskUsedPct: 30.0}, // 1st breach again
	}

	for i, sample := range samples {
		err := service.Evaluate(ctx, sample)
		if err != nil {
			t.Fatalf("Failed to evaluate sample %d: %v", i+1, err)
		}
	}

	// Should have only 1 alert event (after the 2nd sample)
	events, err := store.ListAlertEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list alert events: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 alert event, got %d", len(events))
	}

	if events[0].Value != 80.0 {
		t.Errorf("Expected event value 80.0, got %f", events[0].Value)
	}
}

func TestAlertService_Evaluate_MultipleRules(t *testing.T) {
	service, store := setupTestAlertService(t)
	ctx := context.Background()

	// Create multiple rules
	cpuRule, err := store.CreateAlertRule(ctx, "High CPU", "cpu_pct", "above", 80.0, 2)
	if err != nil {
		t.Fatalf("Failed to create CPU rule: %v", err)
	}

	memRule, err := store.CreateAlertRule(ctx, "High Memory", "mem_used_pct", "above", 90.0, 1)
	if err != nil {
		t.Fatalf("Failed to create memory rule: %v", err)
	}

	// Sample that triggers both rules
	sample := metrics.Metrics{
		CPUPct:      85.0, // Above 80% 
		MemUsedPct:  95.0, // Above 90%
		DiskUsedPct: 30.0,
	}

	// First evaluation - CPU needs 2 samples, Memory needs 1
	err = service.Evaluate(ctx, sample)
	if err != nil {
		t.Fatalf("Failed to evaluate sample 1: %v", err)
	}

	// Check events after first evaluation
	events, err := store.ListAlertEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list alert events after sample 1: %v", err)
	}

	// Should have 1 event (memory rule fired)
	if len(events) != 1 {
		t.Fatalf("Expected 1 alert event after sample 1, got %d", len(events))
	}
	if events[0].RuleID != memRule.ID {
		t.Errorf("Expected memory rule event, got rule ID %d", events[0].RuleID)
	}

	// Second evaluation - CPU should fire now
	err = service.Evaluate(ctx, sample)
	if err != nil {
		t.Fatalf("Failed to evaluate sample 2: %v", err)
	}

	// Check events after second evaluation
	events, err = store.ListAlertEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list alert events after sample 2: %v", err)
	}

	// Should have 2 events now
	if len(events) != 2 {
		t.Fatalf("Expected 2 alert events after sample 2, got %d", len(events))
	}

	// Find the CPU rule event
	var cpuEventFound bool
	for _, event := range events {
		if event.RuleID == cpuRule.ID {
			cpuEventFound = true
			break
		}
	}
	if !cpuEventFound {
		t.Error("Expected CPU rule event to be created")
	}
}

func TestAlertService_UpsertRule(t *testing.T) {
	service, _ := setupTestAlertService(t)
	ctx := context.Background()

	// Test creating a new rule
	rule, err := service.UpsertRule(ctx, 0, "Test Rule", "cpu_pct", "above", 75.0, 2)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	if rule.ID == 0 {
		t.Error("Expected rule ID to be set")
	}
	if rule.Name != "Test Rule" {
		t.Errorf("Expected name 'Test Rule', got '%s'", rule.Name)
	}

	// Test updating the rule
	updatedRule, err := service.UpsertRule(ctx, rule.ID, "Updated Rule", "mem_used_pct", "below", 25.0, 3)
	if err != nil {
		t.Fatalf("Failed to update rule: %v", err)
	}

	if updatedRule.ID != rule.ID {
		t.Errorf("Expected same rule ID %d, got %d", rule.ID, updatedRule.ID)
	}
	if updatedRule.Name != "Updated Rule" {
		t.Errorf("Expected name 'Updated Rule', got '%s'", updatedRule.Name)
	}
	if updatedRule.Metric != "mem_used_pct" {
		t.Errorf("Expected metric 'mem_used_pct', got '%s'", updatedRule.Metric)
	}
}

func TestAlertService_DeleteRule(t *testing.T) {
	service, _ := setupTestAlertService(t)
	ctx := context.Background()

	// Create a rule
	rule, err := service.UpsertRule(ctx, 0, "Test Rule", "cpu_pct", "above", 80.0, 1)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Delete the rule
	err = service.DeleteRule(ctx, rule.ID)
	if err != nil {
		t.Fatalf("Failed to delete rule: %v", err)
	}

	// Verify rule is deleted
	rules, err := service.ListRules(ctx)
	if err != nil {
		t.Fatalf("Failed to list rules: %v", err)
	}
	if len(rules) != 0 {
		t.Errorf("Expected 0 rules after deletion, got %d", len(rules))
	}

	// Verify events are also deleted (cascade)
	events, err := service.ListActiveEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list events: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events after rule deletion, got %d", len(events))
	}

	// Verify rule state is cleaned up
	states := service.GetRuleStates()
	if _, exists := states[rule.ID]; exists {
		t.Error("Expected rule state to be cleaned up")
	}
}

func TestAlertService_AcknowledgeEvent(t *testing.T) {
	service, _ := setupTestAlertService(t)
	ctx := context.Background()

	// Create a rule and trigger an event
	_, err := service.UpsertRule(ctx, 0, "Test Rule", "cpu_pct", "above", 80.0, 1)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	sample := metrics.Metrics{CPUPct: 85.0}
	err = service.Evaluate(ctx, sample)
	if err != nil {
		t.Fatalf("Failed to evaluate sample: %v", err)
	}

	// Get the event
	events, err := service.ListActiveEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Acknowledged {
		t.Error("Expected event to be unacknowledged initially")
	}

	// Acknowledge the event
	err = service.AcknowledgeEvent(ctx, event.ID)
	if err != nil {
		t.Fatalf("Failed to acknowledge event: %v", err)
	}

	// Verify event is acknowledged
	events, err = service.ListActiveEvents(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to list events after ack: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event after ack, got %d", len(events))
	}

	if !events[0].Acknowledged {
		t.Error("Expected event to be acknowledged")
	}
}

func TestAlertService_GetMetricValue(t *testing.T) {
	service, _ := setupTestAlertService(t)

	sample := metrics.Metrics{
		CPUPct:      75.5,
		MemUsedPct:  80.2,
		DiskUsedPct: 45.8,
	}

	tests := []struct {
		metric   string
		expected float64
	}{
		{"cpu_pct", 75.5},
		{"mem_used_pct", 80.2},
		{"disk_used_pct", 45.8},
		{"invalid_metric", 0.0},
	}

	for _, test := range tests {
		value := service.getMetricValue(sample, test.metric)
		if value != test.expected {
			t.Errorf("getMetricValue(%s) = %f, expected %f", test.metric, value, test.expected)
		}
	}
}

func TestAlertService_IsThresholdBreached(t *testing.T) {
	service, _ := setupTestAlertService(t)

	tests := []struct {
		value      float64
		threshold  float64
		comparison string
		expected   bool
	}{
		{85.0, 80.0, "above", true},
		{75.0, 80.0, "above", false},
		{80.0, 80.0, "above", false}, // Equal is not above
		{15.0, 20.0, "below", true},
		{25.0, 20.0, "below", false},
		{20.0, 20.0, "below", false}, // Equal is not below
		{50.0, 40.0, "invalid", false},
	}

	for _, test := range tests {
		result := service.isThresholdBreached(test.value, test.threshold, test.comparison)
		if result != test.expected {
			t.Errorf("isThresholdBreached(%f, %f, %s) = %t, expected %t",
				test.value, test.threshold, test.comparison, result, test.expected)
		}
	}
}