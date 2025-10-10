'use client';

import { useEffect, useState } from 'react';

const SELECTED_MACHINE_KEY = 'lunasentri_selected_machine_id';

export interface UseMachineSelectionReturn {
    /** Currently selected machine ID (null if none selected) */
    selectedMachineId: number | null;

    /** Select a machine by ID */
    selectMachine: (machineId: number | null) => void;

    /** Clear selection */
    clearSelection: () => void;
}

/**
 * Hook for managing machine selection with localStorage persistence.
 * 
 * Features:
 * - Persists selection across page reloads
 * - Syncs across tabs (via storage events)
 * - Null-safe (handles no selection state)
 */
export function useMachineSelection(): UseMachineSelectionReturn {
    const [selectedMachineId, setSelectedMachineId] = useState<number | null>(null);
    const [isHydrated, setIsHydrated] = useState(false);

    // Load from localStorage on mount
    useEffect(() => {
        // SSR safety check
        if (typeof window === "undefined") {
            setIsHydrated(true);
            return;
        }

        const stored = localStorage.getItem(SELECTED_MACHINE_KEY);
        if (stored) {
            const id = parseInt(stored, 10);
            if (!isNaN(id)) {
                setSelectedMachineId(id);
            }
        }
        setIsHydrated(true);
    }, []);

    // Sync selection across tabs
    useEffect(() => {
        // SSR safety check
        if (typeof window === "undefined") return;

        const handleStorageChange = (e: StorageEvent) => {
            if (e.key === SELECTED_MACHINE_KEY) {
                if (e.newValue) {
                    const id = parseInt(e.newValue, 10);
                    if (!isNaN(id)) {
                        setSelectedMachineId(id);
                    } else {
                        // Invalid value, clear selection
                        setSelectedMachineId(null);
                    }
                } else {
                    setSelectedMachineId(null);
                }
            }
        };

        window.addEventListener('storage', handleStorageChange);
        return () => window.removeEventListener('storage', handleStorageChange);
    }, []);

    const selectMachine = (machineId: number | null) => {
        setSelectedMachineId(machineId);
        // SSR safety check
        if (typeof window === "undefined") return;

        if (machineId === null) {
            localStorage.removeItem(SELECTED_MACHINE_KEY);
        } else {
            localStorage.setItem(SELECTED_MACHINE_KEY, machineId.toString());
        }
    };

    const clearSelection = () => {
        selectMachine(null);
    };

    // Don't expose selection until hydrated (prevents SSR mismatch)
    return {
        selectedMachineId: isHydrated ? selectedMachineId : null,
        selectMachine,
        clearSelection,
    };
}
