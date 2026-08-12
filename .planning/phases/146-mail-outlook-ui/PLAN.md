# Phase 146 — Outlook-class Mail UI

> Task: TASK-239 · Status: PLANNED · Depends on: Phases 143, 144

## Ordered tasks

1. Add typed API client/query keys and MSW fixtures matching OpenAPI.
2. Add route shell and responsive folder/list/reader panes; URL is selection truth.
3. Add thread/message reader with sanitized content boundary, blocked remote images,
   attachment states, keyboard/focus behavior and invitation card.
4. Add Plate mail composer wrapper, address chips, reply/forward quoting, signatures,
   alias/CC/BCC/priority, draft autosave conflict state and send-unknown recovery.
5. Add bulk actions, labels/folders, search filters, contact suggestions and realtime
   counter invalidation.
6. Add user/admin settings and engine health surfaces.
7. Add genuine Vitest/RTL units in `frontend/tests/mail/**`, Docker-only verification,
   accessibility pass and docs sync.

Primary production paths are `frontend/src/api/mail.ts`, `frontend/src/modules/mail/**`
and narrow route/sidebar/settings integrations. Do not introduce a duplicate Zustand
mailbox cache.
