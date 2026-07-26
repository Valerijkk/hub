# Performance — Hot Paths / Cache / Bundle / Waterfalls

**Дата:** 2026-07-26  
**Зона:** A13 — hot paths, sync loops, missing cache, unbounded queries, large serialization, frontend bundle, list virtualization, app-start waterfalls  
**Режим:** READ-ONLY (кроме этого отчёта)  
**Ветка:** `cursor-grok-4.5`  
**Контекст волны 1:** N+1 / indexes из `03-tracking.md` (A3-020), `07-infra-domains.md` (A7-017), `08-data-layer.md` (A8-001/002 missing project FK indexes), `04-ai-search-storage.md` (A4-014 unbounded bases records)

## Что проверено

| Область | Фокус |
|---------|--------|
| Tracker / Board / Sprint | `GetBoardData`, `ListIssues`, converters, sprint `WithIssues` |
| Bases / TMS / Dashboard | `GetTable` mutations, flaky N+1, project stats cache miss, velocity loop |
| Notification fan-out | `Dispatch` × prefs/user в циклах consumer |
| Frontend start / bundle | `app-shell`, i18n eager, route `.lazy` coverage, `vite` manualChunks |
| Lists / virtualization | Kanban columns, channels, issues accumulate, messenger chat (ok) |
| Waterfalls | Board list→data, dashboard personal+layout, Gantt per-row links |

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A13-001 | P1 | `backend/internal/tracker/repo_boards.go:56-62` | `GetBoardData`: `WithIssues` **без Limit** на каждый workflow status (+ assignee/labels) | Hot path Kanban: весь проект issues в один ответ/память; при росте досок latency/OOM; усиливается отсутствием FK-индекса project (A8-001) | Per-column `Limit`/`cursor`; hard cap (напр. 200/колонка) + «load more»; Select только brief-полей |
| A13-002 | P1 | `backend/internal/bases/repository.go:123-127`; evidence `bases/service.go:214-242` | `GetTable` грузит **все** records; `CreateRecord`/`UpdateRecord` вызывают полный `GetTable` | Каждая мутация = полный table scan + serialize; уже отмечено A4-014 как pagination gap — здесь **hot-path amplification** | `GetTableMeta` (columns only) для write-path; `GetTableView` с limit/cursor; не грузить records на create/update |
| A13-003 | P1 | `backend/internal/dashboard/repository.go:114-122` (+ `service.go:197`) | `ListProjectIssuesWithStatusAssignee` → `.All(ctx)` без limit (даже при Redis TTL 5m на miss) | Cache miss / cold start тянет **все** issues проекта ради `TasksByStatus` | `GROUP BY status` / `Count` агрегаты в SQL; не материализовать все rows |
| A13-004 | P1 | `frontend/src/modules/tracker/gantt-chart.tsx:113-116` | Каждый `GanttRow` делает `useQuery` → `listIssueLinks(projectId, issue_key)` | Классический **N+1 HTTP**: 50 задач с due_date = 50 параллельных link-запросов при открытии Gantt | Один batch `listIssueLinks` на project / preload links в parent; или backend endpoint dependencies for gantt |
| A13-005 | P1 | `backend/internal/notification/service.go:75-116` + `repository.go:252-279`; fan-out `consumer.go:533-548` | `Dispatch`: до **3× `IsEnabled`** (каждая — 1–2 SQL) + `GetUser` (quiet hours / telegram) **синхронно на каждого** recipient | Call-started / channel mention: M участников → O(M×5) sync DB round-trips в consumer loop | Prefetch prefs map; cache user quiet/telegram; bulk `CreateNotification`; async channel fan-out |
| A13-006 | P1 | `backend/internal/sprint/repository.go:87-93` | `ListSprints` → `WithIssues()` на **все** спринты проекта | Список спринтов тянет полный graph issues (unbounded) | List без issues + `Count`; issues только в `GetSprint` с limit |
| A13-007 | P2 | `backend/internal/tracker/service_issues.go:165-166` + `service_converters.go:50-52` | `ListIssues` сериализует полный `IssueResponse` с `DescriptionJSON` / checklist | List API тащит тяжёлый JSONB description на каждую строку (до 100 items) | List → `IssueBriefResponse` / omit `description_json`; full body только в GetIssue |
| A13-008 | P2 | `backend/internal/tracker/repo_issues.go:206-207` + `:245-250` | Каждый `ListIssues`: `Count` clone + page `All` | Двойной round-trip на hot list; Count дорог без project FK index (A8-001) | Опциональный `total`; `COUNT(*) OVER()`; approximate; или skip total на cursor UX |
| A13-009 | P2 | `backend/internal/dashboard/service.go:250-262` | Velocity: **10 последовательных** `CountProjectCompletedIssuesInRange` (см. также A7-017) | Cache miss = 10 sync DB calls | Один bucketed SQL / `date_trunc('week')` |
| A13-010 | P2 | `backend/internal/tms/service_reports.go:256-257` | Flaky report: `ListTestResults` на каждый run (см. также A3-020) | N+1 PG на отчёт | `Where(HasRunWith(IDIn(...)))` одним запросом |
| A13-011 | P2 | `frontend/.../board/index.tsx:40-61` | Waterfall: `listBoards` → `useEffect` setActiveBoard → только потом `getBoardData` (`enabled: !!activeBoardId`) | Лишний RT до первой отрисовки Kanban | Default board id из list; `Promise.all` / combined endpoint; persist `activeBoardId` |
| A13-012 | P2 | `frontend/.../dashboard.lazy.tsx:77-79` | App-start: `fetchPersonal()` + `loadLayout("personal")` **последовательно** (layout после `.then`) | Два serial API на главной после login | `Promise.all([fetchPersonal(), loadLayout(...)])` |
| A13-013 | P2 | `backend/internal/dashboard/service.go:53-172` vs `:175-186` | Personal dashboard: **нет Redis cache** (project stats кешируются) | Каждый визит главной = пачка sequential repo calls (tasks/calls/runs/deadlines/timelogs/unread) | `dashboard:personal:{userID}` TTL 30–60s; invalidate on issue/notif events |
| A13-014 | P2 | `frontend/src/components/tracker/board-column.tsx:100-114` | Колонка: `issues.map` + `AnimatePresence`/`motion` **без** virtualization | Большие колонки = тяжёлый DOM + layout animation; virtualizer есть только в chat | `@tanstack/react-virtual` per column (как `chat-view.tsx:157`) |
| A13-015 | P2 | `frontend/src/components/layout/app-shell.tsx:11-15`, `:153-155` | `AIChatPanel`, `CommandPalette`, `ShortcutsDialog` — **eager** import в shell (CallPanel уже `lazy`) | Каждый authenticated page тянет AI/search/help UI в initial graph | `lazy(() => import(...))` + Suspense по open-state |
| A13-016 | P2 | `frontend/src/i18n/index.ts:11-28`, `:36-69` | Все locale JSON (ru+en × 9 ns, ~52 KB raw) + все namespaces в `main` через `import "./i18n"` | Bundle bloat на cold start независимо от текущего модуля | `i18next-http-backend` / dynamic `import()` per ns; грузить только `common` + route ns |
| A13-017 | P2 | `frontend/src/routes/**` | **10** `*.lazy.tsx` из **~62** route-файлов; тяжёлые `crm`/`accounting`/`marketplace`/`gantt`/`board`/`messenger` — eager `createFileRoute` | Route tree тащит лишний код до навигации | Перевести heavy modules на `createLazyFileRoute` (как dashboard) |
| A13-018 | P2 | `frontend/src/components/messenger/channel-list.tsx:613-616` | Список каналов без virtualization (полный `map`) | При большом числе DM/каналов — лишний DOM; chat уже virtualized | Virtualize long channel groups |
| A13-019 | P2 | `backend/internal/tracker/repo_issues.go:155-165` | `FindIssueByID` → `WithChildren()` без limit | Issue detail может подтянуть неограниченных children | Cap / separate children endpoint with pagination |
| A13-020 | P3 | `backend/internal/tracker/repo_comments.go:66-74`; `repo_activity.go:42-47` | `ListComments` / `ListActivities` — `.All` без limit | Старые issues → тяжёлый JSON (comments + reactions+users) | Cursor/limit (напр. 50) |
| A13-021 | P3 | `backend/internal/tms/ent_repository.go:666-672` | `ListTestResults` — все results прогона без pagination | Крупный Allure import → огромный API payload | Offset/limit или cursor |
| A13-022 | P3 | `backend/internal/project/repository.go:417-418`, `:465-473` | `ListTeams`/`ListOutgoingCrossProjectLinks` — unbounded `.All` | Рост org/links → memory | Pagination |
| A13-023 | P3 | `backend/internal/search/meilisearch_indexer.go:52-56` | `IndexAllIssues` hard `Limit(5000)` + полный marshal docs | >5k issues молча не индексируются; large JSON POST | Cursor batch upsert (500/chunk) |
| A13-024 | P3 | `frontend/.../issues/index.tsx:57-80`, display map | Cursor accumulate в `allIssues` без windowing | «Load more» раздувает DOM | Virtual list / window на таблицу |
| A13-025 | P3 | `backend/internal/messenger/ent_repository.go:65-68` | `ListUserChannels` — все каналы user без limit | Ок для 10–50, но без cap | Soft cap + archive filter |

