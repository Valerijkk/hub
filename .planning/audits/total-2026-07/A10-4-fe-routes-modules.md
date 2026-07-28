# W10 агент A4 — routes+modules part 4

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус FE: modules wiring, optimistic UI, error boundaries.  
> Список: W10-ASSIGNMENTS.md §A4 (29 файлов).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/routes/_authenticated/intranet/index.tsx` | 18 | 0 | чисто |
| `frontend/src/routes/_authenticated/marketplace.lazy.tsx` | 15 | 0 | чисто |
| `frontend/src/routes/_authenticated/marketplace.tsx` | 6 | 0 | чисто |
| `frontend/src/routes/_authenticated/messenger.tsx` | 58 | 0 | чисто |
| `frontend/src/routes/_authenticated/messenger/$channelId.tsx` | 91 | 2 | замечания |
| `frontend/src/routes/_authenticated/messenger/index.tsx` | 124 | 1 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId.tsx` | 71 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/backlog.tsx` | 22 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/board/index.lazy.tsx` | 382 | 3 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId/board/index.tsx` | 8 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/dashboard.lazy.tsx` | 143 | 1 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId/dashboard.tsx` | 5 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/gantt.lazy.tsx` | 23 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/gantt.tsx` | 8 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/index.lazy.tsx` | 459 | 1 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId/index.tsx` | 10 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/issues/$issueKey.tsx` | 259 | 3 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId/issues/index.tsx` | 336 | 1 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId/issues/new.tsx` | 40 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/settings/integrations.lazy.tsx` | 585 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/settings/integrations.tsx` | 10 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/settings/labels.tsx` | 387 | 1 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId/settings/workflow.tsx` | 396 | 1 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId/sprints/$sprintId.tsx` | 255 | 1 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId/sprints/index.tsx` | 22 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/tms.tsx` | 243 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/autotest.tsx` | 18 | 0 | чисто |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/cases.lazy.tsx` | 835 | 2 | замечания |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/cases.tsx` | 5 | 0 | чисто |

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W10-A4-001 | P1 | `board/index.lazy.tsx:124-140` + `:112-120` (+ `board-column.tsx:52-58`) | DnD: и `monitorForElements.onDrop`, и `BoardColumn` `dropTargetForElements.onDrop` → `handleDrop` вызывают `moveCardMutation.mutate` | Каждый перенос между колонками шлёт **два** `moveIssue`; optimistic `onMutate` гоняется дважды, риск 409/рассинхрона рангов | Оставить один путь: либо только monitor, либо только dropTarget `onDrop`; второй — no-op / убрать |
| W10-A4-002 | P1 | `board/index.lazy.tsx:114-118`; `:133-137` | `rank` всегда `"0|aaaaaa:"`; `_index` из `handleDrop` игнорируется | Optimistic UI и бэкенд получают один и тот же lexorank → порядок карточек ломается / коллизии рангов | Считать rank от соседей в целевой колонке (до/после index); передать index из drop target |
| W10-A4-003 | P1 | `sprints/$sprintId.tsx:83-96` | `getSprint(...).then(setSprint)` без `catch` / Query; при ошибке `sprint === null` навсегда | Бесконечный спиннер «Загрузка...» вместо error/retry; нет abort на unmount | `useQuery` + `QueryStateView` / `isError`; либо `.catch` + error UI |
| W10-A4-004 | P2 | `board/index.lazy.tsx:239-245`; `:147-171` | Чип «Мои задачи» пишет `filters.myTasks`, но `filteredColumns` его **не читает** | Фильтр UI — no-op; пользователь думает, что доска отфильтрована | В filter: сравнить `issue.assignee?.id` с current user id из auth-store; иначе убрать чип |
| W10-A4-005 | P2 | `issues/$issueKey.tsx:150-152` | Диалог удаления: «Это действие нельзя отменить» | ТЗ: soft delete → корзина 30 дней; обзор проекта (`index.lazy.tsx:349-351`) формулирует верно — здесь врёт | Текст как у проекта: «в корзину на 30 дней» |
| W10-A4-006 | P2 | `issues/$issueKey.tsx:210-218` | `IssueAttachments` получает `attachments={[]}` всегда | Вложения с API не показываются; `onAttachmentAdded` только invalidate | Передать `issue.attachments` (или поле DTO); после upload — refetch |
| W10-A4-007 | P2 | `$issueKey.tsx:136-141`; `index.lazy.tsx:147-152`; `messenger/index.tsx:25-53`; `cases.lazy.tsx:184-571`; `dashboard.lazy.tsx:131-139` | `useQuery`/`useChannels`/`zustand fetch` без ветки `isError`: ошибка → «не найдено» / empty / вечный loading | API fail неотличим от пустых данных; ломает triage и UX | `isError` + retry/`QueryStateView`; dashboard — error из store |
| W10-A4-008 | P2 | `messenger/$channelId.tsx:15-65`; `:55` | Локальный `ChatErrorBoundary`: retry = `setState({ error: null })` без remount `key`; дублирует родительский `RouteErrorBoundary` (messenger layout) | После crash «Попробовать снова» часто сразу падает снова (тот же instance); два boundary с разным UX | `key={retryCount}` на children **или** убрать class, оставить `RouteErrorBoundary`; единый retry+reload |
| W10-A4-009 | P3 | `issues/index.tsx:53`; `:153` | `statusFilter` в state и в `resetFilters`, UI Select статуса нет | Мёртвое состояние; «Сбросить» учитывает невидимый фильтр | Добавить Select статуса **или** удалить `statusFilter` |
| W10-A4-010 | P3 | `tms/cases.lazy.tsx:424-438`; `:228-243` | Bulk-панель «Выбрано: N» только «Снять»; нет bulk-действий | Selection UX без outcome — мёртвая фича | Bulk delete/move/status **или** убрать multi-select |
| W10-A4-011 | P3 | `labels.tsx:293-301`; `workflow.tsx:228-236` | Color swatch `<button>` без `aria-label` / `aria-pressed` | AT не знает выбранный цвет (WCAG name/state) | `aria-label={color}`; `aria-pressed={selected}` |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `intranet/index.tsx` | `RouteErrorBoundary module="Intranet"`; `validateSearch.tab`; thin module wire |
| `marketplace.lazy.tsx` + stub | Boundary + MarketplacePage; stub A13-017 OK |
| `messenger.tsx` | Layout + Sheet a11y (`aria-label`, `SheetTitle` sr-only); `RouteErrorBoundary` на Outlet; hooks до return |
| `$projectId.tsx` | Tabs Link + `RouteErrorBoundary module="Трекер"` на Outlet |
| `backlog` / `sprints/index` / `gantt.lazy` / `autotest` / `issues/new` | Thin wrappers → modules; parent boundary покрывает runtime |
| Lazy stubs (`board/index`, `dashboard`, `gantt`, `index`, `integrations`, `cases`) | Пустые `createFileRoute` — code-split канон |
| `integrations.lazy.tsx` | Mutations + toast/`toastApiError`; password inputs; loading/empty; invalidate keys; без optimistic (не требуется) |
| `tms.tsx` | Tabs + `RouteErrorBoundary module="TMS"`; `beforeLoad` redirect на `/tms/cases` |
| Board optimistic path (частично) | `onMutate`/`onError` rollback/`onSettled` invalidate — паттерн верный; ломают double-fire + rank |
| Issue update | `version` в `updateIssue` — optimistic lock на месте; нет optimistic cache (приемлемо) |

**Фокус-вердикт:** критично — **двойной DnD mutate** и **захардкоженный rank** на Kanban; спринт-деталь **висит навечно** при ошибке API. Error boundaries на layout’ах (messenger/tracker/TMS/intranet/marketplace) в целом на месте; слабое место — query error → false empty/not-found и Chat retry без remount.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **29 / 29 / 0** |
| Находок | **11** |
| P0 | 0 |
| P1 | 3 |
| P2 | 5 |
| P3 | 3 |

### Топ-3

1. **W10-A4-001 (P1)** — Kanban: `monitorForElements` + `BoardColumn.onDrop` → двойной `moveIssue` на каждый drop.
2. **W10-A4-002 (P1)** — Kanban: `rank: "0|aaaaaa:"` константа; индекс drop не используется → коллизии порядка.
3. **W10-A4-003 (P1)** — `$sprintId`: `getSprint` без error-path → бесконечный loader при сбое API.
