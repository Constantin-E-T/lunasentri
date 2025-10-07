# LunaSentri Context Notes

> Purpose: keep feature-specific history lightweight, structured, and easy to pick up.

## Conventions
- Use this folder for deep-dive notes that would clutter the main agent logs.
- One file per active workstream (e.g., `alerts-notifications.md`).
- Start with a **TL;DR**, followed by **Decisions**, **Open Items**, and **Next Steps**.
- Link back to relevant roadmap items or PRs.
- Archive closed threads under `archive/` with a date suffix.

## Workflow
1. Update `docs/PLAN.md` and agent logs with high-level status.
2. Capture detailed context here; keep entries concise and scannable.
3. When work ships, move stale context to `archive/` and note the replacement.

## Template
```
# <Feature Name>

**TL;DR** – short summary.

**Decisions**
- Item...

**Open Items**
- Item...

**Next Steps**
- [ ] Checkbox task
```
