"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { MetricsCard } from "@/components/MetricsCard";
import { useSession } from "@/lib/useSession";

export default function Home() {
  const router = useRouter();
  const { status, user, logout } = useSession();

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
            {user?.is_admin && (
              <Link
                href="/users"
                className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-muted-foreground transition-all duration-200 hover:text-foreground hover:border-border"
              >
                Manage Users
              </Link>
            )}
            <Link
              href="/settings"
              className="rounded-full bg-card/40 border border-border/30 px-4 py-2 text-muted-foreground transition-all duration-200 hover:text-foreground hover:border-border"
            >
              Settings
            </Link>
            <span className="text-muted-foreground hidden sm:inline">{user?.email}</span>
            <button
              onClick={logout}
              className="rounded-full bg-accent/20 border border-accent/30 px-4 py-2 text-accent-foreground transition-all duration-200 hover:bg-accent/30 hover:-translate-y-0.5"
            >
              Logout
            </button>
          </div>
        </div>
      </div>

      <div className="flex items-center justify-center min-h-[calc(100vh-82px)] px-4">
        <div className="text-center space-y-6 max-w-3xl w-full">
          <div className="space-y-2">
            <h1 className="text-4xl sm:text-5xl font-semibold tracking-wide text-primary drop-shadow-xl">
              Lunar System Pulse
            </h1>
            <p className="text-muted-foreground text-base sm:text-lg">
              Real-time insight into your infrastructure health with a moonlit touch.
            </p>
          </div>
          <MetricsCard />
        </div>
      </div>
    </div>
  );
}
