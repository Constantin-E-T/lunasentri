/**
 * @jest-environment jsdom
 */

import { renderHook, act, waitFor } from '@testing-library/react';
import { useSession } from '@/lib/useSession';

// Mock the API functions
jest.mock('@/lib/api', () => ({
    fetchCurrentUser: jest.fn(),
    login: jest.fn(),
    logout: jest.fn(),
    register: jest.fn(),
    changePassword: jest.fn(),
}));

// Mock useToast
const mockToast = jest.fn();
jest.mock('@/components/ui/use-toast', () => ({
    useToast: () => ({
        toast: mockToast,
    }),
}));

describe('useSession', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        // Clear any existing event listeners
        if (typeof window !== 'undefined') {
            const listeners = (window as any)._listeners;
            if (listeners && listeners['session-expired']) {
                listeners['session-expired'].forEach((listener: any) => {
                    window.removeEventListener('session-expired', listener);
                });
            }
        }
    });

    test('returns the expected state shape on mount', () => {
        const { result } = renderHook(() => useSession());

        expect(result.current).toHaveProperty('user');
        expect(result.current).toHaveProperty('status');
        expect(result.current).toHaveProperty('login');
        expect(result.current).toHaveProperty('register');
        expect(result.current).toHaveProperty('changePassword');
        expect(result.current).toHaveProperty('logout');

        // Initial state should be loading
        expect(result.current.status).toBe('loading');
        expect(result.current.user).toBeNull();

        // Functions should be callable without throwing
        expect(typeof result.current.login).toBe('function');
        expect(typeof result.current.register).toBe('function');
        expect(typeof result.current.changePassword).toBe('function');
        expect(typeof result.current.logout).toBe('function');
    });

    test('exposes login, register, changePassword, and logout functions', () => {
        const { result } = renderHook(() => useSession());

        // All session functions should be present
        expect(result.current.login).toBeDefined();
        expect(result.current.register).toBeDefined();
        expect(result.current.changePassword).toBeDefined();
        expect(result.current.logout).toBeDefined();

        // They should be functions
        expect(typeof result.current.login).toBe('function');
        expect(typeof result.current.register).toBe('function');
        expect(typeof result.current.changePassword).toBe('function');
        expect(typeof result.current.logout).toBe('function');
    });

    describe('session expiry handling', () => {
        const { logout: mockLogout } = jest.requireMock('@/lib/api');

        beforeEach(() => {
            mockLogout.mockReset();
        });

        test('handles session-expired event by clearing state and showing toast', async () => {
            const { result } = renderHook(() => useSession());

            // Set initial authenticated state
            act(() => {
                // Simulate being authenticated
                (result.current as any).user = { id: 1, email: 'test@example.com', is_admin: false };
                (result.current as any).status = 'authenticated';
            });

            // Simulate session expiry event
            act(() => {
                window.dispatchEvent(new CustomEvent('session-expired'));
            });

            await waitFor(() => {
                // Assert user state is cleared
                expect(result.current.user).toBeNull();
                expect(result.current.status).toBe('unauthenticated');
            });

            // Assert toast was called with correct parameters
            expect(mockToast).toHaveBeenCalledWith({
                title: "Session expired",
                description: "Please log in again.",
                variant: "destructive",
            });

            // Assert best-effort logout was called
            expect(mockLogout).toHaveBeenCalledTimes(1);
        });

        test('handles session-expired event even when logout fails', async () => {
            const { result } = renderHook(() => useSession());

            // Mock logout to reject
            mockLogout.mockRejectedValue(new Error('Network error'));

            // Set initial authenticated state
            act(() => {
                (result.current as any).user = { id: 1, email: 'test@example.com', is_admin: false };
                (result.current as any).status = 'authenticated';
            });

            // Simulate session expiry event
            act(() => {
                window.dispatchEvent(new CustomEvent('session-expired'));
            });

            await waitFor(() => {
                // Assert user state is still cleared despite logout failure
                expect(result.current.user).toBeNull();
                expect(result.current.status).toBe('unauthenticated');
            });

            // Assert toast was still called
            expect(mockToast).toHaveBeenCalledWith({
                title: "Session expired",
                description: "Please log in again.",
                variant: "destructive",
            });

            // Assert logout was attempted
            expect(mockLogout).toHaveBeenCalledTimes(1);
        });

        test('cleans up event listener on unmount', () => {
            const addEventListenerSpy = jest.spyOn(window, 'addEventListener');
            const removeEventListenerSpy = jest.spyOn(window, 'removeEventListener');

            const { unmount } = renderHook(() => useSession());

            // Verify addEventListener was called
            expect(addEventListenerSpy).toHaveBeenCalledWith('session-expired', expect.any(Function));

            // Unmount the hook
            unmount();

            // Verify removeEventListener was called
            expect(removeEventListenerSpy).toHaveBeenCalledWith('session-expired', expect.any(Function));

            addEventListenerSpy.mockRestore();
            removeEventListenerSpy.mockRestore();
        });
    });
});