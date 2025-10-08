"use client";

import { useState, useEffect } from "react";
import type {
  WebhookWithState,
  CreateWebhookRequest,
  UpdateWebhookRequest,
} from "@/lib/alerts/useWebhooks";

interface WebhookFormProps {
  webhook?: WebhookWithState;
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (
    payload: CreateWebhookRequest | UpdateWebhookRequest
  ) => Promise<void>;
}

export function WebhookForm({
  webhook,
  isOpen,
  onClose,
  onSubmit,
}: WebhookFormProps) {
  const isEditing = !!webhook;
  const [url, setUrl] = useState("");
  const [secret, setSecret] = useState("");
  const [isActive, setIsActive] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [validationError, setValidationError] = useState("");

  // Reset form when dialog opens/closes or webhook changes
  useEffect(() => {
    if (isOpen) {
      if (webhook) {
        setUrl(webhook.url);
        setSecret(""); // Don't pre-fill secret for security
        setIsActive(webhook.is_active);
      } else {
        setUrl("");
        setSecret("");
        setIsActive(true);
      }
      setValidationError("");
    }
  }, [isOpen, webhook]);

  function validateForm(): string | null {
    // URL validation
    if (!url.trim()) {
      return "URL is required";
    }

    try {
      const urlObj = new URL(url);
      if (urlObj.protocol !== "https:") {
        return "URL must use HTTPS protocol";
      }
    } catch {
      return "Invalid URL format";
    }

    // Secret validation (only required for create, optional for edit)
    if (!isEditing) {
      if (!secret) {
        return "Secret is required";
      }
      if (secret.length < 16) {
        return "Secret must be at least 16 characters";
      }
      if (secret.length > 128) {
        return "Secret must not exceed 128 characters";
      }
    } else if (secret) {
      // If editing and secret is provided, validate it
      if (secret.length < 16) {
        return "Secret must be at least 16 characters";
      }
      if (secret.length > 128) {
        return "Secret must not exceed 128 characters";
      }
    }

    return null;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setValidationError("");

    const error = validateForm();
    if (error) {
      setValidationError(error);
      return;
    }

    setIsSubmitting(true);

    try {
      if (isEditing) {
        // Build update payload with only changed fields
        const payload: UpdateWebhookRequest = {
          url,
          is_active: isActive,
        };
        // Only include secret if it was changed
        if (secret) {
          payload.secret = secret;
        }
        await onSubmit(payload);
      } else {
        // Create payload requires all fields
        await onSubmit({ url, secret, is_active: isActive });
      }
      onClose();
    } catch (err) {
      setValidationError(
        err instanceof Error ? err.message : "Failed to save webhook"
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="fixed inset-0" onClick={onClose} />

      <div className="relative z-10 w-full max-w-md bg-card/90 border border-border/50 backdrop-blur-xl rounded-lg shadow-xl">
        <div className="p-6">
          <h2 className="text-xl font-semibold text-card-foreground mb-6">
            {isEditing ? "Edit Webhook" : "Add Webhook"}
          </h2>

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* URL Field */}
            <div>
              <label
                htmlFor="webhook-url"
                className="block text-sm font-medium text-card-foreground mb-2"
              >
                Webhook URL
              </label>
              <input
                id="webhook-url"
                type="url"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                disabled={isSubmitting}
                className="w-full px-4 py-3 bg-background/50 border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                placeholder="https://example.com/webhook"
                required
              />
              <p className="text-xs text-muted-foreground mt-1">
                Must use HTTPS protocol
              </p>
            </div>

            {/* Secret Field */}
            <div>
              <label
                htmlFor="webhook-secret"
                className="block text-sm font-medium text-card-foreground mb-2"
              >
                Secret Key {isEditing && "(leave blank to keep current)"}
              </label>
              <input
                id="webhook-secret"
                type="password"
                value={secret}
                onChange={(e) => setSecret(e.target.value)}
                disabled={isSubmitting}
                className="w-full px-4 py-3 bg-background/50 border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed transition-all font-mono"
                placeholder={
                  isEditing
                    ? `••••${webhook?.secret_last_four || ""}`
                    : "mysecretkey12345"
                }
                required={!isEditing}
              />
              <p className="text-xs text-muted-foreground mt-1">
                16-128 characters. Used to verify webhook authenticity.
              </p>
            </div>

            {/* Active Toggle */}
            <div className="flex items-center gap-3">
              <input
                id="webhook-active"
                type="checkbox"
                checked={isActive}
                onChange={(e) => setIsActive(e.target.checked)}
                disabled={isSubmitting}
                className="w-4 h-4 text-primary bg-background/50 border-input rounded focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background disabled:opacity-50 disabled:cursor-not-allowed"
              />
              <label
                htmlFor="webhook-active"
                className="text-sm font-medium text-card-foreground"
              >
                Active (receive notifications)
              </label>
            </div>

            {/* Validation Error */}
            {validationError && (
              <div className="bg-destructive/10 border border-destructive/30 rounded-lg p-3">
                <p className="text-destructive text-sm">{validationError}</p>
              </div>
            )}

            {/* Form Actions */}
            <div className="flex gap-3 pt-2">
              <button
                type="button"
                onClick={onClose}
                disabled={isSubmitting}
                className="flex-1 bg-muted hover:bg-muted/80 disabled:bg-muted/50 disabled:cursor-not-allowed text-foreground font-medium py-3 px-4 rounded-lg transition-all focus:outline-none focus:ring-2 focus:ring-muted focus:ring-offset-2 focus:ring-offset-background"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isSubmitting}
                className="flex-1 bg-primary hover:bg-primary/90 disabled:bg-muted disabled:cursor-not-allowed text-primary-foreground font-medium py-3 px-4 rounded-lg transition-all focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background"
              >
                {isSubmitting
                  ? isEditing
                    ? "Saving..."
                    : "Creating..."
                  : isEditing
                  ? "Save Changes"
                  : "Create Webhook"}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
