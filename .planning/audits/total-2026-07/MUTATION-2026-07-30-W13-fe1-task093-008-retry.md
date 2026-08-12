# MUTATION W13 slot 2 refill — TASK-093 / A2W10-A2-008 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: TASK-093 `[MUTATION CRITERIA]` A2W10-A2-008  
> Prior: [MUTATION-2026-07-30-W13-fe1-task093.md](./MUTATION-2026-07-30-W13-fe1-task093.md) — 008 **NOT_PROVEN**  
> Коммит: **не создавался**.

## Verdict

**A2W10-A2-008 = PROVEN.**  
TASK-093 → **Закрыта** (14 runnable PROVEN incl. 008; GAP_NOTED rows остаются documented, не блокеры).

| ID | Prior | Retry | RED |
|----|-------|-------|-----|
| A2W10-A2-008 | NOT_PROVEN | **PROVEN** | `cases.lazy wires CaseTreePanel into TestCasesPage` |

## Gap closed

Prior unit `it("renders suite tree and all-cases control")` imported `CaseTreePanel` from `@/modules/tms/components/case-tree` — unwiring JSX in `cases.lazy.tsx` still PASS.

Added route/lazy wiring coverage in `case-tree-panel.test.tsx`:

1. **Source-guard** — `readFileSync(cases.lazy.tsx)` asserts `<CaseTreePanel`, import path, and prop wiring (`suites` / `selectedSuiteId` / `onSelectSuite`).
2. **Mount spy** — `vi.mock` wrapper around real `CaseTreePanel`; render `TestCasesPage` from `cases.lazy` → `caseTreeMountSpy` must be called; UI shows «Тест-сьюты» / «Все кейсы».

## Mutation evidence

```text
GREEN baseline:
  npx vitest run …/tms/case-tree-panel.test.tsx
  → 2 passed (~109–121ms)

BREAK: cases.lazy.tsx
  - remove import { CaseTreePanel } …
  - remove <CaseTreePanel suites={…} selectedSuiteId={…} onSelectSuite={…} total={…} />
  - leave MUTATED comment

RED:
  FAIL cases.lazy wires CaseTreePanel into TestCasesPage
  → expected src to match /<CaseTreePanel[\s\n]/
  (component-only it still PASS — confirms prior NOT_PROVEN root cause)

RESTORE: git checkout -- …/cases.lazy.tsx

GREEN:
  case-tree-panel.test.tsx → 2 passed
  cases-list-error.test.tsx → 1 passed (regression neighbor)
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run ../testframework/stages/unit/frontend-test/domains/tms/case-tree-panel.test.tsx
npx vitest run ../testframework/stages/unit/frontend-test/domains/tms/cases-list-error.test.tsx
```

## Production restore

- `cases.lazy.tsx` restored via `git checkout --`
- Leftover MUTATED: нет
- UI/prod wiring intact (`CaseTreePanel` present at L153)

## Handoff

- Verdict **008 PROVEN** · TASK-093 **Закрыта** · README updated  
- Evidence: this file + updated TASK-093 mutation table  
- fe-ver: optional spot-check wiring `it` + restore cleanliness
