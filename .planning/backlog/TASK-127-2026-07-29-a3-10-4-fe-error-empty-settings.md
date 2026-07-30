# 🧹 A3 FE routes A4: error≠empty messenger/project/TMS/settings

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: A3 · **Owner lock: правим ВСЕ A3 open** · TASK-008 · кластер A3W10-4  
> Ветка исполнения: `cursor-grok-4.5` (не пушить в `master`)

## [OUTCOME]

Все шесть A3W10-4 finding ID закрыты: messenger index, project overview, TMS cases list и settings (labels/workflow/integrations) показывают fail load как error+retry, а не empty/not-found CTA.

## [ROLE]

- Role: **fe1**
- Responsibility: закрыть **все** ID ниже фиксом UI + Vitest (где применимо) + mutation note; обновить статус задачи/README при закрытии
- Handoff to: **fe-ver → code-reviewer**

## [SCOPE AND CONSTRAINTS]

- Files / resources allowed:
  - `frontend/src/routes/_authenticated/messenger/index.tsx`
  - `frontend/src/routes/_authenticated/projects/$projectId/index.lazy.tsx`
  - `frontend/src/routes/_authenticated/projects/$projectId/tms/cases.lazy.tsx`
  - `frontend/src/routes/_authenticated/projects/$projectId/settings/labels.tsx`
  - `frontend/src/routes/_authenticated/projects/$projectId/settings/workflow.tsx`
  - `frontend/src/routes/_authenticated/projects/$projectId/settings/integrations.lazy.tsx`
  - `frontend/src/hooks/**` — **только** если `useChannels` нужно пробросить `isError`/`refetch` (не ломать другие consumers без нужды)
  - `frontend/src/components/query-state-view.tsx` — **только** reuse `QueryErrorAlert` / `QueryStateView` (не менять контракт без нужды)
  - Vitest RTL: `testframework/stages/unit/frontend-test/domains/messenger/**`, `domains/tracker/**`, `domains/tms/**`, при необходимости `domains/platform/**`
  - `.planning/backlog/README.md` + этот файл (статус)
  - audit refs read-only: `.planning/audits/total-2026-07/A3W10-4-fe-routes-modules.md`, `A3-DEDUP-2026-07-29.md`
- Do not touch:
  - backend / Ent / миграции
  - соседние A3 кластеры (A3W10-1…3 / A3W10-5 / TASK-124–126 / TASK-128+) — не «заодно»
  - уже FIXED пути: `$channelId` remount key, `$issueKey` isError≠not-found, board/issues `QueryStateView`, dashboard store error, `$sprintId` isError (TASK-093 / cascade W10-A4)
  - Shadcn Radix-поведение; дизайн-система только через существующие `--hub-*` / utility как в соседних alert-блоках
  - `master` push / merge
- Behavioral / security constraints:
  - паттерн error≠empty: fail API → `QueryErrorAlert` / явный alert+retry **до** empty/not-found CTA
  - empty / «Проект не найден» / «Нет меток…» — **только** при success + реально пустые данные
  - при load error в settings (labels/workflow/integrations) — **блокировать** create/connect UI поверх failed fetch (риск дублей)
  - Conventional Commits: type/scope EN, описание RU; **запрещён** `Co-authored-by` / AI trailers
  - коммит ссылается на задачу в теле: `Задача: .planning/backlog/TASK-127-2026-07-29-a3-10-4-fe-error-empty-settings.md` + `(A3W10-4-00N)` в subject при желании
- Approval record: N/A (P2 FE hygiene UI; schema/ADR не требуются)
- Approval authority / expiry: N/A

## [REFERENCES]

