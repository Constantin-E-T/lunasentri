'use client';

import { useSystemInfo } from '@/lib/useSystemInfo';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Server, Cpu, HardDrive, Clock, Monitor, CircuitBoard } from 'lucide-react';

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / (24 * 3600));
  const hours = Math.floor((seconds % (24 * 3600)) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (days > 0) {
    return `${days}d ${hours}h ${minutes}m`;
  } else if (hours > 0) {
    return `${hours}h ${minutes}m`;
  } else {
    return `${minutes}m`;
  }
}

function formatBytes(bytes: number, unit: 'MB' | 'GB'): string {
  if (unit === 'GB') {
    return `${bytes} GB`;
  }
  // Convert MB to GB if > 1024 MB
  if (bytes >= 1024) {
    return `${(bytes / 1024).toFixed(1)} GB`;
  }
  return `${bytes} MB`;
}

function formatTimestamp(timestamp: number): string {
  const date = new Date(timestamp * 1000);
  return date.toLocaleString();
}

export function SystemInfoCard() {
  const { systemInfo, loading, error } = useSystemInfo();

  if (loading) {
    return (
      <Card className="bg-card/70 border border-border/30 backdrop-blur-xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-foreground">
            <Server className="h-5 w-5" />
            System Information
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="animate-pulse space-y-3">
            <div className="h-4 bg-muted/50 rounded w-3/4"></div>
            <div className="h-4 bg-muted/50 rounded w-1/2"></div>
            <div className="h-4 bg-muted/50 rounded w-2/3"></div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="bg-card/70 border border-border/30 backdrop-blur-xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-foreground">
            <Server className="h-5 w-5" />
            System Information
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-destructive text-sm">
            {error}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!systemInfo) {
    return null;
  }

  return (
    <Card className="bg-card/70 border border-border/30 backdrop-blur-xl">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-foreground">
          <Server className="h-5 w-5" />
          System Information
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Hostname & Platform */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Monitor className="h-4 w-4 text-primary" />
              <span className="text-sm text-muted-foreground">Hostname:</span>
              <span className="text-sm font-medium">{systemInfo.hostname}</span>
            </div>
            <div className="flex items-center gap-2">
              <CircuitBoard className="h-4 w-4 text-primary" />
              <span className="text-sm text-muted-foreground">Platform:</span>
              <span className="text-sm font-medium">
                {systemInfo.platform} {systemInfo.platform_version}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-primary" />
              <span className="text-sm text-muted-foreground">Uptime:</span>
              <span className="text-sm font-medium">
                {formatUptime(systemInfo.uptime_s)}
              </span>
            </div>
          </div>

          {/* Hardware Info */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Cpu className="h-4 w-4 text-primary" />
              <span className="text-sm text-muted-foreground">CPU Cores:</span>
              <span className="text-sm font-medium">{systemInfo.cpu_cores}</span>
            </div>
            <div className="flex items-center gap-2">
              <Server className="h-4 w-4 text-primary" />
              <span className="text-sm text-muted-foreground">Memory:</span>
              <span className="text-sm font-medium">
                {formatBytes(systemInfo.memory_total_mb, 'MB')}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <HardDrive className="h-4 w-4 text-primary" />
              <span className="text-sm text-muted-foreground">Disk:</span>
              <span className="text-sm font-medium">
                {formatBytes(systemInfo.disk_total_gb, 'GB')}
              </span>
            </div>
          </div>
        </div>

        {/* Last Boot Time */}
        <div className="pt-2 border-t border-border/30">
          <div className="flex items-center gap-2">
            <Clock className="h-4 w-4 text-primary" />
            <span className="text-sm text-muted-foreground">Last Boot:</span>
            <span className="text-sm font-medium">
              {formatTimestamp(systemInfo.last_boot_time)}
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}