"use client";

import type { AlertEvent, AlertRule } from "@/lib/api";
import {
  getEventSeverity,
  getSeverityStyles,
  getSeverityLabel,
  type MetricType,
} from "@/lib/alerts/severity";

interface AlertEventsListProps {
  events: AlertEvent[];
  rules: AlertRule[];
  onAcknowledge: (id: number) => void;
  loading: boolean;
}

export function AlertEventsList({
  events,
  rules,
  onAcknowledge,
  loading,
}: AlertEventsListProps) {
  const getRuleName = (ruleId: number) => {
    const rule = (rules ?? []).find((r) => r.id === ruleId);
    return rule?.name || `Rule #${ruleId}`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const unacknowledgedEvents = (events ?? []).filter(
    (e: AlertEvent) => !e.acknowledged
  );

  if (loading) {
    return (
      <div className="text-center text-muted-foreground py-8">
        Loading events...
      </div>
    );
  }

  if (unacknowledgedEvents.length === 0) {
    return (
      <div className="text-center text-muted-foreground py-8">
        <div className="text-green-400 text-4xl mb-2">✓</div>
        <div>No active alerts. Your system is healthy!</div>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {unacknowledgedEvents.map((event) => {
        const rule = rules.find((r) => r.id === event.rule_id);
        const severity = rule
          ? getEventSeverity(
              rule.metric as MetricType,
              event.value,
              rule.threshold_pct,
              rule.comparison
            )
          : "warn";
        const severityStyles = getSeverityStyles(severity);
        const severityLabel = getSeverityLabel(severity);

        return (
          <div
            key={event.id}
            className={`p-4 border rounded-md ${severityStyles.background} ${severityStyles.border}`}
          >
            <div className="flex justify-between items-start gap-4">
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <div className={`font-medium ${severityStyles.text}`}>
                    {getRuleName(event.rule_id)}
                  </div>
                  <span
                    className={`text-xs px-2 py-1 rounded-full border ${severityStyles.badge}`}
                  >
                    {severityLabel}
                  </span>
                </div>
                <div className="text-sm text-muted-foreground">
                  Triggered at {formatDate(event.triggered_at)} • Value:{" "}
                  {event.value.toFixed(1)}%
                </div>
              </div>
              <button
                onClick={() => onAcknowledge(event.id)}
                className="text-xs px-3 py-1 bg-card/50 border border-border/50 rounded hover:bg-card/70 transition-colors"
              >
                Acknowledge
              </button>
            </div>
          </div>
        );
      })}
    </div>
  );
}