export interface Metrics {
  cpu_pct: number;
  mem_used_pct: number;
  disk_used_pct: number;
  uptime_s: number;
}

export interface User {
  id: number;
  email: string;
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
