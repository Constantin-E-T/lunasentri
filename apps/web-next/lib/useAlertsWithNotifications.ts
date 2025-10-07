'use client';

import { useEffect, useRef } from 'react';
import { useAlerts } from './useAlerts';
import { useToast } from '@/components/ui/use-toast';
import { useSession } from './useSession';
import type { AlertEvent } from './api';

export interface UseAlertsWithNotificationsReturn extends ReturnType<typeof useAlerts> {
    // Additional functionality for notifications
    lastSeenEventId: number | null;
    newAlertsCount: number;
    markAllAsSeen: () => void;
}

/**
 * Enhanced alerts hook that provides toast notifications for new alert events.
 * 
 * This hook builds upon useAlerts to add:
 * - Toast notifications for previously unseen alert events
 * - Tracking of the last seen event ID
 * - Count of new alerts since last seen
 * - Function to mark all current alerts as seen
 */
export function useAlertsWithNotifications(eventsLimit?: number): UseAlertsWithNotificationsReturn {
    const alertsHook = useAlerts(eventsLimit);
    const { toast } = useToast();
    const { status } = useSession();

    // Track the last seen event ID in localStorage
    const lastSeenEventIdRef = useRef<number | null>(null);
    const isInitialLoadRef = useRef(true);
    const isRefreshInFlightRef = useRef(false);
    const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);

    // Initialize last seen event ID from localStorage
    useEffect(() => {
        try {
            const stored = localStorage.getItem('lunasentri-last-seen-event-id');
            if (stored) {
                lastSeenEventIdRef.current = parseInt(stored, 10);
            }
        } catch (error) {
            // localStorage not available, continue without persistence
            console.warn('localStorage not available for alert notifications');
        }
    }, []);

    // Set up polling for live updates when authenticated
    useEffect(() => {
        if (status !== 'authenticated') {
            // Clear polling if not authenticated
            if (pollingIntervalRef.current) {
                clearInterval(pollingIntervalRef.current);
                pollingIntervalRef.current = null;
            }
            return;
        }

        // Start polling every 10 seconds
        pollingIntervalRef.current = setInterval(async () => {
            // Skip if a refresh is already in progress
            if (isRefreshInFlightRef.current) {
                return;
            }

            try {
                isRefreshInFlightRef.current = true;
                await alertsHook.refresh();
            } catch (error) {
                console.error('Failed to refresh alerts during polling:', error);
            } finally {
                isRefreshInFlightRef.current = false;
            }
        }, 10000); // 10 second interval

        // Cleanup on unmount or auth status change
        return () => {
            if (pollingIntervalRef.current) {
                clearInterval(pollingIntervalRef.current);
                pollingIntervalRef.current = null;
            }
            isRefreshInFlightRef.current = false;
        };
    }, [status, alertsHook.refresh]);

    // Track new events and show notifications
    useEffect(() => {
        if (alertsHook.eventsLoading || !alertsHook.events || alertsHook.events.length === 0) {
            return;
        }

        // Skip notifications on initial load
        if (isInitialLoadRef.current) {
            isInitialLoadRef.current = false;
            // Set the last seen ID to the most recent event on initial load if none exists
            if (lastSeenEventIdRef.current === null && alertsHook.events.length > 0) {
                const latestEvent = alertsHook.events[0];
                lastSeenEventIdRef.current = latestEvent.id;
                try {
                    localStorage.setItem('lunasentri-last-seen-event-id', latestEvent.id.toString());
                } catch (error) {
                    // localStorage not available, continue without persistence
                }
            }
            return;
        }

        // Find new events (events with ID greater than last seen)
        const newEvents = lastSeenEventIdRef.current !== null
            ? alertsHook.events.filter(event => event.id > lastSeenEventIdRef.current!)
            : [];

        // Show toast notifications for new events
        newEvents.forEach((event) => {
            showEventNotification(event);
        });

        // Update last seen event ID if there are new events
        if (newEvents.length > 0) {
            const latestNewEventId = Math.max(...newEvents.map(e => e.id));
            lastSeenEventIdRef.current = latestNewEventId;
            try {
                localStorage.setItem('lunasentri-last-seen-event-id', latestNewEventId.toString());
            } catch (error) {
                // localStorage not available, continue without persistence
            }
        }
    }, [alertsHook.events, alertsHook.eventsLoading, alertsHook.rules, toast]);

    // Helper function to show event notification
    const showEventNotification = (event: AlertEvent) => {
        // Find the rule associated with this event
        const rule = alertsHook.rules?.find(r => r.id === event.rule_id);

        const metricLabels: Record<string, string> = {
            'cpu_pct': 'CPU',
            'mem_used_pct': 'Memory',
            'disk_used_pct': 'Disk'
        };

        const metricName = rule?.metric ? metricLabels[rule.metric] || rule.metric : 'System';
        const ruleName = rule?.name || `Rule ${event.rule_id}`;
        const threshold = rule?.threshold_pct || 0;
        const comparison = rule?.comparison === 'above' ? 'exceeded' : 'dropped below';

        // Determine severity based on metric and threshold
        const severity = getSeverityInfo(rule?.metric || 'unknown', threshold, event.value);

        toast({
            title: `🚨 ${severity.label} Alert`,
            description: `${ruleName}: ${metricName} ${comparison} ${threshold}% (current: ${event.value.toFixed(1)}%)`,
            variant: severity.variant,
            duration: 8000, // Show for 8 seconds
        });
    };

    // Helper function to get severity information
    const getSeverityInfo = (metric: string, threshold: number, currentValue: number) => {
        // Determine severity based on how far we are from threshold
        const difference = Math.abs(currentValue - threshold);

        if (metric === 'cpu_pct' || metric === 'mem_used_pct') {
            if (difference >= 20) {
                return { label: 'Critical', variant: 'destructive' as const };
            } else if (difference >= 10) {
                return { label: 'Warning', variant: 'warning' as const };
            }
        } else if (metric === 'disk_used_pct') {
            if (difference >= 10) {
                return { label: 'Critical', variant: 'destructive' as const };
            } else if (difference >= 5) {
                return { label: 'Warning', variant: 'warning' as const };
            }
        }

        return { label: 'Alert', variant: 'default' as const };
    };

    // Calculate new alerts count
    const newAlertsCount = lastSeenEventIdRef.current !== null && alertsHook.events
        ? alertsHook.events.filter(event => event.id > lastSeenEventIdRef.current!).length
        : 0;

    // Function to mark all current alerts as seen
    const markAllAsSeen = () => {
        if (alertsHook.events && alertsHook.events.length > 0) {
            const latestEventId = Math.max(...alertsHook.events.map(e => e.id));
            lastSeenEventIdRef.current = latestEventId;
            try {
                localStorage.setItem('lunasentri-last-seen-event-id', latestEventId.toString());
            } catch (error) {
                // localStorage not available, continue without persistence
            }
        }
    };

    return {
        ...alertsHook,
        lastSeenEventId: lastSeenEventIdRef.current,
        newAlertsCount,
        markAllAsSeen,
    };
}