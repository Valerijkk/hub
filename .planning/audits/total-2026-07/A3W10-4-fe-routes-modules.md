# A3 W10 агент A4 — routes+modules part 4 (FINAL re-audit)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · ветка `cursor-grok-4.5`  
> Mirror A2: [A2W10-4-fe-routes-modules.md](A2W10-4-fe-routes-modules.md) · Stage A: [A10-4-fe-routes-modules.md](A10-4-fe-routes-modules.md)  
> Периметр: 29 файлов ([W10-ASSIGNMENTS.md](W10-ASSIGNMENTS.md) §A4) · Прочитано: **29/29** · Пропущено: **0**  
> Plan: [A3-PLAN.md](A3-PLAN.md) · Фокус: auth guards, XSS, modules wiring  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Сверка с A2 (A2W10-A4-*)

| A2 ID | Tag A2 | Статус в HEAD | Evidence |
|-------|--------|---------------|----------|
| A2W10-A4-001 | RESIDUAL | **FIXED-OK** | TASK-093: `ChatErrorBoundary` — `retryCount` + `key={retryCount}` remount children (`$channelId.tsx:19–22,58–62,71–72`) |
| A2W10-A4-002 | RESIDUAL | **FIXED-OK** | TASK-093: `$issueKey.tsx` — `isError` + alert/retry (`:53–54,149–167`) отдельно от `!issue` not-found (`:169–175`) |

**REGRESS:** 0

### Cascade FIXED-OK (Stage A W10-A4-* / A2 FIXED-OK)

