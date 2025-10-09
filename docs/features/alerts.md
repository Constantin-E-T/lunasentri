# Alert System

Guide to LunaSentri's intelligent alerting system with flexible rules, event tracking, and acknowledgments.

---

## Overview

LunaSentri provides a comprehensive alert system that monitors metrics and triggers notifications when thresholds are exceeded.

## Alert Rules

### Creating Alert Rules

Alert rules define the conditions that trigger notifications.

**Fields:**

- **Name**: Descriptive name (e.g., "High CPU Usage")
- **Metric**: Metric to monitor (cpu_pct, mem_used_pct, disk_used_pct, network_rx_bytes, network_tx_bytes)
- **Condition**: above or below threshold
- **Threshold**: Numeric value to compare against
- **Consecutive Samples**: Number of consecutive readings before triggering (prevents false alarms)
- **Active**: Enable/disable rule

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/alerts/rules` | List all alert rules |
| `POST` | `/alerts/rules` | Create alert rule |
| `PUT` | `/alerts/rules/:id` | Update alert rule |
| `DELETE` | `/alerts/rules/:id` | Delete alert rule |

## Alert Events

### Event Lifecycle

1. **Triggered**: Rule condition met for consecutive samples
2. **Acknowledged**: Admin marks event as seen
3. **Resolved**: Metric returns to normal (automatic)

### Event Tracking

**Fields:**

- Alert rule details (name, threshold, condition)
- Current metric value
- Trigger timestamp
- Acknowledgment status

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/alerts/events` | List alert events |
| `POST` | `/alerts/events/:id/ack` | Acknowledge alert |

## Notifications

When an alert triggers, notifications are sent to all configured channels:

- **Webhooks**: HTTP POST with alert payload and HMAC signature
- **Telegram**: Formatted message to all active recipients

## UI Features

- Real-time alert status on dashboard
- Alert rule management page
- Event history with filtering
- Acknowledgment workflow
- Active alerts counter

## Best Practices

1. **Set appropriate thresholds** - Too sensitive creates noise, too high misses issues
2. **Use consecutive samples** - Reduces false positives (recommended: 3 samples)
3. **Test notifications** - Verify webhook/Telegram delivery before relying on alerts
4. **Acknowledge alerts** - Keep track of what's been addressed
5. **Review periodically** - Adjust thresholds based on actual usage patterns

---

**Status: Production Ready** 🚨
