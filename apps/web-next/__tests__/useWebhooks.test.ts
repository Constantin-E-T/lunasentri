/**
 * @jest-environment jsdom
 */
import { renderHook, waitFor } from '@testing-library/react';
import { useWebhooks } from '@/lib/alerts/useWebhooks';

// Mock fetch globally
global.fetch = jest.fn();

// Suppress console errors in tests
beforeAll(() => {
    jest.spyOn(console, 'error').mockImplementation();
});

afterAll(() => {
    (console.error as jest.Mock).mockRestore();
});

describe('useWebhooks', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('should fetch webhooks on mount with cooldown fields', async () => {
        const mockWebhooks = [
            {
                id: 1,
                url: 'https://example.com/webhook',
                is_active: true,
                failure_count: 0,
                last_success_at: null,
                last_error_at: null,
                secret_last_four: '2345',
                created_at: '2025-01-01T00:00:00Z',
                updated_at: '2025-01-01T00:00:00Z',
                cooldown_until: null,
                last_attempt_at: null,
            },
        ];

        (global.fetch as jest.Mock).mockResolvedValueOnce({
            ok: true,
            status: 200,
            json: async () => mockWebhooks,
        });

        const { result } = renderHook(() => useWebhooks());

        expect(result.current.loading).toBe(true);

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        expect(result.current.webhooks).toHaveLength(1);
        expect(result.current.webhooks[0].isCoolingDown).toBe(false);
        expect(result.current.webhooks[0].canSendTest).toBe(true);
        expect(result.current.error).toBeNull();
    });

    it('should parse cooldown state correctly', async () => {
        const futureTime = new Date(Date.now() + 60000).toISOString(); // 1 minute in future
        const recentAttempt = new Date(Date.now() - 10000).toISOString(); // 10 seconds ago

        const mockWebhooks = [
            {
                id: 1,
                url: 'https://example.com/webhook',
                is_active: true,
                failure_count: 3,
                last_success_at: null,
                last_error_at: recentAttempt,
                secret_last_four: '2345',
                created_at: '2025-01-01T00:00:00Z',
                updated_at: '2025-01-01T00:00:00Z',
                cooldown_until: futureTime,
                last_attempt_at: recentAttempt,
            },
        ];

        (global.fetch as jest.Mock).mockResolvedValueOnce({
            ok: true,
            status: 200,
            json: async () => mockWebhooks,
        });

        const { result } = renderHook(() => useWebhooks());

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        expect(result.current.webhooks[0].isCoolingDown).toBe(true);
        expect(result.current.webhooks[0].canSendTest).toBe(false);
        expect(result.current.webhooks[0].retryAfterSeconds).toBeGreaterThan(0);
    });

    it('should handle 429 rate limit response', async () => {
        (global.fetch as jest.Mock).mockResolvedValueOnce({
            ok: false,
            status: 429,
            json: async () => ({ error: 'Rate limit active, can retry in 25s' }),
        });

        const { result } = renderHook(() => useWebhooks());

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        expect(result.current.error).toContain('Rate limit');
    });

    it('should set error state on fetch failure', async () => {
        (global.fetch as jest.Mock).mockResolvedValueOnce({
            ok: false,
            status: 500,
            statusText: 'Internal Server Error',
            json: jest.fn().mockRejectedValue(new Error('Failed to parse JSON')),
        });

        const { result } = renderHook(() => useWebhooks());

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        expect(result.current.error).toBeTruthy();
        expect(result.current.webhooks).toEqual([]);
    });

    it('should dispatch session-expired event on 401', async () => {
        const dispatchEventSpy = jest.spyOn(window, 'dispatchEvent');

        (global.fetch as jest.Mock).mockResolvedValueOnce({
            ok: false,
            status: 401,
            statusText: 'Unauthorized',
        });

        renderHook(() => useWebhooks());

        await waitFor(() => {
            expect(dispatchEventSpy).toHaveBeenCalledWith(
                expect.objectContaining({ type: 'session-expired' })
            );
        });

        dispatchEventSpy.mockRestore();
    });

    it('should refresh after sending test webhook', async () => {
        const mockWebhooks = [
            {
                id: 1,
                url: 'https://example.com/webhook',
                is_active: true,
                failure_count: 0,
                last_success_at: null,
                last_error_at: null,
                secret_last_four: '2345',
                created_at: '2025-01-01T00:00:00Z',
                updated_at: '2025-01-01T00:00:00Z',
                cooldown_until: null,
                last_attempt_at: null,
            },
        ];

        // Initial fetch
        (global.fetch as jest.Mock).mockResolvedValueOnce({
            ok: true,
            status: 200,
            json: async () => mockWebhooks,
        });

        const { result } = renderHook(() => useWebhooks());

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        const updatedWebhook = {
            ...mockWebhooks[0],
            last_attempt_at: new Date().toISOString(),
        };

        // Test webhook call
        (global.fetch as jest.Mock).mockResolvedValueOnce({
            ok: true,
            status: 200,
            json: async () => ({ status: 'sent' }),
        });

        // Refresh call after test
        (global.fetch as jest.Mock).mockResolvedValueOnce({
            ok: true,
            status: 200,
            json: async () => [updatedWebhook],
        });

        await result.current.sendTestWebhook(1);

        await waitFor(() => {
            expect(result.current.webhooks[0].last_attempt_at).toBeTruthy();
        });
    });
});
