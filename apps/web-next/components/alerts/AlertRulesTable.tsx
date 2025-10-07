"use client";

import { Badge } from "@/components/ui/badge";
import type { AlertRule } from "@/lib/api";

interface AlertRulesTableProps {
  rules: AlertRule[];
  onEdit: (rule: AlertRule) => void;
  onDelete: (id: number) => void;
  loading: boolean;
}

export function AlertRulesTable({
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