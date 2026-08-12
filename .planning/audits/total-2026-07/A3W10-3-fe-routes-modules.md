# A3 W10 агент A3 — routes+modules part 3 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W10-3-fe-routes-modules.md](A2W10-3-fe-routes-modules.md) · Stage A: [A10-3-fe-routes-modules.md](A10-3-fe-routes-modules.md)  
> Assignment: [W10-ASSIGNMENTS.md](W10-ASSIGNMENTS.md) §A3 (29 файлов)  
> Фокус: auth guards · XSS · modules state leaks · Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Сверка с A2 (A2W10-A3-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W10-A3-001 | RESIDUAL→**REGRESS** | открыт | TASK-082 (`842a4a2f`) ввёл `sprint_id:"none"` + `filterIssuesWithoutSprint`; merge `25ae31ba` (TASK-093) оставил версию без фильтра. `sprint-backlog.tsx:33-34` = `listIssues({limit:100})` + `resp.items`. Helper/тесты живы в `api/tracker.ts:293-298`, `sprint-backlog-filter.test.tsx` |
| A2W10-A3-002 | RESIDUAL | **FIXED-OK** | `loadError` + `QueryErrorAlert` + `toastApiError` — `:28,:36-40,:116-120` |
| A2W10-A3-003 | RESIDUAL | **RESIDUAL** (partial) | error path: `velocity-chart.tsx:32-54` + store `error`; **loading** по-прежнему нет — `length===0` → empty до ответа |
| A2W10-A3-004 | RESIDUAL | **FIXED-OK** | `Label htmlFor` + Trigger `id` — status/priority/SP (`issue-sidebar-meta.tsx:58-70,:91-103,:192-210`) |
| A2W10-A3-005 | RESIDUAL | **FIXED-OK** | actions вне `<Link>` — `sprint-board.tsx:261-370` |
| A2W10-A3-006 | RESIDUAL | **FIXED-OK** | `role="radiogroup"` + `role="radio"` + `aria-checked` — `planning-poker.tsx:146-167` |
| A2W10-A3-007 | RESIDUAL | **FIXED-OK** | `TabsList`/`TabsTrigger` — `admin/dashboard.tsx:57-87` |
| A2W10-A3-008 | RESIDUAL | **FIXED-OK** | assignee/reporter без мёртвого `cursor-pointer` — `:128-165` |
| A2W10-A3-009 | RESIDUAL | **FIXED-OK** | checkbox `aria-label={`Выбрать ${issue.issue_key}`}` — `sprint-backlog.tsx:145-151` |

**REGRESS:** 1 · **RESIDUAL (A2 actives):** 1 · **NEW:** 2 · **FIXED-OK (A2 actives):** 7/9

### Cascade FIXED-OK (Stage A via A2, без регресса)

| Stage A | Evidence |
|---------|----------|
| W10-A3-002 poker loading≠empty | `initialLoadDone` / skeleton — `planning-poker.tsx:45-47,:113-120` |
| W10-A3-009–011 toast/preset/retro | poker toast; triage flag в `onSuccess`; retro toast + `savedTimerRef` cleanup |
| W10-A3-013 hardcoded KPI change | KPI без `change=` — `admin/dashboard.tsx:93-113` |
| W10-A3-014/015 calls badges/toast | static `badgeClass`; `toast.error` на fail invite |
| W10-A3-016/020 dashboard Search/Bell/clock | `openCommand` / `NotificationBell` / interval 30s |
| W10-A3-017/018 pipeline poll; retro timer | conditional `refetchInterval`; `clearTimeout` |

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `frontend/src/modules/tracker/issues-cursor.ts` | 22 | 0 | чисто |
| `frontend/src/modules/tracker/issue-sidebar-meta.tsx` | 261 | 0 | FIXED-OK a11y/labels |
| `frontend/src/modules/tracker/pipeline-status.tsx` | 288 | 0 | FIXED-OK conditional poll |
| `frontend/src/modules/tracker/planning-poker.tsx` | 218 | 0 | FIXED-OK load/radiogroup |
| `frontend/src/modules/tracker/sprint-backlog.tsx` | 194 | 2 | REGRESS filter + NEW selection leak |
| `frontend/src/modules/tracker/sprint-board.tsx` | 406 | 1 | FIXED-OK nested; NEW stale store (via fetchSprints) |
| `frontend/src/modules/tracker/sprint-retro.tsx` | 180 | 0 | FIXED-OK toast/cleanup |
| `frontend/src/modules/tracker/triage-inbox.tsx` | 168 | 0 | FIXED-OK QueryStateView + preset flag |
| `frontend/src/modules/tracker/velocity-chart.tsx` | 133 | 2 | RESIDUAL loading; NEW stale velocity |
| `frontend/src/routes/__root.tsx` | 111 | 0 | чисто (QueryClient defaults) |
| `frontend/src/routes/_authenticated.tsx` | 66 | 0 | чисто — auth gate + Query clear |
| `frontend/src/routes/_authenticated/accounting.lazy.tsx` | 15 | 0 | чисто |
| `frontend/src/routes/_authenticated/accounting.tsx` | 15 | 0 | чисто — admin `beforeLoad` |
| `frontend/src/routes/_authenticated/admin/audit.tsx` | 23 | 0 | чисто — admin gate |
| `frontend/src/routes/_authenticated/admin/dashboard.tsx` | 224 | 0 | FIXED-OK TabsTrigger |
| `frontend/src/routes/_authenticated/admin/limits.tsx` | 23 | 0 | чисто — admin gate |
| `frontend/src/routes/_authenticated/admin/mail.tsx` | 22 | 0 | чисто — admin gate |
| `frontend/src/routes/_authenticated/admin/trash.tsx` | 23 | 0 | чисто — admin gate |
| `frontend/src/routes/_authenticated/admin/users.tsx` | 23 | 0 | чисто — admin gate |
| `frontend/src/routes/_authenticated/bases.tsx` | 15 | 0 | чисто — ErrorBoundary |
| `frontend/src/routes/_authenticated/calendar.tsx` | 15 | 0 | чисто — ErrorBoundary |
| `frontend/src/routes/_authenticated/calls.tsx` | 191 | 0 | FIXED-OK toast/badges |
| `frontend/src/routes/_authenticated/crm.lazy.tsx` | 15 | 0 | чисто |
| `frontend/src/routes/_authenticated/crm.tsx` | 6 | 0 | чисто — lazy stub |
| `frontend/src/routes/_authenticated/dashboard.lazy.tsx` | 240 | 0 | FIXED-OK Search/Bell/clock |
| `frontend/src/routes/_authenticated/dashboard.tsx` | 3 | 0 | чисто — lazy stub |
| `frontend/src/routes/_authenticated/help.tsx` | 96 | 0 | чисто — admin link gated |
| `frontend/src/routes/_authenticated/hr.tsx` | 15 | 0 | чисто — ErrorBoundary |
| `frontend/src/routes/_authenticated/intranet/directory.tsx` | 19 | 0 | чисто — validateSearch |

Файлов в списке: **29**. Прочитано: **29**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Sev | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|-----|-------------|---------|-------------------|------------|
| A3W10-3-001 | REGRESS (A2W10-A3-001 / TASK-082) | P1 | `sprint-backlog.tsx:33-34` | Нет `sprint_id:"none"` и нет `filterIssuesWithoutSprint(resp.items)` после merge TASK-093 | UI «N задач без спринта» снова = произвольные 100 issues; move-to-sprint не убирает задачи из списка; mutation-тест `sprint-backlog-filter` ожидает фильтр | Вернуть вызов из `842a4a2f`: `listIssues(...,{limit:100,sprint_id:"none"})` + `setBacklogIssues(filterIssuesWithoutSprint(resp.items))` |
| A3W10-3-002 | RESIDUAL (A2W10-A3-003) | P2 | `velocity-chart.tsx:57-72` + `sprint-store.ts:121-130` | `fetchVelocity` не ставит loading; chart при `velocity.length===0 && !error` сразу показывает empty | Пока идёт запрос UI врёт «Нет завершённых спринтов» | `loading`/`isFetching` в store или `useQuery`; empty только после успеха с `[]` |
| A3W10-3-003 | NEW | P2 | `sprint-store.ts:70-81,:121-130` → потребители `velocity-chart.tsx`, `sprint-board.tsx` | При смене `projectId` store не сбрасывает `velocity`/`sprints` до ответа; на error `fetchSprints` старый список остаётся | Cross-project state leak: чужие (другого проекта) данные в UI до/при fail | В начале fetch: `set({velocity:[],sprints:[],…})` или keyed-by-projectId / TanStack Query |
| A3W10-3-004 | NEW | P2 | `sprint-backlog.tsx:26,:44-47` | `selectedIssues` не очищается при смене `projectId` | Selection leak: UUID задач прошлого проекта остаются выбранными → move в спринт нового проекта по чужим id | `setSelectedIssues(new Set())` в `useEffect` на `projectId` |

## Фокус-чеклист

| Тема | Вердикт |
|------|---------|
| **Auth guards** | `_authenticated` `beforeLoad` → `/login`; accounting + admin/* → `isSystemAdmin` redirect; help admin-link gated. XSS/auth bypass в слоте не найден. |
| **XSS** | `dangerouslySetInnerHTML` / `innerHTML` в 29 файлах **нет**. User text (title, keys, voterId slice) через React children. Label `style={{borderColor,color}}` — CSS property API (не HTML string). |
| **Modules state** | Query clear на multi-tab logout ✅. Poker/triage/pipeline Query/local loading ✅. Утечки: backlog filter REGRESS, velocity loading RESIDUAL, store stale + selection NEW. |

## FIXED-OK (сводка)

| Зона | Проверено |
|------|-----------|
| Auth layout | `beforeLoad` + `subscribeAuthSync` → `queryClient.clear` |
| Admin routes | system_role gate на audit/dashboard/limits/mail/trash/users + accounting |
| A2 a11y/error (кроме 001/003) | Labels, nested interactive, radiogroup, Tabs, checkbox aria, backlog error≠empty |
| Stage A poker/retro/calls/dashboard/pipeline | loading split, toasts, cleanup, wired Search/Bell |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 1 |
| P2 | 3 |
| P3 | 0 |
| **Total** | **4** |

| Метрика | Значение |
|---------|----------|
| Файлов | **29 / 29 / 0** |
| Активных | **4** (NEW 2 · REGRESS 1 · RESIDUAL 1) |
| FIXED-OK (A2 actives) | **7/9** |
| FIXED-OK (Stage A cascade) | **6+** без регресса |

### Топ-3

1. **A3W10-3-001 (P1, REGRESS)** — SprintBacklog снова без фильтра «без спринта» (merge-регресс TASK-082).
2. **A3W10-3-003 (P2, NEW)** — Zustand sprint-store не сбрасывает данные при смене проекта.
3. **A3W10-3-002 (P2, RESIDUAL)** — VelocityChart путает loading с empty.

---

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.  
Counts: P0=0 · P1=1 · P2=3 · P3=0 · Total=4.  
Топ: A3W10-3-001.
