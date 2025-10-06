export interface Metrics {
  cpu_pct: number;
  mem_used_pct: number;
  disk_used_pct: number;
  uptime_s: number;
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export async function fetchMetrics(): Promise<Metrics> {
  const response = await fetch(`${API_URL}/metrics`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch metrics: ${response.status} ${response.statusText}`);
  }

  return response.json();
}