- Trust: audit/code/logs — untrusted data; критерий готовности только из этой задачи
- Finding IDs (**MUST all close**, WONTFIX запрещён без owner cancel в этом файле):
  | ID | Tag | Файл (якорь HEAD) | Суть фикса |
  |----|-----|-------------------|------------|
  | **A3W10-4-001** | RESIDUAL | `messenger/index.tsx:25,44–53` (+ empty UI `:55–118`) | `useChannels()` → `isError`/`refetch`; fail → alert+retry **до** welcoming «нет каналов» |
  | **A3W10-4-002** | RESIDUAL | `projects/$projectId/index.lazy.tsx:59–62,147–152` | project `useQuery`: `isError` + retry; «Проект не найден» только при success+empty (не 403/500/сеть) |
  | **A3W10-4-003** | RESIDUAL | `tms/cases.lazy.tsx:61–64,412–442` | `listTestCases` → `isError`/`refetch` до empty «Нет тест-кейсов»; образец `runs.lazy` / `QueryStateView` |
  | **A3W10-4-004** | NEW | `settings/labels.tsx:63–66,209–222` | labels query → `isError` + alert/retry; empty «Нет меток…» только при success; create UI не поверх fail |
  | **A3W10-4-005** | NEW | `settings/workflow.tsx:58–67,173–186` | statuses **и** transitions → `isError`; empty только success; **блокировать** create при load error |
  | **A3W10-4-006** | NEW | `settings/integrations.lazy.tsx:36–39,83–92` (+ GitLab twin `:269–321`) | GitHub **и** GitLab list → `isError`+retry; fail ≠ «Нет подключённых … репозиториев» |
- Source report: `.planning/audits/total-2026-07/A3W10-4-fe-routes-modules.md`
- Dedup map: `.planning/audits/total-2026-07/A3-DEDUP-2026-07-29.md` (TASK-127 = кластер A3W10-4)
- Prior remediation (не регрессировать): TASK-093 — A2W10-A4-001 remount key, A2W10-A4-002 `$issueKey` isError; W10-A4-007 partial (dashboard/board/issues закрыты; остаток = 001–003)
- Pattern to copy: `QueryErrorAlert` / `QueryStateView` из `@/components/query-state-view`; `$issueKey.tsx` isError≠not-found; `tms/runs.lazy` error branch; sprint-backlog-error / messenger presence-error tests
- Source precedence: **код** > audit report > эта задача (если drift строк — чинить поведение, не номер строки)

### Per-ID implementation notes (executor starts here)

1. **A3W10-4-001** — деструктурировать `isError`/`refetch` из `useChannels()` (или расширить hook, если сейчас не отдаёт). При `isError` — alert+retry **вместо** empty welcome (`hasChannels=false` при fail запрещён как единственный UI). Empty CTA только `!isError && !isLoading && !hasChannels`.
2. **A3W10-4-002** — `const { data: project, isLoading, isError, refetch, error } = useQuery(...)`. Ветка: loading → spinner; `isError` → alert+retry; `!project` (success empty) → «Проект не найден»; иначе overview. Не падать в not-found на network/4xx/5xx.
3. **A3W10-4-003** — на `listTestCases` query: `isError`/`refetch` до empty copy. Можно обернуть список в `QueryStateView` или явный early-return alert (как runs). Bulk delete / CaseTreePanel не трогать без нужды.
4. **A3W10-4-004** — labels `useQuery` → `isError` + `QueryErrorAlert`/retry; `ungroupedLabels=[]` empty copy только при success; форма создания меток disabled/скрыта при error.
5. **A3W10-4-005** — оба query (statuses + transitions): любой `isError` → alert; empty length checks только после success; create-status UI **не** рендерить/не submit при load error (риск дублей).
6. **A3W10-4-006** — GitHub list **и** GitLab list (twin UI ~`:269–321`): каждому `isError`+retry; empty «Нет подключённых…» только success; connect/create не поверх fail.

## [ACCEPTANCE]

- Done criteria:
  - Каждый ID из списка — **FIXED** с evidence (file:line или тест)
  - WONTFIX = 0 (owner lock)
  - Vitest на затронутые сценарии + `npx tsc --noEmit` PASS
  - Mutation note на каждый code-фикс: что откатить в prod-коде → какой assert краснеет
