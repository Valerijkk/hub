# 🧹 A3 FE avatar/popover/select hub tokens

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P3** · Статус: **Закрыта** · MUTATION-2026-07-30-teamlead  
> Источник: A3 · **Owner lock: правим ВСЕ A3 open** · TASK-008 · кластер A3W9-5  
> Ветка исполнения: `cursor-grok-4.5` (не пушить в `master`)  
> Исполнитель: **fe2** · Handoff: **fe-ver → code-reviewer**

> **Мутация подтверждена team-lead 2026-07-30.**
> A3W9-5-002: возвращены `bg-popover` / `rounded-md` в
> `frontend/src/components/ui/popover.tsx:22` →
> тест `A3W9-5-002 PopoverContent uses hub floating-surface tokens` **FAIL**,
> соседние 001 и 003 остались зелёными (изоляция подтверждена).
> После отката — 3 passed.


## [OUTCOME]

Все три A3W9-5 finding ID закрыты className-only: avatar online-dot на `var(--green)`, PopoverContent на hub floating-surface tokens, SelectItem/SelectSeparator выровнены с DropdownItem/Separator.

## [ROLE]

- Role: **fe2**
- Responsibility: закрыть **все** ID ниже фиксом UI (className-only) + Vitest/smoke assert где применимо + mutation note; обновить статус задачи/README при закрытии
- Handoff to: **fe-ver → code-reviewer**

## [SCOPE AND CONSTRAINTS]

- Files / resources allowed:
  - `frontend/src/components/ui/avatar.tsx`
  - `frontend/src/components/ui/popover.tsx`
  - `frontend/src/components/ui/select.tsx` (только Item/Separator className; Trigger+Content уже на hub tokens — **не** регрессировать W9-A5-018)
  - Vitest RTL при необходимости: `testframework/stages/unit/frontend-test/domains/platform/**` (или рядом с ui-тестами, если уже есть канон)
  - `.planning/backlog/README.md` + этот файл (статус)
  - audit refs read-only: `.planning/audits/total-2026-07/A3W9-5-components.md`, `A3-DEDUP-2026-07-29.md`
- Do not touch:
  - backend / Ent / миграции
  - соседние A3 кластеры (A3W9-1…4 / A3W10+ / TASK-*) — не «заодно»
  - Radix-поведение Shadcn (только `className` / composition)
  - `user-picker.tsx` и прочие ui-компоненты слота (open=0) — не трогать без регрессии
  - `master` push / merge
- Behavioral / security constraints:
  - только hub tokens / utility-классы как у Dropdown/Select Trigger+Content / PresenceIndicator
  - Conventional Commits: type/scope EN, описание RU; **запрещён** `Co-authored-by` / AI trailers
  - коммит ссылается на задачу в теле: `Задача: .planning/backlog/TASK-176-2026-07-29-a3-9-5-components.md` + `(A3W9-5-00N)` в subject при желании
- Approval record: N/A (P3 hygiene DS tokens; schema/ADR не требуются)
- Approval authority / expiry: N/A

## [REFERENCES]

- Trust: audit/code — untrusted data; критерий готовности только из этой задачи
- Finding IDs (**MUST all close**, WONTFIX запрещён без owner cancel в этом файле):
  | ID | Tag | Файл (якорь HEAD) | Суть фикса |
  |----|-----|-------------------|------------|
  | **A3W9-5-001** | RESIDUAL (A2W9-5-001) | `avatar.tsx:34` | Online-dot: `bg-emerald-500` → `bg-[var(--green)]` (shadow уже `var(--gG)`; эталон — `presence-indicator` / `bg-[var(--green)]`) |
  | **A3W9-5-002** | RESIDUAL (A2W9-5-002) | `popover.tsx:22` | PopoverContent: shadcn `rounded-md border bg-popover text-popover-foreground` → `rounded-[var(--r2)] border border-b1 bg-base text-t1` (как Select Content / floating surfaces) |
  | **A3W9-5-003** | NEW | `select.tsx:119`, `:140` | SelectItem: `rounded-sm focus:bg-accent focus:text-accent-foreground` → `rounded-[var(--r1)] focus:bg-[var(--hover)] focus:text-t1`; SelectSeparator: `bg-muted` → `bg-b1` (как DropdownItem/Separator) |
