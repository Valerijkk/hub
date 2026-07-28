# A2-W10 агент A5 — routes+modules part 5 (re-audit)

> TASK-008 · Stage A2 · read-only · 2026-07-28  
> HEAD: `0d6a659f` · Зеркало: [`A10-5-fe-routes-modules.md`](A10-5-fe-routes-modules.md) · [`W10-ASSIGNMENTS.md`](W10-ASSIGNMENTS.md) §A5  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| frontend/src/routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.lazy.tsx | 855 | 1 | RESIDUAL |
| frontend/src/routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.tsx | 9 | 0 | чисто (stub) |
| frontend/src/routes/_authenticated/projects/$projectId/tms/environments.tsx | 15 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/tms/flaky.tsx | 11 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/tms/runs.lazy.tsx | 490 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/projects/$projectId/tms/runs.tsx | 4 | 0 | чисто (stub) |
| frontend/src/routes/_authenticated/projects/$projectId/tms/runs_.$runId.tsx | 315 | 1 | RESIDUAL |
| frontend/src/routes/_authenticated/projects/$projectId/tms/traceability.tsx | 16 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/triage.tsx | 11 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/index.tsx | 308 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/projects/new.tsx | 275 | 0 | чисто |
| frontend/src/routes/_authenticated/settings.tsx | 112 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/settings/ai.tsx | 8 | 0 | чисто |
| frontend/src/routes/_authenticated/settings/appearance.tsx | 12 | 0 | чисто |
| frontend/src/routes/_authenticated/settings/hotkeys.tsx | 61 | 0 | чисто |
| frontend/src/routes/_authenticated/settings/integrations.tsx | 47 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/settings/notifications.tsx | 18 | 0 | чисто |
| frontend/src/routes/_authenticated/settings/profile.lazy.tsx | 475 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/settings/profile.tsx | 7 | 0 | чисто (stub) |
| frontend/src/routes/_authenticated/settings/security.tsx | 69 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/trash.tsx | 459 | 1 | RESIDUAL |
| frontend/src/routes/call.$code.lazy.tsx | 301 | 0 | FIXED-OK |
| frontend/src/routes/call.$code.tsx | 7 | 0 | чисто (stub) |
| frontend/src/routes/index.tsx | 11 | 0 | чисто |
| frontend/src/routes/login.tsx | 154 | 0 | чисто |
| frontend/src/routes/onboarding.lazy.tsx | 878 | 0 | FIXED-OK |
| frontend/src/routes/onboarding.tsx | 7 | 0 | чисто (stub) |
| frontend/src/routes/status.tsx | 17 | 0 | чисто |

Файлов в списке: 28. Прочитано: 28. Пропущено: 0.

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | A1 ref |
|----|-----|-----|-------------|----------|--------|
| A2W10-A5-001 | RESIDUAL | P1 | trash.tsx:455–460 | «Удалить навсегда» → `toast.success` без API; данные остаются | W10-A5-002 |
| A2W10-A5-002 | RESIDUAL | P2 | cases_.$caseKey.lazy.tsx:260 | `if (isLoading \|\| !tc)` без `isError` → вечный shimmer на fail | W10-A5-005 |
| A2W10-A5-003 | RESIDUAL | P2 | runs_.$runId.tsx:124 | `if (isLoading \|\| !run)` без `isError` → вечный shimmer | W10-A5-006 |

## FIXED-OK

| A1 ID | Что проверено |
|-------|----------------|
| W10-A5-001 | trash: `listTrash` |
| W10-A5-003 | settings/integrations: hint + Link, нет `projectId=""` |
| W10-A5-004 | «Очистить все» удалена |
| W10-A5-007 | runs.lazy: `isError` + retry |
| W10-A5-008 | profile: init form в `useEffect` |
| W10-A5-009 | projects index: static `INFO_STEP_BADGE` |
| W10-A5-010–012 | onboarding: file input, timezone register, accent store |
| W10-A5-013 | GuestCallSession без full CallResponse cast |
| W10-A5-014–016 | placeholder path gone; security `<Link>`; avatar opens file |
| W10-A5-018 | колонка «Удалено» |
| W10-A5-019 | `existingIds` Set вынесен до map (IIFE) |

## Сводка

| Метрика | Значение |
|---------|----------:|
| Файлов | 28 / 28 / 0 |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 3 |
| P0 / P1 / P2 / P3 (open) | 0 / 1 / 2 / 0 |
