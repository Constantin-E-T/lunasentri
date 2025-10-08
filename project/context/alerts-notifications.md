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
- **NEW**: Includes `cooldown_until` and `last_attempt_at` for rate limiting visibility
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

- [x] Frontend: build notifications settings UI to list/add/edit/delete webhooks and trigger test payloads.
- [x] Backend: add `POST /notifications/webhooks/{id}/test` endpoint to trigger a signed test payload for the owning user.
- [ ] Enhanced monitoring: rate limiting + circuit breaker patterns for webhook delivery.
- [ ] Additional channels: email, Slack, Telegram integrations.

**Test Webhook Endpoint**

**POST `/notifications/webhooks/{id}/test`**

Sends a test webhook notification to verify the configuration. Requires authentication and webhook ownership.

**Request:** No request body required.

**Success Response (200):**

```json
{
  "status": "sent",
  "webhook_id": 1,
  "triggered_at": "2025-10-08T12:34:56Z"
}
```

**Test Payload Example:**

```json
{
  "rule_id": 0,
  "rule_name": "Test Webhook",
  "metric": "cpu_pct",
  "comparison": "above",
  "threshold_pct": 80.0,
  "trigger_after": 1,
  "value": 85.5,
  "triggered_at": "2025-10-08T12:34:56Z",
  "event_id": 0
}
```

**Error Responses:**

- `401 Unauthorized`: User not authenticated
- `400 Bad Request`: Invalid webhook ID or webhook is inactive
- `404 Not Found`: Webhook not found or belongs to another user
- `502 Bad Gateway`: Test webhook delivery failed

**Security:**

- Validates user ownership of webhook
- Requires webhook to be active
- Uses 10-second timeout for delivery
- Updates success/failure tracking like normal alerts

**Example curl:**

```bash
# Send test webhook
curl -X POST http://localhost:8080/notifications/webhooks/1/test \
  -H "Cookie: lunasentri_session=..."
```

**Frontend Implementation (Completed)**

The webhook management UI has been fully implemented in the Settings page:

**Components Created:**

- `lib/alerts/useWebhooks.ts` - React hook for webhook CRUD operations with proper error handling
- `components/settings/notifications/WebhookList.tsx` - Displays webhooks with status pills and action buttons
- `components/settings/notifications/WebhookForm.tsx` - Modal form for creating/editing webhooks
- `components/settings/notifications/WebhookEmptyState.tsx` - Encourages first webhook setup
- `components/settings/notifications/DeleteWebhookDialog.tsx` - Confirmation dialog for deletion

**Features:**

- ✅ List all user webhooks with status (Active/Inactive) and failure tracking
- ✅ Create new webhooks with HTTPS URL validation and 16-128 char secret requirements
- ✅ Edit existing webhooks (URL, secret rotation, active toggle)
- ✅ Delete webhooks with confirmation dialog
- ✅ Send test payloads to verify webhook configuration
- ✅ Toast notifications for all operations (success/error feedback)
- ✅ Proper error handling with inline validation
- ✅ Glassmorphism design matching LunaSentri aesthetic

**User Experience:**

- Status pills show webhook health (emerald for active, amber for failures, red for high failures)
- Secret masking displays last 4 characters only (e.g., `••••2345`)
- Relative timestamps for last success/error (e.g., "2h ago", "5m ago")
- Form validation prevents HTTP URLs and enforces secret length requirements
- Empty state provides helpful onboarding information

**Testing:**

- ✅ Unit tests for `useWebhooks` hook covering success, error, and auth expiry cases
- ✅ All existing tests continue to pass
- ✅ Build verification successful with Turbopack