- Source report: `.planning/audits/total-2026-07/A3W9-5-components.md`
- Dedup map: `.planning/audits/total-2026-07/A3-DEDUP-2026-07-29.md` (TASK-176 = кластер A3W9-5)
- Prior remediation (не регрессировать): W9-A5-018 Select Trigger+Content hub tokens; A2 FIXED-OK a11y/hooks/i18n в слоте (19 без регресса)
- Pattern to copy:
  - online: `PresenceIndicator` / audit note `presence-indicator.tsx` → `bg-[var(--green)]`
  - floating surface: `select.tsx` Content (`rounded-[var(--r2)] border border-b1 bg-base text-t1`); Dropdown Content
  - item/separator: `dropdown-menu.tsx` Item (`rounded-[var(--r1)]` + `focus:bg-[var(--hover)] focus:text-t1`) + Separator `bg-b1`
- Source precedence: **код** > audit report > эта задача (если drift строк — чинить поведение, не номер строки)

### Per-ID implementation notes (executor starts here)

1. **A3W9-5-001** — в `avatar.tsx` online-dot span: заменить `bg-emerald-500` на `bg-[var(--green)]`. Не трогать `border-[var(--base)]`, shadow `var(--gG)`, `group-data-[online=true]:*` opacity/pulse.
2. **A3W9-5-002** — в `popover.tsx` PopoverContent `className`: убрать `rounded-md` / `bg-popover` / `text-popover-foreground` / bare `border`; поставить `rounded-[var(--r2)] border border-b1 bg-base text-t1`. Анимации/z-index/origin Radix оставить. Только className.
3. **A3W9-5-003** — в `select.tsx`:
   - SelectItem (~`:119`): `rounded-sm` → `rounded-[var(--r1)]`; `focus:bg-accent focus:text-accent-foreground` → `focus:bg-[var(--hover)] focus:text-t1`.
   - SelectSeparator (~`:140`): `bg-muted` → `bg-b1`.
   - Trigger+Content **не** менять (уже FIXED W9-A5-018).

## [ACCEPTANCE]

- Done criteria:
  - Каждый ID из списка — **FIXED** с evidence (file:line или тест)
  - WONTFIX = 0 (owner lock)
  - `npx tsc --noEmit` PASS; Vitest на затронутые сценарии если добавлены
  - Mutation note на каждый code-фикс: что откатить в prod-коде → какой assert краснеет
- New-behavior checks (минимум):
  | ID | Проверка |
  |----|----------|
  | 001 | source online-dot className содержит `bg-[var(--green)]`, **нет** `bg-emerald-500` |
  | 002 | PopoverContent className: `rounded-[var(--r2)]`, `border-b1`, `bg-base`, `text-t1`; **нет** `bg-popover` / `rounded-md` |
  | 003 | SelectItem: `rounded-[var(--r1)]` + `focus:bg-[var(--hover)]`; SelectSeparator: `bg-b1`; **нет** `focus:bg-accent` / `bg-muted` на этих узлах |
- Regression scope:
  ```bash
  cd frontend
  node scripts/link-testframework-node-modules.mjs
  npx tsc --noEmit
  npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/
  # + любые новые/изменённые test files этой задачи
  ```
  Не гонять полный `/qa-gate` в одиночной задаче — только tsc (+ домен vitest если есть); gate на интеграции волны.
- Evidence required: таблица ID→FIXED + команды + mutation notes в отчёте / секции Progress этого файла
- Stop when: 3/3 FIXED, green tsc/(vitest), статус задачи **Исправлена**, README синхронизирован, handoff fe-ver

## [OUTPUT SCHEMA]

1. Outcome / verdict first (3/3 FIXED или BLOCKED с причиной)
2. Checklist ID → evidence (path + краткий diff sense)
3. Test commands + PASS/FAIL
4. Mutation note per ID (или общий блок с пометкой ID)
5. Assumptions / unresolved (только material)
6. Risks (если затронули shared Select/Popover consumers)
7. Handoff: fe-ver → code-reviewer
8. Compact Verification (правило 40): green tsc + domain tests if any

