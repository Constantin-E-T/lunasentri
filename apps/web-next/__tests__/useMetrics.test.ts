/**
 * @jest-environment jsdom
 */

import { renderHook } from '@testing-library/react';
import { useMetrics } from '../lib/useMetrics';

jest.mock('../lib/api', () => ({
  fetchMetrics: jest.fn(),
}));

describe('useMetrics (smoke)', () => {
  it('returns the expected state shape on mount', () => {
    const { result } = renderHook(() => useMetrics());

    expect(result.current).toEqual(
      expect.objectContaining({
        metrics: null,
        error: null,
        loading: true,
        connectionType: 'disconnected',
        lastUpdate: null,
        retry: expect.any(Function),
      }),
    );
  });

  it('exposes a retry function without throwing', () => {
    const { result } = renderHook(() => useMetrics());

    expect(() => result.current.retry()).not.toThrow();
  });
});
