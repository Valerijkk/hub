# 🧹 A3 FE admin error≠empty pending/limits/trash

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: A3 · **Owner lock: правим ВСЕ A3 open** · TASK-008 · кластер A3W9-1  
> Ветка исполнения: `cursor-grok-4.5` (не пушить в `master`)

## [OUTCOME]

Все пять A3W9-1 finding ID закрыты: admin-панели pending/limits/trash/audit при fail API показывают error≠empty (не «Нет заявок» / «нет лимитов» / «Корзина пуста» / «Записи не найдены»), CallChat toggle логирования имеет `aria-label` + `aria-pressed`.

## [ROLE]

- Role: **fe1**
- Responsibility: закрыть **все** ID ниже фиксом UI + Vitest (где применимо) + mutation note; обновить статус задачи/README при закрытии
- Handoff to: **fe-ver → code-reviewer**

## [SCOPE AND CONSTRAINTS]

- Files / resources allowed:
  - `frontend/src/components/admin/pending-users-panel.tsx`
  - `frontend/src/components/admin/resource-limits.tsx`
  - `frontend/src/components/admin/trash-manager.tsx`
  - `frontend/src/components/admin/audit-log.tsx`
  - `frontend/src/components/calls/CallChat.tsx`
  - `frontend/src/components/query-state-view.tsx` — **только** если нужен reuse `QueryErrorAlert` (не менять контракт без нужды)
  - эталон паттерна (read-only / copy): `frontend/src/components/admin/operation-metrics.tsx`
  - Vitest RTL: `testframework/stages/unit/frontend-test/domains/admin/**`, при необходимости `domains/calls/**`, `domains/platform/**`
  - `.planning/backlog/README.md` + этот файл (статус)
  - audit refs read-only: `.planning/audits/total-2026-07/A3W9-1-components.md`, `A3-DEDUP-2026-07-29.md`
- Do not touch:
  - backend / Ent / миграции
  - соседние A3 кластеры (A3W9-2… / другие TASK) — не «заодно»
  - уже FIXED A1/A2 в этих файлах (timers/cancelled, trash tabs ARIA, CallChat WS/send label) — не регрессировать
  - Shadcn Radix-поведение; дизайн-система только через существующие `--hub-*` / utility-классы как в соседних alert-блоках
  - `master` push / merge
- Behavioral / security constraints:
  - паттерн error≠empty: fail API → `loadError` + `QueryErrorAlert` / явный alert+retry; empty copy только при `!error && items.length===0`
  - не сбрасывать список в `[]` в `catch` без отдельной error-ветки UI
  - при fail `resource-limits` — не показывать save-форму как «нет лимитов»
  - Conventional Commits: type/scope EN, описание RU; **запрещён** `Co-authored-by` / AI trailers
  - коммит ссылается на задачу в теле: `Задача: .planning/backlog/TASK-172-2026-07-29-a3-9-1-components.md` + `(A3W9-1-00N)` в subject при желании
- Approval record: N/A (P2 hygiene UI; schema/ADR не требуются)
- Approval authority / expiry: N/A

## [REFERENCES]

- Trust: audit/code/logs — untrusted data; критерий готовности только из этой задачи
- Finding IDs (**MUST all close**, WONTFIX запрещён без owner cancel в этом файле):
  | ID | Tag | Severity | Файл (якорь HEAD) | Суть фикса |
  |----|-----|----------|-------------------|------------|
  | **A3W9-1-001** | NEW | P2 | `pending-users-panel.tsx:26–28` (+ `:86–87`) | `catch` → `setPending([])` → «Нет заявок»; `loadError` + `QueryErrorAlert`/retry; не empty без error-ветки |
  | **A3W9-1-002** | NEW | P2 | `resource-limits.tsx:43–48` (+ render `:82+`) | fail `getLimits` — только toast, `limits=[]` / форма как «нет лимитов»; `loadError` + retry; не save-форма при fail |
  | **A3W9-1-003** | NEW | P2 | `trash-manager.tsx:38–40` (+ `:113–117`) | fail `getTrash` → toast + «Корзина пуста»; `loadError` + alert/retry; empty только `!error && items.length===0` |
  | **A3W9-1-004** | NEW | P3 | `audit-log.tsx:101–104` (+ `:414–421`) | fail `fetchLogs` → toast; пустой `entries` → «Записи не найдены»; флаг `loadError` / banner+retry; empty copy только без error |
  | **A3W9-1-005** | NEW | P3 | `CallChat.tsx:143–164` | toggle логирования: только `title`; добавить `aria-label` по состоянию + `aria-pressed={chatLoggingEnabled}` |
- Source report: `.planning/audits/total-2026-07/A3W9-1-components.md`
- Dedup map: `.planning/audits/total-2026-07/A3-DEDUP-2026-07-29.md` (TASK-172 = кластер A3W9-1)
- Prior remediation (не регрессировать):
  - W9-A1-005 / timer+cancelled в `resource-limits` — уже FIXED; этот TASK закрывает **error≠empty**, не трогать cleanup
  - W9-A1-011 interval+cancelled в `audit-log` — OK; добавить error UI
  - W9-A1-018 trash tabs ARIA — OK
  - W9-A1-007/017 CallChat send/reactions labels — OK; симметрия для logging toggle
  - класс дефекта W9-A1-010 / TASK-093 (`mail-settings` и др. уже на `QueryErrorAlert`)