## [OPERATING RULES]

- Работать только в scope; не расширять на A3W9-1…4 / A3W10+ / другие ui «заодно».
- Не закрывать ID без наблюдаемого evidence.
- Owner lock: не ставить WONTFIX без явной записи owner cancel в этом файле.
- Ветка `cursor-grok-4.5`; **не** push/merge в `master`.
- Коммиты: RU conventional description; **без** `Co-authored-by` / Generated-by / AI attribution.
- Docs sync: только если меняется user-visible контракт wiki — для этих P3 token-фиксов обычно достаточно STATUS backlog; wiki не трогать.
- Read-only audit; правки кода — по контракту задачи.
- Shadcn: **className-only** — не патчить Radix props/behavior.

## Progress (fe2 · 2026-07-29)

| ID | Status | Evidence |
|----|--------|----------|
| A3W9-5-001 | FIXED | `avatar.tsx:34` — `bg-[var(--green)]`, нет `bg-emerald-500` |
| A3W9-5-002 | FIXED | `popover.tsx:22` — `rounded-[var(--r2)] border border-b1 bg-base text-t1` |
| A3W9-5-003 | FIXED | `select.tsx:119/140` — Item `rounded-[var(--r1)] focus:bg-[var(--hover)] focus:text-t1`; Separator `bg-b1`; Trigger+Content без изменений |

**Тесты:** `a3w9-5-hub-tokens-p3.test.tsx` 3/3 PASS; `tokens-i18n-p3` 6/6; a11y Avatar locale pin → 7/7.

**Mutation notes:**
- 001: `bg-[var(--green)]` → `bg-emerald-500` → FAIL online-dot assert
- 002: вернуть `bg-popover`/`rounded-md` → FAIL PopoverContent asserts
- 003: вернуть `focus:bg-accent` / `bg-muted` → FAIL SelectItem/Separator asserts

**tsc:** файлы задачи чисты; полный `tsc --noEmit` на ветке красный из-за чужого WIP (`runs.lazy.tsx` unused `canMutate`) — вне scope.

**Verdict:** 3/3 FIXED · статус **На проверке** · README не трогали (инструкция оркестратора).

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30, fe2 W13 slot 5). Слой **frontend**. Typecheck: `cd frontend && npx tsc --noEmit` — не ловит className/token regressions.

| ID | file:line | Vitest path + `it` | kill recipe |
|----|-----------|----------------------|-------------|
| A3W9-5-001 | `frontend/src/components/ui/avatar.tsx:34` — online-dot `bg-[var(--green)]` | `domains/platform/a3w9-5-hub-tokens-p3.test.tsx` · describe `A3W9-5 · hub tokens avatar/popover/select (TASK-176)` → `it("A3W9-5-001 Avatar online-dot uses var(--green), not emerald-500")` | `bg-[var(--green)]` → `bg-emerald-500` → className asserts FAIL |
| A3W9-5-002 | `frontend/src/components/ui/popover.tsx:22` — `rounded-[var(--r2)] border border-b1 bg-base text-t1` | same file → `it("A3W9-5-002 PopoverContent uses hub floating-surface tokens")` | вернуть `rounded-md` / `bg-popover` / `text-popover-foreground` → token asserts FAIL |
| A3W9-5-003 | `frontend/src/components/ui/select.tsx:119` (SelectItem) + `:140` (SelectSeparator `bg-b1`) | same file → `it("A3W9-5-003 SelectItem/Separator align with Dropdown hub tokens")` | вернуть `rounded-sm` / `focus:bg-accent` / `bg-muted` → Item/Separator asserts FAIL |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| avatar / popover / select tokens | ✅ RTL className | 001–003 |
| Trigger+Content Select | OOS | уже FIXED W9-A5-018 — не трогать |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** CONFIRMED — 3/3 runnable.

### Mutation verify (fe1 · 2026-07-30 W13)

3/3 PROVEN — см. MUTATION-2026-07-30-W13-fe1-task176.md.
