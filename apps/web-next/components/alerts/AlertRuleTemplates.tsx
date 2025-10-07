"use client";

import type { CreateAlertRuleRequest } from "@/lib/api";

interface AlertRuleTemplate {
  label: string;
  description: string;
  data: CreateAlertRuleRequest;
}

interface AlertRuleTemplatesProps {
  onSelectTemplate: (template: CreateAlertRuleRequest) => void;
}

const templates: AlertRuleTemplate[] = [
  {
    label: "CPU Hot (>85%)",
    description: "Alert when CPU usage is critically high",
    data: {
      name: "CPU Critical Alert",
      metric: "cpu_pct" as const,
      comparison: "above" as const,
      threshold_pct: 85,
      trigger_after: 3,
    },
  },
  {
    label: "Memory Tight (>90%)",
    description: "Alert when memory usage is critically high",
    data: {
      name: "Memory Critical Alert",
      metric: "mem_used_pct" as const,
      comparison: "above" as const,
      threshold_pct: 90,
      trigger_after: 3,
    },
  },
  {
    label: "Disk Warning (>80%)",
    description: "Alert when disk usage is getting high",
    data: {
      name: "Disk Usage Warning",
      metric: "disk_used_pct" as const,
      comparison: "above" as const,
      threshold_pct: 80,
      trigger_after: 5,
    },
  },
  {
    label: "CPU Warning (>60%)",
    description: "Early warning for elevated CPU usage",
    data: {
      name: "CPU Usage Warning",
      metric: "cpu_pct" as const,
      comparison: "above" as const,
      threshold_pct: 60,
      trigger_after: 5,
    },
  },
  {
    label: "Memory Warning (>70%)",
    description: "Early warning for elevated memory usage",
    data: {
      name: "Memory Usage Warning",
      metric: "mem_used_pct" as const,
      comparison: "above" as const,
      threshold_pct: 70,
      trigger_after: 5,
    },
  },
];

export function AlertRuleTemplates({
  onSelectTemplate,
}: AlertRuleTemplatesProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {templates.map((template, index) => (
        <button
          key={index}
          onClick={() => onSelectTemplate(template.data)}
          className="text-left p-4 rounded-lg border border-border/50 bg-card/50 hover:bg-card/70 transition-colors"
        >
          <h4 className="font-medium text-primary">{template.label}</h4>
          <p className="text-sm text-muted-foreground mt-1">
            {template.description}
          </p>
        </button>
      ))}
    </div>
  );
}
