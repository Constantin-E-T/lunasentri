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

/**
 * Centralized request helper that handles authentication errors.
 * Dispatches a 'session-expired' event on 401/403 responses.
 */
async function request<T>(input: RequestInfo | URL, init?: RequestInit): Promise<T> {
  const response = await fetch(input, {
    ...init,
    credentials: 'include', // Always include cookies for authentication
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });

  // Handle authentication errors
  if (response.status === 401 || response.status === 403) {
    // Dispatch session expired event for useSession to handle
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('session-expired'));
    }
    throw new Error('Session expired');
  }

  if (!response.ok) {
    throw new Error(`Request failed: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

/**
 * Request helper for endpoints that don't return JSON content.
 */
async function requestVoid(input: RequestInfo | URL, init?: RequestInit): Promise<void> {
  const response = await fetch(input, {
    ...init,
    credentials: 'include', // Always include cookies for authentication
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });

  // Handle authentication errors
  if (response.status === 401 || response.status === 403) {
    // Dispatch session expired event for useSession to handle
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('session-expired'));
    }
    throw new Error('Session expired');
  }

  if (!response.ok) {
    throw new Error(`Request failed: ${response.status} ${response.statusText}`);
  }
}

export async function fetchMetrics(): Promise<Metrics> {
  return request<Metrics>(`${API_URL}/metrics`);
}

export async function fetchSystemInfo(): Promise<SystemInfo> {
  return request<SystemInfo>(`${API_URL}/system/info`);
}

export async function login(email: string, password: string): Promise<User> {
  try {
    return await request<User>(`${API_URL}/auth/login`, {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  } catch (error) {
    // Handle specific login errors
    if (error instanceof Error && error.message === 'Request failed: 401 Unauthorized') {
      throw new Error('Invalid email or password');
    }
    throw error;
  }
}

export async function logout(): Promise<void> {
  // Don't use request helper for logout - it shouldn't trigger session expiry
  const response = await fetch(`${API_URL}/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  });

  if (!response.ok && response.status !== 204) {
    throw new Error(`Logout failed: ${response.status} ${response.statusText}`);
  }
}

export async function register(email: string, password: string): Promise<CreateUserResponse> {
  return request<CreateUserResponse>(`${API_URL}/auth/register`, {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
}

export async function fetchCurrentUser(): Promise<User> {
  return request<User>(`${API_URL}/auth/me`);
}

export async function listUsers(): Promise<User[]> {
  return request<User[]>(`${API_URL}/auth/users`);
}

export async function createUser(data: CreateUserRequest): Promise<CreateUserResponse> {
  return request<CreateUserResponse>(`${API_URL}/auth/users`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function deleteUser(id: number): Promise<void> {
  return requestVoid(`${API_URL}/auth/users/${id}`, {
    method: 'DELETE',
  });
}

export async function changePassword(currentPassword: string, newPassword: string): Promise<void> {
  return requestVoid(`${API_URL}/auth/change-password`, {
    method: 'POST',
    body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
  });
}

// Alert Rules API

export async function listAlertRules(): Promise<AlertRule[]> {
  return request<AlertRule[]>(`${API_URL}/alerts/rules`);
}

export async function createAlertRule(rule: CreateAlertRuleRequest): Promise<AlertRule> {
  return request<AlertRule>(`${API_URL}/alerts/rules`, {
    method: 'POST',
    body: JSON.stringify(rule),
  });
}

export async function updateAlertRule(id: number, rule: CreateAlertRuleRequest): Promise<AlertRule> {
  return request<AlertRule>(`${API_URL}/alerts/rules/${id}`, {
    method: 'PUT',
    body: JSON.stringify(rule),
  });
}

export async function deleteAlertRule(id: number): Promise<void> {
  return requestVoid(`${API_URL}/alerts/rules/${id}`, {
    method: 'DELETE',
  });
}

// Alert Events API

export async function listAlertEvents(limit?: number): Promise<AlertEvent[]> {
  const url = new URL(`${API_URL}/alerts/events`);
  if (limit) {
    url.searchParams.set('limit', limit.toString());
  }

  return request<AlertEvent[]>(url.toString());
}

export async function ackAlertEvent(id: number): Promise<void> {
  return requestVoid(`${API_URL}/alerts/events/${id}/ack`, {
    method: 'POST',
  });
}

// Email Notifications API

export interface EmailRecipient {
  id: number;
  user_id: number;
  email: string;
  is_active: boolean;
  created_at: string;
  last_attempt_at?: string;
  last_success_at?: string;
  last_error_at?: string;
  failure_count: number;
  cooldown_until?: string;
}

export interface CreateEmailRecipientRequest {
  email: string;
}

export interface UpdateEmailRecipientRequest {
  email?: string;
  is_active?: boolean;
}

export interface TestEmailResponse {
  status: string;
  recipient_id: number;
  triggered_at: string;
}

export async function listEmailRecipients(): Promise<EmailRecipient[]> {
  return request<EmailRecipient[]>(`${API_URL}/notifications/emails`);
}

export async function createEmailRecipient(recipient: CreateEmailRecipientRequest): Promise<EmailRecipient> {
  return request<EmailRecipient>(`${API_URL}/notifications/emails`, {
    method: 'POST',
    body: JSON.stringify(recipient),
  });
}

export async function updateEmailRecipient(id: number, recipient: UpdateEmailRecipientRequest): Promise<EmailRecipient> {
  return request<EmailRecipient>(`${API_URL}/notifications/emails/${id}`, {
    method: 'PUT',
    body: JSON.stringify(recipient),
  });
}

export async function deleteEmailRecipient(id: number): Promise<void> {
  return requestVoid(`${API_URL}/notifications/emails/${id}`, {
    method: 'DELETE',
  });
}

export async function testEmailRecipient(id: number): Promise<TestEmailResponse> {
  return request<TestEmailResponse>(`${API_URL}/notifications/emails/${id}/test`, {
    method: 'POST',
  });
}

// Telegram Notifications API

export interface TelegramRecipient {
  id: number;
  chat_id: string;
  is_active: boolean;
  created_at: string;
  last_success_at?: string;
  failure_count: number;
}

export interface CreateTelegramRecipientRequest {
  chat_id: string;
}

export interface UpdateTelegramRecipientRequest {
  chat_id?: string;
  is_active?: boolean;
}

export interface TestTelegramResponse {
  status: string;
  recipient_id: number;
  triggered_at: string;
}

export async function listTelegramRecipients(): Promise<TelegramRecipient[]> {
  return request<TelegramRecipient[]>(`${API_URL}/notifications/telegram`);
}

export async function createTelegramRecipient(recipient: CreateTelegramRecipientRequest): Promise<TelegramRecipient> {
  return request<TelegramRecipient>(`${API_URL}/notifications/telegram`, {
    method: 'POST',
    body: JSON.stringify(recipient),
  });
}

export async function updateTelegramRecipient(id: number, recipient: UpdateTelegramRecipientRequest): Promise<TelegramRecipient> {
  return request<TelegramRecipient>(`${API_URL}/notifications/telegram/${id}`, {
    method: 'PUT',
    body: JSON.stringify(recipient),
  });
}

export async function deleteTelegramRecipient(id: number): Promise<void> {
  return requestVoid(`${API_URL}/notifications/telegram/${id}`, {
    method: 'DELETE',
  });
}

export async function testTelegramRecipient(id: number): Promise<TestTelegramResponse> {
  return request<TestTelegramResponse>(`${API_URL}/notifications/telegram/${id}/test`, {
    method: 'POST',
  });
}

