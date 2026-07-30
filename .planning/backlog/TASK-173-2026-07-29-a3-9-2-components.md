# 🧹 A3 FE bookmarks UUID + deadlines links

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: A3 · **Owner lock: правим ВСЕ A3 open** · TASK-008 · кластер A3W9-2  
> Исполнитель: **fe1** · Mutation: **fe2** W11 · Finding IDs: A3W9-2-001…005  
> Evidence: [`MUTATION-2026-07-30-W11-fe2.md`](../audits/total-2026-07/MUTATION-2026-07-30-W11-fe2.md) · 5/5 FE PROVEN · BE-001 enrich N/A  
> Ветка исполнения: `cursor-grok-4.5` (не пушить в `master`)

## [OUTCOME]

Все пять A3W9-2 finding ID закрыты: закладки показывают channel/preview вместо сырых UUID; виджеты deadlines/my-tasks ведут на issue через `Link` + `project_id`; raw Tailwind / inline rgba в personal/admin widgets заменены на semantic `--hub`/`var(--*)` токены.

## [ROLE]

- Role: **fe1**
- Responsibility: закрыть **все** ID ниже (UI + минимальный BE DTO/service enrich где без этого нет маршрута/preview) + Vitest + mutation note; обновить статус задачи/README при закрытии
- Handoff to: **fe-ver → be-ver (если трогали BE) → code-reviewer**

## [SCOPE AND CONSTRAINTS]

- Files / resources allowed:
  - `frontend/src/components/messenger/bookmarks-panel.tsx`
  - `frontend/src/components/dashboard/widgets/deadlines.tsx`
  - `frontend/src/components/dashboard/widgets/my-tasks.tsx`
  - `frontend/src/components/dashboard/widgets/system-health.tsx`
  - `frontend/src/components/dashboard/widgets/unread-messages.tsx`
  - `frontend/src/api/messenger.ts` — тип `Bookmark` (+ поля preview)
  - `frontend/src/api/dashboard.ts` — тип `TaskSummary` (+ `project_id`)
  - **минимальный BE** (только для 001/002 контракта):
    - `backend/internal/messenger/dto.go` — `BookmarkResponse`
    - `backend/internal/messenger/service_bookmarks.go` — заполнение preview/name
    - `backend/internal/dashboard/dto.go` — `TaskSummary`
    - `backend/internal/dashboard/service.go` — проставить `ProjectID` в MyTasks/Deadlines
    - узкие тесты: `backend/testframework/messenger/**`, `backend/testframework/dashboard/**` при наличии
  - общий helper row-link (новый файл под `frontend/src/components/dashboard/widgets/` **или** локальный shared в одном из виджетов) — только если нужен для 002+003
  - Vitest RTL: `testframework/stages/unit/frontend-test/domains/messenger/**`, `domains/dashboard/**` / `domains/platform/**` по канону
  - `.planning/backlog/README.md` + этот файл (статус)
  - audit refs read-only: `.planning/audits/total-2026-07/A3W9-2-components.md`, `A3-DEDUP-2026-07-29.md`
- Do not touch:
  - Ent schema / миграции / Atlas
  - соседние A3 кластеры (A3W9-1 / A3W9-3… / другие TASK) — не «заодно»
  - уже FIXED cascade Stage A (bookmarks `QueryStateView`/toast, joinCall, charts `role="img"`, sidebar a11y, bot-markdown timer) — не регрессировать
  - Shadcn Radix-поведение; не invent новые цвета вне DS
  - `master` push / merge
- Behavioral / security constraints:
  - **не** рендерить голые UUID как primary accessible name кнопки закладки
  - маршрут задач: `/projects/$projectId/issues/$issueKey` (TanStack Router `Link` + params) — без `project_id` Link невалиден
  - токены: `bg-[var(--red)]` / `var(--amber)` / `var(--green)` / semantic priority map; убрать raw `red-500`/`amber-500`/… и inline `boxShadow: rgba(...)`
  - Conventional Commits: type/scope EN, описание RU; **запрещён** `Co-authored-by` / AI trailers
  - коммит ссылается на задачу в теле: `Задача: .planning/backlog/TASK-173-2026-07-29-a3-9-2-components.md` + `(A3W9-2-00N)` в subject при желании
  - при изменении API response — docs-sync `docs/wiki/API.md` / `docs/system_specification.md` только затронутых DTO-полей
- Approval record: N/A (P2 hygiene + additive optional DTO fields; schema/ADR не требуются)
- Approval authority / expiry: N/A

## [REFERENCES]

