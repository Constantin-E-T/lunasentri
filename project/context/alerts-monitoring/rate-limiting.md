# Webhook Rate Limiting & Circuit Breakers

**TL;DR** – Webhooks retry automatically today (1s/2s/4s). ✅ **COMPLETED**: Added rate limiting and circuit breaker protection for webhook deliveries.

**Goals** ✅ Complete
- Avoid spamming a failing webhook (back off quickly after repeated failures).
- Provide visibility (failure counts, lockout timers) to users and ops.
- Keep implementation simple enough to fit the current SQLite-backed service.

**Implementation Complete** ✅

1. **Storage Schema**: Added `cooldown_until` and `last_attempt_at` columns to webhooks table.
2. **Rate Limiting**: Enforces 30-second minimum interval between delivery attempts.
3. **Circuit Breaker**: After 3 failures within 10 minutes, enters 15-minute cooldown.
4. **API Behavior**: Returns 429 status for rate-limited requests with descriptive messages.
5. **Comprehensive Logging**: All rate limit and cooldown events are logged.

**Decisions Made**

- **Failure Threshold**: 3 failures within 10-minute window triggers cooldown
- **Cooldown Duration**: 15 minutes after reaching failure threshold
- **Rate Limit**: Minimum 30 seconds between delivery attempts
- **API Response**: 429 status code for both cooldown and rate limit violations
- **Storage**: Added two new nullable columns to existing webhooks table

**API Response Changes**

Webhook list responses now include:
```json
{
  "id": 1,
  "url": "https://example.com/webhook",
  "is_active": true,
  "cooldown_until": "2025-01-01T15:00:00Z",
  "last_attempt_at": "2025-01-01T14:30:00Z",
  "failure_count": 2,
  "last_success_at": "2025-01-01T14:00:00Z",
  "last_error_at": "2025-01-01T14:30:00Z"
}
```

**Curl Examples**

```bash
# Test webhook during cooldown (returns 429)
curl -X POST http://localhost:8080/notifications/webhooks/1/test \
  -H "Cookie: lunasentri_session=..." \
  -v

# Expected response:
# HTTP/1.1 429 Too Many Requests
# {"error":"Webhook in cooldown until 2025-01-01T15:00:00Z"}

# Test webhook too soon after last attempt (returns 429)
curl -X POST http://localhost:8080/notifications/webhooks/1/test \
  -H "Cookie: lunasentri_session=..." \
  -v

# Expected response:
# HTTP/1.1 429 Too Many Requests  
# {"error":"Rate limit active, can retry in 25s"}
```

**Logging Examples**

```
[WEBHOOK] throttled webhook=123 reason=cooldown until=2025-01-01T15:00:00Z
[WEBHOOK] rate_limited webhook=123 delay=25s
[WEBHOOK] cooldown webhook=123 until=2025-01-01T15:00:00Z
[WEBHOOK] delivered webhook=123 url=example.com status=200 attempt=1
```
