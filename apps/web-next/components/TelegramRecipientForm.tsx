"use client";

import { useState, FormEvent } from "react";
import { Button } from "@/components/ui/button";

interface TelegramRecipientFormProps {
  onSubmit: (chatId: string) => Promise<void>;
  onCancel: () => void;
}

export function TelegramRecipientForm({
  onSubmit,
  onCancel,
}: TelegramRecipientFormProps) {
  const [chatId, setChatId] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [validationError, setValidationError] = useState("");

  function validateChatId(value: string): string | null {
    if (!value.trim()) {
      return "Chat ID is required";
    }

    // Chat ID should be numeric and typically 9-10 digits
    if (!/^\d{6,15}$/.test(value.trim())) {
      return "Chat ID must be a numeric value (6-15 digits)";
    }

    return null;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();

    const error = validateChatId(chatId);
    if (error) {
      setValidationError(error);
      return;
    }

    setIsSubmitting(true);
    setValidationError("");

    try {
      await onSubmit(chatId.trim());
      setChatId("");
    } catch (err) {
      // Error handling is done in parent component
      setValidationError(
        err instanceof Error ? err.message : "Failed to add recipient"
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="w-full max-w-lg mx-4 rounded-2xl border border-border/40 bg-card/95 backdrop-blur-xl shadow-2xl">
        <div className="px-6 py-4 border-b border-border/40">
          <h3 className="text-xl font-semibold text-foreground">
            Add Telegram Recipient
          </h3>
          <p className="text-sm text-muted-foreground mt-1">
            Enter the Telegram chat ID to receive alert notifications
          </p>
        </div>

        <form onSubmit={handleSubmit} className="px-6 py-6 space-y-4">
          <div className="space-y-2">
            <label
              htmlFor="chatId"
              className="text-sm font-medium text-foreground"
            >
              Telegram Chat ID
            </label>
            <input
              id="chatId"
              type="text"
              value={chatId}
              onChange={(e) => {
                setChatId(e.target.value);
                setValidationError("");
              }}
              placeholder="e.g., 123456789"
              className="w-full px-4 py-2 rounded-lg border border-border/40 bg-background/50 text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-[#0088cc]/50 focus:border-[#0088cc]/50"
              disabled={isSubmitting}
              autoFocus
            />
            {validationError && (
              <p className="text-sm text-destructive">{validationError}</p>
            )}
            <p className="text-xs text-muted-foreground">
              Get your chat ID from @userinfobot on Telegram
            </p>
          </div>

          <div className="flex gap-3 justify-end pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={onCancel}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isSubmitting}
              className="bg-[#0088cc] hover:bg-[#0088cc]/90 text-white"
            >
              {isSubmitting ? "Adding..." : "Add Recipient"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
