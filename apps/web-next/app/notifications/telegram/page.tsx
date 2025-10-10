"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useSession } from "@/lib/useSession";
import { useAlertsWithNotifications } from "@/lib/alerts";
import { Navbar } from "@/components/Navbar";
import { useToast } from "@/components/ui/use-toast";
import type {
  TelegramRecipient,
  CreateTelegramRecipientRequest,
  UpdateTelegramRecipientRequest,
} from "@/lib/api";
import {
  listTelegramRecipients,
  createTelegramRecipient,
  updateTelegramRecipient,
  deleteTelegramRecipient,
  testTelegramRecipient,
} from "@/lib/api";
import { TelegramSetupGuide } from "@/components/TelegramSetupGuide";
import { TelegramRecipientForm } from "@/components/TelegramRecipientForm";
import { TelegramRecipientTable } from "@/components/TelegramRecipientTable";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

export default function TelegramNotificationsPage() {
  const router = useRouter();
  const { status, user, logout } = useSession();
  const { toast } = useToast();
  const { events, newAlertsCount } = useAlertsWithNotifications(10);

  // Count unacknowledged events
  const unacknowledgedCount =
    events?.filter((e) => !e.acknowledged).length || 0;

  const [recipients, setRecipients] = useState<TelegramRecipient[]>([]);
  const [loading, setLoading] = useState(true);
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingRecipient, setEditingRecipient] = useState<
    TelegramRecipient | undefined
  >();

  // Redirect if unauthenticated
  useEffect(() => {
    if (status === "unauthenticated") {
      router.push("/login");
    }
  }, [status, router]);

  // Fetch recipients
  useEffect(() => {
    if (status === "authenticated") {
      loadRecipients();
    }
  }, [status]);

  async function loadRecipients() {
    try {
      setLoading(true);
      const data = await listTelegramRecipients();
      setRecipients(data);
    } catch (err) {
      toast({
        title: "Error",
        description: "Failed to load Telegram recipients",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  }

  async function handleCreate(chatId: string) {
    try {
      const newRecipient = await createTelegramRecipient({ chat_id: chatId });
      setRecipients([...recipients, newRecipient]);
      setIsFormOpen(false);
      toast({
        title: "Success",
        description: "Telegram recipient added successfully",
      });
    } catch (err) {
      toast({
        title: "Error",
        description:
          err instanceof Error ? err.message : "Failed to add recipient",
        variant: "destructive",
      });
      throw err;
    }
  }

  async function handleToggleActive(id: number, isActive: boolean) {
    try {
      const updated = await updateTelegramRecipient(id, {
        is_active: isActive,
      });
      setRecipients(recipients.map((r) => (r.id === id ? updated : r)));
      toast({
        title: "Success",
        description: `Recipient ${isActive ? "activated" : "deactivated"}`,
      });
    } catch (err) {
      toast({
        title: "Error",
        description: "Failed to update recipient",
        variant: "destructive",
      });
    }
  }

  async function handleTest(id: number) {
    try {
      await testTelegramRecipient(id);
      toast({
        title: "Test message sent",
        description: "Check your Telegram for the test message",
      });
      // Reload to get updated last_success_at
      await loadRecipients();
    } catch (err) {
      toast({
        title: "Error",
        description: "Failed to send test message",
        variant: "destructive",
      });
    }
  }

  async function handleDelete(id: number) {
    try {
      await deleteTelegramRecipient(id);
      setRecipients(recipients.filter((r) => r.id !== id));
      toast({
        title: "Success",
        description: "Recipient deleted successfully",
      });
    } catch (err) {
      toast({
        title: "Error",
        description: "Failed to delete recipient",
        variant: "destructive",
      });
    }
  }

  // Show loading state while checking authentication
  if (status === "loading" || loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-muted-foreground animate-pulse">Loading...</div>
      </div>
    );
  }

  // Don't render if not authenticated
  if (status !== "authenticated") {
    return null;
  }

  const activeCount = recipients.filter((r) => r.is_active).length;
  const inactiveCount = recipients.length - activeCount;

  return (
    <div className="min-h-screen">
      <Navbar
        user={user}
        unacknowledgedCount={unacknowledgedCount}
        newAlertsCount={newAlertsCount}
        onLogout={logout}
      />

      {/* Main Content */}
      <div className="px-4 py-8">
        <div className="max-w-7xl mx-auto space-y-8">
          {/* Header */}
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="space-y-2">
              <h1 className="text-4xl sm:text-5xl font-semibold tracking-wide text-primary drop-shadow-xl flex items-center gap-3">
                <span>📱</span>
                <span>Telegram Notifications</span>
              </h1>
              <p className="text-muted-foreground text-base sm:text-lg">
                Receive instant alerts on Telegram when system thresholds are
                exceeded
              </p>
            </div>
            <div className="flex items-center gap-3">
              {recipients.length > 0 && (
                <div className="flex gap-2">
                  <Badge
                    variant="default"
                    className="bg-[#0088cc] hover:bg-[#0088cc]/90"
                  >
                    {activeCount} Active
                  </Badge>
                  {inactiveCount > 0 && (
                    <Badge variant="secondary">{inactiveCount} Inactive</Badge>
                  )}
                </div>
              )}
              <Button
                onClick={() => setIsFormOpen(true)}
                className="bg-[#0088cc] hover:bg-[#0088cc]/90 text-white"
              >
                + Add Recipient
              </Button>
            </div>
          </div>

          {/* Setup Guide */}
          <TelegramSetupGuide />

          {/* Add Form Dialog */}
          {isFormOpen && (
            <TelegramRecipientForm
              onSubmit={handleCreate}
              onCancel={() => setIsFormOpen(false)}
            />
          )}

          {/* Recipients Table or Empty State */}
          {recipients.length === 0 ? (
            <div className="rounded-2xl border border-border/40 bg-card/40 backdrop-blur-xl p-12 text-center space-y-4">
              <div className="text-6xl">📱</div>
              <h3 className="text-xl font-semibold text-foreground">
                No Telegram Recipients
              </h3>
              <p className="text-muted-foreground max-w-md mx-auto">
                Add your Telegram chat ID to start receiving alert
                notifications. Follow the setup guide above to get your chat ID.
              </p>
              <Button
                onClick={() => setIsFormOpen(true)}
                className="bg-[#0088cc] hover:bg-[#0088cc]/90 text-white mt-4"
              >
                Add Your First Recipient
              </Button>
            </div>
          ) : (
            <TelegramRecipientTable
              recipients={recipients}
              onToggleActive={handleToggleActive}
              onTest={handleTest}
              onDelete={handleDelete}
            />
          )}
        </div>
      </div>
    </div>
  );
}
