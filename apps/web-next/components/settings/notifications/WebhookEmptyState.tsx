"use client";

interface WebhookEmptyStateProps {
  onAddWebhook: () => void;
}

export function WebhookEmptyState({ onAddWebhook }: WebhookEmptyStateProps) {
  return (
    <div className="bg-card/50 backdrop-blur-xl rounded-lg p-12 border border-border/30 text-center">
      <div className="max-w-md mx-auto">
        <div className="text-5xl mb-4">🔔</div>
        <h3 className="text-lg font-semibold text-foreground mb-2">
          No Webhooks Configured
        </h3>
        <p className="text-sm text-muted-foreground mb-6">
          Webhooks let you receive real-time alert notifications at your own
          endpoints. Configure your first webhook to get started.
        </p>
        <button
          onClick={onAddWebhook}
          className="bg-primary hover:bg-primary/90 text-primary-foreground font-medium py-3 px-6 rounded-lg transition-all focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background"
        >
          Add Your First Webhook
        </button>
        <div className="mt-6 pt-6 border-t border-border/30">
          <p className="text-xs text-muted-foreground">
            Webhooks are signed with HMAC-SHA256 for security. You'll need to
            verify the{" "}
            <code className="px-1 py-0.5 bg-muted/50 rounded">
              X-LunaSentri-Signature
            </code>{" "}
            header.
          </p>
        </div>
      </div>
    </div>
  );
}
