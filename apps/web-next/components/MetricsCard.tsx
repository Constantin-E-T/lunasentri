"use client";

import { useMetrics } from "@/lib/useMetrics";

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);

  if (days > 0) {
    return `${days}d ${hours}h ${mins}m`;
  }
  if (hours > 0) {
    return `${hours}h ${mins}m ${secs}s`;
  }
  if (mins > 0) {
    return `${mins}m ${secs}s`;
  }
  return `${secs}s`;
}

function ConnectionStatus({
  type,
  lastUpdate,
}: {
  type: string;
  lastUpdate: Date | null;
}) {
  const getStatusInfo = () => {
    switch (type) {
      case "websocket":
        return {
          icon: "🔗",
          label: "Live (WebSocket)",
          color: "text-green-400",
        };
      case "polling":
        return { icon: "🔄", label: "Polling (5s)", color: "text-yellow-400" };
      default:
        return { icon: "⚠️", label: "Disconnected", color: "text-red-400" };
    }
  };

  const { icon, label, color } = getStatusInfo();
  const timeAgo = lastUpdate
    ? `${Math.round((Date.now() - lastUpdate.getTime()) / 1000)}s ago`
    : "Never";

  return (
    <div className="flex items-center justify-between text-xs">
      <span className={`${color} flex items-center gap-1`}>
        <span>{icon}</span>
        <span>{label}</span>
      </span>
      <span className="text-slate-500">{timeAgo}</span>
    </div>
  );
}

export function MetricsCard() {
  const { metrics, error, loading, connectionType, lastUpdate, retry } =
    useMetrics();

  if (loading) {
    return (
      <div className="max-w-xl mx-auto rounded-2xl bg-card/70 border border-border/30 backdrop-blur-xl shadow-xl p-8">
        <div className="flex items-center justify-center text-muted-foreground animate-pulse">
          Loading metrics...
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-xl mx-auto rounded-2xl bg-destructive/10 border border-destructive/30 backdrop-blur-xl shadow-xl p-8">
        <div className="text-center space-y-4">
          <p className="text-destructive text-sm tracking-wide">⚠️ Connection Error</p>
          <p className="text-muted-foreground text-xs">{error}</p>
          <button
            onClick={retry}
            className="text-xs px-4 py-2 rounded-full border border-destructive/40 bg-destructive/20 text-destructive-foreground transition-all duration-200 hover:bg-destructive/30 hover:-translate-y-0.5"
          >
            Retry Connection
          </button>
        </div>
      </div>
    );
  }

  if (!metrics) {
    return null;
  }

  return (
    <div className="max-w-xl mx-auto rounded-2xl bg-card/70 border border-border/30 backdrop-blur-xl shadow-2xl p-8 transition-all duration-300 hover:-translate-y-1 hover:shadow-[0_20px_60px_rgba(20,40,120,0.35)]">
      <div className="flex justify-between items-center mb-6">
        <div className="text-left">
          <p className="text-xs uppercase tracking-[0.25em] text-muted-foreground">Orbital Feed</p>
          <h2 className="text-2xl font-semibold text-foreground">Live Metrics</h2>
        </div>
        {connectionType === "websocket" && (
          <div
            className="w-3 h-3 rounded-full bg-accent shadow-[0_0_12px_rgba(64,228,255,0.65)] animate-pulse"
            title="Live WebSocket connection"
          />
        )}
      </div>
      <div className="space-y-5">
        <MetricRow label="CPU" value={metrics.cpu_pct} />
        <MetricRow label="Memory" value={metrics.mem_used_pct} />
        <MetricRow label="Disk" value={metrics.disk_used_pct} />
        <div className="pt-4 border-t border-border/40">
          <div className="flex justify-between items-center">
            <span className="text-muted-foreground text-sm">Uptime</span>
            <span className="text-foreground font-medium tracking-wide">
              {formatUptime(metrics.uptime_s)}
            </span>
          </div>
        </div>
      </div>
      <div className="mt-6 pt-4 border-t border-border/40">
        <ConnectionStatus type={connectionType} lastUpdate={lastUpdate} />
      </div>
    </div>
  );
}

interface MetricRowProps {
  readonly label: string;
  readonly value: number;
}

function MetricRow({ label, value }: MetricRowProps) {
  const getColor = (val: number) => {
    if (val >= 80) return "bg-red-500";
    if (val >= 60) return "bg-yellow-500";
    return "bg-green-500";
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-2">
        <span className="text-muted-foreground text-sm tracking-wide uppercase">{label}</span>
        <span className="text-foreground font-semibold">{value.toFixed(1)}%</span>
      </div>
      <div className="w-full bg-border/40 rounded-full h-2 overflow-hidden">
        <div
          className={`h-full ${getColor(value)} transition-all duration-500 ease-out`}
          style={{ width: `${Math.min(value, 100)}%` }}
        />
      </div>
    </div>
  );
}
