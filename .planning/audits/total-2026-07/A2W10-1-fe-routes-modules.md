# A2-W10 агент A1 — routes+modules part 1 (re-audit)

> TASK-008 · Stage A2 · read-only · 2026-07-28  
> HEAD: `0d6a659f` · Зеркало: [`A10-1-fe-routes-modules.md`](A10-1-fe-routes-modules.md) · [`W10-ASSIGNMENTS.md`](W10-ASSIGNMENTS.md) §A1  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| frontend/src/modules/accounting/accounting-page.tsx | 362 | 1 | RESIDUAL |
| frontend/src/modules/admin/backups.tsx | 204 | 1 | RESIDUAL |
| frontend/src/modules/bases/bases-page.tsx | 455 | 0 | FIXED-OK |
| frontend/src/modules/calendar/unified-calendar.tsx | 430 | 0 | чисто |
| frontend/src/modules/calls/calendar-agenda-view.tsx | 92 | 0 | чисто |
| frontend/src/modules/calls/calendar-day-view.tsx | 118 | 0 | чисто |
| frontend/src/modules/calls/calendar-month-view.tsx | 149 | 0 | чисто |
| frontend/src/modules/calls/calendar-types.ts | 184 | 0 | чисто |
| frontend/src/modules/calls/calendar-week-view.tsx | 162 | 0 | чисто |
| frontend/src/modules/calls/call-calendar.tsx | 490 | 0 | FIXED-OK |
| frontend/src/modules/calls/create-call-dialog.tsx | 529 | 0 | FIXED-OK |
| frontend/src/modules/calls/recording-controls.tsx | 118 | 0 | FIXED-OK |
| frontend/src/modules/calls/recordings-list.tsx | 191 | 1 | RESIDUAL |
| frontend/src/modules/calls/scheduled-calls.tsx | 292 | 0 | FIXED-OK |
| frontend/src/modules/crm/crm-page.tsx | 355 | 1 | RESIDUAL |
| frontend/src/modules/crm/deals-pipeline.tsx | 141 | 0 | чисто |
| frontend/src/modules/hr/leave-requests.tsx | 265 | 0 | FIXED-OK |
| frontend/src/modules/intranet/intranet-hub.tsx | 62 | 0 | FIXED-OK |
| frontend/src/modules/intranet/kudos-feed.tsx | 193 | 0 | FIXED-OK |
| frontend/src/modules/intranet/news-feed.tsx | 169 | 0 | чисто |
| frontend/src/modules/intranet/org-directory.tsx | 249 | 0 | чисто |
| frontend/src/modules/intranet/random-coffee.tsx | 119 | 0 | FIXED-OK |
| frontend/src/modules/marketplace/marketplace-page.tsx | 480 | 0 | FIXED-OK |
| frontend/src/modules/messenger/link-preview.tsx | 216 | 0 | чисто |
| frontend/src/modules/messenger/message-bubble.tsx | 202 | 0 | FIXED-OK |
| frontend/src/modules/settings/ai-privacy.tsx | 85 | 1 | RESIDUAL |
| frontend/src/modules/settings/api-keys.tsx | 200 | 1 | RESIDUAL |
| frontend/src/modules/settings/appearance.tsx | 132 | 0 | чисто |
| frontend/src/modules/settings/project-templates.tsx | 151 | 1 | RESIDUAL |

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | A1 ref |
|----|-----|-----|-------------|----------|--------|
| A2W10-A1-001 | RESIDUAL | P2 | accounting-page.tsx:32–52 | Пять `useQuery` без `isError`/`isLoading` в UI — fail = пустые таблицы | W10-A1-005 |
| A2W10-A1-002 | RESIDUAL | P2 | backups.tsx:28–36 | `fetchBackups` `try/finally` без `catch`; create/restore/delete без toast | W10-A1-008 |
| A2W10-A1-003 | RESIDUAL | P2 | crm-page.tsx:214 | Контакты без `isError`; `contacts.length === 0` = empty | W10-A1-007 |
| A2W10-A1-004 | RESIDUAL | P2 | recordings-list.tsx:43,59 | `catch {}` на load/download — empty/silent fail | W10-A1-014 |
| A2W10-A1-005 | RESIDUAL | P2 | api-keys.tsx:28–36,42–54 | Timer cleanup ✅; list/create/revoke всё ещё без catch/toast | W10-A1-009 |
| A2W10-A1-006 | RESIDUAL | P2 | ai-privacy.tsx:14–24,46+ | Toast на fail load есть, но форма всё равно рендерится с default `allowExternal=true` | W10-A1-011 |
| A2W10-A1-007 | RESIDUAL | P2 | project-templates.tsx:84–110 | Loading есть; `api.get` без catch/error UI → пустая сетка | W10-A1-012 |

## FIXED-OK (A1 закрыты в коде)

| A1 ID | Что проверено |
|-------|----------------|
| W10-A1-001/002/003 | create-call: Select `projectId` + `isValid` + `onError`; ScheduledCalls → `useQuery` + invalidate |
| W10-A1-004 | call-calendar: `isLoading`/`isError` |
| W10-A1-006 | bases: `basesError` |
| W10-A1-010 | api-keys: `copyTimerRef` + cleanup |
| W10-A1-015 | recording-controls: `toastApiError` |
| W10-A1-016 | intranet-hub: `useSearch` + `navigate({ search })` |
| W10-A1-017/018 | leave/kudos: `isError` |
| W10-A1-019/020/021 | coffee/marketplace: `onError`/`isError` |
| W10-A1-022 | message-bubble: `extractUrls(displayText)` ∪ contentText |

## Сводка

| Метрика | Значение |
|---------|----------:|
| Файлов | 29 / 29 / 0 |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 7 |
| FIXED-OK (ref) | 14 A1 IDs |
| P0 / P1 / P2 / P3 (open) | 0 / 0 / 7 / 0 |
