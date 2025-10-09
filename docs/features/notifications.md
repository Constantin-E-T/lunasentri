# Notifications System

Guide to LunaSentri's multi-channel notification system for alert delivery.

---

## Overview

LunaSentri supports two notification channels: **Webhooks** and **Telegram**. Both integrate with the alert system to deliver real-time notifications.

## Webhook Notifications

### Features

- **HMAC-SHA256 Signatures**: Secure payload verification
- **Exponential Backoff**: 3 retry attempts with 1s, 2s, 4s delays
- **Circuit Breaker**: 15-minute cooldown after 3 failures
- **Rate Limiting**: 30-second minimum between test attempts

### Setup

1. Navigate to Settings → Notifications
2. Click "Add Webhook"
3. Enter HTTPS URL
4. Generate secret (16-128 characters)
5. Toggle active status
6. Click "Test" to verify

### Payload Format

```json
{
  "rule": {
    "id": 1,
    "name": "High CPU Usage",
    "description": "CPU exceeded 80%"
  },
  "event": {
    "id": 123,
    "rule_id": 1,
    "value": 85.5,
    "triggered_at": "2025-10-09T12:00:00Z"
  },
  "timestamp": "2025-10-09T12:00:00Z"
}
```

### Signature Verification

```
X-LunaSentri-Signature: sha256=<hex_encoded_hmac>
```

Verify using HMAC-SHA256 with your webhook secret.

## Telegram Notifications

### Setup

**Admin (One-time):**

1. Message @BotFather on Telegram
2. Send `/newbot` and follow prompts
3. Copy bot token
4. Set `TELEGRAM_BOT_TOKEN` environment variable
5. Restart backend

**Users:**

1. Start the bot on Telegram
2. Get chat ID from bot updates
3. Navigate to Settings → Telegram
4. Add chat ID
5. Click "Test" to verify

### Message Format

```
🚨 *LunaSentri Alert*

*Rule:* High CPU Usage
*Metric:* cpu_pct
*Condition:* above 80.0%
*Current Value:* 85.5%
*Triggered:* 2025-10-09 15:04:05

_Alert triggered after 3 consecutive samples_
```

## Managing Notifications

### Webhooks

- **View all webhooks** with status indicators
- **Test delivery** before alerts trigger
- **Enable/disable** without deleting
- **Monitor failures** with count tracking
- **Circuit breaker** prevents spam to failing endpoints

### Telegram

- **Add multiple chat IDs** per user
- **Test messages** to verify configuration
- **Enable/disable** individual recipients
- **Failure tracking** with success timestamps

## API Reference

See detailed API documentation:

- [Webhook Implementation](implementation/webhook-notifications.md)
- [Telegram Implementation](implementation/telegram-notifications.md)

---

**Status: Production Ready** 📬
