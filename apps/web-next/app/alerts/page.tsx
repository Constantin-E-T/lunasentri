"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useSession } from "@/lib/useSession";
import { useAlertsWithNotifications } from "@/lib/useAlertsWithNotifications";
import { CreateAlertRuleRequest, AlertRule, AlertEvent } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

// Alert Rule Form Component
interface AlertRuleFormProps {
  rule?: AlertRule;
  onSubmit: (rule: CreateAlertRuleRequest) => Promise<void>;
  onCancel: () => void;
  isEditing?: boolean;
}

function AlertRuleForm({
  rule,
  onSubmit,
  onCancel,
  isEditing = false,
}: AlertRuleFormProps) {
  const [formData, setFormData] = useState<CreateAlertRuleRequest>({
    name: rule?.name || "",
    metric: rule?.metric || "cpu_pct",
    threshold_pct: rule?.threshold_pct || 80,
    comparison: rule?.comparison || "above",
    trigger_after: rule?.trigger_after || 3,
  });
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      await onSubmit(formData);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4 z-50">
      <Card className="max-w-md w-full bg-card/90 border-border/50 backdrop-blur-xl">
        <CardHeader>
          <CardTitle>
            {isEditing ? "Edit Alert Rule" : "Create Alert Rule"}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Name</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                className="w-full px-3 py-2 bg-card/50 border border-border/50 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Metric</label>
              <select
                value={formData.metric}
                onChange={(e) =>
                  setFormData({ ...formData, metric: e.target.value as any })
                }
                className="w-full px-3 py-2 bg-card/50 border border-border/50 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
              >
                <option value="cpu_pct">CPU %</option>
                <option value="mem_used_pct">Memory %</option>
                <option value="disk_used_pct">Disk %</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">
                Comparison
              </label>
              <select
                value={formData.comparison}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    comparison: e.target.value as any,
                  })
                }
                className="w-full px-3 py-2 bg-card/50 border border-border/50 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
              >
                <option value="above">Above</option>
                <option value="below">Below</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">
                Threshold (%)
              </label>
              <input
                type="number"
                min="0"
                max="100"
                value={formData.threshold_pct}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    threshold_pct: Number(e.target.value),
                  })
                }
                className="w-full px-3 py-2 bg-card/50 border border-border/50 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">
                Trigger After (consecutive samples)
              </label>
              <input
                type="number"
                min="1"
                value={formData.trigger_after}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    trigger_after: Number(e.target.value),
                  })
                }
                className="w-full px-3 py-2 bg-card/50 border border-border/50 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                required
              />
            </div>

            <div className="flex gap-2 pt-4">
              <button
                type="submit"
                disabled={isSubmitting}
                className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 disabled:opacity-50"
              >
                {isSubmitting ? "Saving..." : isEditing ? "Update" : "Create"}
              </button>
              <button
                type="button"
                onClick={onCancel}
                className="flex-1 px-4 py-2 bg-card/50 border border-border/50 rounded-md hover:bg-card/70"
              >
                Cancel
              </button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

// Alert Rules Table Component
interface AlertRulesTableProps {
  rules: AlertRule[];
  onEdit: (rule: AlertRule) => void;
  onDelete: (id: number) => void;
  loading: boolean;
}

