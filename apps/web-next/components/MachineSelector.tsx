"use client";

import Link from "next/link";
import { useState, useRef, useEffect } from "react";
import type { Machine } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { ChevronDown, Server, Settings } from "lucide-react";

interface MachineSelectorProps {
  machines: Machine[];
  selectedMachineId: number | null;
  onSelectMachine: (machineId: number | null) => void;
  loading?: boolean;
}

export function MachineSelector({
  machines,
  selectedMachineId,
  onSelectMachine,
  loading = false,
}: MachineSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isOpen]);

  const selectedMachine = machines.find((m) => m.id === selectedMachineId);

  const getStatusColor = (status: string) => {
    return status === "online"
      ? "bg-green-500/20 text-green-500 border-green-500/30"
      : "bg-gray-500/20 text-gray-500 border-gray-500/30";
  };

  if (loading) {
    return (
      <div className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-sm text-muted-foreground">
        <div className="animate-pulse">Loading machines...</div>
      </div>
    );
  }

  if (machines.length === 0) {
    return (
      <Link
        href="/machines"
        className="rounded-full bg-primary/20 border border-primary/30 px-4 py-2 text-primary text-sm transition-all duration-200 hover:bg-primary/30 hover:-translate-y-0.5 flex items-center gap-2"
      >
        <Server className="h-4 w-4" />
        <span>Register Machine</span>
      </Link>
    );
  }

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-sm transition-all duration-200 hover:text-foreground hover:border-border flex items-center gap-2 min-w-[180px]"
      >
        <Server className="h-4 w-4" />
        {selectedMachine ? (
          <>
            <span className="flex-1 text-left truncate">
              {selectedMachine.name}
            </span>
            <Badge
              className={`${getStatusColor(
                selectedMachine.status
              )} text-xs px-1.5 py-0`}
            >
              {selectedMachine.status}
            </Badge>
          </>
        ) : (
          <span className="flex-1 text-left text-muted-foreground">
            Select machine...
          </span>
        )}
        <ChevronDown
          className={`h-4 w-4 transition-transform ${
            isOpen ? "rotate-180" : ""
          }`}
        />
      </button>

      {isOpen && (
        <div className="absolute top-full right-0 mt-2 w-64 rounded-lg bg-card border border-border shadow-lg z-50 overflow-hidden">
          <div className="max-h-80 overflow-y-auto">
            {machines.map((machine) => (
              <button
                key={machine.id}
                onClick={() => {
                  onSelectMachine(machine.id);
                  setIsOpen(false);
                }}
                className={`w-full px-4 py-3 text-left hover:bg-accent/50 transition-colors border-b border-border/30 last:border-b-0 ${
                  machine.id === selectedMachineId ? "bg-accent/30" : ""
                }`}
              >
                <div className="flex items-center justify-between gap-2">
                  <div className="flex-1 min-w-0">
                    <div className="font-medium truncate">{machine.name}</div>
                    {machine.hostname && (
                      <div className="text-xs text-muted-foreground truncate">
                        {machine.hostname}
                      </div>
                    )}
                  </div>
                  <Badge
                    className={`${getStatusColor(
                      machine.status
                    )} text-xs px-1.5 py-0 shrink-0`}
                  >
                    {machine.status}
                  </Badge>
                </div>
              </button>
            ))}
          </div>

          <Link
            href="/machines"
            onClick={() => setIsOpen(false)}
            className="flex items-center gap-2 w-full px-4 py-3 text-sm text-primary hover:bg-accent/50 transition-colors border-t border-border"
          >
            <Settings className="h-4 w-4" />
            <span>Manage Machines</span>
          </Link>
        </div>
      )}
    </div>
  );
}
