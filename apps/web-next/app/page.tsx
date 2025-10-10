"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { MetricsCard } from "@/components/MetricsCard";
import { SystemInfoCard, ActiveAlertsCard } from "@/components/dashboard";
import { MachineSelector } from "@/components/MachineSelector";
import { Navbar } from "@/components/Navbar";
import { useSession } from "@/lib/useSession";
import { useAlertsWithNotifications } from "@/lib/alerts";
import { useMachines } from "@/lib/useMachines";
import { useMachineSelection } from "@/lib/useMachineSelection";

export default function Home() {
  const router = useRouter();
  const { status, user, logout } = useSession();
  const { events, rules, newAlertsCount } = useAlertsWithNotifications(10); // Fetch limited events for badge count
  const { machines, loading: machinesLoading } = useMachines();
  const { selectedMachineId, selectMachine } = useMachineSelection();

  // Count unacknowledged events from live data
  const unacknowledgedCount =
    events?.filter((e) => !e.acknowledged).length || 0;

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

  // Don't render dashboard if not authenticated (will redirect)
  if (status !== "authenticated") {
    return null;
  }

  return (
    <div className="min-h-screen">
      <Navbar
        user={user}
        unacknowledgedCount={unacknowledgedCount}
        newAlertsCount={newAlertsCount}
        onLogout={logout}
      />

      <div className="px-4 py-8">
        <div className="max-w-7xl mx-auto">
          <div className="space-y-8">
            {/* Hero Header with Machine Selector */}
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
              <div className="space-y-2">
                <h1 className="text-4xl sm:text-5xl font-semibold tracking-wide text-primary drop-shadow-xl">
                  Lunar System Pulse
                </h1>
                <p className="text-muted-foreground text-base sm:text-lg">
                  Real-time insight into your infrastructure health with a
                  moonlit touch.
                </p>
              </div>
              <div className="flex-shrink-0">
                <MachineSelector
                  machines={machines || []}
                  selectedMachineId={selectedMachineId}
                  onSelectMachine={selectMachine}
                />
              </div>
            </div>

            {/* Responsive Grid Layout */}
            <div className="grid gap-6 lg:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
              {/* Left Column */}
              <div className="space-y-6">
                <MetricsCard machineId={selectedMachineId ?? undefined} />
              </div>

              {/* Right Column */}
              <div className="space-y-6">
                <SystemInfoCard machineId={selectedMachineId ?? undefined} />
                <ActiveAlertsCard events={events} rules={rules} />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