| Stage A / A2 | Evidence (файл:строка) |
|--------------|------------------------|
| W10-A4-001 | board: единственный `onDrop`→`handleDrop` (`board/index.lazy.tsx:114–130,305`); нет `monitorForElements` double-fire |
| W10-A4-002 | `computeInsertRank(neighbors, index)` (`:122`) |
| W10-A4-003 | `$sprintId`: `useQuery` + `isError` + retry (`:84–114`) |
| W10-A4-004 | `filterBoardIssue(…, filters, currentUserId)` (`:40,137–139`) |
| W10-A4-005 | soft-delete copy «корзину на 30 дней» (`$issueKey.tsx:184`) |
| W10-A4-006 | `IssueAttachments` без `attachments={[]}` — локальный список + invalidate (`:242–251`) |
| W10-A4-007 (частично) | закрыты: `$issueKey` isError; `dashboard.lazy` store `error`+retry (`:134–147`); board/issues list через `QueryStateView`. **Остаток** → A3W10-4-001…003 |
| W10-A4-008 | = A2W10-A4-001 → FIXED-OK (remount key) |
| W10-A4-009 | `statusFilter` удалён из `issues/index.tsx` |
| W10-A4-010 | bulk delete на cases (`cases.lazy.tsx:105–123,279–295`) |
| W10-A4-011 | labels/workflow: `aria-label`/`aria-pressed` на color swatch (`labels.tsx:300–301`, `workflow.tsx:235–236`) |

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `frontend/src/routes/_authenticated/intranet/index.tsx` | 18 | 0 | чисто — Boundary + validateSearch.tab |
| `frontend/src/routes/_authenticated/marketplace.lazy.tsx` | 15 | 0 | чисто — Boundary + MarketplacePage |
| `frontend/src/routes/_authenticated/marketplace.tsx` | 6 | 0 | чисто — lazy stub |
| `frontend/src/routes/_authenticated/messenger.tsx` | 58 | 0 | чисто — layout Sheet a11y + RouteErrorBoundary |
| `frontend/src/routes/_authenticated/messenger/$channelId.tsx` | 100 | 0 | FIXED-OK remount key |
| `frontend/src/routes/_authenticated/messenger/index.tsx` | 124 | 1 | RESIDUAL — fail→welcome empty |
| `frontend/src/routes/_authenticated/projects/$projectId.tsx` | 71 | 0 | чисто — tabs + RouteErrorBoundary |
| `frontend/src/routes/_authenticated/projects/$projectId/backlog.tsx` | 22 | 0 | чисто — thin → SprintBacklog |
| `frontend/src/routes/_authenticated/projects/$projectId/board/index.lazy.tsx` | 356 | 0 | FIXED-OK DnD/rank/myTasks + QueryStateView |
| `frontend/src/routes/_authenticated/projects/$projectId/board/index.tsx` | 8 | 0 | чисто — lazy stub |
| `frontend/src/routes/_authenticated/projects/$projectId/dashboard.lazy.tsx` | 158 | 0 | FIXED-OK store error≠empty |
| `frontend/src/routes/_authenticated/projects/$projectId/dashboard.tsx` | 5 | 0 | чисто — lazy stub |
| `frontend/src/routes/_authenticated/projects/$projectId/gantt.lazy.tsx` | 23 | 0 | чисто — thin → GanttChart |
| `frontend/src/routes/_authenticated/projects/$projectId/gantt.tsx` | 8 | 0 | чисто — lazy stub |
| `frontend/src/routes/_authenticated/projects/$projectId/index.lazy.tsx` | 459 | 1 | RESIDUAL — fail→«Проект не найден» |
| `frontend/src/routes/_authenticated/projects/$projectId/index.tsx` | 10 | 0 | чисто — lazy stub |
| `frontend/src/routes/_authenticated/projects/$projectId/issues/$issueKey.tsx` | 291 | 0 | FIXED-OK isError + soft-delete copy + attachments |
| `frontend/src/routes/_authenticated/projects/$projectId/issues/index.tsx` | 333 | 0 | FIXED-OK QueryStateView; statusFilter gone |
| `frontend/src/routes/_authenticated/projects/$projectId/issues/new.tsx` | 40 | 0 | чисто — IssueForm create |
| `frontend/src/routes/_authenticated/projects/$projectId/settings/integrations.lazy.tsx` | 592 | 1 | NEW — list fail→«Нет подключённых…» |
| `frontend/src/routes/_authenticated/projects/$projectId/settings/integrations.tsx` | 10 | 0 | чисто — lazy stub |
| `frontend/src/routes/_authenticated/projects/$projectId/settings/labels.tsx` | 389 | 1 | NEW — list fail→«Нет меток…» |
| `frontend/src/routes/_authenticated/projects/$projectId/settings/workflow.tsx` | 398 | 1 | NEW — list fail→«Нет статусов…» |
| `frontend/src/routes/_authenticated/projects/$projectId/sprints/$sprintId.tsx` | 285 | 0 | FIXED-OK useQuery+isError |
| `frontend/src/routes/_authenticated/projects/$projectId/sprints/index.tsx` | 22 | 0 | чисто — thin → SprintBoard |
| `frontend/src/routes/_authenticated/projects/$projectId/tms.tsx` | 243 | 0 | чисто — tabs + beforeLoad redirect + Boundary |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/autotest.tsx` | 18 | 0 | чисто — thin → AutotestSetup |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/cases.lazy.tsx` | 706 | 1 | RESIDUAL — fail→«Нет тест-кейсов»; bulk delete OK |
| `frontend/src/routes/_authenticated/projects/$projectId/tms/cases.tsx` | 5 | 0 | чисто — lazy stub |

