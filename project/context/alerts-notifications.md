# Alerts & Notification Channels

**TL;DR** – Dashboard is real-time; webhook notification system implemented with full CRUD endpoints.

**Decisions**

- ✅ **COMPLETED**: Per-user webhook delivery as the first outbound channel
- ✅ **COMPLETED**: Signed payloads + secret verification for every webhook POST  
- ✅ **COMPLETED**: User-scoped webhook storage with secure secret hashing
- ✅ **COMPLETED**: Exponential backoff retry logic with failure tracking
- ✅ **COMPLETED**: Comprehensive test coverage for all components

**Implementation Complete**

- ✅ **Backend**: Webhook persistence, notifier, signed delivery, retry + logging
- ✅ **HTTP API**: Full CRUD endpoints for webhook management  
- ✅ **Storage**: User-scoped webhook table with secret hashing and failure tracking
- ✅ **Delivery**: HMAC-signed payloads with 3-attempt retry and exponential backoff
- ✅ **Security**: HTTPS-only URLs, 16-128 character secrets, safe logging
- ✅ **Testing**: Comprehensive test suite covering all functionality

**API Endpoints**

All webhook endpoints require authentication and are user-scoped:

**GET `/notifications/webhooks`**

- Returns list of current user's webhooks
- Response includes safe fields (id, url, is_active, failure tracking, secret_last_four)
- Secret hashes are never exposed

**POST `/notifications/webhooks`**

```json
{
  "url": "https://example.com/webhook",
  "secret": "mysecretkey12345",
  "is_active": true
}
```

- Creates new webhook with validation
- Requires HTTPS URL and 16-128 character secret
- Auto-generates secure hash and tracks secret's last 4 characters

**PUT `/notifications/webhooks/{id}`**

```json
{
  "url": "https://updated.com/webhook", 
  "secret": "newsecretkey12345",
  "is_active": false
}
```

- Updates existing webhook (user ownership verified)
- All fields optional - omitted fields remain unchanged
- New secret rotates hash and updates last 4 characters

**DELETE `/notifications/webhooks/{id}`**

- Removes webhook after verifying user ownership
- Returns 204 No Content on success

**Webhook Payload Format**

```json
{
  "rule_id": 1,
  "rule_name": "High CPU Usage",
  "metric": "cpu_pct", 
  "comparison": "above",
  "threshold_pct": 80.0,
  "trigger_after": 2,
  "value": 85.5,
  "triggered_at": "2025-01-01T12:00:00Z",
  "event_id": 123
}
```

**Security Headers**

- `Content-Type: application/json`
- `X-LunaSentri-Signature: sha256=<hmac-hex>` - HMAC-SHA256 of payload
- `User-Agent: LunaSentri-Webhook/1.0`

**Delivery Behavior**

- **Immediate**: First attempt on alert trigger
- **Retry Logic**: Up to 3 attempts with exponential backoff (1s, 2s, 4s)
- **Success Criteria**: HTTP 2xx response codes  
- **Failure Tracking**: Increments failure count and records last error time
- **Logging**: Structured logs with webhook ID and domain (never full URLs or secrets)

**Example curl commands**

```bash
# List webhooks
curl -X GET http://localhost:8080/notifications/webhooks \
  -H "Cookie: lunasentri_session=..." 

# Create webhook  
curl -X POST http://localhost:8080/notifications/webhooks \
  -H "Content-Type: application/json" \
  -H "Cookie: lunasentri_session=..." \
  -d '{"url":"https://example.com/webhook","secret":"mysecretkey12345"}'

# Update webhook
curl -X PUT http://localhost:8080/notifications/webhooks/1 \
  -H "Content-Type: application/json" \
  -H "Cookie: lunasentri_session=..." \
  -d '{"is_active":false}'

# Delete webhook
curl -X DELETE http://localhost:8080/notifications/webhooks/1 \
  -H "Cookie: lunasentri_session=..."
```

**Next Steps**

- [ ] Frontend: Settings UI for managing webhook URLs/secrets + test payload action
- [ ] Enhanced monitoring: Rate limiting and circuit breaker patterns  
- [ ] Additional channels: Email, Slack, Telegram integrations
