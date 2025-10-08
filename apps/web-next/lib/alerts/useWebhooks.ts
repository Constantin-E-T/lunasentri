'use client';

import { useEffect, useState, useCallback } from 'react';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// TypeScript types matching API responses
export interface Webhook {
    id: number;
    url: string;
    is_active: boolean;
    failure_count: number;
    last_success_at: string | null;
    last_error_at: string | null;
    secret_last_four: string;
    created_at: string;
    updated_at: string;
}

export interface CreateWebhookRequest {
    url: string;
    secret: string;
    is_active: boolean;
}

export interface UpdateWebhookRequest {
    url?: string;
    secret?: string;
    is_active?: boolean;
}

/**
 * Centralized request helper for webhook endpoints
 */
async function webhookRequest<T>(
    input: RequestInfo | URL,
    init?: RequestInit
): Promise<T> {
    const response = await fetch(input, {
        ...init,
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json',
            ...init?.headers,
        },
    });

    // Handle authentication errors
    if (response.status === 401 || response.status === 403) {
        if (typeof window !== 'undefined') {
            window.dispatchEvent(new CustomEvent('session-expired'));
        }
        throw new Error('Session expired');
    }

    if (!response.ok) {
        // Try to parse error message from response
        try {
            const errorData = await response.json();
            throw new Error(errorData.error || `Request failed: ${response.status}`);
        } catch {
            throw new Error(`Request failed: ${response.status} ${response.statusText}`);
        }
    }

    // Handle 204 No Content responses
    if (response.status === 204) {
        return undefined as T;
    }

    return response.json();
}

// API functions
export async function listWebhooks(): Promise<Webhook[]> {
    return webhookRequest<Webhook[]>(`${API_URL}/notifications/webhooks`);
}

export async function createWebhook(
    payload: CreateWebhookRequest
): Promise<Webhook> {
    return webhookRequest<Webhook>(`${API_URL}/notifications/webhooks`, {
        method: 'POST',
        body: JSON.stringify(payload),
    });
}

export async function updateWebhook(
    id: number,
    payload: UpdateWebhookRequest
): Promise<Webhook> {
    return webhookRequest<Webhook>(`${API_URL}/notifications/webhooks/${id}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
    });
}

export async function deleteWebhook(id: number): Promise<void> {
    return webhookRequest<void>(`${API_URL}/notifications/webhooks/${id}`, {
        method: 'DELETE',
    });
}

export async function sendTestWebhook(id: number): Promise<void> {
    return webhookRequest<void>(
        `${API_URL}/notifications/webhooks/${id}/test`,
        {
            method: 'POST',
        }
    );
}

// Custom hook for managing webhooks
export interface UseWebhooksReturn {
    webhooks: Webhook[];
    loading: boolean;
    error: string | null;
    createWebhook: (payload: CreateWebhookRequest) => Promise<void>;
    updateWebhook: (id: number, payload: UpdateWebhookRequest) => Promise<void>;
    deleteWebhook: (id: number) => Promise<void>;
    sendTestWebhook: (id: number) => Promise<void>;
    refresh: () => Promise<void>;
}

export function useWebhooks(): UseWebhooksReturn {
    const [webhooks, setWebhooks] = useState<Webhook[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // Fetch webhooks
    const refresh = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);
            const data = await listWebhooks();
            setWebhooks(data);
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to fetch webhooks';
            setError(message);
            // Don't re-throw on initial load to prevent unhandled promise rejection
        } finally {
            setLoading(false);
        }
    }, []);

    // Create webhook
    const create = useCallback(
        async (payload: CreateWebhookRequest) => {
            try {
                await createWebhook(payload);
                await refresh();
            } catch (err) {
                const message = err instanceof Error ? err.message : 'Failed to create webhook';
                setError(message);
                throw err;
            }
        },
        [refresh]
    );

    // Update webhook
    const update = useCallback(
        async (id: number, payload: UpdateWebhookRequest) => {
            try {
                await updateWebhook(id, payload);
                await refresh();
            } catch (err) {
                const message = err instanceof Error ? err.message : 'Failed to update webhook';
                setError(message);
                throw err;
            }
        },
        [refresh]
    );

    // Delete webhook
    const remove = useCallback(
        async (id: number) => {
            try {
                await deleteWebhook(id);
                await refresh();
            } catch (err) {
                const message = err instanceof Error ? err.message : 'Failed to delete webhook';
                setError(message);
                throw err;
            }
        },
        [refresh]
    );

    // Send test webhook
    const sendTest = useCallback(async (id: number) => {
        try {
            await sendTestWebhook(id);
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to send test webhook';
            setError(message);
            throw err;
        }
    }, []);

    // Load webhooks on mount
    useEffect(() => {
        refresh();
    }, [refresh]);

    return {
        webhooks,
        loading,
        error,
        createWebhook: create,
        updateWebhook: update,
        deleteWebhook: remove,
        sendTestWebhook: sendTest,
        refresh,
    };
}
