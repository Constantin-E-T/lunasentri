"use client";

import * as React from "react";

export interface DropdownMenuProps {
  children: React.ReactNode;
}

export function DropdownMenu({ children }: DropdownMenuProps) {
  const [open, setOpen] = React.useState(false);

  return (
    <div className="relative">
      {React.Children.map(children, (child) => {
        if (React.isValidElement(child)) {
          return React.cloneElement(child as React.ReactElement<any>, {
            open,
            setOpen,
          });
        }
        return child;
      })}
    </div>
  );
}

export interface DropdownMenuTriggerProps {
  children: React.ReactNode;
  className?: string;
  open?: boolean;
  setOpen?: (open: boolean) => void;
}

export function DropdownMenuTrigger({
  children,
  className = "",
  open,
  setOpen,
}: DropdownMenuTriggerProps) {
  return (
    <button
      onClick={() => setOpen?.(!open)}
      className={className}
      type="button"
    >
      {children}
    </button>
  );
}

export interface DropdownMenuContentProps {
  children: React.ReactNode;
  align?: "start" | "end";
  className?: string;
  open?: boolean;
  setOpen?: (open: boolean) => void;
}

export function DropdownMenuContent({
  children,
  align = "start",
  className = "",
  open,
  setOpen,
}: DropdownMenuContentProps) {
  const contentRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        contentRef.current &&
        !contentRef.current.contains(event.target as Node)
      ) {
        setOpen?.(false);
      }
    }

    if (open) {
      document.addEventListener("mousedown", handleClickOutside);
      return () => {
        document.removeEventListener("mousedown", handleClickOutside);
      };
    }
  }, [open, setOpen]);

  if (!open) return null;

  return (
    <div
      ref={contentRef}
      className={`absolute top-full mt-2 min-w-[180px] rounded-lg border border-border/40 bg-card/95 backdrop-blur-xl shadow-xl z-50 transition-all duration-200 ${
        align === "end" ? "right-0" : "left-0"
      } ${className}`}
      style={{
        animation: "fadeIn 0.15s ease-out",
      }}
    >
      <div className="p-1">{children}</div>
    </div>
  );
}

export interface DropdownMenuItemProps {
  children: React.ReactNode;
  onClick?: () => void;
  className?: string;
  asChild?: boolean;
}

export function DropdownMenuItem({
  children,
  onClick,
  className = "",
  asChild,
}: DropdownMenuItemProps) {
  if (asChild) {
    return (
      <div className={`block w-full ${className}`}>
        {React.Children.map(children, (child) => {
          if (React.isValidElement(child)) {
            return React.cloneElement(child as React.ReactElement<any>, {
              className: `flex items-center w-full px-3 py-2 text-sm rounded-md hover:bg-accent/20 transition-colors ${
                (child.props as any).className || ""
              }`,
            });
          }
          return child;
        })}
      </div>
    );
  }

  return (
    <button
      onClick={onClick}
      className={`flex items-center w-full px-3 py-2 text-sm rounded-md hover:bg-accent/20 transition-colors ${className}`}
    >
      {children}
    </button>
  );
}

export function DropdownMenuSeparator() {
  return <div className="h-px bg-border/40 my-1" />;
}
