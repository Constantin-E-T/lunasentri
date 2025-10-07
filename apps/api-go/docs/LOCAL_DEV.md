# LunaSentri API - Local Development Guide

## Overview

LunaSentri API provides server monitoring with user authentication and alert management capabilities.

## Environment Setup

### Required Environment Variables

```bash
# JWT Secret (required)
export AUTH_JWT_SECRET=$(python3 -c "import secrets; print(secrets.token_urlsafe(32))")

# Optional configurations
export DB_PATH="./data/lunasentri.db"
export CORS_ALLOWED_ORIGIN="http://localhost:3000"
export ACCESS_TOKEN_TTL="15m"
export PASSWORD_RESET_TTL="1h"
export SECURE_COOKIE="false"  # Only for development

# Optional admin bootstrap
export ADMIN_EMAIL="admin@example.com"
export ADMIN_PASSWORD="your-secure-password"
```

### Starting the Server

```bash
cd apps/api-go
go run main.go
```

The server will start on port 8080 with the following endpoints available.

## API Endpoints

### Authentication

All endpoints except `/health` require authentication via session cookies.

#### Login

```bash
POST /auth/login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "password123"
}
```

#### Logout

```bash
POST /auth/logout
```

### Metrics

#### Get Current Metrics

```bash
GET /metrics
Authorization: Required

Response:
{
    "cpu_pct": 45.2,
    "mem_used_pct": 67.8,
    "disk_used_pct": 23.1,
    "uptime_s": 3600.5
}
```

#### WebSocket Metrics Stream

```bash
GET /ws
Authorization: Required

# Streams metrics every 3 seconds
```

### System Info API

#### Get System Information

```bash
GET /system/info

Response:
{
    "hostname": "server01",
    "platform": "linux",
    "platform_version": "ubuntu 22.04",
    "kernel_version": "5.15.0-72-generic",
    "uptime_s": 3600,
    "cpu_cores": 4,
    "memory_total_mb": 8192,
    "disk_total_gb": 256,
    "last_boot_time": 1640995200
}
```

Example using curl:

```bash
curl -X GET http://localhost:8080/system/info
```

### Alert Management

#### List Alert Rules

```bash
GET /alerts/rules
Authorization: Required

Response:
[
    {
        "id": 1,
        "name": "High CPU Usage",
        "metric": "cpu_pct",
        "threshold_pct": 80.0,
        "comparison": "above",
        "trigger_after": 3,
        "created_at": "2025-10-07T01:00:00Z",
        "updated_at": "2025-10-07T01:00:00Z"
    }
]
```

#### Create Alert Rule

```bash
POST /alerts/rules
Authorization: Required
Content-Type: application/json

{
    "name": "High CPU Usage",
    "metric": "cpu_pct",
    "threshold_pct": 80.0,
    "comparison": "above",
    "trigger_after": 3
}

# Valid metrics: "cpu_pct", "mem_used_pct", "disk_used_pct"
# Valid comparisons: "above", "below"
# threshold_pct: 0-100
# trigger_after: >= 1
```

#### Update Alert Rule

```bash
PUT /alerts/rules/{id}
Authorization: Required
Content-Type: application/json

{
    "name": "Very High CPU Usage",
    "metric": "cpu_pct",
    "threshold_pct": 90.0,
    "comparison": "above",
    "trigger_after": 2
}
```

#### Delete Alert Rule

```bash
DELETE /alerts/rules/{id}
Authorization: Required
```

#### List Alert Events

```bash
GET /alerts/events?limit=50
Authorization: Required

Response:
[
    {
        "id": 1,
        "rule_id": 1,
        "triggered_at": "2025-10-07T01:30:00Z",
        "value": 85.2,
        "acknowledged": false,
        "acknowledged_at": null
    }
]

# Events are ordered by: unacknowledged first, then by triggered_at DESC
# Default limit: 50
```

#### Acknowledge Alert Event

```bash
POST /alerts/events/{id}/ack
Authorization: Required
```

## Alert System Behavior

### How Alerts Work

1. **Rule Evaluation**: Alert rules are evaluated against every metric sample (both HTTP `/metrics` requests and WebSocket streams)

2. **Consecutive Breaches**: Alerts only fire after `trigger_after` consecutive samples breach the threshold

3. **Recovery**: When a metric returns to normal levels, the consecutive breach counter is reset

4. **Logging**: All alert events are logged with format:

   ```
   [ALERT] {rule_name} {comparison} {threshold}% for {trigger_after} samples (value={actual_value}) - Event ID: {id}
   ```

### Example Alert Scenarios

#### High CPU Alert

```json
{
    "name": "High CPU Usage",
    "metric": "cpu_pct", 
    "threshold_pct": 80.0,
    "comparison": "above",
    "trigger_after": 3
}
```

This rule will fire an alert when CPU usage is above 80% for 3 consecutive metric samples.

#### Low Memory Alert

```json
{
    "name": "Low Available Memory",
    "metric": "mem_used_pct",
    "threshold_pct": 20.0, 
    "comparison": "below",
    "trigger_after": 2
}
```

This rule will fire when memory usage drops below 20% for 2 consecutive samples.

### Testing Alert Rules

1. Create a test rule with low threshold:

   ```bash
   curl -X POST http://localhost:8080/alerts/rules \
     -H "Content-Type: application/json" \
     -d '{
       "name": "Test CPU Alert",
       "metric": "cpu_pct",
       "threshold_pct": 1.0,
       "comparison": "above", 
       "trigger_after": 2
     }'
   ```

2. Make several `/metrics` requests or connect via WebSocket to trigger the rule

3. Check server logs for alert messages

4. List events to see fired alerts:

   ```bash
   curl http://localhost:8080/alerts/events
   ```

## Development Notes

- Database migrations run automatically on startup
- Alert rule state is maintained in memory and refreshed every 30 seconds
- Foreign key constraints are enabled for proper cascade deletes
- All alert endpoints require authentication
- Alert evaluation is integrated into both HTTP and WebSocket metric collection

## Database Schema

The alert system adds two new tables:

- `alert_rules`: Stores alert rule definitions
- `alert_events`: Stores fired alert events with foreign key to rules

Rule deletion cascades to delete related events.
