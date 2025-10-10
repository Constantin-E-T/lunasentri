'use client';

import { useCallback, useEffect, useState } from 'react';
import { listMachines, registerMachine, type Machine, type RegisterMachineRequest, type RegisterMachineResponse } from './api';

export interface UseMachinesReturn {
    /** List of machines */
    machines: Machine[];

    /** Loading state (initial load) */
    loading: boolean;

    /** Error state */
    error: string | null;

    /** Refresh machine list */
    refresh: () => Promise<void>;

    /** Register a new machine */
    register: (data: RegisterMachineRequest) => Promise<RegisterMachineResponse>;

    /** Get machine by ID */
    getMachine: (id: number) => Machine | undefined;
}

/**
 * Hook for managing machines.
 * 
 * Features:
 * - Automatic loading of machines on mount
 * - Manual refresh capability
 * - Machine registration with API key return
 * - Error handling
 */
export function useMachines(): UseMachinesReturn {
    const [machines, setMachines] = useState<Machine[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const refresh = useCallback(async () => {
        try {
            setError(null);
            const data = await listMachines();
            setMachines(data);
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to load machines';
            setError(message);
            console.error('Failed to refresh machines:', err);
        }
    }, []);

    const register = useCallback(async (data: RegisterMachineRequest): Promise<RegisterMachineResponse> => {
        try {
            const response = await registerMachine(data);
            // Refresh machine list after registration
            await refresh();
            return response;
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to register machine';
            setError(message);
            throw err;
        }
    }, [refresh]);

    const getMachine = useCallback((id: number): Machine | undefined => {
        return machines.find(m => m.id === id);
    }, [machines]);

    // Load machines on mount
    useEffect(() => {
        const loadMachines = async () => {
            setLoading(true);
            await refresh();
            setLoading(false);
        };

        loadMachines();
    }, [refresh]);

    return {
        machines,
        loading,
        error,
        refresh,
        register,
        getMachine,
    };
}
