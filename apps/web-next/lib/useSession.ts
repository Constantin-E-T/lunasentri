'use client';

import { useEffect, useState, useCallback } from 'react';
import { fetchCurrentUser, login as apiLogin, logout as apiLogout, register as apiRegister, changePassword as apiChangePassword, type User } from './api';
import { useToast } from '@/components/ui/use-toast';

export type SessionStatus = 'loading' | 'authenticated' | 'unauthenticated';

export interface Session {
  user: User | null;
  status: SessionStatus;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  changePassword: (currentPassword: string, newPassword: string) => Promise<void>;
  logout: () => Promise<void>;
}

export function useSession(): Session {
  const [user, setUser] = useState<User | null>(null);
  const [status, setStatus] = useState<SessionStatus>('loading');
  const { toast } = useToast();

  // Handle session expiry
  const handleSessionExpiry = useCallback(async () => {
    // Clear user state
    setUser(null);
    setStatus('unauthenticated');

    // Show toast notification
    toast({
      title: "Session expired",
      description: "Please log in again.",
      variant: "destructive",
    });

    // Best effort logout to clean up server session
    try {
      await apiLogout();
    } catch (error) {
      // Ignore logout failures during session expiry
      console.warn('Failed to logout during session expiry:', error);
    }
  }, [toast]);

  // Listen for session expiry events
  useEffect(() => {
    const handleSessionExpired = () => {
      handleSessionExpiry();
    };

    if (typeof window !== 'undefined') {
      window.addEventListener('session-expired', handleSessionExpired);

      return () => {
        window.removeEventListener('session-expired', handleSessionExpired);
      };
    }
  }, [handleSessionExpiry]);

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

  const register = useCallback(async (email: string, password: string) => {
    try {
      // Register the user
      const registeredUser = await apiRegister(email, password);

      // After successful registration, try to log them in automatically
      // by fetching the current user (registration endpoint should set session cookie)
      try {
        const currentUser = await fetchCurrentUser();
        setUser(currentUser);
        setStatus('authenticated');
      } catch {
        // If fetching current user fails, the registration didn't auto-login
        // Keep them unauthenticated and they'll need to login manually
        setUser(null);
        setStatus('unauthenticated');
      }
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

  const changePassword = useCallback(async (currentPassword: string, newPassword: string) => {
    try {
      await apiChangePassword(currentPassword, newPassword);
      // Password change successful - keep user logged in
      // Session remains valid, no need to update user state
    } catch (error) {
      throw error; // Re-throw so caller can handle error display
    }
  }, []);

  return {
    user,
    status,
    login,
    register,
    changePassword,
    logout,
  };
}
