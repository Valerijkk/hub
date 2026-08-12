# TASK-239 — Outlook-class mail UI

> Date: **2026-08-13** · Priority: **P0** · Status: **Open** · Phase: **146**

## Outcome

Deliver the responsive three-pane mail workspace, thread reader, rich composer, attachment
workflow, labels/folders/bulk actions, search, autocomplete, invitation card, realtime
counters, and settings surfaces.

## Owned paths

- `frontend/src/api/mail.ts`
- `frontend/src/modules/mail/**`
- mail-owned route/sidebar/settings/admin additions
- `frontend/tests/mail/**`

## Acceptance

- Desktop/medium/mobile navigation follows the approved design and keeps selection in URL.
- TanStack Query is server truth; transient reducer state does not duplicate mailbox cache.
- Composer autosave coalesces edits, detects revision conflict, and never blind-retries send.
- Plate editor is reused through a mail wrapper with deterministic HTML/plain serialization.
- Remote content is blocked by default; sanitized reader never fetches it directly.
- RSVP card covers current, cancelled, stale-sequence, conflict, untrusted and recurrence
  scope states with accessible announcements.
- Keyboard/focus/touch/accessibility contract is implemented with Vitest/RTL units in
  `frontend/tests/mail/**` and no legacy static-contract tests.
