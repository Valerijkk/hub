# A2-W10 агент A4 — routes+modules part 4 (re-audit)

> TASK-008 · Stage A2 · read-only · 2026-07-28  
> HEAD: `0d6a659f` · Зеркало: [`A10-4-fe-routes-modules.md`](A10-4-fe-routes-modules.md) · [`W10-ASSIGNMENTS.md`](W10-ASSIGNMENTS.md) §A4  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| frontend/src/routes/_authenticated/intranet/index.tsx | 16 | 0 | чисто |
| frontend/src/routes/_authenticated/marketplace.lazy.tsx | 13 | 0 | чисто |
| frontend/src/routes/_authenticated/marketplace.tsx | 5 | 0 | чисто |
| frontend/src/routes/_authenticated/messenger.tsx | 51 | 0 | чисто |
| frontend/src/routes/_authenticated/messenger/$channelId.tsx | 83 | 1 | RESIDUAL |
| frontend/src/routes/_authenticated/messenger/index.tsx | 117 | 0 | FIXED-OK* |
| frontend/src/routes/_authenticated/projects/$projectId.tsx | 65 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/backlog.tsx | 19 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/board/index.lazy.tsx | 336 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/projects/$projectId/board/index.tsx | 7 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/dashboard.lazy.tsx | 146 | 0 | FIXED-OK* |
| frontend/src/routes/_authenticated/projects/$projectId/dashboard.tsx | 4 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/gantt.lazy.tsx | 21 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/gantt.tsx | 7 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/index.lazy.tsx | 438 | 0 | FIXED-OK* |
| frontend/src/routes/_authenticated/projects/$projectId/index.tsx | 9 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/issues/$issueKey.tsx | 243 | 1 | RESIDUAL |
| frontend/src/routes/_authenticated/projects/$projectId/issues/index.tsx | 319 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/projects/$projectId/issues/new.tsx | 37 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/settings/integrations.lazy.tsx | 566 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/settings/integrations.tsx | 9 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/settings/labels.tsx | 366 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/projects/$projectId/settings/workflow.tsx | 374 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/projects/$projectId/sprints/$sprintId.tsx | 268 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/projects/$projectId/sprints/index.tsx | 19 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/tms.tsx | 235 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/tms/autotest.tsx | 15 | 0 | чисто |
| frontend/src/routes/_authenticated/projects/$projectId/tms/cases.lazy.tsx | 849 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/projects/$projectId/tms/cases.tsx | 4 | 0 | чисто |

\* частичный: error≠empty в части query-путей закрыт sweep’ами; точечный gap на issue detail ниже.

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | A1 ref |
|----|-----|-----|-------------|----------|--------|
| A2W10-A4-001 | RESIDUAL | P2 | messenger/$channelId.tsx:52–55 | `ChatErrorBoundary` retry = `setState({error:null})` без `key` remount children | W10-A4-008 |
| A2W10-A4-002 | RESIDUAL | P2 | issues/$issueKey.tsx:50,104 | `useQuery` issue: есть `isLoading`, нет `isError` → fail может выглядеть как not-found/empty | W10-A4-007 |

## FIXED-OK

| A1 ID | Что проверено |
|-------|----------------|
| W10-A4-001/002 | board: один `onDrop`→`handleDrop`; `computeInsertRank` |
| W10-A4-003 | `$sprintId`: `useQuery` + `isError` |
| W10-A4-004 | `filterBoardIssue(…, myTasks, currentUserId)` |
| W10-A4-005 | soft-delete copy «корзину на 30 дней» |
| W10-A4-006 | `IssueAttachments` без `attachments={[]}` (локальный список + invalidate) |
| W10-A4-009 | `statusFilter` удалён |
| W10-A4-010 | bulk delete на cases |
| W10-A4-011 | labels/workflow: `aria-label`/`aria-pressed` |

## Сводка

| Метрика | Значение |
|---------|----------:|
| Файлов | 29 / 29 / 0 |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 2 |
| P0 / P1 / P2 / P3 (open) | 0 / 0 / 2 / 0 |
