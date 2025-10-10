# UI/UX Guardrails – Dialog & Error Patterns

**Effective:** 2025-10-10  
**Applies to:** `apps/web-next/**`

## Non-Negotiables

- **Browser dialogs are forbidden.** Never call `alert()`, `confirm()`, or `prompt()`.
- **All confirmations use `@/components/ui/dialog`.** Destructive actions must include item name, irreversible warning, and destructive confirm button.
- **Errors render inline.** Use styled error containers inside the form/modal; do not close the UI on failure.
- **Async work shows progress.** Buttons flip to "Saving…", "Deleting…", etc., and remain disabled until the request resolves.
- **Data parity.** Every database field must have create → view → edit coverage in the UI.

See `docs/development/UI_GUIDELINES.md` for code snippets and visual examples.

## Staging QA Checklist

1. Log in to staging, visit `/machines`.
2. Register a machine, trigger validation error (blank name) → error must appear inline.
3. Delete a machine → dialog copy references the machine name, confirm button shows "Deleting…" while pending.
4. Repeat the pattern on `/users` (create + delete) and any other destructive flow touched in the release.
5. Verify no native dialogs appear in the browser devtools console (search for `alert`/`confirm`).

Once this checklist passes, the release is ready for production.

