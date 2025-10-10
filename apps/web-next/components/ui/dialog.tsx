import * as React from "react";

export interface DialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  children: React.ReactNode;
}

export interface DialogContentProps {
  className?: string;
  children: React.ReactNode;
}

export interface DialogHeaderProps {
  className?: string;
  children: React.ReactNode;
}

export interface DialogFooterProps {
  className?: string;
  children: React.ReactNode;
}

export interface DialogTitleProps {
  className?: string;
  children: React.ReactNode;
}

export interface DialogDescriptionProps {
  className?: string;
  children: React.ReactNode;
}

export function Dialog({ open, onOpenChange, children }: DialogProps) {
  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
      onClick={() => onOpenChange?.(false)}
    >
      {children}
    </div>
  );
}

export function DialogContent({
  className = "",
  children,
}: DialogContentProps) {
  return (
    <div
      className={`relative bg-card border border-border rounded-lg shadow-lg p-6 w-full max-w-lg max-h-[90vh] overflow-y-auto ${className}`}
      onClick={(e) => e.stopPropagation()}
    >
      {children}
    </div>
  );
}

export function DialogHeader({ className = "", children }: DialogHeaderProps) {
  return <div className={`space-y-2 mb-4 ${className}`}>{children}</div>;
}

export function DialogFooter({ className = "", children }: DialogFooterProps) {
  return (
    <div className={`flex justify-end gap-2 mt-6 ${className}`}>{children}</div>
  );
}

export function DialogTitle({ className = "", children }: DialogTitleProps) {
  return <h2 className={`text-xl font-semibold ${className}`}>{children}</h2>;
}

export function DialogDescription({
  className = "",
  children,
}: DialogDescriptionProps) {
  return (
    <p className={`text-sm text-muted-foreground ${className}`}>{children}</p>
  );
}
