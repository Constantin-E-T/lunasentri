'use client';

import { useEffect, useState, useRef, useCallback } from 'react';
import { fetchMetrics, type Metrics } from './api';

export interface MetricSample {
    timestamp: number;
    cpu_pct: number;
    mem_used_pct: number;
    disk_used_pct: number;
    uptime_s: number;
}

export interface UseMetricsOptions {
    /**
     * Machine ID to fetch metrics for.
     * If not provided, falls back to localhost metrics (if LOCAL_HOST_METRICS is enabled).
     */
    machineId?: number;

    /**
     * WebSocket URL for real-time streaming.
     * If not provided, falls back to polling.
     */
    wsUrl?: string;

    /**
     * Polling interval in milliseconds when WebSocket is unavailable.
     * @default 5000
     */
    pollInterval?: number;

    /**
     * Maximum number of WebSocket reconnection attempts.
     * @default 3
     */
    maxReconnectAttempts?: number;

    /**
     * WebSocket reconnection delay in milliseconds.
     * @default 2000
     */
    reconnectDelay?: number;
}

export interface UseMetricsReturn {
    /** Current metrics data */
    metrics: Metrics | null;

    /** Error state */
    error: string | null;

    /** Loading state (initial load only) */
    loading: boolean;

    /** Connection type being used */
    connectionType: 'websocket' | 'polling' | 'disconnected';

    /** Last update timestamp */
    lastUpdate: Date | null;

    /** Manual retry function */
    retry: () => void;

    /** Rolling buffer of metrics history (last 60 samples) */
    history: MetricSample[];
}

const DEFAULT_OPTIONS: Required<Omit<UseMetricsOptions, 'machineId'>> = {
    wsUrl: '',
    pollInterval: 5000,
    maxReconnectAttempts: 3,
    reconnectDelay: 2000,
};

/**
 * Hook for fetching metrics with WebSocket streaming and polling fallback.
 * 
 * Features:
 * - Machine-specific or localhost metrics
 * - Automatic WebSocket connection with fallback to polling
 * - Automatic reconnection on WebSocket failures
 * - Graceful degradation when WebSocket is unavailable
 * - Real-time updates every ~3s via WebSocket
 * - Manual retry capability
 * 
 * @param options - Configuration options including optional machineId
 */