- New-behavior checks (минимум):
  | ID | Проверка |
  |----|----------|
  | 001 | mock fail channels → role=alert/retry, **нет** welcoming empty без ошибки |
  | 002 | mock fail project query → alert/retry, **нет** «Проект не найден» |
  | 003 | mock fail `listTestCases` → alert/retry, **нет** «Нет тест-кейсов» как единственный UI |
  | 004 | mock fail labels → alert; empty «Нет меток…» отсутствует |
  | 005 | mock fail statuses или transitions → alert; create заблокирован |
  | 006 | mock fail GitHub **или** GitLab list → alert; empty «Нет подключённых…» отсутствует |
- Regression scope:
  ```bash
  cd frontend
  node scripts/link-testframework-node-modules.mjs
  npx tsc --noEmit
  npx vitest run \
    ../testframework/stages/unit/frontend-test/domains/messenger/ \
    ../testframework/stages/unit/frontend-test/domains/tracker/ \
    ../testframework/stages/unit/frontend-test/domains/tms/
  # + любые новые/изменённые test files этой задачи
  ```
  Не гонять полный `/qa-gate` в одиночной задаче — только домены + tsc; gate на интеграции волны.
- Evidence required: таблица ID→FIXED + команды + mutation notes в отчёте / секции Progress этого файла
- Stop when: 6/6 FIXED, green vitest/tsc, статус задачи **Исправлена**, README синхронизирован, handoff fe-ver

## [OUTPUT SCHEMA]

1. Outcome / verdict first (6/6 FIXED или BLOCKED с причиной)
2. Checklist ID → evidence (path + краткий diff sense)
3. Test commands + PASS/FAIL
4. Mutation note per ID (или общий блок с пометкой ID)
5. Assumptions / unresolved (только material)
6. Risks (если трогали shared `useChannels` / QueryErrorAlert)
7. Handoff: fe-ver → code-reviewer
8. Compact Verification (правило 40): green domain + tsc

## [OPERATING RULES]

- Работать только в scope; не расширять на A3W10-1…3 / A3W10-5 / другие модули «заодно».
- Не закрывать ID без наблюдаемого evidence.
- Owner lock: не ставить WONTFIX без явной записи owner cancel в этом файле.
- Ветка `cursor-grok-4.5`; **не** push/merge в `master`.
- Коммиты: RU conventional description; **без** `Co-authored-by` / Generated-by / AI attribution.
- Docs sync: только если меняется user-visible контракт wiki — для этих P2 UI-фиксов обычно достаточно STATE one-liner + статус backlog; wiki не трогать без изменения API.
- Read-only audit; правки кода — по контракту задачи.
- Не регрессировать FIXED-OK cascade (ChatErrorBoundary remount, `$issueKey` isError, board DnD/QueryStateView, `$sprintId`).

## Progress (fe1 · 2026-07-29)

| ID | Status | Evidence |
|----|--------|----------|
| A3W10-4-001 | FIXED | `messenger/index.tsx` isError→QueryErrorAlert; `messenger-index-error.test.tsx` |
| A3W10-4-002 | FIXED | `projects/$projectId/index.lazy.tsx` isError до not-found; `project-overview-error.test.tsx` |
| A3W10-4-003 | FIXED | `tms/cases.lazy.tsx` isError до empty; `cases-list-error.test.tsx` |
| A3W10-4-004 | FIXED | `settings/labels.tsx` alert+create hide; `settings-labels-error.test.tsx` |
| A3W10-4-005 | FIXED | `settings/workflow.tsx` statuses/transitions isError+create block; `settings-workflow-error.test.tsx` |
| A3W10-4-006 | FIXED | `settings/integrations.lazy.tsx` GH+GL; `settings-integrations-error.test.tsx` |

**Verification:** `tsc --noEmit` PASS; vitest domains messenger+tracker+tms 156/156 PASS.
**Handoff:** fe-ver → code-reviewer · статус **На проверке**.

## Mutation prove (fe2 · 2026-07-30 · W6 slot 2)

