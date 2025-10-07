# Alerts Page Refactor & Delete UX

**TL;DR** – `apps/web-next/app/alerts/page.tsx` has been split into focused modules and now uses a shadcn-style delete dialog. Future work should keep the composition lean.

**Decisions**
- Extract UI pieces into components under `apps/web-next/components/alerts/` (form, templates, events list, rules table, delete dialog).
- Centralize modal/dialog logic using shadcn/ui primitives (`AlertDialog` or `Dialog`) instead of `window.confirm`.
- Keep data fetching within a dedicated hook where practical (`useAlertsManagement`).

**Component Map**
- `apps/web-next/components/alerts/AlertRuleForm.tsx`
- `apps/web-next/components/alerts/AlertRuleTemplates.tsx`
- `apps/web-next/components/alerts/AlertRulesTable.tsx`
- `apps/web-next/components/alerts/AlertEventsList.tsx`
- `apps/web-next/components/alerts/DeleteAlertDialog.tsx`
- `apps/web-next/lib/alerts/useAlertsManagement.ts`

**Notes**
- Delete confirmation uses `DeleteAlertDialog` (shadcn `AlertDialog`) with toast feedback.
- `useAlertsManagement` centralizes CRUD + modal state.
- Page composition down to ~376 lines; continue moving new UI into feature folder.

**Follow-ups**
- Add targeted component tests if we extend functionality (templates, delete dialog, etc.).
