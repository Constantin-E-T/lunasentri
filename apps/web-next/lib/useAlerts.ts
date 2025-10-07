'use client';

import { useEffect, useState, useCallback } from 'react';
import {
    listAlertRules,
    createAlertRule,
    updateAlertRule,
    deleteAlertRule,
    listAlertEvents,
    ackAlertEvent,
    type AlertRule,
    type AlertEvent,
    type CreateAlertRuleRequest,
} from './api';

export interface UseAlertsReturn {
    // Data
    rules: AlertRule[];
    events: AlertEvent[];

    // Loading states
    rulesLoading: boolean;
    eventsLoading: boolean;

    // Error states
    rulesError: string | null;
    eventsError: string | null;

    // Actions
    createRule: (rule: CreateAlertRuleRequest) => Promise<void>;
    updateRule: (id: number, rule: CreateAlertRuleRequest) => Promise<void>;
    deleteRule: (id: number) => Promise<void>;
    acknowledgeEvent: (id: number) => Promise<void>;
    refresh: () => Promise<void>;
    refreshRules: () => Promise<void>;
    refreshEvents: () => Promise<void>;
}

export function useAlerts(eventsLimit?: number): UseAlertsReturn {
    const [rules, setRules] = useState<AlertRule[]>([]);
    const [events, setEvents] = useState<AlertEvent[]>([]);

    const [rulesLoading, setRulesLoading] = useState(true);
    const [eventsLoading, setEventsLoading] = useState(true);

    const [rulesError, setRulesError] = useState<string | null>(null);
    const [eventsError, setEventsError] = useState<string | null>(null);

    // Fetch alert rules
    const refreshRules = useCallback(async () => {
        try {
            setRulesLoading(true);
            setRulesError(null);
            const rulesData = await listAlertRules();
            setRules(rulesData);
        } catch (error) {
            setRulesError(error instanceof Error ? error.message : 'Failed to fetch alert rules');
        } finally {
            setRulesLoading(false);
        }
    }, []);

    // Fetch alert events
    const refreshEvents = useCallback(async () => {
        try {
            setEventsLoading(true);
            setEventsError(null);
            const eventsData = await listAlertEvents(eventsLimit);
            setEvents(eventsData);
        } catch (error) {
            setEventsError(error instanceof Error ? error.message : 'Failed to fetch alert events');
        } finally {
            setEventsLoading(false);
        }
    }, [eventsLimit]);

    // Refresh both rules and events
    const refresh = useCallback(async () => {
        await Promise.all([refreshRules(), refreshEvents()]);
    }, [refreshRules, refreshEvents]);

    // Create a new alert rule
    const createRule = useCallback(async (rule: CreateAlertRuleRequest) => {
        try {
            await createAlertRule(rule);
            await refreshRules(); // Refresh to get the new rule with ID
        } catch (error) {
            throw error; // Re-throw so caller can handle error display
        }
    }, [refreshRules]);

    // Update an existing alert rule
    const updateRule = useCallback(async (id: number, rule: CreateAlertRuleRequest) => {
        try {
            const updatedRule = await updateAlertRule(id, rule);

            // Optimistically update the local state
            setRules(prev => prev.map(r => r.id === id ? updatedRule : r));
        } catch (error) {
            // Refresh rules to revert optimistic update on error
            await refreshRules();
            throw error; // Re-throw so caller can handle error display
        }
    }, [refreshRules]);

    // Delete an alert rule
    const deleteRule = useCallback(async (id: number) => {
        try {
            await deleteAlertRule(id);

            // Optimistically remove from local state
            setRules(prev => prev.filter(r => r.id !== id));

            // Also refresh events since related events may be deleted (cascade)
            await refreshEvents();
        } catch (error) {
            // Refresh rules to revert optimistic update on error
            await refreshRules();
            throw error; // Re-throw so caller can handle error display
        }
    }, [refreshRules, refreshEvents]);

    // Acknowledge an alert event
    const acknowledgeEvent = useCallback(async (id: number) => {
        try {
            await ackAlertEvent(id);

            // Optimistically update the local state
            setEvents(prev => prev.map(e =>
                e.id === id
                    ? { ...e, acknowledged: true, acknowledged_at: new Date().toISOString() }
                    : e
            ));
        } catch (error) {
            // Refresh events to revert optimistic update on error
            await refreshEvents();
            throw error; // Re-throw so caller can handle error display
        }
    }, [refreshEvents]);

    // Initial load
    useEffect(() => {
        refresh();
    }, [refresh]);

    return {
        rules,
        events,
        rulesLoading,
        eventsLoading,
        rulesError,
        eventsError,
        createRule,
        updateRule,
        deleteRule,
        acknowledgeEvent,
        refresh,
        refreshRules,
        refreshEvents,
    };
}