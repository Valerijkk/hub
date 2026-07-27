# FIX-W11 — API contract FE↔BE (Block C)

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** fe1  
**Правила:** 1 commit / finding · minimal FE/BE mismatch fix · mutation/regression where applicable · **no Ent schema** · no push master

Источник: `11-api-contract.md` (A11-004…A11-013) + spot-check critical domains.

## Findings A11-004…A11-013

| ID | Sev | Status | SHA | Notes |
|----|-----|--------|-----|-------|
| A11-004 | P1 | FIXED | `7687f20` | Removed dead `loginTelegram` + `TelegramAuthRequest`. Follow-up assert tweak `375783b`. TZ Magic-Link-only. |
| A11-005 | P1 | FIXED | `ebfc6b0` | Removed dead `updateChannelNotifyMode` + FE `notify_mode` on member DTO. BE route absent; mute/unmute remain. Full notify_mode API = DEFERRED (needs BE handler; Ent field exists). |
| A11-006 | P1 | FIXED | `29526f6` | FE normalizes `YYYY-MM-DD` → RFC3339 UTC before PUT/onboarding; date input slices ISO. |
| A11-007 | P1 | FIXED | `8bb37ab` | `api.getBlob` + TMS export/run-export use blob transport (no `JSON.parse`). |
| A11-008 | P2 | FIXED | `0a4ddb7` | Dropped false `IssueResponse.attachments`; issue page passes `[]` (upload still via `/files`). List-via-files UI = DEFERRED. |
| A11-009 | P2 | FIXED | `b3cf6df` | `updateContact` / `deleteContact` → PUT/DELETE `/crm/contacts/:id`. |
| A11-010 | P2 | FIXED | `0d4e3ff` | `updateBase` / `deleteBase` → PUT/DELETE `/bases/:baseId`. |
| A11-011 | P2 | FIXED | `35310c7` | `updateListing` / `deleteListing` → PUT/DELETE `/marketplace/listings/:id`. |
| A11-012 | P2 | FIXED | `d631197` | `listDashboardTemplates` / `applyDashboardTemplate`. |
| A11-013 | P2 | FIXED | `9b32eaf` | `getRegistrationMode` / `setRegistrationMode` in `admin-org.ts`. |

## Adjacent A11 (out of 004…013 scope, spot-check)

| ID | Status | SHA / note |
|----|--------|------------|
| A11-001 | FIXED (prior) | `da3fe4b` — `/invites/join` |
| A11-002 | FIXED (prior) | `1d6d81a` / `c5be0c4` — FormData multipart |
| A11-003 | FIXED (parallel) | `5112ab2` — TelegramButton removed |
| A11-014 | DEFERRED | TZ doc drift (`error`/`request_id` vs live `code`/`message`/`request_id`); BE↔FE already agree |

## Critical domains sample (methods / fields)

| Domain | Verdict | Notes |
|--------|---------|-------|
| auth | OK after fixes | Magic-link / me / sessions / refresh match. Telegram login clients gone. Birthday date-only fixed. |
| tracker | OK after fixes | Join path OK. IssueResponse no longer lies about attachments. Boards/issues paths mirror BE. |
| tms | OK after fixes | Import multipart + export blob both use non-JSON transport. |
| messenger | OK after fixes | Notify-mode dead client removed. Mute/unmute/star paths match registered routes. |
| calls | OK (sample) | guest-join, instant, join/leave/end, raise-hand, chat, recording, scheduled-calls, conflicts — paths align with `calls/handler.go`. Webhook external (N/A FE). |

### Residual unused BE (sample, not bugs)

Teams CRUD, project trash/unarchive, RBAC `/roles`, imex HTTP, intranet coffee stats, messenger join/leave/members — still no FE wrappers (same as audit unused table). **DEFERRED** product work, not contract 404s from live UI.

## Tests

```text
npx vitest run …/auth …/tms …/messenger …/tracker …/dashboard
  + registration-mode / crm-contact / bases-update / marketplace-update
→ 72 tests passed in touched domains
```

Known unrelated FAIL: `telegram-button-absent.test.ts` (A11-003) — Vite transform resolves missing module before reject assertion (parallel residual).

## DEFERRED / N/A

| Item | Status | Why |
|------|--------|-----|
| BE `notify_mode` write API | DEFERRED | Needs handler/service; Ent field only; FE dead client removed |
| Issue attachments list in GET issue | DEFERRED | No BE DTO field; upload via `/files` works |
| Full 286+ route matrix dump | N/A | Sampled critical domains + closed known A11 IDs |
| A11-014 TZ wording | DEFERRED | Docs-only; do not break live AppError shape |
| Schema changes | N/A | Forbidden this wave |
