# A2-W10 агент A3 — routes+modules part 3 (re-audit)

> TASK-008 · Stage A2 · read-only · 2026-07-28  
> HEAD: `0d6a659f` · Зеркало: [`A10-3-fe-routes-modules.md`](A10-3-fe-routes-modules.md) · [`W10-ASSIGNMENTS.md`](W10-ASSIGNMENTS.md) §A3  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| frontend/src/modules/tracker/issues-cursor.ts | 22 | 0 | чисто |
| frontend/src/modules/tracker/issue-sidebar-meta.tsx | 242 | 2 | RESIDUAL |
| frontend/src/modules/tracker/pipeline-status.tsx | 288 | 0 | FIXED-OK |
| frontend/src/modules/tracker/planning-poker.tsx | 212 | 1 | RESIDUAL |
| frontend/src/modules/tracker/sprint-backlog.tsx | 165 | 2 | RESIDUAL |
| frontend/src/modules/tracker/sprint-board.tsx | 377 | 1 | RESIDUAL |
| frontend/src/modules/tracker/sprint-retro.tsx | 171 | 0 | FIXED-OK |
| frontend/src/modules/tracker/triage-inbox.tsx | 156 | 0 | FIXED-OK |
| frontend/src/modules/tracker/velocity-chart.tsx | 103 | 1 | RESIDUAL |
| frontend/src/routes/__root.tsx | 106 | 0 | чисто |
| frontend/src/routes/_authenticated.tsx | 60 | 0 | чисто |
| frontend/src/routes/_authenticated/accounting.lazy.tsx | 13 | 0 | чисто |
| frontend/src/routes/_authenticated/accounting.tsx | 14 | 0 | чисто |
| frontend/src/routes/_authenticated/admin/audit.tsx | 21 | 0 | чисто |
| frontend/src/routes/_authenticated/admin/dashboard.tsx | 215 | 1 | RESIDUAL |
| frontend/src/routes/_authenticated/admin/limits.tsx | 21 | 0 | чисто |
| frontend/src/routes/_authenticated/admin/mail.tsx | 20 | 0 | чисто |
| frontend/src/routes/_authenticated/admin/trash.tsx | 21 | 0 | чисто |
| frontend/src/routes/_authenticated/admin/users.tsx | 21 | 0 | чисто |
| frontend/src/routes/_authenticated/bases.tsx | 13 | 0 | чисто |
| frontend/src/routes/_authenticated/calendar.tsx | 13 | 0 | чисто |
| frontend/src/routes/_authenticated/calls.tsx | 183 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/crm.lazy.tsx | 13 | 0 | чисто |
| frontend/src/routes/_authenticated/crm.tsx | 5 | 0 | чисто |
| frontend/src/routes/_authenticated/dashboard.lazy.tsx | 224 | 0 | FIXED-OK |
| frontend/src/routes/_authenticated/dashboard.tsx | 2 | 0 | чисто |
| frontend/src/routes/_authenticated/help.tsx | 89 | 0 | чисто |
| frontend/src/routes/_authenticated/hr.tsx | 13 | 0 | чисто |
| frontend/src/routes/_authenticated/intranet/directory.tsx | 17 | 0 | чисто |

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | A1 ref |
|----|-----|-----|-------------|----------|--------|
| A2W10-A3-001 | RESIDUAL | P1 | sprint-backlog.tsx:28–32 | `listIssues({limit:100})` без фильтра «без спринта»; DTO/`api` без `sprint_id`; UI врёт «N задач без спринта» | W10-A3-001 |
| A2W10-A3-002 | RESIDUAL | P2 | sprint-backlog.tsx:30–32 | `catch` глотает ошибку → empty «Бэклог пуст» | W10-A3-004 |
| A2W10-A3-003 | RESIDUAL | P2 | velocity-chart.tsx:26–32 | Zustand `fetchVelocity`; `length===0` сразу = empty | W10-A3-003 |
| A2W10-A3-004 | RESIDUAL | P2 | issue-sidebar-meta.tsx:57–64 | Подписи «Статус»/«Приоритет» — `<div>`, не `Label`/`htmlFor` | W10-A3-005 |
| A2W10-A3-005 | RESIDUAL | P2 | sprint-board.tsx:255–351 | `<Button>` «Начать/Завершить» внутри `<Link>` — nested interactive | W10-A3-007 |
| A2W10-A3-006 | RESIDUAL | P2 | planning-poker.tsx (vote cards) | Loading/error path ✅; карточки голоса без `aria-pressed`/`radiogroup` | W10-A3-008 |
| A2W10-A3-007 | RESIDUAL | P2 | admin/dashboard.tsx:58–91 | Кастомные `<button>` вкладок, не `TabsTrigger` / нет `role="tab"` | W10-A3-012 |
| A2W10-A3-008 | RESIDUAL | P3 | issue-sidebar-meta.tsx:226 | Блок с `cursor-pointer` (исполнитель/добавить) без handler | W10-A3-019 |
| A2W10-A3-009 | RESIDUAL | P2 | sprint-backlog.tsx (~checkbox rows) | Checkbox выбора задачи без `aria-label` / label | W10-A3-006 |

## FIXED-OK

| A1 ID | Что проверено |
|-------|----------------|
| W10-A3-002 | poker: `initialLoadDone` / skeleton до empty |
| W10-A3-009–011 | poker toast; preset flag в `onSuccess`; retro toast+timer cleanup |
| W10-A3-013 | hardcoded KPI `change=` убраны |
| W10-A3-014/015 | calls: static badgeClass; `toast.error` на fail invite |
| W10-A3-016/020 | dashboard: Search/Bell wired; clock interval |
| W10-A3-017/018 | pipeline conditional poll; retro `clearTimeout` |

## Сводка

| Метрика | Значение |
|---------|----------:|
| Файлов | 29 / 29 / 0 |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 9 |
| P0 / P1 / P2 / P3 (open) | 0 / 1 / 7 / 1 |
