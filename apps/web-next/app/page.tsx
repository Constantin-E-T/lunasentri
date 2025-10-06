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
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 to-slate-800">
        <div className="text-slate-400">Loading...</div>
      </div>
    );
  }

  // Don't render dashboard if not authenticated (will redirect)
  if (status !== "authenticated") {
    return null;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 to-slate-800">
      {/* Header with logout */}
      <div className="border-b border-slate-700/50 bg-slate-800/30 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center gap-3">
            <span className="text-2xl">🌙</span>
            <span className="text-white font-semibold">LunaSentri</span>
          </div>
          <div className="flex items-center gap-4">
            {user?.is_admin && (
              <Link
                href="/users"
                className="text-sm text-slate-300 hover:text-white transition-colors"
              >
                Manage Users
              </Link>
            )}
            <span className="text-sm text-slate-400">{user?.email}</span>
            <button
              onClick={logout}
              className="text-sm text-slate-300 hover:text-white bg-slate-700/50 hover:bg-slate-700 px-4 py-2 rounded-lg transition-colors"
            >
              Logout
            </button>
          </div>
        </div>
      </div>

      {/* Dashboard Content */}
      <div className="flex items-center justify-center min-h-[calc(100vh-73px)]">
        <div className="text-center px-4">
          <h1 className="text-6xl font-bold text-white mb-4">🌙 LunaSentri</h1>
          <p className="text-xl text-slate-300 mb-8">
            Lightweight Server Monitoring Dashboard
          </p>
          <MetricsCard />
        </div>
      </div>
    </div>
  );
}
