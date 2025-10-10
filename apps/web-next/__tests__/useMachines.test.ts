/**
 * @jest-environment jsdom
 */

import { renderHook, act, waitFor } from '@testing-library/react';
import { useMachines } from '../lib/useMachines';
import * as api from '../lib/api';

// Mock the API module
jest.mock('../lib/api');

const mockApi = api as jest.Mocked<typeof api>;

describe('useMachines', () => {
    const mockMachines: api.Machine[] = [
        {
            id: 1,
            user_id: 1,
            name: 'production-server',
            hostname: 'web-1.example.com',
            status: 'online',
            last_seen: new Date().toISOString(),
            created_at: new Date().toISOString(),
        },
        {
            id: 2,
            user_id: 1,
            name: 'staging-server',
            hostname: 'staging.example.com',
            status: 'offline',
            last_seen: new Date(Date.now() - 10 * 60 * 1000).toISOString(), // 10 minutes ago
            created_at: new Date().toISOString(),
        },
    ];

    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('should load machines on mount', async () => {
        mockApi.listMachines.mockResolvedValue(mockMachines);

        const { result } = renderHook(() => useMachines());

        // Initially loading
        expect(result.current.loading).toBe(true);
        expect(result.current.machines).toEqual([]);

        // Wait for loading to complete
        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        // Should have machines
        expect(result.current.machines).toEqual(mockMachines);
        expect(result.current.error).toBeNull();
        expect(mockApi.listMachines).toHaveBeenCalledTimes(1);
    });

    it('should handle errors when loading machines', async () => {
        const errorMessage = 'Failed to load machines';
        mockApi.listMachines.mockRejectedValue(new Error(errorMessage));

        const { result } = renderHook(() => useMachines());

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        expect(result.current.machines).toEqual([]);
        expect(result.current.error).toBe(errorMessage);
    });

    it('should refresh machines', async () => {
        mockApi.listMachines.mockResolvedValue(mockMachines);

        const { result } = renderHook(() => useMachines());

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        // Clear mock to verify refresh call
        mockApi.listMachines.mockClear();
        mockApi.listMachines.mockResolvedValue([mockMachines[0]]);

        await act(async () => {
            await result.current.refresh();
        });

        expect(mockApi.listMachines).toHaveBeenCalledTimes(1);
        expect(result.current.machines).toEqual([mockMachines[0]]);
    });

    it('should register a new machine', async () => {
        mockApi.listMachines.mockResolvedValue([]);
        const registerResponse: api.RegisterMachineResponse = {
            id: 3,
            name: 'new-server',
            hostname: 'new.example.com',
            api_key: 'test-api-key-12345',
            created_at: new Date().toISOString(),
        };
        mockApi.registerMachine.mockResolvedValue(registerResponse);

        const { result } = renderHook(() => useMachines());

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        // Register machine
        const newMachineData = {
            name: 'new-server',
            hostname: 'new.example.com',
        };

        let response: api.RegisterMachineResponse | undefined;

        // Mock the refresh call after registration
        const newMachine: api.Machine = {
            id: registerResponse.id,
            user_id: 1,
            name: registerResponse.name,
            hostname: registerResponse.hostname,
            status: 'offline',
            last_seen: '',
            created_at: registerResponse.created_at,
        };
        mockApi.listMachines.mockResolvedValue([newMachine]);

        await act(async () => {
            response = await result.current.register(newMachineData);
        });

        expect(mockApi.registerMachine).toHaveBeenCalledWith(newMachineData);
        expect(response).toEqual(registerResponse);
        expect(result.current.machines).toEqual([newMachine]);
    });

    it('should handle registration errors', async () => {
        mockApi.listMachines.mockResolvedValue([]);
        const errorMessage = 'Machine name is required';
        mockApi.registerMachine.mockRejectedValue(new Error(errorMessage));

        const { result } = renderHook(() => useMachines());

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        const newMachineData = {
            name: '',
        };

        let caughtError: Error | undefined;
        await act(async () => {
            try {
                await result.current.register(newMachineData);
            } catch (err) {
                caughtError = err as Error;
            }
        });

        expect(caughtError).toBeDefined();
        expect(caughtError?.message).toBe(errorMessage);

        // Wait for error state to be set
        await waitFor(() => {
            expect(result.current.error).toBe(errorMessage);
        });
    });

    it('should get machine by ID', async () => {
        mockApi.listMachines.mockResolvedValue(mockMachines);

        const { result } = renderHook(() => useMachines());

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        const machine = result.current.getMachine(1);
        expect(machine).toEqual(mockMachines[0]);

        const nonExistent = result.current.getMachine(999);
        expect(nonExistent).toBeUndefined();
    });
});
