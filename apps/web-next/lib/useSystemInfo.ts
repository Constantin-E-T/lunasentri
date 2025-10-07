'use client';

import { useEffect, useState, useRef } from 'react';
import { fetchSystemInfo, type SystemInfo } from './api';

export interface UseSystemInfoReturn {
  /** Current system info data */
  systemInfo: SystemInfo | null;

  /** Error state */
  error: string | null;

  /** Loading state */
  loading: boolean;

  /** Manually refetch system info */
  refetch: () => Promise<void>;
}

export function useSystemInfo(): UseSystemInfoReturn {
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const isMountedRef = useRef(true);

  const fetchData = async () => {
    try {
      setError(null);
      const data = await fetchSystemInfo();
      
      if (isMountedRef.current) {
        setSystemInfo(data);
      }
    } catch (err) {
      console.error('Failed to fetch system info:', err);
      if (isMountedRef.current) {
        setError(err instanceof Error ? err.message : 'Failed to fetch system info');
      }
    } finally {
      if (isMountedRef.current) {
        setLoading(false);
      }
    }
  };

  const refetch = async () => {
    setLoading(true);
    await fetchData();
  };

  useEffect(() => {
    isMountedRef.current = true;
    fetchData();

    return () => {
      isMountedRef.current = false;
    };
  }, []);

  return {
    systemInfo,
    error,
    loading,
    refetch,
  };
}