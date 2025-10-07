# Alerts & Notification Channels

**TL;DR** – Dashboard is real-time; next focus is outbound notifications (user-scoped webhooks, later email/Telegram).

**Decisions**
- Prioritize per-user webhook delivery as the first outbound channel.
- Treat notification fan-out via notifier interface; future channels plug into the same pipeline.
- Require signed payloads + secret verification for every webhook POST.

**Open Items**
- Design DB schema for user-owned webhook configs.
- Define alert payload contract and signature headers.
- Decide retry/backoff policy and dead-letter handling.

**Next Steps**
- [ ] Draft backend task: webhook storage, notifier integration, signed delivery.
- [ ] Draft frontend task: user settings UI for managing webhook URLs.
- [ ] Add testing plan covering webhook delivery + signature validation.
