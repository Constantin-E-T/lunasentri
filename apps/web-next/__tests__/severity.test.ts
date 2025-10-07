/**
 * Tests for alert severity utility
 * @jest-environment jsdom
 */

import {
    getMetricSeverity,
    getSeverityStyles,
    getSeverityLabel,
    getEventSeverity,
    type AlertSeverity,
    type MetricType
} from '@/lib/alerts/severity';

describe('Alert Severity Utility', () => {
    describe('getMetricSeverity', () => {
        test('returns correct severity for CPU metric above thresholds', () => {
            expect(getMetricSeverity('cpu_pct', 50)).toBe('ok');
            expect(getMetricSeverity('cpu_pct', 65)).toBe('warn');
            expect(getMetricSeverity('cpu_pct', 90)).toBe('critical');
        });

        test('returns correct severity for memory metric above thresholds', () => {
            expect(getMetricSeverity('mem_used_pct', 60)).toBe('ok');
            expect(getMetricSeverity('mem_used_pct', 75)).toBe('warn');
            expect(getMetricSeverity('mem_used_pct', 95)).toBe('critical');
        });

        test('returns correct severity for disk metric above thresholds', () => {
            expect(getMetricSeverity('disk_used_pct', 70)).toBe('ok');
            expect(getMetricSeverity('disk_used_pct', 85)).toBe('warn');
            expect(getMetricSeverity('disk_used_pct', 98)).toBe('critical');
        });

        test('handles custom thresholds correctly', () => {
            // Value above custom threshold should be warning/critical
            expect(getMetricSeverity('cpu_pct', 85, 'above', 80)).toBe('warn');
            expect(getMetricSeverity('cpu_pct', 101, 'above', 80)).toBe('critical'); // 25%+ over threshold

            // Value below custom threshold should be ok
            expect(getMetricSeverity('cpu_pct', 75, 'above', 80)).toBe('ok');
        });

        test('handles below comparisons correctly', () => {
            // For below comparisons, being far below threshold is bad
            expect(getMetricSeverity('cpu_pct', 45, 'below', 60)).toBe('warn'); // 25% below
            expect(getMetricSeverity('cpu_pct', 30, 'below', 60)).toBe('critical'); // 50% below
            expect(getMetricSeverity('cpu_pct', 70, 'below', 60)).toBe('ok'); // above threshold is good
        });
    });

    describe('getSeverityStyles', () => {
        test('returns correct styles for each severity level', () => {
            const okStyles = getSeverityStyles('ok');
            expect(okStyles.text).toContain('emerald');
            expect(okStyles.progressBar).toBe('bg-emerald-500');

            const warnStyles = getSeverityStyles('warn');
            expect(warnStyles.text).toContain('amber');
            expect(warnStyles.progressBar).toBe('bg-amber-500');

            const criticalStyles = getSeverityStyles('critical');
            expect(criticalStyles.text).toContain('red');
            expect(criticalStyles.progressBar).toBe('bg-red-500');
        });

        test('includes all required style properties', () => {
            const styles = getSeverityStyles('warn');
            expect(styles).toHaveProperty('background');
            expect(styles).toHaveProperty('text');
            expect(styles).toHaveProperty('border');
            expect(styles).toHaveProperty('badge');
            expect(styles).toHaveProperty('button');
            expect(styles).toHaveProperty('progressBar');
            expect(styles).toHaveProperty('glow');
        });
    });

    describe('getSeverityLabel', () => {
        test('returns correct labels for each severity', () => {
            expect(getSeverityLabel('ok')).toBe('Good');
            expect(getSeverityLabel('warn')).toBe('Warning');
            expect(getSeverityLabel('critical')).toBe('Critical');
        });
    });

    describe('getEventSeverity', () => {
        test('calculates event severity based on rule configuration', () => {
            // CPU above 85% (threshold) should be warning
            expect(getEventSeverity('cpu_pct', 87, 85, 'above')).toBe('warn');

            // CPU way above threshold should be critical  
            expect(getEventSeverity('cpu_pct', 107, 85, 'above')).toBe('critical');

            // CPU below threshold should be ok
            expect(getEventSeverity('cpu_pct', 80, 85, 'above')).toBe('ok');
        });
    });
});