"use client";

import { useState } from "react";
import type { TelegramRecipient } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

interface TelegramRecipientTableProps {
  recipients: TelegramRecipient[];
  onToggleActive: (id: number, isActive: boolean) => Promise<void>;
  onTest: (id: number) => Promise<void>;
  onDelete: (id: number) => Promise<void>;
}

export function TelegramRecipientTable({
  recipients,
  onToggleActive,
  onTest,
  onDelete,
}: TelegramRecipientTableProps) {
  const [loadingStates, setLoadingStates] = useState<{
    [key: number]: "toggle" | "test" | "delete" | null;
  }>({});
  const [deleteConfirm, setDeleteConfirm] = useState<number | null>(null);

  async function handleToggle(id: number, currentActive: boolean) {
    setLoadingStates({ ...loadingStates, [id]: "toggle" });
    try {
      await onToggleActive(id, !currentActive);
    } finally {
      setLoadingStates({ ...loadingStates, [id]: null });
    }
  }

  async function handleTest(id: number) {
    setLoadingStates({ ...loadingStates, [id]: "test" });
    try {
      await onTest(id);
    } finally {
      setLoadingStates({ ...loadingStates, [id]: null });
    }
  }

  async function handleDelete(id: number) {
    setLoadingStates({ ...loadingStates, [id]: "delete" });
    try {
      await onDelete(id);
      setDeleteConfirm(null);
    } finally {
      setLoadingStates({ ...loadingStates, [id]: null });
    }
  }

  function formatDate(dateString?: string): string {
    if (!dateString) return "Never";
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  }

  return (
    <div className="rounded-2xl border border-border/40 bg-card/40 backdrop-blur-xl overflow-hidden">
      {/* Desktop Table */}
      <div className="hidden md:block overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border/40 bg-card/60">
              <th className="text-left px-6 py-4 text-sm font-semibold text-foreground">
                Chat ID
              </th>
              <th className="text-left px-6 py-4 text-sm font-semibold text-foreground">
                Status
              </th>
              <th className="text-left px-6 py-4 text-sm font-semibold text-foreground">
                Last Success
              </th>
              <th className="text-left px-6 py-4 text-sm font-semibold text-foreground">
                Failures
              </th>
              <th className="text-right px-6 py-4 text-sm font-semibold text-foreground">
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            {recipients.map((recipient) => {
              const isLoading = loadingStates[recipient.id];
              const isDeleting = deleteConfirm === recipient.id;

              return (
                <tr
                  key={recipient.id}
                  className="border-b border-border/20 hover:bg-card/60 transition-colors"
                >
                  <td className="px-6 py-4">
                    <code className="text-sm font-mono text-foreground bg-muted/40 px-2 py-1 rounded">
                      {recipient.chat_id}
                    </code>
                  </td>
                  <td className="px-6 py-4">
                    {recipient.is_active ? (
                      <Badge
                        variant="default"
                        className="bg-[#0088cc] hover:bg-[#0088cc]/90"
                      >
                        Active
                      </Badge>
                    ) : (
                      <Badge variant="secondary">Inactive</Badge>
                    )}
                  </td>
                  <td className="px-6 py-4 text-sm text-muted-foreground">
                    {formatDate(recipient.last_success_at)}
                  </td>
                  <td className="px-6 py-4">
                    {recipient.failure_count > 0 ? (
                      <Badge variant="destructive">
                        {recipient.failure_count}
                      </Badge>
                    ) : (
                      <span className="text-sm text-muted-foreground">0</span>
                    )}
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center justify-end gap-2">
                      {isDeleting ? (
                        <>
                          <span className="text-sm text-muted-foreground">
                            Confirm delete?
                          </span>
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => handleDelete(recipient.id)}
                            disabled={!!isLoading}
                          >
                            {isLoading === "delete" ? "Deleting..." : "Yes"}
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => setDeleteConfirm(null)}
                            disabled={!!isLoading}
                          >
                            No
                          </Button>
                        </>
                      ) : (
                        <>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleTest(recipient.id)}
                            disabled={!!isLoading || !recipient.is_active}
                            className="text-[#0088cc] border-[#0088cc]/30 hover:bg-[#0088cc]/10"
                          >
                            {isLoading === "test" ? "Sending..." : "Test"}
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() =>
                              handleToggle(recipient.id, recipient.is_active)
                            }
                            disabled={!!isLoading}
                          >
                            {isLoading === "toggle"
                              ? "..."
                              : recipient.is_active
                              ? "Disable"
                              : "Enable"}
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => setDeleteConfirm(recipient.id)}
                            disabled={!!isLoading}
                            className="text-destructive border-destructive/30 hover:bg-destructive/10"
                          >
                            Delete
                          </Button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {/* Mobile Cards */}
      <div className="md:hidden divide-y divide-border/20">
        {recipients.map((recipient) => {
          const isLoading = loadingStates[recipient.id];
          const isDeleting = deleteConfirm === recipient.id;

          return (
            <div key={recipient.id} className="p-4 space-y-3">
              <div className="flex items-start justify-between">
                <div className="space-y-1">
                  <div className="text-xs text-muted-foreground">Chat ID</div>
                  <code className="text-sm font-mono text-foreground bg-muted/40 px-2 py-1 rounded">
                    {recipient.chat_id}
                  </code>
                </div>
                {recipient.is_active ? (
                  <Badge
                    variant="default"
                    className="bg-[#0088cc] hover:bg-[#0088cc]/90"
                  >
                    Active
                  </Badge>
                ) : (
                  <Badge variant="secondary">Inactive</Badge>
                )}
              </div>

              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <div className="text-xs text-muted-foreground mb-1">
                    Last Success
                  </div>
                  <div className="text-foreground">
                    {formatDate(recipient.last_success_at)}
                  </div>
                </div>
                <div>
                  <div className="text-xs text-muted-foreground mb-1">
                    Failures
                  </div>
                  {recipient.failure_count > 0 ? (
                    <Badge variant="destructive">
                      {recipient.failure_count}
                    </Badge>
                  ) : (
                    <span className="text-foreground">0</span>
                  )}
                </div>
              </div>

              {isDeleting ? (
                <div className="space-y-2">
                  <p className="text-sm text-muted-foreground">
                    Are you sure you want to delete this recipient?
                  </p>
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="destructive"
                      onClick={() => handleDelete(recipient.id)}
                      disabled={!!isLoading}
                      className="flex-1"
                    >
                      {isLoading === "delete" ? "Deleting..." : "Yes, Delete"}
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => setDeleteConfirm(null)}
                      disabled={!!isLoading}
                      className="flex-1"
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleTest(recipient.id)}
                    disabled={!!isLoading || !recipient.is_active}
                    className="flex-1 text-[#0088cc] border-[#0088cc]/30 hover:bg-[#0088cc]/10"
                  >
                    {isLoading === "test" ? "Sending..." : "Test"}
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() =>
                      handleToggle(recipient.id, recipient.is_active)
                    }
                    disabled={!!isLoading}
                    className="flex-1"
                  >
                    {isLoading === "toggle"
                      ? "..."
                      : recipient.is_active
                      ? "Disable"
                      : "Enable"}
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setDeleteConfirm(recipient.id)}
                    disabled={!!isLoading}
                    className="text-destructive border-destructive/30 hover:bg-destructive/10"
                  >
                    🗑️
                  </Button>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
