export interface Metrics {
  cpu_pct: number;
  mem_used_pct: number;
  disk_used_pct: number;
  uptime_s: number;
}

export interface SystemInfo {
  hostname: string;
  platform: string;
  platform_version: string;
  kernel_version: string;
  uptime_s: number;
  cpu_cores: number;
  memory_total_mb: number;
  disk_total_gb: number;
  last_boot_time: number;
}

export interface User {
  id: number;
  email: string;
  is_admin: boolean;
  created_at?: string;
}

export interface AlertRule {
  id: number;
  name: string;
  metric: 'cpu_pct' | 'mem_used_pct' | 'disk_used_pct';
  threshold_pct: number;
  comparison: 'above' | 'below';
  trigger_after: number;
  created_at: string;
  updated_at: string;
}

export interface AlertEvent {
  id: number;
  rule_id: number;
  triggered_at: string;
  value: number;
  acknowledged: boolean;
  acknowledged_at?: string;
}

export interface CreateAlertRuleRequest {
  name: string;
  metric: 'cpu_pct' | 'mem_used_pct' | 'disk_used_pct';
  threshold_pct: number;
  comparison: 'above' | 'below';
  trigger_after: number;
}

export interface CreateUserRequest {
  email: string;
  password?: string;
}

export interface CreateUserResponse {
  id: number;
  email: string;
  is_admin: boolean;
  created_at: string;
  temp_password?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export async function fetchMetrics(): Promise<Metrics> {
  const response = await fetch(`${API_URL}/metrics`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Include cookies for authentication
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch metrics: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

export async function fetchSystemInfo(): Promise<SystemInfo> {
  const response = await fetch(`${API_URL}/system/info`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    // Note: endpoint is public, but keeping credentials for consistency
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch system info: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

export async function login(email: string, password: string): Promise<User> {
  const response = await fetch(`${API_URL}/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Important: allows cookies to be set
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new Error('Invalid email or password');
    }
    throw new Error(`Login failed: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

export async function logout(): Promise<void> {
  const response = await fetch(`${API_URL}/auth/logout`, {
    method: 'POST',
    credentials: 'include', // Include cookies for authentication
  });

  if (!response.ok && response.status !== 204) {
    throw new Error(`Logout failed: ${response.status} ${response.statusText}`);
  }
}

export async function register(email: string, password: string): Promise<CreateUserResponse> {
  const response = await fetch(`${API_URL}/auth/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Include cookies for authentication
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => null);
    const message = errorData?.error || `Registration failed: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  return response.json();
}

export async function fetchCurrentUser(): Promise<User> {
  const response = await fetch(`${API_URL}/auth/me`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Include cookies for authentication
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch current user: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

export async function listUsers(): Promise<User[]> {
  const response = await fetch(`${API_URL}/auth/users`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Include cookies for authentication
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch users: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

export async function createUser(data: CreateUserRequest): Promise<CreateUserResponse> {
  const response = await fetch(`${API_URL}/auth/users`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Include cookies for authentication
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => null);
    const message = errorData?.error || `Failed to create user: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  return response.json();
}

export async function deleteUser(id: number): Promise<void> {
  const response = await fetch(`${API_URL}/auth/users/${id}`, {
    method: 'DELETE',
    credentials: 'include', // Include cookies for authentication
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => null);
    const message = errorData?.error || `Failed to delete user: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }
}

export async function changePassword(currentPassword: string, newPassword: string): Promise<void> {
  const response = await fetch(`${API_URL}/auth/change-password`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Include cookies for authentication
    body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => null);
    let message: string;

    if (response.status === 401) {
      message = errorData?.error || 'Current password is incorrect';
    } else if (response.status === 400) {
      message = errorData?.error || 'New password does not meet requirements';
    } else {
      message = errorData?.error || `Failed to change password: ${response.status} ${response.statusText}`;
    }

    throw new Error(message);
  }
}

// Alert Rules API

export async function listAlertRules(): Promise<AlertRule[]> {
  const response = await fetch(`${API_URL}/alerts/rules`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Include cookies for authentication
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch alert rules: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

export async function createAlertRule(rule: CreateAlertRuleRequest): Promise<AlertRule> {
  const response = await fetch(`${API_URL}/alerts/rules`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Include cookies for authentication
    body: JSON.stringify(rule),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => null);
    const message = errorData?.error || `Failed to create alert rule: ${response.status} ${response.statusText}`;
    throw new Error(message);
  }

  return response.json();
}

export async function updateAlertRule(id: number, rule: CreateAlertRuleRequest): Promise<AlertRule> {
  const response = await fetch(`${API_URL}/alerts/rules/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Include cookies for authentication
    body: JSON.stringify(rule),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => null);
    let message: string;

    if (response.status === 404) {
      message = 'Alert rule not found';
    } else {
      message = errorData?.error || `Failed to update alert rule: ${response.status} ${response.statusText}`;
    }

    throw new Error(message);
  }

  return response.json();
}

export async function deleteAlertRule(id: number): Promise<void> {
  const response = await fetch(`${API_URL}/alerts/rules/${id}`, {
    method: 'DELETE',
    credentials: 'include', // Include cookies for authentication
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => null);
    let message: string;

    if (response.status === 404) {
      message = 'Alert rule not found';
    } else {
      message = errorData?.error || `Failed to delete alert rule: ${response.status} ${response.statusText}`;
    }

    throw new Error(message);
  }
}

// Alert Events API

export async function listAlertEvents(limit?: number): Promise<AlertEvent[]> {
  const url = new URL(`${API_URL}/alerts/events`);
  if (limit) {
    url.searchParams.set('limit', limit.toString());
  }

  const response = await fetch(url.toString(), {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include', // Include cookies for authentication
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch alert events: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

export async function ackAlertEvent(id: number): Promise<void> {
  const response = await fetch(`${API_URL}/alerts/events/${id}/ack`, {
    method: 'POST',
    credentials: 'include', // Include cookies for authentication
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => null);
    let message: string;

    if (response.status === 404) {
      message = 'Alert event not found or already acknowledged';
    } else {
      message = errorData?.error || `Failed to acknowledge alert event: ${response.status} ${response.statusText}`;
    }

    throw new Error(message);
  }
}
