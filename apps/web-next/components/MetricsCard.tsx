'use client';

import { useEffect, useState } from 'react';
import { fetchMetrics, type Metrics } from '@/lib/api';

const POLL_INTERVAL_MS = 5000; // 5 seconds

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

export function MetricsCard() {
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let isMounted = true;
    let timeoutId: NodeJS.Timeout;

    const loadMetrics = async () => {
      try {
        const data = await fetchMetrics();
        if (isMounted) {
          setMetrics(data);
          setError(null);
          setLoading(false);
        }
      } catch (err) {
        if (isMounted) {
          setError(err instanceof Error ? err.message : 'Failed to load metrics');
          setLoading(false);
        }
      } finally {
        if (isMounted) {
          timeoutId = setTimeout(loadMetrics, POLL_INTERVAL_MS);
        }
      }
    };

    loadMetrics();

    return () => {
      isMounted = false;
      clearTimeout(timeoutId);
    };
  }, []);

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
        <div className="text-center">
          <p className="text-red-400 text-sm mb-2">⚠️ Error</p>
          <p className="text-slate-400 text-xs">{error}</p>
        </div>
      </div>
    );
  }

  if (!metrics) {
    return null;
  }

  return (
    <div className="bg-slate-700/50 backdrop-blur-sm rounded-lg p-6 max-w-md mx-auto">
      <h2 className="text-xl font-semibold text-white mb-4">Live Metrics</h2>
      <div className="space-y-4">
        <MetricRow label="CPU" value={metrics.cpu_pct} />
        <MetricRow label="Memory" value={metrics.mem_used_pct} />
        <MetricRow label="Disk" value={metrics.disk_used_pct} />
        <div className="pt-2 border-t border-slate-600">
          <div className="flex justify-between items-center">
            <span className="text-slate-400 text-sm">Uptime</span>
            <span className="text-white font-medium">{formatUptime(metrics.uptime_s)}</span>
          </div>
        </div>
      </div>
      <div className="mt-4 text-center">
        <span className="text-xs text-slate-500">Updates every 5s</span>
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
    if (val >= 80) return 'bg-red-500';
    if (val >= 60) return 'bg-yellow-500';
    return 'bg-green-500';
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
