'use client';

import { useEffect, useState, useCallback } from 'react';
import { fetchCurrentUser, login as apiLogin, logout as apiLogout, type User } from './api';

export type SessionStatus = 'loading' | 'authenticated' | 'unauthenticated';

export interface Session {
  user: User | null;
  status: SessionStatus;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

export function useSession(): Session {
  const [user, setUser] = useState<User | null>(null);
  const [status, setStatus] = useState<SessionStatus>('loading');

  // Check authentication status on mount
  useEffect(() => {
    async function checkAuth() {
      try {
        const currentUser = await fetchCurrentUser();
        setUser(currentUser);
        setStatus('authenticated');
      } catch (error) {
        setUser(null);
        setStatus('unauthenticated');
      }
    }

    checkAuth();
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    try {
      const loggedInUser = await apiLogin(email, password);
      setUser(loggedInUser);
      setStatus('authenticated');
    } catch (error) {
      setUser(null);
      setStatus('unauthenticated');
      throw error; // Re-throw so caller can handle error display
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await apiLogout();
    } finally {
      // Clear user state even if logout fails
      setUser(null);
      setStatus('unauthenticated');
    }
  }, []);

  return {
    user,
    status,
    login,
    logout,
  };
}
