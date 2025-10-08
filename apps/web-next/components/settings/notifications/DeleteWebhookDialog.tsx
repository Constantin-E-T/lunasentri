"use client";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

interface DeleteWebhookDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  webhookUrl?: string;
  loading?: boolean;
}

export function DeleteWebhookDialog({
  isOpen,
  onClose,
  onConfirm,
  webhookUrl,
  loading = false,
}: DeleteWebhookDialogProps) {
  return (
    <AlertDialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Webhook</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete the webhook for{" "}
            {webhookUrl ? (
              <code className="px-1 py-0.5 bg-muted/50 rounded text-xs">
                {webhookUrl}
              </code>
            ) : (
              "this endpoint"
            )}
            ? This action cannot be undone and you will stop receiving
            notifications at this URL.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={loading ? undefined : onClose}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            disabled={loading}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {loading ? "Deleting..." : "Delete"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
