# Session Expiry UX

**TL;DR** – When the auth cookie expires the UI keeps running and all fetches 401; we need a shared handler to auto-logout, redirect to `/login`, and notify the user.

**Decisions**

- Centralize fetch logic in `apps/web-next/lib/api.ts` so 401/403 responses broadcast a `session-expired` event.
- `useSession` listens for that event, clears auth state, triggers redirect, and shows a toast (“Session expired. Please log in again.”).
- All API helpers must go through the central request wrapper.

**Open Items**

- [x] Consider adding automated test coverage to simulate session expiry in integration tests.

**Next Steps**

- [x] Implement frontend changes to enforce automatic logout + notification on session expiry.
- [x] Add regression check (build + manual test) and document behavior in `project/logs/agent-b.md`.
- [x] Add automated test coverage for session expiry flow to prevent regressions.
