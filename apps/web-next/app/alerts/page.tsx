"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useSession } from "@/lib/useSession";
import { useAlertsWithNotifications } from "@/lib/alerts";
import { useAlertsManagement } from "@/lib/alerts/useAlertsManagement";
import type { CreateAlertRuleRequest, AlertRule, AlertEvent } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  AlertRuleForm,
  AlertRuleTemplates,
  AlertRulesTable,
  AlertEventsList,
  DeleteAlertDialog,
} from "@/components/alerts";

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

  const {
    createRule: managedCreateRule,
    updateRule: managedUpdateRule,
    deleteRule: managedDeleteRule,
    acknowledgeEvent: managedAcknowledgeEvent,
  } = useAlertsManagement();

  const [showForm, setShowForm] = useState(false);
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null);
  const [templateData, setTemplateData] =
    useState<CreateAlertRuleRequest | null>(null);
  const [message, setMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  // Delete dialog state
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [ruleToDelete, setRuleToDelete] = useState<{
    id: number;
    name: string;
  } | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);

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
      await managedCreateRule(ruleData);
      setShowForm(false);
      setTemplateData(null);
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

  const handleTemplateSelect = (template: CreateAlertRuleRequest) => {
    setTemplateData(template);
    setShowForm(true);
  };

  const handleFormCancel = () => {
    setShowForm(false);
    setTemplateData(null);
  };

  const handleUpdateRule = async (ruleData: CreateAlertRuleRequest) => {
    if (!editingRule) return;

    try {
      await managedUpdateRule(editingRule.id, ruleData);
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

  const handleDeleteClick = (id: number) => {
    const rule = rules?.find((r) => r.id === id);
    if (rule) {
      setRuleToDelete({ id, name: rule.name });
      setShowDeleteDialog(true);
    }
  };

  const handleDeleteConfirm = async () => {
    if (!ruleToDelete) return;

    setDeleteLoading(true);
    try {
      await managedDeleteRule(ruleToDelete.id);
      setShowDeleteDialog(false);
      setRuleToDelete(null);
      setMessage({ type: "success", text: "Alert rule deleted successfully!" });
      setTimeout(() => setMessage(null), 3000);
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Failed to delete rule",
      });
      setTimeout(() => setMessage(null), 5000);
    } finally {
      setDeleteLoading(false);
    }
  };

  const handleDeleteCancel = () => {
    setShowDeleteDialog(false);
    setRuleToDelete(null);
  };

  const handleAcknowledgeEvent = async (id: number) => {
    try {
      await managedAcknowledgeEvent(id);
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
              href="/notifications/telegram"
              className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-muted-foreground transition-all duration-200 hover:text-foreground hover:border-border"
            >
              Telegram Alerts
            </Link>
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

          {/* Quick Templates */}
          <Card className="bg-card/70 border-border/30 backdrop-blur-xl">
            <CardHeader>
              <CardTitle>Quick Templates</CardTitle>
              <p className="text-sm text-muted-foreground">
                Create common alert rules with one click
              </p>
            </CardHeader>
            <CardContent>
              <AlertRuleTemplates onSelectTemplate={handleTemplateSelect} />
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
                  onDelete={handleDeleteClick}
                  loading={rulesLoading}
                />
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Form Modals */}
      {showForm && (
        <AlertRuleForm
          rule={
            templateData
              ? ({
                  ...templateData,
                  id: 0,
                  created_at: "",
                  updated_at: "",
                } as AlertRule)
              : undefined
          }
          onSubmit={handleCreateRule}
          onCancel={handleFormCancel}
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

      {/* Delete Confirmation Dialog */}
      <DeleteAlertDialog
        isOpen={showDeleteDialog}
        onClose={handleDeleteCancel}
        onConfirm={handleDeleteConfirm}
        ruleName={ruleToDelete?.name}
        loading={deleteLoading}
      />
    </div>
  );
}
