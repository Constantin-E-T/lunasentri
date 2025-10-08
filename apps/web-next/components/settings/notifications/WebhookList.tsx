"use client";

import type { Webhook } from "@/lib/alerts/useWebhooks";

// Simple time ago helper
function timeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000);

  if (seconds < 60) return "just now";
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  if (seconds < 604800) return `${Math.floor(seconds / 86400)}d ago`;
  return `${Math.floor(seconds / 604800)}w ago`;
}

interface WebhookListProps {
  webhooks: Webhook[];
  onEdit: (webhook: Webhook) => void;
  onDelete: (webhook: Webhook) => void;
  onTest: (webhook: Webhook) => void;
}

export function WebhookList({
  webhooks,
  onEdit,
  onDelete,
  onTest,
}: WebhookListProps) {
  return (
    <div className="space-y-3">
      {webhooks.map((webhook) => (
        <div
          key={webhook.id}
          className="bg-card/50 backdrop-blur-xl rounded-lg p-4 border border-border/30 hover:border-border/50 transition-colors"
        >
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 min-w-0">
              {/* URL and Status */}
              <div className="flex items-center gap-3 mb-2">
                <h3 className="text-sm font-medium text-foreground truncate">
                  {webhook.url}
                </h3>
                <StatusPill
                  isActive={webhook.is_active}
                  failureCount={webhook.failure_count}
                />
              </div>

              {/* Metadata */}
              <div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
                <div className="flex items-center gap-1">
                  <span className="text-muted-foreground/70">Secret:</span>
                  <code className="px-1.5 py-0.5 bg-muted/50 rounded font-mono">
                    ••••{webhook.secret_last_four}
                  </code>
                </div>

                {webhook.last_success_at && (
                  <div className="flex items-center gap-1">
                    <span className="text-chart-4/70">✓</span>
                    <span>
                      Last success {timeAgo(new Date(webhook.last_success_at))}
                    </span>
                  </div>
                )}

                {webhook.last_error_at && (
                  <div className="flex items-center gap-1">
                    <span className="text-destructive/70">✗</span>
                    <span>
                      Last error {timeAgo(new Date(webhook.last_error_at))}
                    </span>
                  </div>
                )}

                {webhook.failure_count > 0 && (
                  <div className="flex items-center gap-1">
                    <span
                      className={
                        webhook.failure_count >= 3
                          ? "text-destructive/70"
                          : "text-chart-2/70"
                      }
                    >
                      ⚠
                    </span>
                    <span>
                      {webhook.failure_count} failure
                      {webhook.failure_count !== 1 ? "s" : ""}
                    </span>
                  </div>
                )}
              </div>
            </div>

            {/* Actions */}
            <div className="flex items-center gap-2">
              <button
                onClick={() => onTest(webhook)}
                className="px-3 py-1.5 text-xs font-medium bg-primary/10 hover:bg-primary/20 text-primary rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background"
                title="Send test payload"
              >
                Test
              </button>
              <button
                onClick={() => onEdit(webhook)}
                className="px-3 py-1.5 text-xs font-medium bg-card hover:bg-muted text-foreground rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background"
              >
                Edit
              </button>
              <button
                onClick={() => onDelete(webhook)}
                className="px-3 py-1.5 text-xs font-medium bg-destructive/10 hover:bg-destructive/20 text-destructive rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-destructive focus:ring-offset-2 focus:ring-offset-background"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

function StatusPill({
  isActive,
  failureCount,
}: {
  isActive: boolean;
  failureCount: number;
}) {
  if (!isActive) {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-muted/50 text-muted-foreground">
        Inactive
      </span>
    );
  }

  if (failureCount >= 3) {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-destructive/10 text-destructive">
        Active • High Failures
      </span>
    );
  }

  if (failureCount > 0) {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-chart-2/10 text-chart-2">
        Active • {failureCount} Failure{failureCount !== 1 ? "s" : ""}
      </span>
    );
  }

  return (
    <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-chart-4/10 text-chart-4">
      Active
    </span>
  );
}
