# Webhook Notifications Service

## Overview

This package implements a robust webhook notification system for LunaSentri alerts with the following features:

- **HMAC-SHA256 Signature Verification**: Each webhook payload is signed with the webhook's secret
- **Exponential Backoff Retry Logic**: Up to 3 attempts with 1s, 2s, 4s delays
- **Context Cancellation Support**: Respects context deadlines and cancellation
- **Comprehensive Testing**: Full test suite with mock HTTP servers and storage

## Core Components

### 1. AlertNotifier Interface

```go
type AlertNotifier interface {
    Send(ctx context.Context, rule storage.AlertRule, event storage.AlertEvent) error
}
```

Clean interface for dependency injection into the alerts service.

### 2. Notifier Struct

The main webhook notification service with dependencies:

- `storage.Store`: Database operations for webhooks and failure tracking
- `http.Client`: HTTP client with 5-second timeout
- `log.Logger`: Structured logging

### 3. Webhook Payload Format

```json
{
    "rule": {
        "id": 1,
        "name": "High CPU Usage",
        "description": "CPU usage exceeded threshold"
    },
    "event": {
        "id": 123,
        "rule_id": 1,
        "value": 85.5,
        "triggered_at": "2025-01-01T12:00:00Z"
    },
    "timestamp": "2025-01-01T12:00:00Z"
}
```

### 4. HMAC Signature

Each request includes the `X-LunaSentri-Signature` header:

```
X-LunaSentri-Signature: sha256=<hex_encoded_hmac>
```

The HMAC is computed using SHA-256 with the webhook's secret over the JSON payload.

## Retry Logic

1. **Initial Attempt**: Immediate webhook delivery
2. **Retry 1**: 1 second delay after first failure
3. **Retry 2**: 2 seconds delay after second failure
4. **Final Failure**: Mark webhook delivery as failed after 3 attempts

Exponential backoff helps with temporary network issues while respecting webhook endpoints.

## Error Handling

- **HTTP Errors**: Non-2xx status codes trigger retries
- **Network Errors**: Connection timeouts and DNS failures trigger retries
- **Context Cancellation**: Immediate termination when context is cancelled
- **Permanent Failures**: Logged after all retry attempts exhausted

## Integration

The notifier integrates with the alerts service through dependency injection:

```go
// In main.go
notifier := notifications.NewNotifier(store, logger)
alertsService := alerts.NewService(store, logger, notifier)
```

When alerts fire, notifications are sent asynchronously with a 30-second timeout.

## Testing

Comprehensive test suite covers:

- ✅ Successful webhook delivery
- ✅ Retry logic with exponential backoff
- ✅ Maximum retry attempts exceeded
- ✅ Context cancellation during retries
- ✅ HMAC signature generation and verification
- ✅ No active webhooks scenario
- ✅ Inactive webhooks filtering

All tests use mock HTTP servers and storage for isolated testing.

## Usage Example

```go
// Create notifier
notifier := notifications.NewNotifier(store, logger)

// Send alert notification
rule := storage.AlertRule{ID: 1, Name: "High CPU"}
event := storage.AlertEvent{ID: 123, RuleID: 1, Value: 85.5}

ctx := context.Background()
err := notifier.Send(ctx, rule, event)
if err != nil {
    log.Printf("Failed to send notification: %v", err)
}
```

## Security Considerations

- Webhook secrets are stored as hashed values in the database
- HMAC signatures prevent payload tampering
- 5-second HTTP timeout prevents hanging requests
- Context cancellation enables request cleanup
