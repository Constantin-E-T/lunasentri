"use client";

import { useState, useEffect } from "react";
import type { WebhookWithState } from "@/lib/alerts/useWebhooks";

// Simple time ago helper
function timeAgo(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000);

  if (seconds < 60) return "just now";
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  if (seconds < 604800) return `${Math.floor(seconds / 86400)}d ago`;
  return `${Math.floor(seconds / 604800)}w ago`;
}

// Format cooldown time
function formatCooldownTime(date: Date): string {
  return date.toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  });
}

interface WebhookListProps {
  webhooks: WebhookWithState[];
  onEdit: (webhook: WebhookWithState) => void;
  onDelete: (webhook: WebhookWithState) => void;
  onTest: (webhook: WebhookWithState) => void;
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
        <WebhookCard
          key={webhook.id}
          webhook={webhook}
          onEdit={onEdit}
          onDelete={onDelete}
          onTest={onTest}
        />
      ))}
    </div>
  );
}

interface WebhookCardProps {
  webhook: WebhookWithState;
  onEdit: (webhook: WebhookWithState) => void;
  onDelete: (webhook: WebhookWithState) => void;
  onTest: (webhook: WebhookWithState) => void;
}

function WebhookCard({ webhook, onEdit, onDelete, onTest }: WebhookCardProps) {
  const [retryAfter, setRetryAfter] = useState(webhook.retryAfterSeconds);

  // Update retry countdown every second
  useEffect(() => {
    if (!retryAfter || retryAfter <= 0) return;

    const timer = setInterval(() => {
      setRetryAfter((prev) => (prev && prev > 0 ? prev - 1 : null));
    }, 1000);

    return () => clearInterval(timer);
  }, [retryAfter]);

  const testButtonDisabled = !webhook.canSendTest;
  const testButtonTooltip = webhook.isCoolingDown
    ? `Cooling down until ${formatCooldownTime(webhook.cooldownUntil!)}`
    : retryAfter
    ? `Rate limited, retry in ${retryAfter}s`
    : "Send test payload";

  return (
    <div className="bg-card/50 backdrop-blur-xl rounded-lg p-4 border border-border/30 hover:border-border/50 transition-colors">
      {/* Cooldown banner */}
      {webhook.isCoolingDown && (
        <div className="mb-3 p-2 bg-destructive/10 border border-destructive/20 rounded-md text-xs text-destructive flex items-center gap-2">
          <span>🚫</span>
          <span>
            Circuit breaker active: cooling down until{" "}
            {formatCooldownTime(webhook.cooldownUntil!)}
          </span>
        </div>
      )}

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
              isCoolingDown={webhook.isCoolingDown}
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

            {webhook.last_attempt_at && (
              <div className="flex items-center gap-1">
                <span className="text-muted-foreground/70">⏱</span>
                <span>
                  Last attempt {timeAgo(new Date(webhook.last_attempt_at))}
                </span>
              </div>
            )}

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
                      : "text-amber-500/70"
                  }
                >
                  ⚠
                </span>
                <span
                  className={
                    webhook.failure_count >= 3
                      ? "text-destructive"
                      : "text-amber-500"
                  }
                >
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
            disabled={testButtonDisabled}
            className="px-3 py-1.5 text-xs font-medium bg-primary/10 hover:bg-primary/20 text-primary rounded-md transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background disabled:opacity-50 disabled:cursor-not-allowed"
            title={testButtonTooltip}
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
  );
}

function StatusPill({
  isActive,
  failureCount,
  isCoolingDown,
}: {
  isActive: boolean;
  failureCount: number;
  isCoolingDown: boolean;
}) {
  if (!isActive) {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-muted/50 text-muted-foreground">
        Inactive
      </span>
    );
  }

  if (isCoolingDown) {
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-destructive/20 text-destructive border border-destructive/30">
        Cooling Down
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
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-amber-500/10 text-amber-500">
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
