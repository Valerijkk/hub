# FIX-W13 — FE dead Telegram login + UX/a11y states

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** fe1  
**Правила:** 1 commit / finding · no push master · Hub tokens / Shadcn className-only

Сфокус: **A11-003** (мёртвый TelegramButton) + быстрые high-value loading/empty/error и a11y без redesign.

---

## Итог

| ID | Status | SHA | Notes |
|----|--------|-----|-------|
| A11-003 | FIXED | `5112ab2` | Deleted `telegram-button.tsx`; login hint → i18n `auth.magicLinkOnly`; absence test (Vite-safe polish `555252e`) |
| A11-004 | FIXED* | `7687f20` / `375783b` | *Code already removed prior wave; W13 added regression `auth-api-telegram-absent.test.ts` |
| A10-006 | FIXED | `04d98f8` | Admin users: `useQuery` + `QueryStateView` (error ≠ empty); invite fetch cancelled; search `aria-label` (test polish `4ae5e20`) |
| A10-005 | FIXED | `bba4666` | Channel list: `isError` + retry; Bookmarks/CreateChannel `aria-label` (A10-013 partial) |
| A10-010 | FIXED | `ed8cff6` | Mention autocomplete: `role="option"` + `aria-selected` + listbox `aria-label` |
| A10-007 | DEFERRED | — | notification-store → InfiniteQuery (larger refactor) |
| A10-008 | FIXED | `015cc4c` (A13-014) | Kanban column virtualization (`@tanstack/react-virtual` at 24+ cards) — closed via A13-014 alias |
| A10-009 | FIXED | `bbdb5fb` | `time-tracker.tsx`: narrow `useTimerStore` selectors; tick effect deps = `[isTimerForThisIssue, tick]` |
| A10-011 | PARTIAL | `a3e35cc` | Gantt wrapped in `QueryStateView` (error ≠ empty); notifications still zustand — full rollout deferred |
| A10-012 | FIXED | `866fb7f` | Deleted unused `modules/tracker/kanban-board.tsx` (zero importers; canon = board route) |
| A10-013 | FIXED | `a3e35cc` | Mute indicator + star `aria-label` (title kept); bookmarks remove label; Bookmarks/CreateChannel already done W13 |
| A10-014 | FIXED | pending | Gantt: `@tanstack/react-virtual` rows at ≥24 + cursor `useInfiniteQuery` (auto-fetch pages) + «показано N из M»; Vitest threshold assert |
| A15-004 | DEFERRED | — | REQ LOGN-05 / wiki drift (docs, not FE code) |
| BE LoginWithTelegram* | DEFERRED | — | Backend service/DTO — out of fe1 scope (cross-cut A18-010) |

---

## A11-003 detail

- Removed forbidden Telegram login UI (ADR-AUTH-01 / TZ Magic Link only).
- i18n: `common.auth.magicLinkOnly` + `telegramLoginRemoved` (ru/en); login page uses Magic Link copy.
- Notification Telegram bind (`telegram-link-card`) **kept** — not login.

## A10-006 detail

`GET /admin/users` failures no longer render as «Пользователи не найдены». Mutations patch Query cache via `setQueryData`.

## Tests run

```
npx vitest run …/auth/telegram-button-absent.test.ts
  …/auth/auth-api-telegram-absent.test.ts
  …/admin/user-management.test.tsx
  …/messenger/channel-list-error.test.tsx
  …/messenger/mention-autocomplete-a11y.test.tsx
→ 5 files / 9 tests PASS
```

`npx tsc --noEmit` (frontend) PASS after A10-006 edit.
