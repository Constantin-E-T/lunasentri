/**
 * Alert severity utility for consistent styling and thresholds across the app
 */

export type AlertSeverity = 'ok' | 'warn' | 'critical';

export type MetricType = 'cpu_pct' | 'mem_used_pct' | 'disk_used_pct';

export interface SeverityThresholds {
    warn: number;
    critical: number;
}

/**
 * Default severity thresholds per metric type
 */
const DEFAULT_THRESHOLDS: Record<MetricType, SeverityThresholds> = {
    cpu_pct: {
        warn: 60,
        critical: 85,
    },
    mem_used_pct: {
        warn: 70,
        critical: 90,
    },
    disk_used_pct: {
        warn: 80,
        critical: 95,
    },
};

/**
 * Get the severity level for a metric value
 * 
 * @param metric - The metric type
 * @param value - The current metric value (0-100)
 * @param comparison - Whether we're checking above or below thresholds (defaults to 'above')
 * @param threshold - Custom threshold for rule-based severity (optional)
 * @returns The severity level
 */
export function getMetricSeverity(
    metric: MetricType,
    value: number,
    comparison: 'above' | 'below' = 'above',
    threshold?: number
): AlertSeverity {
    const thresholds = DEFAULT_THRESHOLDS[metric];

    if (comparison === 'above') {
        // For custom thresholds (e.g., from alert rules), use those to determine severity
        if (threshold !== undefined) {
            // If value exceeds threshold, determine severity based on how far above it is
            if (value >= threshold) {
                const percentageOverThreshold = ((value - threshold) / threshold) * 100;
                // If 25% or more over threshold, it's critical; otherwise warn
                return percentageOverThreshold >= 25 ? 'critical' : 'warn';
            }
            return 'ok';
        }

        // Use default thresholds
        if (value >= thresholds.critical) return 'critical';
        if (value >= thresholds.warn) return 'warn';
        return 'ok';
    } else {
        // For "below" comparisons (e.g., CPU drops below expected level)
        // Reverse the logic - being far below threshold is critical
        if (threshold !== undefined) {
            if (value <= threshold) {
                const percentageBelowThreshold = ((threshold - value) / threshold) * 100;
                // If 50% or more below threshold, it's critical; if 25% or more, it's warn
                if (percentageBelowThreshold >= 50) return 'critical';
                if (percentageBelowThreshold >= 25) return 'warn';
                return 'warn'; // Any amount below threshold is at least a warning
            }
            return 'ok';
        }

        // Use inverted default thresholds for below comparisons
        // Critical if way below normal operating range
        const invertedCritical = 100 - thresholds.critical;
        const invertedWarn = 100 - thresholds.warn;

        if (value <= invertedCritical) return 'critical';
        if (value <= invertedWarn) return 'warn';
        return 'ok';
    }
}

/**
 * Get Tailwind CSS classes for severity styling
 * Includes background, text, and border classes with accessible contrast
 * 
 * @param severity - The severity level
 * @returns Object with CSS classes for different UI elements
 */
export function getSeverityStyles(severity: AlertSeverity) {
    switch (severity) {
        case 'critical':
            return {
                // Red styling for critical alerts
                background: 'bg-red-500/20',
                text: 'text-red-400',
                border: 'border-red-500/30',
                badge: 'bg-red-500/20 text-red-400 border-red-500/30',
                button: 'bg-red-500/20 hover:bg-red-500/30 text-red-400 border-red-500/30',
                progressBar: 'bg-red-500',
                glow: 'shadow-[0_0_12px_rgba(239,68,68,0.4)]', // red-500 with alpha
            };
        case 'warn':
            return {
                // Amber/yellow styling for warnings
                background: 'bg-amber-500/20',
                text: 'text-amber-400',
                border: 'border-amber-500/30',
                badge: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
                button: 'bg-amber-500/20 hover:bg-amber-500/30 text-amber-400 border-amber-500/30',
                progressBar: 'bg-amber-500',
                glow: 'shadow-[0_0_12px_rgba(245,158,11,0.4)]', // amber-500 with alpha
            };
        case 'ok':
        default:
            return {
                // Green styling for normal/ok state
                background: 'bg-emerald-500/20',
                text: 'text-emerald-400',
                border: 'border-emerald-500/30',
                badge: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
                button: 'bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-400 border-emerald-500/30',
                progressBar: 'bg-emerald-500',
                glow: 'shadow-[0_0_12px_rgba(16,185,129,0.4)]', // emerald-500 with alpha
            };
    }
}

/**
 * Get a human-readable label for severity
 * 
 * @param severity - The severity level
 * @returns Human-readable label
 */
export function getSeverityLabel(severity: AlertSeverity): string {
    switch (severity) {
        case 'critical':
            return 'Critical';
        case 'warn':
            return 'Warning';
        case 'ok':
        default:
            return 'Good';
    }
}

/**
 * Get severity for an alert event based on rule configuration
 * 
 * @param metric - The metric type from the rule
 * @param value - The event value
 * @param threshold - The rule threshold
 * @param comparison - The rule comparison type
 * @returns The severity level for this event
 */
export function getEventSeverity(
    metric: MetricType,
    value: number,
    threshold: number,
    comparison: 'above' | 'below'
): AlertSeverity {
    return getMetricSeverity(metric, value, comparison, threshold);
}