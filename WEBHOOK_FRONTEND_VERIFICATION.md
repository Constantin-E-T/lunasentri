# Frontend Webhook Rate Limiting - Manual Verification Guide

## Testing the Cooldown & Rate Limiting UI

### Prerequisites

1. Backend running with rate limiting enabled
2. Frontend dev server running: `pnpm --filter web-next dev`
3. Logged in user account
4. At least one webhook configured

### Test Cases

#### Test 1: Rate Limiting (30-second minimum between attempts)

**Steps:**

1. Navigate to Settings page
2. Find an active webhook
3. Click "Send Test" button
4. Immediately try to click "Send Test" again

**Expected Behavior:**

- ✅ Button becomes disabled
- ✅ Tooltip shows: "Rate limited, retry in Xs"
- ✅ Countdown updates every second (25s, 24s, 23s...)
- ✅ "Last attempt" timestamp shows "just now" or "1s ago"
- ✅ Button re-enables after 30 seconds

**Screenshot Areas:**

- Disabled button with tooltip
- Live countdown timer
- Last attempt timestamp

---

#### Test 2: Circuit Breaker Cooldown (after 3 failures)

**Steps:**

1. Configure webhook to point to a failing endpoint (e.g., `https://httpstat.us/500`)
2. Trigger 3 failures within 10 minutes (either via test button or actual alerts)
3. Observe webhook card after 3rd failure

**Expected Behavior:**

- ✅ Red banner appears: "🚫 Circuit breaker active: cooling down until HH:MM"
- ✅ Status badge changes to "Cooling Down" with red border
- ✅ "Send Test" button disabled
- ✅ Tooltip shows: "Cooling down until HH:MM"
- ✅ After 15 minutes, cooldown clears automatically

**Screenshot Areas:**

- Cooldown banner (red alert at top of card)
- "Cooling Down" status badge
- Disabled button with cooldown tooltip

---

#### Test 3: Failure Count Visual Indicators

**Steps:**

1. Create multiple webhooks with different failure counts:
   - Webhook A: 0 failures
   - Webhook B: 1-2 failures
   - Webhook C: 3+ failures (not in cooldown yet)

**Expected Behavior:**

**Webhook A (0 failures):**

- ✅ Green badge: "Active"
- ✅ No warning icons

**Webhook B (1-2 failures):**

- ✅ Amber/yellow badge: "Active • X failures"
- ✅ Amber ⚠ icon next to failure count

**Webhook C (3+ failures):**

- ✅ Red badge: "Active • High Failures"
- ✅ Red ⚠ icon next to failure count

**Screenshot Areas:**

- Side-by-side comparison of different status badges
- Failure count indicators with color-coded icons

---

#### Test 4: 429 Error Handling

**Steps:**

1. Get webhook into cooldown state (3 failures)
2. Try to send test via API or button
3. Check browser console and toast notifications

**Expected Behavior:**

- ✅ Toast notification appears: "Test failed" with backend error message
- ✅ Error includes text like "Webhook in cooldown until..." or "Rate limit active, can retry in Xs"
- ✅ Console shows no unhandled promise rejections
- ✅ Webhook list refreshes to show updated cooldown state

**Screenshot Areas:**

- Toast notification with 429 error message
- Browser console (should be clean)

---

#### Test 5: Last Attempt Timestamp

**Steps:**

1. Send test webhook
2. Wait 2 minutes
3. Observe "Last attempt" field

**Expected Behavior:**

- ✅ Initially shows "just now"
- ✅ Updates to "2m ago" after 2 minutes
- ✅ Clock icon (⏱) displayed next to timestamp
- ✅ Distinct from "Last success" and "Last error" fields

**Screenshot Areas:**

- Metadata section showing all three timestamps (attempt, success, error)

---

## UI Component Reference

### Cooldown Banner

```
┌─────────────────────────────────────────────────────────┐
│ 🚫 Circuit breaker active: cooling down until 14:35    │  ← Red background
└─────────────────────────────────────────────────────────┘
```

### Status Badges

```
┌────────────┐  ┌──────────────────────┐  ┌────────────────────────┐
│   Active   │  │ Active • 2 failures  │  │   Cooling Down        │
└────────────┘  └──────────────────────┘  └────────────────────────┘
   (Green)           (Amber)                    (Red border)
```

### Button States

```
┌─────────┐  ← Enabled (blue/primary color)
│  Test   │
└─────────┘

┌─────────┐  ← Disabled (grayed out, shows tooltip on hover)
│  Test   │
└─────────┘
```

### Metadata Row

```
Secret: ••••2345  |  ⏱ Last attempt 2m ago  |  ✓ Last success 5m ago  |  ⚠ 2 failures
```

---

## Common Issues & Troubleshooting

### Issue: Button not disabling during rate limit

**Check:**

- `last_attempt_at` field in API response
- Browser console for computed `retryAfterSeconds` value
- Component re-render (countdown timer should trigger updates)

### Issue: Cooldown banner not appearing

**Check:**

- `cooldown_until` field in API response (should be future ISO timestamp)
- Backend correctly setting cooldown after 3 failures
- `isCoolingDown` computed property in React DevTools

### Issue: Countdown not updating

**Check:**

- Browser console for errors
- `useEffect` cleanup (timer should re-initialize when `retryAfter` changes)
- Component mounting/unmounting cycles

---

## Backend Integration Points

The frontend expects these fields from `GET /notifications/webhooks`:

```json
{
  "id": 1,
  "url": "https://example.com/webhook",
  "is_active": true,
  "cooldown_until": "2025-10-08T14:35:00Z",    // ← Required (null if not cooling down)
  "last_attempt_at": "2025-10-08T14:20:00Z",   // ← Required (null if never attempted)
  "failure_count": 3,
  "last_success_at": "2025-10-08T14:00:00Z",
  "last_error_at": "2025-10-08T14:20:00Z",
  "secret_last_four": "2345",
  "created_at": "2025-10-01T00:00:00Z",
  "updated_at": "2025-10-08T14:20:00Z"
}
```

The frontend expects 429 responses from `POST /notifications/webhooks/{id}/test`:

```json
HTTP/1.1 429 Too Many Requests
Content-Type: application/json

{
  "error": "Webhook in cooldown until 2025-10-08T14:35:00Z"
}
```

OR

```json
{
  "error": "Rate limit active, can retry in 25s"
}
```
