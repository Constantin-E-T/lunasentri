"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { AlertRule, CreateAlertRuleRequest } from "@/lib/api";

interface AlertRuleFormProps {
  rule?: AlertRule;
  onSubmit: (rule: CreateAlertRuleRequest) => Promise<void>;
  onCancel: () => void;
  isEditing?: boolean;
}

export function AlertRuleForm({
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
