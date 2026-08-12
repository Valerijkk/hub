# MUTATION W13 slot 2 refill — TASK-176 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/archive/2026-07-29/TASK-176-2026-07-29-a3-9-5-components.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**3/3 PROVEN** · TASK-176 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| A3W9-5-001 | **PROVEN** | `A3W9-5-001 Avatar online-dot uses var(--green), not emerald-500` |
| A3W9-5-002 | **PROVEN** | `A3W9-5-002 PopoverContent uses hub floating-surface tokens` |
| A3W9-5-003 | **PROVEN** | `A3W9-5-003 SelectItem/Separator align with Dropdown hub tokens` |

Production restored (manual restore; `git diff` UI files empty). Leftover MUTATED: **нет**. Mock-gap: нет (RTL на prod `avatar` / `popover` / `select`).

`npx tsc --noEmit` — N/A (не primary per criteria; не ловит className/token regressions).

## Per-ID evidence

### A3W9-5-001 — Avatar online-dot hub green

```text
BREAK: avatar.tsx:34 — bg-[var(--green)] → bg-emerald-500
→ FAIL A3W9-5-001 Avatar online-dot uses var(--green), not emerald-500
     AssertionError: expected '… bg-emerald-500 …' to match /bg-\[var\(--green\)\]/
RESTORE → PASS (baseline 3/3 after full restore)
```

Guard: `frontend/src/components/ui/avatar.tsx:34`  
Vitest: `domains/platform/a3w9-5-hub-tokens-p3.test.tsx` · describe `A3W9-5 · hub tokens avatar/popover/select (TASK-176)`

### A3W9-5-002 — PopoverContent floating-surface tokens

```text
BREAK: popover.tsx:22 — вернуть rounded-md / bg-popover / text-popover-foreground
  (убрать rounded-[var(--r2)] border-b1 bg-base text-t1)
→ FAIL A3W9-5-002 PopoverContent uses hub floating-surface tokens
     AssertionError: expected '… rounded-md border bg-popover …' to match /rounded-\[var\(--r2\)\]/
RESTORE → PASS
```

Guard: `frontend/src/components/ui/popover.tsx:22`  
Vitest: same file · `it("A3W9-5-002 …")`

### A3W9-5-003 — SelectItem/Separator hub tokens

```text
BREAK: select.tsx SelectItem — rounded-sm + focus:bg-accent focus:text-accent-foreground;
  SelectSeparator — bg-muted (вместо bg-b1)
→ FAIL A3W9-5-003 SelectItem/Separator align with Dropdown hub tokens
     AssertionError: expected '… rounded-sm … focus:bg-accent …' to match /rounded-\[var\(--r1\)\]/
RESTORE → PASS 3/3
```

Guard: `frontend/src/components/ui/select.tsx:119` (Item) + `:140` (Separator)  
Vitest: same file · `it("A3W9-5-003 …")`  
Trigger+Content Select — OOS (W9-A5-018), не трогали.

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

# RED (после BREAK) / GREEN (после RESTORE)
npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/a3w9-5-hub-tokens-p3.test.tsx
# per-ID: -t "A3W9-5-001" | -t "A3W9-5-002" | -t "A3W9-5-003"
```

GREEN (после restore): Test Files 1 passed · Tests 3 passed · ~2.0s

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| avatar online-dot tokens | ✅ RTL className | PROVEN |
| PopoverContent floating tokens | ✅ RTL className | PROVEN |
| SelectItem/Separator tokens | ✅ RTL className | PROVEN |
| Trigger+Content Select | OOS / N/A | уже FIXED W9-A5-018 |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** CONFIRMED — 3/3 runnable PROVEN.

## Handoff

- Evidence: этот файл · TASK-176 **Закрыта** · README sync  
- fe-ver → code-reviewer (fix code already FIXED; mutation PROVEN)  
- Коммит: не создавался