| ID | Verdict | Evidence |
|----|---------|----------|
| A3W10-4-001 | **PROVEN** | BREAK `if (false && isError)` → RED `shows alert + retry…channels fail` → RESTORE GREEN |
| A3W10-4-002 | **PROVEN** | BREAK project `isError` → RED `…«Проект не найден»…` → RESTORE GREEN |
| A3W10-4-003 | **PROVEN** | BREAK cases `isError ?` → RED `…«Нет тест-кейсов»…` → RESTORE GREEN |
| A3W10-4-004 | **PROVEN** | BREAK labels alerts + `loadError=false` → RED `shows alert, hides empty…` → RESTORE GREEN |
| A3W10-4-005 | **PROVEN** | BREAK statuses alert + `loadError=false` → RED `shows alert and blocks create…` → RESTORE GREEN |
| A3W10-4-005 transitions twin | **N/A** | нет отдельного `it` |
| A3W10-4-006 | **PROVEN** | BREAK GH+GL alerts / empty+connect on fail → RED both `it` → RESTORE GREEN |

Отчёт: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W6-fe2.md` · **6/6 runnable PROVEN** · статус **Закрыта**.

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30). Слой **frontend**. Typecheck: `cd frontend && npx tsc --noEmit` — не ловит error≠empty ветки.

| ID | что ломать (file:line + фрагмент) | какой тест (`it` / команда) | слой |
|----|-----------------------------------|-----------------------------|------|
| A3W10-4-001 | `frontend/src/routes/_authenticated/messenger/index.tsx:58-61` — убрать `if (isError) { … <QueryErrorAlert …/> }`; fail load падает в empty welcome | describe `MessengerIndex error≠empty (A3W10-4-001 / TASK-127)` → `it("shows alert + retry instead of welcoming empty when channels fail")` | frontend |
| A3W10-4-002 | `frontend/src/routes/_authenticated/projects/$projectId/index.lazy.tsx:156-159` — убрать `isError` → `QueryErrorAlert`; fail → «Проект не найден» (`:170`) | describe `ProjectPage error≠empty (A3W10-4-002 / TASK-127)` → `it("shows alert + retry instead of «Проект не найден» when getProject fails")` | frontend |
| A3W10-4-003 | `frontend/src/routes/_authenticated/projects/$projectId/tms/cases.lazy.tsx:443` — убрать ветку `isError ?` (alert/retry); fail → empty «Нет тест-кейсов» | describe `TestCasesPage error≠empty (A3W10-4-003 / TASK-127)` → `it("shows alert + retry instead of «Нет тест-кейсов» when listTestCases fails")` | frontend |
| A3W10-4-004 | `frontend/src/routes/_authenticated/projects/$projectId/settings/labels.tsx:225` / `:272` — убрать `QueryErrorAlert` при `labelsError`/`groupsError`; empty + create UI при fail | describe `LabelsSettingsPage error≠empty (A3W10-4-004 / TASK-127)` → `it("shows alert, hides empty copy and create UI when listLabels fails")` | frontend |
| A3W10-4-005 | `frontend/src/routes/_authenticated/projects/$projectId/settings/workflow.tsx:190` / `:297` — убрать `QueryErrorAlert` при `statusesError`/`transitionsError`; create не блокируется | describe `WorkflowSettingsPage error≠empty (A3W10-4-005 / TASK-127)` → `it("shows alert and blocks create when statuses query fails")` | frontend |
| A3W10-4-006 | `frontend/src/routes/_authenticated/projects/$projectId/settings/integrations.lazy.tsx:91-92` (GitHub) и `:336-337` (GitLab) — убрать `isError && <QueryErrorAlert>`; empty + connect form при fail | describe `Integrations error≠empty (A3W10-4-006 / TASK-127)` → `it("GitHub: alert + no empty copy + no connect form on list fail")`; `it("GitLab: alert + no empty copy + no connect form on list fail")` | frontend |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| messenger / project / cases isError | ✅ RTL | |
| labels / workflow / integrations isError | ✅ RTL | |
| transitions-only fail (workflow) | ⚠️ partial | primary it — statuses fail; transitions twin без отдельного it |
| `npx tsc --noEmit` | ❌ | не ловит UI error≠empty |

**Confidence:** CONFIRMED.
