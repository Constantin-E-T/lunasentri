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
      <div className="bg-slate-700/50 backdrop-blur-sm rounded-lg p-6 max-w-md mx-auto">
        <div className="flex items-center justify-center">
          <div className="animate-pulse text-slate-400">Loading metrics...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-slate-700/50 backdrop-blur-sm rounded-lg p-6 max-w-md mx-auto border border-red-500/30">
        <div className="text-center space-y-3">
          <p className="text-red-400 text-sm mb-2">⚠️ Connection Error</p>
          <p className="text-slate-400 text-xs">{error}</p>
          <button
            onClick={retry}
            className="px-3 py-1 bg-red-600/20 hover:bg-red-600/30 border border-red-500/30 rounded text-red-400 text-xs transition-colors"
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
    <div className="bg-slate-700/50 backdrop-blur-sm rounded-lg p-6 max-w-md mx-auto">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold text-white">Live Metrics</h2>
        {connectionType === "websocket" && (
          <div
            className="w-2 h-2 bg-green-400 rounded-full animate-pulse"
            title="Live WebSocket connection"
          />
        )}
      </div>
      <div className="space-y-4">
        <MetricRow label="CPU" value={metrics.cpu_pct} />
        <MetricRow label="Memory" value={metrics.mem_used_pct} />
        <MetricRow label="Disk" value={metrics.disk_used_pct} />
        <div className="pt-2 border-t border-slate-600">
          <div className="flex justify-between items-center">
            <span className="text-slate-400 text-sm">Uptime</span>
            <span className="text-white font-medium">
              {formatUptime(metrics.uptime_s)}
            </span>
          </div>
        </div>
      </div>
      <div className="mt-4 pt-3 border-t border-slate-600">
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
      <div className="flex justify-between items-center mb-1">
        <span className="text-slate-400 text-sm">{label}</span>
        <span className="text-white font-medium">{value.toFixed(1)}%</span>
      </div>
      <div className="w-full bg-slate-800 rounded-full h-2 overflow-hidden">
        <div
          className={`h-full ${getColor(value)} transition-all duration-300`}
          style={{ width: `${Math.min(value, 100)}%` }}
        />
      </div>
    </div>
  );
}
