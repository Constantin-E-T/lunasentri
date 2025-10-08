"use client";

import { useState, FormEvent, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useSession } from "@/lib/useSession";
import { useWebhooks } from "@/lib/alerts/useWebhooks";
import type {
  WebhookWithState,
  CreateWebhookRequest,
  UpdateWebhookRequest,
} from "@/lib/alerts/useWebhooks";
import {
  WebhookList,
  WebhookForm,
  WebhookEmptyState,
  DeleteWebhookDialog,
} from "@/components/settings/notifications";
import { useToast } from "@/components/ui/use-toast";

export default function SettingsPage() {
  const router = useRouter();
  const { status, changePassword } = useSession();
  const { toast } = useToast();

  // Password change state
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmNewPassword, setConfirmNewPassword] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Webhook management state
  const {
    webhooks,
    loading: webhooksLoading,
    error: webhooksError,
    createWebhook,
    updateWebhook,
    deleteWebhook,
    sendTestWebhook,
    refresh: refreshWebhooks,
  } = useWebhooks();

  const [isWebhookFormOpen, setIsWebhookFormOpen] = useState(false);
  const [editingWebhook, setEditingWebhook] = useState<
    WebhookWithState | undefined
  >();
  const [deletingWebhook, setDeletingWebhook] = useState<
    WebhookWithState | undefined
  >();
  const [isDeleting, setIsDeleting] = useState(false);

  // Redirect if unauthenticated
  useEffect(() => {
    if (status === "unauthenticated") {
      router.push("/login");
    }
  }, [status, router]);

  // Webhook handlers
  function handleAddWebhook() {
    setEditingWebhook(undefined);
    setIsWebhookFormOpen(true);
  }

  function handleEditWebhook(webhook: WebhookWithState) {
    setEditingWebhook(webhook);
    setIsWebhookFormOpen(true);
  }

  async function handleWebhookSubmit(
    payload: CreateWebhookRequest | UpdateWebhookRequest
  ) {
    try {
      if (editingWebhook) {
        await updateWebhook(editingWebhook.id, payload as UpdateWebhookRequest);
        toast({
          title: "Webhook updated",
          description: "Your webhook has been updated successfully.",
        });
      } else {
        await createWebhook(payload as CreateWebhookRequest);
        toast({
          title: "Webhook created",
          description: "Your webhook has been created successfully.",
        });
      }
      setIsWebhookFormOpen(false);
      setEditingWebhook(undefined);
    } catch (err) {
      toast({
        title: "Error",
        description:
          err instanceof Error ? err.message : "Failed to save webhook",
        variant: "destructive",
      });
      throw err;
    }
  }

  function handleDeleteWebhook(webhook: WebhookWithState) {
    setDeletingWebhook(webhook);
  }

  async function confirmDeleteWebhook() {
    if (!deletingWebhook) return;

    setIsDeleting(true);
    try {
      await deleteWebhook(deletingWebhook.id);
      toast({
        title: "Webhook deleted",
        description: "The webhook has been deleted successfully.",
      });
      setDeletingWebhook(undefined);
    } catch (err) {
      toast({
        title: "Error",
        description:
          err instanceof Error ? err.message : "Failed to delete webhook",
        variant: "destructive",
      });
    } finally {
      setIsDeleting(false);
    }
  }

  async function handleTestWebhook(webhook: WebhookWithState) {
    try {
      await sendTestWebhook(webhook.id);
      toast({
        title: "Test sent",
        description: `Test payload sent to ${webhook.url}`,
      });
    } catch (err) {
      toast({
        title: "Test failed",
        description:
          err instanceof Error ? err.message : "Failed to send test webhook",
        variant: "destructive",
      });
    }
  }

  // Password change validation and handlers

  function validateForm(): string | null {
    if (!currentPassword.trim()) {
      return "Current password is required";
    }
    if (newPassword.length < 8) {
      return "New password must be at least 8 characters long";
    }
    if (newPassword !== confirmNewPassword) {
      return "New password confirmation does not match";
    }
    if (newPassword === currentPassword) {
      return "New password must be different from current password";
    }
    return null;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSuccess("");

    // Client-side validation
    const validationError = validateForm();
    if (validationError) {
      setError(validationError);
      return;
    }

    setIsSubmitting(true);

    try {
      await changePassword(currentPassword, newPassword);
      setSuccess("Password updated successfully");
      // Clear form
      setCurrentPassword("");
      setNewPassword("");
      setConfirmNewPassword("");
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to change password"
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  // Show loading state while checking authentication
  if (status === "loading") {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-muted-foreground animate-pulse">Loading...</div>
      </div>
    );
  }

  // Don't show settings if not authenticated (will redirect)
  if (status !== "authenticated") {
    return null;
  }

  return (
    <div className="min-h-screen">
      {/* Header */}
      <div className="border-b border-border/40 bg-card/40 backdrop-blur-xl">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center gap-3">
            <Link
              href="/"
              className="flex items-center gap-3 hover:opacity-80 transition-opacity"
            >
              <span className="text-2xl">🌙</span>
              <span className="text-primary font-semibold">LunaSentri</span>
            </Link>
          </div>
          <div className="flex items-center gap-4">
            <Link
              href="/"
              className="text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              Dashboard
            </Link>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-2xl mx-auto px-4 py-12">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-primary mb-2">Settings</h1>
          <p className="text-muted-foreground">Manage your account settings</p>
        </div>

        {/* Change Password Card */}
        <div className="bg-card/70 backdrop-blur-xl rounded-xl p-8 shadow-2xl border border-border/30">
          <h2 className="text-xl font-semibold text-card-foreground mb-6">
            Change Password
          </h2>

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Current Password Field */}
            <div>
              <label
                htmlFor="currentPassword"
                className="block text-sm font-medium text-card-foreground mb-2"
              >
                Current Password
              </label>
              <input
                id="currentPassword"
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                required
                disabled={isSubmitting}
                className="w-full px-4 py-3 bg-background/50 border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                placeholder="••••••••"
                autoComplete="current-password"
              />
            </div>

            {/* New Password Field */}
            <div>
              <label
                htmlFor="newPassword"
                className="block text-sm font-medium text-card-foreground mb-2"
              >
                New Password
              </label>
              <input
                id="newPassword"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
                disabled={isSubmitting}
                className="w-full px-4 py-3 bg-background/50 border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                placeholder="••••••••"
                autoComplete="new-password"
                minLength={8}
              />
              <p className="text-xs text-muted-foreground mt-1">
                Must be at least 8 characters long and different from current
                password
              </p>
            </div>

            {/* Confirm New Password Field */}
            <div>
              <label
                htmlFor="confirmNewPassword"
                className="block text-sm font-medium text-card-foreground mb-2"
              >
                Confirm New Password
              </label>
              <input
                id="confirmNewPassword"
                type="password"
                value={confirmNewPassword}
                onChange={(e) => setConfirmNewPassword(e.target.value)}
                required
                disabled={isSubmitting}
                className="w-full px-4 py-3 bg-background/50 border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                placeholder="••••••••"
                autoComplete="new-password"
              />
            </div>

            {/* Success Message */}
            {success && (
              <div className="bg-chart-4/10 border border-chart-4/30 rounded-lg p-4">
                <p className="text-chart-4 text-sm">{success}</p>
              </div>
            )}

            {/* Error Message */}
            {error && (
              <div className="bg-destructive/10 border border-destructive/30 rounded-lg p-4">
                <p className="text-destructive text-sm">{error}</p>
              </div>
            )}

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full bg-primary hover:bg-primary/90 disabled:bg-muted disabled:cursor-not-allowed text-primary-foreground font-medium py-3 px-4 rounded-lg transition-all focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background"
            >
              {isSubmitting ? "Changing password..." : "Change Password"}
            </button>
          </form>
        </div>

        {/* Additional Settings Placeholder */}
        <div className="mt-8 bg-card/70 backdrop-blur-xl rounded-xl p-8 shadow-2xl border border-border/30">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold text-card-foreground mb-1">
                Notifications
              </h2>
              <p className="text-sm text-muted-foreground">
                Manage webhook endpoints for alert notifications
              </p>
            </div>
            {webhooks.length > 0 && (
              <button
                onClick={handleAddWebhook}
                className="bg-primary hover:bg-primary/90 text-primary-foreground font-medium py-2 px-4 rounded-lg transition-all focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 focus:ring-offset-background"
              >
                Add Webhook
              </button>
            )}
          </div>

          {webhooksLoading ? (
            <div className="text-center py-8">
              <div className="text-muted-foreground animate-pulse">
                Loading webhooks...
              </div>
            </div>
          ) : webhooksError ? (
            <div className="bg-destructive/10 border border-destructive/30 rounded-lg p-4">
              <p className="text-destructive text-sm">{webhooksError}</p>
            </div>
          ) : webhooks.length === 0 ? (
            <WebhookEmptyState onAddWebhook={handleAddWebhook} />
          ) : (
            <WebhookList
              webhooks={webhooks}
              onEdit={handleEditWebhook}
              onDelete={handleDeleteWebhook}
              onTest={handleTestWebhook}
            />
          )}
        </div>
      </div>

      {/* Webhook Form Dialog */}
      <WebhookForm
        webhook={editingWebhook}
        isOpen={isWebhookFormOpen}
        onClose={() => {
          setIsWebhookFormOpen(false);
          setEditingWebhook(undefined);
        }}
        onSubmit={handleWebhookSubmit}
      />

      {/* Delete Confirmation Dialog */}
      <DeleteWebhookDialog
        isOpen={!!deletingWebhook}
        onClose={() => setDeletingWebhook(undefined)}
        onConfirm={confirmDeleteWebhook}
        webhookUrl={deletingWebhook?.url}
        loading={isDeleting}
      />
    </div>
  );
}
