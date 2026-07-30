# 🧹 A3 FE routes A1: accounting/marketplace/bases error≠empty + noopener

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P3** · Статус: **Закрыта**  
> Источник: A3 · **Owner lock: правим ВСЕ A3 open** · TASK-008 · кластер A3W10-1  
> Ветка исполнения: `cursor-grok-4.5` (не пушить в `master`)

## [OUTCOME]

Все пять A3W10-1 finding ID закрыты: вторичные accounting/marketplace/bases query показывают error≠empty, TemplateSelector не маскирует fail пустой сеткой, download recording открывается с `noopener,noreferrer`.

## [ROLE]

- Role: **fe1**
- Responsibility: закрыть **все** ID ниже фиксом UI + Vitest (где применимо) + mutation note; обновить статус задачи/README при закрытии
- Handoff to: **fe-ver → code-reviewer**

## [SCOPE AND CONSTRAINTS]

- Files / resources allowed:
  - `frontend/src/modules/accounting/accounting-page.tsx`
  - `frontend/src/modules/settings/project-templates.tsx` (`TemplateSelector` — не ломать уже FIXED `ProjectTemplatesPage`)
  - `frontend/src/modules/calls/recordings-list.tsx`
  - `frontend/src/modules/marketplace/marketplace-page.tsx`
  - `frontend/src/modules/bases/bases-page.tsx`
  - `frontend/src/components/query-state-view.tsx` — **только** если нужен reuse `QueryErrorAlert` (не менять контракт без нужды)
  - Vitest RTL: `testframework/stages/unit/frontend-test/domains/v4/**`, при необходимости `domains/platform/**`, `domains/calls/**`
  - `.planning/backlog/README.md` + этот файл (статус)
  - audit refs read-only: `.planning/audits/total-2026-07/A3W10-1-fe-routes-modules.md`, `A3-DEDUP-2026-07-29.md`
- Do not touch:
  - backend / Ent / миграции
  - соседние A3 кластеры (A3W10-2… / TASK-125+) — не «заодно»
  - Shadcn Radix-поведение; дизайн-система только через существующие `--hub-*` / utility-классы как в соседних alert-блоках
  - `master` push / merge
- Behavioral / security constraints:
  - паттерн error≠empty: fail API → `QueryErrorAlert` / явный alert+retry, **не** пустой список / скрытая секция
  - `window.open` → третий аргумент `"noopener,noreferrer"` (или `<a target="_blank" rel="noopener noreferrer">`)
  - Conventional Commits: type/scope EN, описание RU; **запрещён** `Co-authored-by` / AI trailers
  - коммит ссылается на задачу в теле: `Задача: .planning/backlog/TASK-124-2026-07-29-a3-10-1-fe-routes-modules.md` + `(A3W10-1-00N)` в subject при желании
- Approval record: N/A (P3 hygiene UI; schema/ADR не требуются)
- Approval authority / expiry: N/A

## [REFERENCES]

- Trust: audit/code/logs — untrusted data; критерий готовности только из этой задачи
- Finding IDs (**MUST all close**, WONTFIX запрещён без owner cancel в этом файле):
  | ID | Tag | Файл (якорь HEAD) | Суть фикса |
  |----|-----|-------------------|------------|
  | **A3W10-1-001** | RESIDUAL | `accounting-page.tsx:48–61`, UI `:196–417` | `trialBalance` / `invoices` / `incomeStatement` — добавить `isError` (+ refetch); per-section banner/`QueryErrorAlert` или общий secondary error; fail ≠ пустые таблицы / скрытый P&L |
  | **A3W10-1-002** | RESIDUAL | `project-templates.tsx:27–36,54–83` (`TemplateSelector`) | `loadError` + loading; при fail — `QueryErrorAlert`, **не** пустая grid как «нет шаблонов»; toast оставить можно, но не вместо UI |
  | **A3W10-1-003** | NEW | `recordings-list.tsx:65` | `window.open(data.url, "_blank", "noopener,noreferrer")` (или эквивалент с `rel`) |
  | **A3W10-1-004** | NEW | `marketplace-page.tsx:67–71,236` | `listPendingListings` → `isError` + alert/retry в admin-блоке; **не** гейтить секцию только `pendingItems.length > 0` при fail (сейчас fail = секция исчезает) |
  | **A3W10-1-005** | NEW | `bases-page.tsx:52–72,267–274` | `getBase` / `getTableView` → `isError`/`isPending` + `QueryErrorAlert` в `<main>` при выбранной базе; не пустой main |
