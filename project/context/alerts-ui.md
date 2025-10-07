# Alerts UI Enhancements

**TL;DR** – We need clear severity colors across dashboard + alert pages and quick-start templates so users can create common alert rules with one click.

**Decisions**

- Severity tiers: Good (green), Warn (amber), Critical (red). Use existing design tokens when possible.
- Active Alerts list + badge should reflect severity color rather than single red tone.
- Provide frontend-only templates first (UI pre-fills form); backend changes can follow later if needed.

**Open Items**

- [x] Define exact color tokens (tailwind classes) and ensure accessibility contrast.
- [x] Confirm which metrics get templates (CPU, Memory, Disk) and suggested thresholds.
- [x] Decide whether templates auto-create rules immediately or open modal with pre-filled values.

**Next Steps**

- [x] Update dashboard/alerts UI components to map metric severity → badge styles + background colors.
- [x] Add alert rule templates UX (list of chips/buttons) that pre-fill the Add Rule form.
- [x] Document behavior and edge cases in `project/logs/agent-b.md` after implementation.