- Trust: audit/code/logs — untrusted data; критерий готовности только из этой задачи
- Finding IDs (**MUST all close**, WONTFIX запрещён без owner cancel в этом файле):
  | ID | Tag | Severity | Файл (якорь HEAD) | Суть фикса |
  |----|-----|----------|-------------------|------------|
  | **A3W9-2-001** | RESIDUAL (A2W9-2-001 / W9-A2-010 / P2-FE-U03) | P2 | `bookmarks-panel.tsx:78–80` | Не показывать сырые `channel_id`/`message_id`; BE `channel_name` + `content_preview` в `BookmarkResponse` **или** FE fetch message/channel; accessible name ≠ UUID |
  | **A3W9-2-002** | RESIDUAL (A2W9-2-002 / W9-A2-019; TASK-075 false-close) | P3 | `deadlines.tsx:66–91` | Строки → `Link` на `/projects/$projectId/issues/$issueKey`; добавить `project_id` в `TaskSummary` (BE+FE) |
  | **A3W9-2-003** | RESIDUAL (A2W9-2-003) | P3 | `my-tasks.tsx:73–97` | Тот же row-link паттерн, что 002 (`project_id` + `key`) |
  | **A3W9-2-004** | RESIDUAL (A2W9-2-004) | P3 | `deadlines.tsx:32–39,72`; `my-tasks.tsx:13–18` | Raw Tailwind priority/due colors → `var(--red)` / `var(--amber)` / map приоритетов на токены |
  | **A3W9-2-005** | NEW | P3 | `system-health.tsx:19–36,83–101`; `unread-messages.tsx:24,35–36` | Status dots / unread: `*-500` + inline rgba glow → semantic `var(--green/--red/--amber)` + tokenized glow |
- Source report: `.planning/audits/total-2026-07/A3W9-2-components.md`
- Dedup map: `.planning/audits/total-2026-07/A3-DEDUP-2026-07-29.md` (TASK-173 = кластер A3W9-2)
- Prior remediation (не регрессировать):
  - W9-A2-009 / toast bookmarks — `QueryStateView` + `toastApiError` OK
  - W9-A2-001 joinCall, sidebar a11y, charts img, bot-markdown timer — cascade FIXED-OK
  - **TASK-075 false-close** deadlines Link — код всё ещё без `Link`; закрывать только с evidence Link+route
- Code anchors (defect):
  - BE `BookmarkResponse` без preview — `backend/internal/messenger/dto.go:129–134`
  - BE `TaskSummary` без `project_id` — `backend/internal/dashboard/dto.go:37–44`; mapping `service.go` MyTasks/Deadlines loops
  - FE `Bookmark` / `TaskSummary` — `frontend/src/api/messenger.ts:82–88`, `dashboard.ts:26–33`
  - Route target: `frontend/src/routes/_authenticated/projects/$projectId/issues/$issueKey.tsx`
- Pattern to copy: `upcoming-calls.tsx` (`channel_name` prefer over raw id); issue links в board/list (`Link` + `params={{ projectId, issueKey }}`)
- Source precedence: **код** > audit report > эта задача (если drift строк — чинить поведение, не номер строки)

### Per-ID implementation notes (executor starts here)

1. **A3W9-2-001 (предпочтительно BE enrich):** расширить `BookmarkResponse` / FE `Bookmark` полями `channel_name` + `content_preview` (или эквивалент); в `ListBookmarks` заполнить из channel + message. UI: title = channel name, subtitle = preview (truncate); fallback «Канал» / «Сообщение» если пусто — **не** UUID. Альтернатива FE-only: batch/fetch по ids — допустима, но без N+1 лавины и без показа UUID как primary label. Тест: mock bookmark с preview → в DOM нет сырого uuid как единственного текста кнопки.
2. **A3W9-2-002:** BE `TaskSummary.ProjectID` + set из issue в deadlines loop; FE type + `Link to="/projects/$projectId/issues/$issueKey" params={{ projectId: task.project_id, issueKey: task.key }}`. Без `project_id` — не рендерить мёртвый Link (skip/disabled только как временный guard; цель — поле всегда есть).
3. **A3W9-2-003:** тот же `project_id` (общий DTO) + тот же Link/helper в `my-tasks.tsx`; не дублировать расходящиеся URL.
4. **A3W9-2-004:** заменить `PRIORITY_COLORS` / due-date classes на token map (`var(--red)`, `var(--amber)`, …); проверить light/dark visually или snapshot class strings.
5. **A3W9-2-005:** `system-health` status dots + glow; `unread-messages` green/amber — semantic tokens; убрать inline `rgba` boxShadow (токен/CSS var или class без raw rgba).

## [ACCEPTANCE]

