---
applyTo: "apps/web-next/**"
---

- Use pnpm, not npm.
- App Router; components typed; charts via `recharts`.
- Read API base from `process.env.NEXT_PUBLIC_API_URL`.

## UI/UX Guidelines

### User Confirmations

- **NEVER use `alert()` or `confirm()` for user interactions** - these are browser-native dialogs that break the UI flow
- Always use proper Dialog components from `@/components/ui/dialog` for confirmations
- For destructive actions (delete, etc.), use a confirmation Dialog with:
  - Clear warning message about what will be deleted/changed
  - "Cannot be undone" warning if applicable
  - Cancel button (outline variant)
  - Confirm button (destructive variant for delete operations)
  - Loading state while action is processing

### Error Handling

- Display errors inline within modals/forms, not as alerts
- Use error state styling: `bg-destructive/20 border border-destructive/30 text-destructive`
- Keep modals open on error so users can retry or see what went wrong

### Data Display

- If a field exists in the data model (database), it should be:
  1. Included in create/registration forms
  2. Displayed in list/card views (conditionally if optional)
  3. Editable in update/edit forms
- Don't create database fields without UI support for viewing/editing them