export function useMetrics(options: UseMetricsOptions = {}): UseMetricsReturn {
    const opts = { ...DEFAULT_OPTIONS, ...options };
    const { machineId } = options;

    // Determine WebSocket URL with machine_id parameter if provided
    const wsUrl = opts.wsUrl || (() => {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
        const baseWsUrl = apiUrl.replace(/^http/, 'ws') + '/ws';
        if (machineId !== undefined) {
            return `${baseWsUrl}?machine_id=${machineId}`;
        }
        return baseWsUrl;
    })();

    // State
    const [metrics, setMetrics] = useState<Metrics | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(true);
    const [connectionType, setConnectionType] = useState<'websocket' | 'polling' | 'disconnected'>('disconnected');
    const [lastUpdate, setLastUpdate] = useState<Date | null>(null);
    const [history, setHistory] = useState<MetricSample[]>([]);

    // Refs for cleanup and connection management
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectAttemptsRef = useRef(0);
    const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    const pollTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    const isMountedRef = useRef(true);

    // Helper function to update metrics and history
    const updateMetricsAndHistory = useCallback((newMetrics: Metrics) => {
        const sample: MetricSample = {
            timestamp: Date.now(),
            cpu_pct: newMetrics.cpu_pct,
            mem_used_pct: newMetrics.mem_used_pct,
            disk_used_pct: newMetrics.disk_used_pct,
            uptime_s: newMetrics.uptime_s,
        };

        setMetrics(newMetrics);
        setHistory(prev => {
            const newHistory = [...prev, sample];
            // Keep only the last 60 samples
            return newHistory.slice(-60);
        });
        setError(null);
        setLoading(false);
        setLastUpdate(new Date());
    }, []);

    // WebSocket message handler
    const handleWebSocketMessage = useCallback((event: MessageEvent) => {
        try {
            const data: Metrics = JSON.parse(event.data);
            if (isMountedRef.current) {
                updateMetricsAndHistory(data);
                setConnectionType('websocket');
            }
        } catch (err) {
            console.error('Failed to parse WebSocket message:', err);
            if (isMountedRef.current) {
                setError('Invalid data received from server');
            }
        }
    }, [updateMetricsAndHistory]);

    // WebSocket error handler
    const handleWebSocketError = useCallback((event: Event) => {
        console.warn('WebSocket error:', event);
        if (isMountedRef.current) {
            setConnectionType('disconnected');
        }
    }, []);

    // WebSocket close handler with reconnection logic
    const handleWebSocketClose = useCallback((event: CloseEvent) => {
        console.log('WebSocket closed:', event.code, event.reason);

        if (!isMountedRef.current) return;

        setConnectionType('disconnected');
        wsRef.current = null;

        // Attempt reconnection if not at max attempts
        if (reconnectAttemptsRef.current < opts.maxReconnectAttempts) {
            reconnectAttemptsRef.current++;
            console.log(`WebSocket reconnection attempt ${reconnectAttemptsRef.current}/${opts.maxReconnectAttempts}`);

            reconnectTimeoutRef.current = setTimeout(() => {
                if (isMountedRef.current) {
                    connectWebSocket();
                }
            }, opts.reconnectDelay);
        } else {
            // Max reconnection attempts reached, fallback to polling
            console.log('Max WebSocket reconnection attempts reached, falling back to polling');
            startPolling();
        }
    }, [opts.maxReconnectAttempts, opts.reconnectDelay]);

    // WebSocket connection function
    const connectWebSocket = useCallback(() => {
        if (!isMountedRef.current || wsRef.current?.readyState === WebSocket.OPEN) {
            return;
        }

        try {
            const ws = new WebSocket(wsUrl);
            wsRef.current = ws;

            ws.onopen = () => {
                console.log('WebSocket connected');
                if (isMountedRef.current) {
                    reconnectAttemptsRef.current = 0; // Reset reconnection attempts
                    setConnectionType('websocket');
                    setError(null);
                }
            };

            ws.onmessage = handleWebSocketMessage;
            ws.onerror = handleWebSocketError;
            ws.onclose = handleWebSocketClose;

        } catch (err) {
            console.error('Failed to create WebSocket connection:', err);
            if (isMountedRef.current) {
                setError('Failed to establish WebSocket connection');
                startPolling();
            }
        }
    }, [wsUrl, handleWebSocketMessage, handleWebSocketError, handleWebSocketClose]);

    // Polling function
    const poll = useCallback(async () => {
        if (!isMountedRef.current) return;

        try {
            const data = await fetchMetrics(machineId);
            if (isMountedRef.current) {
                updateMetricsAndHistory(data);
                setConnectionType('polling');
            }
        } catch (err) {
            if (isMountedRef.current) {
                // Check if it's a 401 authentication error
                const errorMessage = err instanceof Error ? err.message : 'Failed to load metrics';
                if (errorMessage.includes('401')) {
                    setError('Please log in to view metrics');
                } else {
                    setError(errorMessage);
                }
                setLoading(false);
            }
        }

        // Schedule next poll only if still in polling mode
        if (isMountedRef.current) {
            pollTimeoutRef.current = setTimeout(poll, opts.pollInterval);
        }
    }, [opts.pollInterval, updateMetricsAndHistory, machineId]);

    // Start polling fallback
    const startPolling = useCallback(() => {
        if (pollTimeoutRef.current) {
            clearTimeout(pollTimeoutRef.current);
        }
        poll();
    }, [poll]);

    // Manual retry function
    const retry = useCallback(() => {
        setError(null);
        setLoading(true);
        reconnectAttemptsRef.current = 0;

        // Clear existing connections/timeouts
        if (wsRef.current) {
            wsRef.current.close();
            wsRef.current = null;
        }
        if (pollTimeoutRef.current) {
            clearTimeout(pollTimeoutRef.current);
            pollTimeoutRef.current = null;
        }
        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = null;
        }

        // Try WebSocket first
        connectWebSocket();
    }, [connectWebSocket]);

    // Initialize connection on mount
    useEffect(() => {
        isMountedRef.current = true;

        // Try WebSocket first, fallback to polling if it fails
        connectWebSocket();

        // Cleanup on unmount
        return () => {
            isMountedRef.current = false;

            if (wsRef.current) {
                wsRef.current.close();
            }
            if (pollTimeoutRef.current) {
                clearTimeout(pollTimeoutRef.current);
            }
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
        };
    }, [connectWebSocket]);

    return {
        metrics,
        error,
        loading,
        connectionType,
        lastUpdate,
        retry,
        history,
    };
}