## Чисто

- **Messenger messages:** cursor + `Limit` в Mongo (`mongo_repository.go:108-110`); UI virtualized (`chat-view.tsx:157`).
- **ListIssues API:** cursor + max limit 100 (`repo_issues.go:230-233`).
- **Board DTO:** brief issues (`toIssueBriefResponse`), не full description.
- **Project dashboard cache:** Redis `dashboard:project:{id}` TTL 5m (`dashboard/service.go:177-186`).
- **RBAC / link preview:** Redis cache (`rbac/service.go:38-49`, `github/preview.go:41-86`).
- **Vite splitting:** `manualChunks` для tanstack/plate/livekit/charts/i18n (`vite.config.ts:28-59`); chart widgets `React.lazy` (`widget-grid.tsx:50-68`); `CallPanel` lazy в shell.
- **Personal dashboard sub-queries:** limits на my tasks/calls/runs (`dashboard/repository.go:56-65` и др.).
- **QueryClient default `staleTime: 30_000`** (`__root.tsx:8-10`) снижает refetch-шум.

## Оценка зоны: 5/10

**Обоснование:** Хороший фундамент на «видимых» путях (пагинация issues/messages, Redis на project stats/RBAC/preview, vite chunks, chat virtualization). Оценка снижена **горячими** путями: unbounded Kanban/`GetTable`/sprint list, sync notification fan-out, Gantt N+1 HTTP, personal dashboard без cache + serial FE waterfalls, eager shell/i18n/routes. После limit на board/bases, batch notification prefs, brief list DTO, lazy shell+routes и virtualized columns зона могла бы быть ~7.5/10.

### Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 6 |
| P2 | 13 |
| P3 | 6 |
| **Total** | **25** |

**Score: 5/10**