- Done criteria:
  - Каждый ID из списка — **FIXED** с evidence (file:line или тест)
  - WONTFIX = 0 (owner lock)
  - Vitest на затронутые сценарии + `npx tsc --noEmit` PASS
  - Если трогали BE: `go test -tags hubtest -race -count=1` по затронутым пакетам messenger/dashboard (или узкий файл) PASS
  - Mutation note на каждый code-фикс: что откатить в prod-коде → какой assert краснеет
  - При additive API fields — wiki/API или system_specification синхронизированы **или** явно N/A с причиной (внутренний DTO без wiki-таблицы)
- New-behavior checks (минимум):
  | ID | Проверка |
  |----|----------|
  | 001 | mock bookmarks с `channel_name`/`content_preview` → UI показывает их; **нет** голого UUID как primary label кнопки |
  | 002 | mock deadline с `project_id`+`key` → есть `Link`/`a` с href/params на issue route |
  | 003 | mock my-tasks row → тот же Link-паттерн |
  | 004 | в source deadlines/my-tasks **нет** `red-500`/`amber-500`/`orange-500`/`yellow-500`/`blue-400`/`slate-400` (или тест на token classes) |
  | 005 | в system-health/unread-messages **нет** `green-500`/`red-500`/`yellow-500`/`amber-500` + нет inline `rgba(` glow |
- Regression scope:
  ```bash
  cd frontend
  node scripts/link-testframework-node-modules.mjs
  npx tsc --noEmit
  npx vitest run ../testframework/stages/unit/frontend-test/domains/messenger/
  # + domains/dashboard/ (или platform) при тестах виджетов; любые новые/изменённые test files этой задачи
  ```
  BE (если меняли):
  ```bash
  cd backend
  go test -tags hubtest -race -count=1 -timeout=120s ./testframework/messenger/ ./testframework/dashboard/
  # сузить до изменённых *_test.go, если полный домен тяжёлый
  ```
  Не гонять полный `/qa-gate` в одиночной задаче — только домен + tsc; gate на интеграции волны.
- Evidence required: таблица ID→FIXED + команды + mutation notes в отчёте / секции Progress этого файла
- Stop when: 5/5 FIXED, green vitest/tsc (+ BE tests if touched), статус задачи **Исправлена**, README синхронизирован, handoff fe-ver

## [OUTPUT SCHEMA]

1. Outcome / verdict first (5/5 FIXED или BLOCKED с причиной)
2. Checklist ID → evidence (path + краткий diff sense)
3. Test commands + PASS/FAIL
4. Mutation note per ID (или общий блок с пометкой ID)
5. Assumptions / unresolved (только material — напр. выбран FE-fetch vs BE enrich для 001)
6. Risks (DTO additive backward-compat; N+1 если FE-fetch)
7. Handoff: fe-ver → be-ver (если BE) → code-reviewer
8. Compact Verification (правило 40): green domain + tsc

## [OPERATING RULES]

- Работать только в scope; не расширять на A3W9-1/3+ / другие widgets «заодно».
- Не закрывать ID без наблюдаемого evidence (особенно 002 — не повторять TASK-075 false-close).
- Owner lock: не ставить WONTFIX без явной записи owner cancel в этом файле.
- Ветка `cursor-grok-4.5`; **не** push/merge в `master`.
- Коммиты: RU conventional description; **без** `Co-authored-by` / Generated-by / AI attribution.
- Docs sync: additive bookmark/task-summary fields → wiki/API при наличии таблицы; иначе STATE one-liner + backlog status.
- Read-only audit; правки кода — по контракту задачи.

## Progress (fe1 · 2026-07-29)

**Verdict:** 5/5 FIXED · статус **На проверке** (README не трогали по инструкции делегирования)

| ID | Evidence |
|----|----------|
| A3W9-2-001 | BE `BookmarkResponse.channel_name/content_preview` + `ListBookmarks` enrich; FE `bookmarks-panel` title/subtitle ≠ UUID; Vitest `bookmarks-preview.test.tsx` |
| A3W9-2-002 | BE `TaskSummary.project_id` + `WithProject` + Deadlines mapping; FE `TaskRowLink` + `deadlines.tsx` Link; Vitest link attrs |
| A3W9-2-003 | тот же DTO + `my-tasks.tsx` через `TaskRowLink` |
| A3W9-2-004 | `var(--red/--amber/--orange/--blue/--t3)` в deadlines/my-tasks; source-guard тест |
| A3W9-2-005 | system-health/unread → `var(--green/--red/--amber)` + `shadow-[…var(--gG/--gR)]`; нет inline rgba |

