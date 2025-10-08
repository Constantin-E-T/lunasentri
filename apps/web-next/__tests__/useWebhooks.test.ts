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

    it('should fetch webhooks on mount', async () => {
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

        expect(result.current.webhooks).toEqual(mockWebhooks);
        expect(result.current.error).toBeNull();
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
});