Файлов в списке: **29**. Прочитано: **29**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A3W10-4-001 | RESIDUAL | P2 | `messenger/index.tsx:25,44–53` (+ empty UI `:55–118`) | `useChannels()` без `isError`: при fail `isLoading=false` и `hasChannels=false` → welcoming «нет каналов» | API/сеть неотличимы от пустого аккаунта; ломает triage (W10-A4-007 остаток; A2 пометил FIXED-OK* без закрытия) | Деструктурировать `isError`/`refetch`; ветка alert+retry до empty CTA |
| A3W10-4-002 | RESIDUAL | P2 | `projects/$projectId/index.lazy.tsx:59–62,147–152` | `useQuery` project: есть `isLoading`, нет `isError` → fail падает в `!project` → «Проект не найден» | 403/500/сеть выглядят как 404; W10-A4-007 остаток | `isError` + retry UI; not-found только при success+empty |
| A3W10-4-003 | RESIDUAL | P2 | `tms/cases.lazy.tsx:61–64,412–442` | `listTestCases` query без `isError`: fail → `cases=[]` → empty «Нет тест-кейсов» | Ошибка списка кейсов маскируется empty state (W10-A4-007 остаток) | `isError`/`refetch` до empty; по аналогии с `runs.lazy` / QueryStateView |
| A3W10-4-004 | NEW | P2 | `settings/labels.tsx:63–66,209–222` | labels query без `isError`: fail → `ungroupedLabels=[]` → «Нет меток без группы» | Settings: fail load ≠ empty; форма создания доступна поверх «пустого» списка | `isError` + QueryErrorAlert/retry; empty только при success |
| A3W10-4-005 | NEW | P2 | `settings/workflow.tsx:58–67,173–186` | statuses/transitions без `isError`: fail → пустой список (empty copy только при `length===0`, не при `undefined`) + форма create остаётся | Fail load неотличим от «ещё не настроили»; риск дублей статусов поверх failed fetch | `isError` на обоих query; блокировать create UI при load error |
| A3W10-4-006 | NEW | P2 | `settings/integrations.lazy.tsx:36–39,83–92` (+ GitLab twin `:269–321`) | GitHub/GitLab list без `isError`: fail → «Нет подключённых … репозиториев» | Пользователь думает, что интеграций нет, и может пересоздать | `isError`+retry для обоих list query |

## FIXED-OK (сводка)

| ID | Evidence |
|----|----------|
| A2W10-A4-001 / W10-A4-008 | `$channelId.tsx` remount via `retryCount` key |
| A2W10-A4-002 | `$issueKey.tsx` isError ≠ not-found |
| W10-A4-001…006, 009…011 | см. cascade — без регрессии на HEAD |
| W10-A4-007 (dashboard/board/issues list/issue detail) | закрытые пути; остаток в 001–003 |

## Фокус: auth / XSS / modules

| Зона | Вердикт |
|------|---------|
| **Auth guards** | Все 29 под `_authenticated/*`. Layouts: messenger / project / TMS — `RouteErrorBoundary`. FE role: `$issueKey` `canMutateProject` перед delete; `index.lazy` archive/delete только `owner\|admin`. Authz канала/проекта — на BE (ожидаемо). |
| **XSS** | В периметре **нет** `dangerouslySetInnerHTML` / `innerHTML` / `eval`. `error.message` в ChatErrorBoundary — React text escape. User content (issue title/desc) через компоненты модулей, не raw HTML в этих routes. |
| **Modules wiring** | Thin wrappers корректны (backlog/gantt/sprints/autotest → modules). Lazy stubs канон. TMS `beforeLoad` redirect `/tms`→`/tms/cases`. Intranet/marketplace Boundary+page. |

## Чисто (дополнительно)

- Board optimistic: `onMutate` / rollback / `onSettled` invalidate; single drop path + lexorank.
- Issue update: `version` optimistic lock в mutate.
- Integrations: password inputs для token/secret; mutations toast/`toastApiError`; без optimistic (OK).
- Cases: bulk delete + a11y pagination/select; CaseTreePanel вне scope residual (TASK-093 A2W10-A2-008 FIXED).

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 6 |
| P3 | 0 |
| **Total** | **6** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active → closed) | 2 (`A2W10-A4-001`, `A2W10-A4-002`) |
| FIXED-OK (Stage A cascade) | 10 (W10-A4-001…006,008…011; 007 partial) |
| RESIDUAL | 3 (`A3W10-4-001`…`003`) |
| NEW | 3 (`A3W10-4-004`…`006`) |
| REGRESS | 0 |

**Топ IDs:** `A3W10-4-001` (messenger empty≠error) · `A3W10-4-002` (project overview) · `A3W10-4-003` (TMS cases list).