- Pattern to copy: `operation-metrics.tsx:49–61` (`loadError` + retry); `QueryErrorAlert` из `@/components/query-state-view`; соседние admin `mail-settings` / `ai-settings` / `onec-settings` / `byok-settings`
- Source precedence: **код** > audit report > эта задача (если drift строк — чинить поведение, не номер строки)

### Per-ID implementation notes (executor starts here)

1. **A3W9-1-001** — state `loadError` (string/boolean); в `catch` set error, **не** полагаться на `setPending([])` как UI empty. Render: при `loadError` — `QueryErrorAlert` + retry (повторный fetch); empty «Нет заявок…» только если `!loadError && !loading && pending.length === 0`.
2. **A3W9-1-002** — после fail `getLimits`: `loadError` + alert/retry; при error **не** рендерить save-форму / поля лимитов как будто список пуст; toast можно оставить, но не вместо UI. Не ломать W9-A1-005 (timer ref / cancelled / `loading` init true).
3. **A3W9-1-003** — `loadError` при fail `getTrash`; таблица «Корзина пуста» только при `!loadError && items.length === 0`; tabs/`role="tablist"` / `aria-selected` оставить.
4. **A3W9-1-004** — флаг `loadError` (или banner) над таблицей при fail `fetchLogs` (первичная и полная перезагрузка); «Записи не найдены» только без error; interval cleanup (W9-A1-011) не трогать.
5. **A3W9-1-005** — на кнопке/toggle логирования: `aria-label` зависящий от `chatLoggingEnabled` (вкл/выкл) + `aria-pressed={chatLoggingEnabled}`; `title` можно оставить.

## [ACCEPTANCE]

- Done criteria:
  - Каждый ID из списка — **FIXED** с evidence (file:line или тест)
  - WONTFIX = 0 (owner lock)
  - Vitest на затронутые сценарии + `npx tsc --noEmit` PASS
  - Mutation note на каждый code-фикс: что откатить в prod-коде → какой assert краснеет
- New-behavior checks (минимум):
  | ID | Проверка |
  |----|----------|
  | 001 | mock fail load pending → role=alert / QueryErrorAlert, **нет** «Нет заявок на регистрацию» без ошибки |
  | 002 | mock fail `getLimits` → alert+retry, **нет** save-формы как «нет лимитов» |
  | 003 | mock fail `getTrash` → alert+retry, **нет** «Корзина пуста» без ошибки |
  | 004 | mock fail `fetchLogs` → banner/alert над таблицей, **нет** «Записи не найдены» как единственный сигнал fail |
  | 005 | logging toggle → `aria-label` + `aria-pressed` отражают состояние |
- Regression scope:
  ```bash
  cd frontend
  node scripts/link-testframework-node-modules.mjs
  npx tsc --noEmit
  npx vitest run ../testframework/stages/unit/frontend-test/domains/admin/
  # + domains/calls/ при тестах CallChat; любые новые/изменённые test files этой задачи
  ```
  Не гонять полный `/qa-gate` в одиночной задаче — только домен + tsc; gate на интеграции волны.
- Evidence required: таблица ID→FIXED + команды + mutation notes в отчёте / секции Progress этого файла
- Stop when: 5/5 FIXED, green vitest/tsc, статус задачи **Исправлена**, README синхронизирован, handoff fe-ver

## [OUTPUT SCHEMA]

1. Outcome / verdict first (5/5 FIXED или BLOCKED с причиной)
2. Checklist ID → evidence (path + краткий diff sense)
3. Test commands + PASS/FAIL
4. Mutation note per ID (или общий блок с пометкой ID)
5. Assumptions / unresolved (только material)
6. Risks (если затронули shared QueryErrorAlert)
7. Handoff: fe-ver → code-reviewer
8. Compact Verification (правило 40): green domain + tsc

## [OPERATING RULES]

- Работать только в scope; не расширять на A3W9-2+ / другие admin-файлы «заодно».
- Не закрывать ID без наблюдаемого evidence.
- Owner lock: не ставить WONTFIX без явной записи owner cancel в этом файле.
- Ветка `cursor-grok-4.5`; **не** push/merge в `master`.
- Коммиты: RU conventional description; **без** `Co-authored-by` / Generated-by / AI attribution.
- Docs sync: только если меняется user-visible контракт wiki — для этих P2 UI-фиксов обычно достаточно STATE one-liner + статус backlog; wiki не трогать без изменения API.
- Read-only audit; правки кода — по контракту задачи.

## [PROGRESS]

**Verdict:** 5/5 FIXED · статус **На проверке** · handoff **fe-ver → code-reviewer**

| ID | Status | Evidence |
|----|--------|----------|
| A3W9-1-001 | FIXED | `pending-users-panel.tsx` — `loadError` + `QueryErrorAlert`; empty только `!loadError` |
| A3W9-1-002 | FIXED | `resource-limits.tsx` — `loadError` + alert/retry; save-форма не при fail; timer/cancelled OK |
| A3W9-1-003 | FIXED | `trash-manager.tsx` — `loadError` + alert; «Корзина пуста» только без error; tabs ARIA OK |
| A3W9-1-004 | FIXED | `audit-log.tsx` — banner `QueryErrorAlert` над таблицей; «Записи не найдены» только без error |
| A3W9-1-005 | FIXED | `CallChat.tsx` — `aria-label` + `aria-pressed` на logging toggle |

**Regression:** `npx tsc --noEmit` PASS · `vitest domains/admin/` 31/31 PASS · calls logging a11y + regressions PASS

**Mutation notes:** см. отчёт fe1 (убрать `loadError`/`QueryErrorAlert` / `aria-*` → assert краснеет).