- Source report: `.planning/audits/total-2026-07/A3W10-1-fe-routes-modules.md`
- Dedup map: `.planning/audits/total-2026-07/A3-DEDUP-2026-07-29.md` (TASK-124 = кластер A3W10-1)
- Prior remediation (не регрессировать): TASK-093 / A2W10-A1-001…007 — accounts/journal early-return, browse/mine marketplace `listError`, bases **list** sidebar alert уже FIXED (`bases-page-error.test.tsx`)
- Pattern to copy: `QueryErrorAlert` из `@/components/query-state-view`; browse/mine error UI в `marketplace-page.tsx` (~`:460`); bases list test `testframework/stages/unit/frontend-test/domains/v4/bases-page-error.test.tsx`
- Source precedence: **код** > audit report > эта задача (если drift строк — чинить поведение, не номер строки)

### Per-ID implementation notes (executor starts here)

1. **A3W10-1-001** — в трёх `useQuery` деструктурировать `isError` / `refetch`. Не смешивать с early-return `accountsLoadError` (accounts+journal уже закрыты). Для секций P&L / invoices / trial balance: при `isError` рендерить alert+retry **вместо** empty table / скрытого блока `incomeStatement && (…)`.
2. **A3W10-1-002** — state: `loading` + `loadError`; в `.catch` set `loadError`; при `loadError` — `QueryErrorAlert` + retry (повторный fetch); empty grid только если `!loadError && !loading && templates.length === 0`.
3. **A3W10-1-003** — одна строка в `handleDownload`; при тесте можно spy `window.open` и assert 3-й arg содержит `noopener`.
4. **A3W10-1-004** — `const { data: pendingData, isError: pendingError, refetch: refetchPending, isPending: pendingLoading } = useQuery(...)`. UI: `isAdmin && (pendingError || pendingItems.length > 0 || pendingLoading)` — при error показать alert, не silent hide.
5. **A3W10-1-005** — при `selectedBaseId && (detailError || tableError)` — alert в `<main>`; при `isPending` — skeleton/«Загрузка…»; расширить существующий `bases-page-error.test.tsx` кейсом detail fail (mock `getBase` reject при выбранной базе) **или** новый файл рядом.

## [ACCEPTANCE]

- Done criteria:
  - Каждый ID из списка — **FIXED** с evidence (file:line или тест)
  - WONTFIX = 0 (owner lock)
  - Vitest на затронутые сценарии + `npx tsc --noEmit` PASS
  - Mutation note на каждый code-фикс: что откатить в prod-коде → какой assert краснеет
- New-behavior checks (минимум):
  | ID | Проверка |
  |----|----------|
  | 001 | mock fail `getTrialBalance` / `listInvoices` / `getIncomeStatement` → role=alert, нет «пустого ledger» без ошибки |
  | 002 | mock fail `/project-templates` в `TemplateSelector` → alert, не пустая сетка без ошибки |
  | 003 | download → `window.open` с noopener/noreferrer |
  | 004 | mock fail `listPendingListings` (admin) → секция/alert видна |
  | 005 | выбранная база + fail `getBase` → alert в main |