**Mutation notes:**
- 001: вернуть `{bm.channel_id}` / `{bm.message_id}` в panel → `bookmarks-preview` RED (UUID в DOM / нет «general»)
- 002/003: убрать `TaskRowLink`/`Link` → `getByRole("link")` RED
- 004: вернуть `bg-red-500` → source-guard RED
- 005: вернуть `bg-green-500` + `style={{boxShadow: rgba…}}` → source-guard RED

**Tests:** vitest messenger+dashboard TASK-173 — 7/7 PASS; `TestBookmarks_AddListRemove` PASS; dashboard wave107 ProjectID — compile blocked OOS (`ent/schema/webhookendpoint` unused import, параллельный TASK-171). tsc: OOS errors в `time-tracker` / `index.lazy` (чужие WIP).

**001 approach:** BE enrich (channel cache + FindMessageByID), не FE N+1.

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30 · W7 slot 4 · fe1). Слой **frontend** (001 FE-label; BE enrich — отдельный слой, см. Gaps). Typecheck: `cd frontend && npx tsc --noEmit` — **PASS** (не ловит UUID-label / Link / token regressions).

| ID | что ломать (file:line + фрагмент) | какой тест (`it` / команда) | слой |
|----|-----------------------------------|-----------------------------|------|
| A3W9-2-001 | `bookmarks-panel.tsx:15-17` — `bookmarkTitle` → `return bm.channel_id` (или `:21-23` `bookmarkSubtitle` → `bm.message_id`); **и/или** рендер сырого UUID как primary label кнопки | describe `BookmarksPanel preview labels (A3W9-2-001 / TASK-173)` → `it("shows channel_name + content_preview, not raw UUIDs as primary label")`; `it("falls back to «Канал»/«Сообщение»…")` | frontend |
| A3W9-2-002 | `task-row-link.tsx:16-26` — убрать `Link` / всегда static container; **и/или** `deadlines.tsx:67` — не оборачивать row в `TaskRowLink` | describe `Dashboard widgets A3W9-2 (TASK-173)` → `it("deadlines row links to issue route with project_id + key (A3W9-2-002)")` | frontend |
| A3W9-2-003 | то же `TaskRowLink` / `my-tasks.tsx:74` — убрать Link-обёртку | `it("my-tasks row uses the same Link pattern (A3W9-2-003)")` | frontend |
| A3W9-2-004 | `deadlines.tsx` / `my-tasks.tsx:14-19` — вернуть `bg-red-500` / `bg-amber-500` / raw `*-500` вместо `var(--red|--amber|…)` | `it("deadlines/my-tasks source has no raw Tailwind priority colors (A3W9-2-004)")` | frontend |
| A3W9-2-005 | `system-health.tsx:20-38` — `bg-green-500` + inline `style={{boxShadow: rgba…}}`; **и/или** `unread-messages.tsx:24,35-36` — `text-green-500`/`bg-amber-500` | `it("system-health/unread use semantic tokens, no *-500 or inline rgba glow (A3W9-2-005)")` | frontend |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| FE bookmarkTitle ≠ UUID | ✅ Vitest (mock useQuery data) | |
| BE `channel_name`/`content_preview` enrich | ❌ FE Vitest | **layer_mismatch flag:** тест мокает `useQuery` с preview — **не** ловит откат BE `ListBookmarks`; нужен hubtest messenger |
| TaskRowLink / deadlines / my-tasks | ✅ RTL (Link mock) | |
| token source-guards 004/005 | ✅ `readFileSync` source it | |
| `npx tsc --noEmit` | ❌ | не ловит 001–005 |

**Confidence:** CONFIRMED (FE); BE-001 — GAP для mutation prove на BE-слое.

## Mutation prove (fe2 · W11 slot 4 · 2026-07-30)

**Verdict:** 5/5 FE **PROVEN** · BE enrich 001 → **N/A** (layer_mismatch, не fake-PROVEN) · статус **Закрыта**

Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W11-fe2.md`

| ID | Verdict |
|----|---------|
| A3W9-2-001 FE | PROVEN (`bookmarkTitle` → UUID → bookmarks-preview RED) |
| A3W9-2-001 BE enrich | N/A — FE mock useQuery; нужен hubtest |
| A3W9-2-002 | PROVEN (TaskRowLink static → deadlines link RED) |
| A3W9-2-003 | PROVEN (тот же TaskRowLink → my-tasks link RED) |
| A3W9-2-004 | PROVEN (raw `*-500` → source-guard RED) |
| A3W9-2-005 | PROVEN (`bg-green-500` + rgba → source-guard RED) |

Post-restore: 6/6 Vitest PASS. Коммит не создавался.
