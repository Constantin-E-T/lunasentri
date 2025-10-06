/**
 * @jest-environment jsdom
 */

import { renderHook } from '@testing-library/react';
import { useSession } from '@/lib/useSession';

// Mock the API functions
jest.mock('@/lib/api', () => ({
    fetchCurrentUser: jest.fn(),
    login: jest.fn(),
    logout: jest.fn(),
    register: jest.fn(),
}));

describe('useSession', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    test('returns the expected state shape on mount', () => {
        const { result } = renderHook(() => useSession());

        expect(result.current).toHaveProperty('user');
        expect(result.current).toHaveProperty('status');
        expect(result.current).toHaveProperty('login');
        expect(result.current).toHaveProperty('register');
        expect(result.current).toHaveProperty('logout');

        // Initial state should be loading
        expect(result.current.status).toBe('loading');
        expect(result.current.user).toBeNull();

        // Functions should be callable without throwing
        expect(typeof result.current.login).toBe('function');
        expect(typeof result.current.register).toBe('function');
        expect(typeof result.current.logout).toBe('function');
    });

    test('exposes login, register, and logout functions', () => {
        const { result } = renderHook(() => useSession());

        // All session functions should be present
        expect(result.current.login).toBeDefined();
        expect(result.current.register).toBeDefined();
        expect(result.current.logout).toBeDefined();

        // They should be functions
        expect(typeof result.current.login).toBe('function');
        expect(typeof result.current.register).toBe('function');
        expect(typeof result.current.logout).toBe('function');
    });
});