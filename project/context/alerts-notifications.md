# Alerts & Notification Channels

**TL;DR** – Dashboard is real-time; next focus is outbound notifications (user-scoped webhooks, later email/Telegram).

**Decisions**
- Prioritize per-user webhook delivery as the first outbound channel.
- Treat notification fan-out via notifier interface; future channels plug into the same pipeline.
- Require signed payloads + secret verification for every webhook POST.

**Open Items**
- Webhook storage: table scoped by user (url, secret_hash, last_success_at, last_error_at, is_active).
- Payload contract: include rule + event fields (`rule_id`, `rule_name`, `metric`, `comparison`, `threshold_pct`, `value`, `triggered_at`, `server_id?` future) and HMAC signature header (`X-LunaSentri-Signature`).
- Delivery policy: immediate attempt with 2 retries (exponential backoff), log failures for visibility.

**Next Steps**
- [ ] Backend: implement webhook persistence, notifier, signed delivery, retry + logging.
- [ ] Frontend: settings UI for managing webhook URLs/secrets + test payload action.
- [ ] QA: integration tests for signature verification + notifier fan-out.
