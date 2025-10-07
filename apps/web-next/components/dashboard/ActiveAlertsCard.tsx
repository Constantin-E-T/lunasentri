"use client";

import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  AlertTriangle,
  ChevronRight,
  Clock,
  Cpu,
  HardDrive,
  Server,
} from "lucide-react";
import { useAlertsWithNotifications } from "@/lib/alerts";
import type { AlertEvent, AlertRule } from "@/lib/alerts";

interface ActiveAlertsCardProps {
  events?: AlertEvent[];
  rules?: AlertRule[];
}

function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();

  const diffMinutes = Math.floor(diffMs / (1000 * 60));
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffMinutes < 1) {
    return "Just now";
  } else if (diffMinutes < 60) {
    return `${diffMinutes}m ago`;
  } else if (diffHours < 24) {
    return `${diffHours}h ago`;
  } else {
    return `${diffDays}d ago`;
  }
}

function getMetricIcon(metric: string) {
  switch (metric) {
    case "cpu_pct":
      return <Cpu className="h-4 w-4" />;
    case "mem_used_pct":
      return <Server className="h-4 w-4" />;
    case "disk_used_pct":
      return <HardDrive className="h-4 w-4" />;
    default:
      return <AlertTriangle className="h-4 w-4" />;
  }
}

function getMetricBadgeStyle(metric: string) {
  switch (metric) {
    case "cpu_pct":
      return "bg-destructive/20 text-destructive border-destructive/30";
    case "mem_used_pct":
      return "bg-amber-500/20 text-amber-400 border-amber-500/30";
    case "disk_used_pct":
      return "bg-primary/20 text-primary border-primary/30";
    default:
      return "bg-muted/20 text-muted-foreground border-muted/30";
  }
}

export function ActiveAlertsCard({
  events: propEvents,
  rules: propRules,
}: ActiveAlertsCardProps) {
  // Use hook if no events/rules provided via props
  const alertsHook = useAlertsWithNotifications(10);

  const events = propEvents || alertsHook.events || [];
  const rules = propRules || alertsHook.rules || [];

  // Get unacknowledged events (latest 3)
  const activeEvents = events
    .filter((event) => !event.acknowledged)
    .slice(0, 3);

  return (
    <Card className="w-full bg-card/70 border border-border/30 backdrop-blur-xl shadow-2xl transition-all duration-300 hover:-translate-y-1 hover:shadow-[0_20px_60px_rgba(20,40,120,0.35)]">
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center gap-2 text-foreground">
          <AlertTriangle className="h-5 w-5" />
          Active Alerts
          {activeEvents.length > 0 && (
            <span className="inline-flex items-center justify-center w-5 h-5 text-xs font-medium text-white bg-destructive rounded-full">
              {activeEvents.length}
            </span>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {activeEvents.length === 0 ? (
          <div className="text-center py-8 space-y-3">
            <div className="text-muted-foreground text-sm">
              No active alerts at the moment
            </div>
            <p className="text-xs text-muted-foreground/70">
              Your system is running smoothly
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {activeEvents.map((event) => {
              const rule = rules.find((r) => r.id === event.rule_id);
              const metric = rule?.metric || "unknown";

              return (
                <div
                  key={event.id}
                  className="flex items-start gap-3 p-3 rounded-lg bg-card/40 border border-border/20"
                >
                  <div
                    className={`flex items-center justify-center w-8 h-8 rounded-full border ${getMetricBadgeStyle(
                      metric
                    )}`}
                  >
                    {getMetricIcon(metric)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <h4 className="text-sm font-medium text-foreground truncate">
                        {rule?.name || `Rule ${event.rule_id}`}
                      </h4>
                      <div className="flex items-center gap-1 text-xs text-muted-foreground">
                        <Clock className="h-3 w-3" />
                        {formatRelativeTime(event.triggered_at)}
                      </div>
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">
                      {rule?.metric === "cpu_pct" && "CPU"}
                      {rule?.metric === "mem_used_pct" && "Memory"}
                      {rule?.metric === "disk_used_pct" && "Disk"}
                      {rule?.comparison === "above"
                        ? " exceeded "
                        : " dropped below "}
                      {rule?.threshold_pct}% (current: {event.value.toFixed(1)}
                      %)
                    </p>
                  </div>
                </div>
              );
            })}
          </div>
        )}

        <Link
          href="/alerts"
          className="flex items-center justify-center gap-2 text-sm text-primary hover:text-primary/80 transition-colors duration-200 group"
        >
          <span>View all alerts</span>
          <ChevronRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-1" />
        </Link>
      </CardContent>
    </Card>
  );
}