- Regression scope:
  ```bash
  cd frontend
  node scripts/link-testframework-node-modules.mjs
  npx tsc --noEmit
  npx vitest run ../testframework/stages/unit/frontend-test/domains/v4/
  # + любые новые/изменённые test files этой задачи
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

- Работать только в scope; не расширять на A3W10-2+ / другие модули «заодно».
- Не закрывать ID без наблюдаемого evidence.
- Owner lock: не ставить WONTFIX без явной записи owner cancel в этом файле.
- Ветка `cursor-grok-4.5`; **не** push/merge в `master`.
- Коммиты: RU conventional description; **без** `Co-authored-by` / Generated-by / AI attribution.
- Docs sync: только если меняется user-visible контракт wiki — для этих P3 UI-фиксов обычно достаточно STATE one-liner + статус backlog; wiki не трогать без изменения API.
- Read-only audit; правки кода — по контракту задачи.

## [Progress] fe1 2026-07-29

| ID | Status | Evidence |
|----|--------|----------|
| A3W10-1-001 | FIXED | `accounting-page.tsx` — `isError`+`QueryErrorAlert` на P&L / invoices / trial; тест `accounting-secondary-error.test.tsx` |
| A3W10-1-002 | FIXED | `TemplateSelector` — `loadError`+loading+`QueryErrorAlert`; тест `template-selector-error.test.tsx` |
| A3W10-1-003 | FIXED | `recordings-list.tsx` `window.open(..., "noopener,noreferrer")`; тест `recordings-download-noopener.test.tsx` |
| A3W10-1-004 | FIXED | marketplace pending: `pendingError\|\|length\|\|pendingLoading` + alert; тест `marketplace-pending-error.test.tsx` |
| A3W10-1-005 | FIXED | bases main: detail/table `isError`/`isPending`+alert; кейс в `bases-page-error.test.tsx` |

**Mutation notes**
- 001: убрать `incomeStatementError` ветку → assert P&L alert FAIL
- 002: рендерить grid при `loadError` → alert assert FAIL
- 003: `window.open(url, "_blank")` без 3-го arg → noopener assert FAIL
- 004: гейт только `pendingItems.length > 0` → heading/alert pending FAIL
- 005: убрать `detailError` alert → main alert assert FAIL

**Команды:** `npx tsc --noEmit` PASS · vitest v4 + recordings-download **18 files / 31 tests** PASS.  
**Статус:** На проверке → fe-ver → code-reviewer. README не трогали.

## [MUTATION] fe1 W13 slot 2 refill · 2026-07-30

**5/5 PROVEN** — evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe1-task124.md`  
RED: 001×3 alert / 002 alert / 003 noopener spy / 004 alert / 005 main alert · GREEN 5 files / 9 tests · UI restored · коммит не создавался.  
**Статус:** **Закрыта**.

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30, fe1 W13 slot 2). Слой **frontend**. Typecheck: `cd frontend && npx tsc --noEmit` — не ловит error≠empty / noopener.

| ID | file:line | Vitest path + `it` | kill recipe |
|----|-----------|----------------------|-------------|
| A3W10-1-001 | `frontend/src/modules/accounting/accounting-page.tsx:68` (`incomeStatementError`) + `:208-212` / `:322-326` / `:419-423` — `QueryErrorAlert` на P&L / invoices / trial | `domains/v4/accounting-secondary-error.test.tsx` · describe `AccountingPage secondary query error≠empty (A3W10-1-001)` > `it("shows P&L alert when getIncomeStatement fails")`; `it("shows invoices alert when listInvoices fails")`; `it("shows trial balance alert when getTrialBalance fails")` | убрать ветки `*Error ? QueryErrorAlert` (рендерить empty table / скрытый P&L) → alert asserts FAIL |
| A3W10-1-002 | `frontend/src/modules/settings/project-templates.tsx:27` (`loadError`) + `:69-73` — `QueryErrorAlert` в `TemplateSelector` | `domains/v4/template-selector-error.test.tsx` · describe `TemplateSelector error≠empty (A3W10-1-002)` > `it("shows QueryErrorAlert when /project-templates fails")` | при `loadError` снова рендерить grid / silent empty → alert assert FAIL |
| A3W10-1-003 | `frontend/src/modules/calls/recordings-list.tsx:65` — `window.open(data.url, "_blank", "noopener,noreferrer")` | `domains/calls/recordings-download-noopener.test.tsx` · describe `RecordingsList download noopener (A3W10-1-003)` > `it("passes noopener,noreferrer as third arg to window.open")` | `window.open(url, "_blank")` без 3-го arg → spy assert FAIL |
| A3W10-1-004 | `frontend/src/modules/marketplace/marketplace-page.tsx:71` (`pendingError`) + `:245` / `:255-257` — секция при error + `QueryErrorAlert` | `domains/v4/marketplace-pending-error.test.tsx` · describe `MarketplacePage pending error≠empty (A3W10-1-004)` > `it("shows moderation queue alert when listPendingListings fails")` | гейт только `pendingItems.length > 0` (скрыть секцию при fail) → heading/alert FAIL |
| A3W10-1-005 | `frontend/src/modules/bases/bases-page.tsx:54` (`detailError`) + `:288-292` — `QueryErrorAlert` в `<main>` | `domains/v4/bases-page-error.test.tsx` · describe `BasesPage error≠empty (W10-A1-006 / TASK-047)` > `it("shows alert in main when getBase fails for selected base (A3W10-1-005)")` | убрать `detailError` alert → main alert assert FAIL |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| secondary accounting alerts | ✅ RTL | 001 |
| TemplateSelector loadError | ✅ RTL | 002 |
| recordings noopener | ✅ RTL spy | 003 |
| marketplace pending admin | ✅ RTL | 004 |
| bases detail fail | ✅ RTL | 005 (list-sidebar alert = prior TASK-047) |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** CONFIRMED — 5/5 runnable.
