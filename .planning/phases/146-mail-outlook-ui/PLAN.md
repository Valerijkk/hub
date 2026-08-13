# Phase 146 — Outlook-class Mail UI

> Task: TASK-239 · Status: PLANNED · Depends on: Phases 143, 144

## Ordered tasks

1. Extend the shared API client with typed stable per-request headers for every create/bulk
   `Idempotency-Key` and
   `If-Match`, then add typed mail API/query keys. Use repository-native fetch/client mocks
   for genuine units unless a separately approved dependency change introduces MSW.
2. Add `/mail` redirect, folder/thread/search routes, owner-scoped draft/message deep links
   and responsive folder/list/reader panes; URL is selection truth. Coordinate the
   calendar leaf-route split instead of masking it with mail-local state.
3. Surface account provisioning/error plus projection bootstrapping/degraded/rebuilding
   state conservatively across folder,
   thread and search pages. Keep current owner-filtered rows visible, label them stale or
   in progress, and let Retry invalidate queries without pretending to control the worker.
4. Add thread/message reader with sanitized content boundary, explicit token-only blocked-
   remote-image action,
   attachment states, keyboard/focus behavior and invitation card.
5. Add Plate mail composer wrapper, address chips, reply/forward quoting, signatures,
   alias/CC/BCC/priority, draft autosave conflict state and submission polling/retry that
   recovers `unknown` without a second send.
6. Add typed message/thread bulk actions with explicit classifier outcome, full label/folder
   management (including versioned `creating`/`unknown`/`deleting` recovery states), body/
   filename search filters, contact suggestions and realtime invalidation.
7. Add user settings/signatures/rules deployment and admin engine health/domain/per-user
   alias surfaces.
8. Add genuine Vitest/RTL units in `frontend/tests/mail/**`, Docker-only verification,
   accessibility pass and docs sync.

Primary production paths are `frontend/src/api/mail.ts`, `frontend/src/modules/mail/**`
and narrow route/sidebar/settings integrations. Do not introduce a duplicate Zustand
mailbox cache.
