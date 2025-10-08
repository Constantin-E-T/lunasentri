# Webhook Rate Limiting & Cooldown - Frontend Implementation

**Date:** 2025-10-08  
**Agent:** Frontend (Agent B)  
**Status:** ✅ Complete

## Summary

Enhanced the webhook management UI to display and respect the backend's rate limiting and circuit breaker functionality. Users can now see when webhooks are in cooldown, understand rate limits, and receive clear feedback about why test sends may be blocked.

## Changes Made

### 1. Hook Updates (`lib/alerts/useWebhooks.ts`)

**New Fields:**

- `cooldown_until`: ISO timestamp when cooldown expires
- `last_attempt_at`: ISO timestamp of last delivery attempt

**New Computed Properties (via `WebhookWithState`):**

- `isCoolingDown`: Boolean - circuit breaker is active
- `cooldownUntil`: Date object when cooldown expires (or null)
- `retryAfterSeconds`: Number of seconds until rate limit expires (or null)
- `canSendTest`: Boolean - true only if not cooling down AND rate limit passed

**Enhanced Error Handling:**

- Dedicated 429 status handling with error message parsing
- Automatic refresh after test webhook (even on failure) to update state

### 2. UI Enhancements (`components/settings/notifications/WebhookList.tsx`)

**Visual Indicators:**

- **Cooldown Banner:** Red alert banner at top of card showing cooldown end time
- **Status Badges:**
  - "Cooling Down" - red border and background when circuit breaker active
  - "Active • X failures" - amber for 1-2 failures, red for ≥3
  - "Active" - green for healthy webhooks
  - "Inactive" - gray for disabled webhooks

**New Metadata Display:**

- Last attempt timestamp (e.g., "Last attempt 2m ago")
- Failure count with color coding (amber/red based on severity)
- Last success/error timestamps (existing, preserved)

**Smart Button Behavior:**

- "Send Test" button disabled when:
  - Webhook in cooldown (shows "Cooling down until HH:MM" tooltip)
  - Rate limit active (shows "Rate limited, retry in Xs" tooltip with live countdown)
- Live countdown timer updates every second for rate limit display

### 3. Type Safety Updates

Updated all components to use `WebhookWithState` instead of `Webhook`:

- `components/settings/notifications/WebhookForm.tsx`
- `app/settings/page.tsx`

## Testing

### Unit Tests (`__tests__/useWebhooks.test.ts`)

**New Tests Added:**

1. ✅ Fetch webhooks with cooldown field parsing
2. ✅ Parse cooldown state correctly (derives computed properties)
3. ✅ Handle 429 rate limit response with error message
4. ✅ Refresh after sending test webhook (updates state)

**Test Results:**

- 6/6 webhook hook tests passing
- 22/22 total frontend tests passing
- Build successful with TypeScript strict mode

### Manual Verification Checklist

The following should be verified with a running backend:

- [ ] Webhook enters cooldown after 3 failures within 10 minutes
- [ ] Cooldown banner displays with correct end time
- [ ] "Send Test" button disabled during cooldown with tooltip
- [ ] Rate limit countdown updates every second
- [ ] "Send Test" button re-enables after 30s rate limit expires
- [ ] 429 error messages surfaced in toast notifications
- [ ] Status badges update colors based on failure count (amber 1-2, red ≥3)
- [ ] Last attempt timestamp updates after each test send

## User Experience Flow

### Scenario: Webhook Failing Repeatedly

1. **0 Failures:** Green "Active" badge, "Send Test" enabled
2. **1-2 Failures:** Amber "Active • X failures" badge, "Send Test" still enabled (respecting 30s rate limit)
3. **3 Failures (within 10 min):**
   - Red "Cooling Down" badge
   - Red cooldown banner appears: "🚫 Circuit breaker active: cooling down until HH:MM"
   - "Send Test" button disabled with tooltip
4. **After 15 min cooldown:** Badge returns to appropriate state based on current failure count

### Scenario: Rate Limiting

1. User clicks "Send Test"
2. Backend responds, `last_attempt_at` updated
3. If user tries again <30s later:
   - "Send Test" disabled
   - Tooltip shows: "Rate limited, retry in 25s" (counts down)
4. After 30s: Button re-enabled automatically

## Files Changed

```
apps/web-next/
├── lib/alerts/useWebhooks.ts                                  # Added cooldown fields + 429 handling
├── components/settings/notifications/
│   ├── WebhookList.tsx                                        # Cooldown UI + rate limit countdown
│   └── WebhookForm.tsx                                        # Type update
├── app/settings/page.tsx                                      # Type update
└── __tests__/useWebhooks.test.ts                              # 4 new tests
```

## Documentation Updated

- `project/context/alerts-notifications.md` - Added cooldown UI section to frontend implementation
- `project/logs/agent-b.md` - New entry documenting this work

## Commands Run

```bash
# Test webhook hook
pnpm --filter web-next test -- __tests__/useWebhooks.test.ts

# Build verification
pnpm --filter web-next build

# Full test suite
pnpm --filter web-next test
```

All commands succeeded with no errors.

## Next Steps

**For Users:**

- No migration needed - changes are backward compatible
- New fields gracefully handle null values from older backend versions

**For Developers:**

- Manual verification recommended with backend returning 429 responses
- Consider adding E2E tests for cooldown scenarios if Playwright/Cypress added to project

**For Backend Team:**

- Ensure API responses include `cooldown_until` and `last_attempt_at` fields
- Verify 429 responses include descriptive `error` message in JSON body
