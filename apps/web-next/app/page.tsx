"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { MetricsCard } from "@/components/MetricsCard";
import { SystemInfoCard, ActiveAlertsCard } from "@/components/dashboard";
import { useSession } from "@/lib/useSession";
import { useAlertsWithNotifications } from "@/lib/alerts";
import { Badge } from "@/components/ui/badge";

export default function Home() {
  const router = useRouter();
  const { status, user, logout } = useSession();
  const { events, rules, newAlertsCount } = useAlertsWithNotifications(10); // Fetch limited events for badge count

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
      <div className="border-b border-border/40 bg-card/40 backdrop-blur-xl">
        <div className="max-w-6xl mx-auto px-6 py-4 flex flex-wrap gap-4 justify-between items-center">
          <div className="flex items-center gap-3 text-primary">
            <span className="text-2xl">🌙</span>
            <span className="font-semibold tracking-wide">LunaSentri</span>
          </div>
          <div className="flex items-center gap-3 text-sm">
            {(unacknowledgedCount > 0 || newAlertsCount > 0) && (
              <Link
                href="/alerts"
                className="rounded-full bg-destructive/20 border border-destructive/30 px-4 py-2 text-destructive transition-all duration-200 hover:bg-destructive/30 hover:-translate-y-0.5 flex items-center gap-2"
              >
                <span>Alerts</span>
                <div className="flex items-center gap-1">
                  <Badge
                    variant="destructive"
                    className="text-xs px-1.5 py-0.5"
                  >
                    {unacknowledgedCount}
                  </Badge>
                  {newAlertsCount > 0 && (
                    <Badge
                      variant="default"
                      className="text-xs px-1.5 py-0.5 bg-blue-500"
                    >
                      {newAlertsCount} new
                    </Badge>
                  )}
                </div>
              </Link>
            )}
            {user?.is_admin && (
              <>
                <Link
                  href="/alerts"
                  className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-muted-foreground transition-all duration-200 hover:text-foreground hover:border-border"
                >
                  Alerts
                </Link>
                <Link
                  href="/users"
                  className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-muted-foreground transition-all duration-200 hover:text-foreground hover:border-border"
                >
                  Manage Users
                </Link>
                <Link
                  href="/notifications/telegram"
                  className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-muted-foreground transition-all duration-200 hover:text-foreground hover:border-border"
                >
                  Telegram Alerts
                </Link>
              </>
            )}
            <Link
              href="/settings"
              className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-muted-foreground transition-all duration-200 hover:text-foreground hover:border-border"
            >
              Settings
            </Link>
            <Link
              href="/notifications/telegram"
              className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-muted-foreground transition-all duration-200 hover:text-foreground hover:border-border"
            >
              Telegram Alerts
            </Link>
            <span className="text-muted-foreground hidden sm:inline">
              {user?.email}
            </span>
            <button
              onClick={logout}
              className="rounded-full bg-accent/20 border border-accent/30 px-4 py-2 text-accent-foreground transition-all duration-200 hover:bg-accent/30 hover:-translate-y-0.5"
            >
              Logout
            </button>
          </div>
        </div>
      </div>

      <div className="min-h-[calc(100vh-82px)] px-4 py-8">
        <div className="max-w-6xl mx-auto">
          <div className="space-y-8">
            {/* Hero Header */}
            <div className="space-y-2">
              <h1 className="text-4xl sm:text-5xl font-semibold tracking-wide text-primary drop-shadow-xl">
                Lunar System Pulse
              </h1>
              <p className="text-muted-foreground text-base sm:text-lg">
                Real-time insight into your infrastructure health with a moonlit
                touch.
              </p>
            </div>

            {/* Responsive Grid Layout */}
            <div className="grid gap-6 lg:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
              {/* Left Column */}
              <div className="space-y-6">
                <MetricsCard />
              </div>

              {/* Right Column */}
              <div className="space-y-6">
                <SystemInfoCard />
                <ActiveAlertsCard events={events} rules={rules} />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