function AlertRulesTable({
  rules,
  onEdit,
  onDelete,
  loading,
}: AlertRulesTableProps) {
  const formatMetric = (metric: string) => {
    switch (metric) {
      case "cpu_pct":
        return "CPU %";
      case "mem_used_pct":
        return "Memory %";
      case "disk_used_pct":
        return "Disk %";
      default:
        return metric;
    }
  };

  if (loading) {
    return (
      <div className="text-center text-muted-foreground py-8">
        Loading rules...
      </div>
    );
  }

  if (rules.length === 0) {
    return (
      <div className="text-center text-muted-foreground py-8">
        No alert rules configured. Create your first rule to get started.
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-border/30">
            <th className="text-left py-3 px-4 font-medium">Name</th>
            <th className="text-left py-3 px-4 font-medium hidden sm:table-cell">
              Metric
            </th>
            <th className="text-left py-3 px-4 font-medium">Threshold</th>
            <th className="text-left py-3 px-4 font-medium hidden md:table-cell">
              Trigger After
            </th>
            <th className="text-left py-3 px-4 font-medium">Actions</th>
          </tr>
        </thead>
        <tbody>
          {rules.map((rule) => (
            <tr key={rule.id} className="border-b border-border/20">
              <td className="py-3 px-4">
                <div>
                  <div className="font-medium">{rule.name}</div>
                  <div className="text-sm text-muted-foreground sm:hidden">
                    {formatMetric(rule.metric)} {rule.comparison}{" "}
                    {rule.threshold_pct}%
                  </div>
                </div>
              </td>
              <td className="py-3 px-4 hidden sm:table-cell">
                {formatMetric(rule.metric)}
              </td>
              <td className="py-3 px-4">
                <Badge variant="outline">
                  {rule.comparison} {rule.threshold_pct}%
                </Badge>
              </td>
              <td className="py-3 px-4 hidden md:table-cell">
                {rule.trigger_after}
              </td>
              <td className="py-3 px-4">
                <div className="flex gap-2">
                  <button
                    onClick={() => onEdit(rule)}
                    className="text-xs px-2 py-1 bg-card/50 border border-border/50 rounded hover:bg-card/70 transition-colors"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => onDelete(rule.id)}
                    className="text-xs px-2 py-1 bg-destructive/20 border border-destructive/30 text-destructive rounded hover:bg-destructive/30 transition-colors"
                  >
                    Delete
                  </button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// Alert Events List Component
interface AlertEventsListProps {
  events: AlertEvent[];
  rules: AlertRule[];
  onAcknowledge: (id: number) => void;
  loading: boolean;
}

function AlertEventsList({
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
      {unacknowledgedEvents.map((event) => (
        <div
          key={event.id}
          className="p-4 bg-destructive/10 border border-destructive/30 rounded-md"
        >
          <div className="flex justify-between items-start gap-4">
            <div className="flex-1">
              <div className="font-medium text-destructive">
                {getRuleName(event.rule_id)}
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
      ))}
    </div>
  );
}

export default function AlertsPage() {
  const router = useRouter();
  const { status, user } = useSession();
  const {
    rules,
    events,
    rulesLoading,
    eventsLoading,
    rulesError,
    eventsError,
    createRule,
    updateRule,
    deleteRule,
    acknowledgeEvent,
    newAlertsCount,
    markAllAsSeen,
  } = useAlertsWithNotifications(50);

  const [showForm, setShowForm] = useState(false);
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null);
  const [message, setMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  // Redirect to login if unauthenticated
  useEffect(() => {
    if (status === "unauthenticated") {
      router.push("/login");
    }
  }, [status, router]);

  // Show loading state while checking authentication
  if (status === "loading") {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-muted-foreground animate-pulse">Loading...</div>
      </div>
    );
  }

  // Don't render if not authenticated (will redirect)
  if (status !== "authenticated") {
    return null;
  }

  const handleCreateRule = async (ruleData: CreateAlertRuleRequest) => {
    try {
      await createRule(ruleData);
      setShowForm(false);
      setMessage({ type: "success", text: "Alert rule created successfully!" });
      setTimeout(() => setMessage(null), 3000);
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Failed to create rule",
      });
      setTimeout(() => setMessage(null), 5000);
    }
  };

  const handleUpdateRule = async (ruleData: CreateAlertRuleRequest) => {
    if (!editingRule) return;

    try {
      await updateRule(editingRule.id, ruleData);
      setEditingRule(null);
      setMessage({ type: "success", text: "Alert rule updated successfully!" });
      setTimeout(() => setMessage(null), 3000);
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Failed to update rule",
      });
      setTimeout(() => setMessage(null), 5000);
    }
  };

  const handleDeleteRule = async (id: number) => {
    if (
      !confirm(
        "Are you sure you want to delete this alert rule? This will also delete all related events."
      )
    ) {
      return;
    }

    try {
      await deleteRule(id);
      setMessage({ type: "success", text: "Alert rule deleted successfully!" });
      setTimeout(() => setMessage(null), 3000);
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Failed to delete rule",
      });
      setTimeout(() => setMessage(null), 5000);
    }
  };

  const handleAcknowledgeEvent = async (id: number) => {
    try {
      await acknowledgeEvent(id);
      setMessage({ type: "success", text: "Alert acknowledged!" });
      setTimeout(() => setMessage(null), 3000);
    } catch (error) {
      setMessage({
        type: "error",
        text:
          error instanceof Error
            ? error.message
            : "Failed to acknowledge alert",
      });
      setTimeout(() => setMessage(null), 5000);
    }
  };

  const unacknowledgedCount = (events ?? []).filter(
    (e: AlertEvent) => !e.acknowledged
  ).length;

  return (
    <div className="min-h-screen">
      {/* Navigation */}
      <div className="border-b border-border/40 bg-card/40 backdrop-blur-xl">
        <div className="max-w-6xl mx-auto px-6 py-4 flex flex-wrap gap-4 justify-between items-center">
          <div className="flex items-center gap-3 text-primary">
            <Link href="/" className="flex items-center gap-3">
              <span className="text-2xl">🌙</span>
              <span className="font-semibold tracking-wide">LunaSentri</span>
            </Link>
          </div>
          <div className="flex items-center gap-3 text-sm">
            <Link
              href="/"
              className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-muted-foreground transition-all duration-200 hover:text-foreground hover:border-border"
            >
              Dashboard
            </Link>
            {user?.is_admin && (
              <Link
                href="/users"
                className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-muted-foreground transition-all duration-200 hover:text-foreground hover:border-border"
              >
                Manage Users
              </Link>
            )}
            <Link
              href="/settings"
              className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-muted-foreground transition-all duration-200 hover:text-foreground hover:border-border"
            >
              Settings
            </Link>
            <span className="text-muted-foreground hidden sm:inline">
              {user?.email}
            </span>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-6xl mx-auto px-6 py-8">
        <div className="space-y-6">
          {/* Header */}
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-semibold text-primary">
                Alert Management
              </h1>
              <p className="text-muted-foreground">
                Monitor system metrics and manage alert rules
              </p>
            </div>
            {unacknowledgedCount > 0 && (
              <Badge variant="destructive" className="px-3 py-1">
                {unacknowledgedCount} Active Alert
                {unacknowledgedCount > 1 ? "s" : ""}
              </Badge>
            )}
          </div>

          {/* Message */}
          {message && (
            <div
              className={`p-4 rounded-md border ${
                message.type === "success"
                  ? "bg-green-100/10 border-green-500/30 text-green-400"
                  : "bg-red-100/10 border-red-500/30 text-red-400"
              }`}
            >
              {message.text}
            </div>
          )}

          {/* Active Events */}
          <Card className="bg-card/70 border-border/30 backdrop-blur-xl">
            <CardHeader>
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-2">
                  <CardTitle>Active Alerts</CardTitle>
                  {unacknowledgedCount > 0 && (
                    <Badge variant="destructive">{unacknowledgedCount}</Badge>
                  )}
                  {newAlertsCount > 0 && (
                    <Badge variant="default" className="bg-blue-500">
                      {newAlertsCount} new
                    </Badge>
                  )}
                </div>
                {newAlertsCount > 0 && (
                  <button
                    onClick={markAllAsSeen}
                    className="px-3 py-1 text-sm bg-card/40 border border-border/30 rounded-md text-muted-foreground transition-all duration-200 hover:text-foreground hover:border-border"
                  >
                    Mark all as seen
                  </button>
                )}
              </div>
            </CardHeader>
            <CardContent>
              {eventsError ? (
                <div className="text-red-400 text-center py-8">
                  {eventsError}
                </div>
              ) : (
                <AlertEventsList
                  events={events ?? []}
                  rules={rules ?? []}
                  onAcknowledge={handleAcknowledgeEvent}
                  loading={eventsLoading}
                />
              )}
            </CardContent>
          </Card>

          {/* Alert Rules */}
          <Card className="bg-card/70 border-border/30 backdrop-blur-xl">
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>Alert Rules</CardTitle>
                <button
                  onClick={() => setShowForm(true)}
                  className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
                >
                  Add Rule
                </button>
              </div>
            </CardHeader>
            <CardContent>
              {rulesError ? (
                <div className="text-red-400 text-center py-8">
                  {rulesError}
                </div>
              ) : (
                <AlertRulesTable
                  rules={rules ?? []}
                  onEdit={setEditingRule}
                  onDelete={handleDeleteRule}
                  loading={rulesLoading}
                />
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Form Modal */}
      {showForm && (
        <AlertRuleForm
          onSubmit={handleCreateRule}
          onCancel={() => setShowForm(false)}
        />
      )}

      {editingRule && (
        <AlertRuleForm
          rule={editingRule}
          isEditing={true}
          onSubmit={handleUpdateRule}
          onCancel={() => setEditingRule(null)}
        />
      )}
    </div>
  );
}
