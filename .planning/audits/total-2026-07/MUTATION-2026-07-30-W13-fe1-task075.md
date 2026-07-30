# MUTATION W13 slot 2 refill — TASK-075 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/archive/legacy/TASK-075-fe-p3-dead-ui-state.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался**.

## Verdict

**9/11 PROVEN · 2 NOT_PROVEN** · TASK-075 → **Открыта**

| ID | Verdict | RED |
|----|---------|-----|
| W9-A2-019 | **PROVEN** | `deadlines row links to issue route with project_id + key (A3W9-2-002)` |
| W9-A3-023 | **PROVEN** | `mention-autocomplete Esc sets suppressed` |
| W9-A4-012 | **PROVEN** | `issue-attachments empty copy does not promise DnD` |
| W10-A2-025 | **PROVEN** | `run-execution-panel has no dead stepResults state` |
| W10-A3-019 | **NOT_PROVEN** | `issue-sidebar-meta assignee/reporter are not cursor-pointer` — baseline RED |
| W10-A4-009 | **PROVEN** | `issues list has no dead statusFilter state` |
| W10-A4-010 | **PROVEN** | `tms cases bulk bar has delete action` |
| W10-A5-014 | **PROVEN** | `settings NavItem has no unused placeholder branch` |
| W10-A5-016 | **NOT_PROVEN** | kill recipe (avatar `onClick` only) → exit 0; source-guard still sees button `click()` |
| W10-A5-017 | **PROVEN** | same `it` — `not.toMatch(/department:/)` / `const statusDuration = "1d"` |
| W11-A3-016 | **PROVEN** | `tms-store is filters-only (no server-cache fields)` |

Production restored (`git checkout -- frontend/src`). Leftover MUTATED: **нет**.

## Per-ID evidence

### W9-A2-019 — deadlines TaskRowLink — **PROVEN**

```text
BREAK: deadlines.tsx — <TaskRowLink> → <div> (closing tag too)
→ FAIL deadlines row links… (A3W9-2-002)
     Unable to find role="link"
RESTORE → PASS
```

### W9-A3-023 — mention Esc suppressed — **PROVEN**

```text
BREAK: mention-autocomplete.tsx — Escape: drop setSuppressed(true), keep return true
→ FAIL mention-autocomplete Esc sets suppressed
     expected src toMatch /setSuppressed\(true\)/
RESTORE → PASS
```

### W9-A4-012 — attachments empty copy — **PROVEN**

```text
BREAK: issue-attachments.tsx — «Нажмите «Добавить файл»» → «Перетащите файлы…»
→ FAIL issue-attachments empty copy does not promise DnD
     expected src not.toMatch /Перетащите/
RESTORE → PASS
```

### W10-A2-025 — stepResults dead state — **PROVEN**

```text
BREAK: run-execution-panel.tsx — add useState stepResults (unused)
→ FAIL run-execution-panel has no dead stepResults state
     expected src not.toMatch /stepResults/
RESTORE → PASS
```

### W10-A3-019 — assignee/reporter cursor-pointer — **NOT_PROVEN**

```text
BASELINE (no mutation): already FAIL
  expect(src).not.toMatch(/Исполнитель[\s\S]{0,200}cursor-pointer/)
  Comment on L128: «Исполнитель — read-only, без мёртвого cursor-pointer …»
  → comment self-matches within 200 chars of «Исполнитель»
UI blocks (assignee/reporter) correctly lack cursor-pointer.
Restore cannot GREEN → mutation cycle incomplete.

Fix needed (out of mutation scope): rename comment (drop substring cursor-pointer)
and/or tighten regex to className on the meta row only.
```

### W10-A4-009 — statusFilter — **PROVEN**

```text
BREAK: issues/index.tsx — add statusFilter/setStatusFilter useState
→ FAIL issues list has no dead statusFilter state
RESTORE → PASS
```

### W10-A4-010 — TMS bulk delete — **PROVEN**

```text
BREAK: cases.lazy.tsx — rename bulkDeleteMutation + replace «Удалить» label
→ FAIL tms cases bulk bar has delete action
RESTORE → PASS
```

### W10-A5-014 — NavItem placeholder — **PROVEN**

```text
BREAK: settings.tsx — add placeholder?: + if (placeholder) return null
→ FAIL settings NavItem has no unused placeholder branch
RESTORE → PASS
```

### W10-A5-016 — avatar onClick — **NOT_PROVEN**

```text
BREAK (kill recipe): remove onClick from avatar button only; leave «Загрузить фото» button
→ GREEN (exit 0) — source-guard still matches fileInputRef.current?.click() on remaining button

ALT (not criteria): strip ALL fileInputRef.current?.click()
→ FAIL expected toMatch /fileInputRef\.current\?\.click\(\)/
→ proves presence guard, but stated kill recipe does not RED
```

### W10-A5-017 — department / statusDuration — **PROVEN**

```text
BREAK: profile.lazy.tsx — setForm({ … department: "", … })
→ FAIL not.toMatch(/department:/)
RESTORE → PASS

ALT kill: const statusDuration = "1d" → const [statusDuration] = useState("1d")
→ FAIL expected toMatch /const statusDuration = "1d"/
RESTORE → PASS
```

### W11-A3-016 — tms-store filters-only — **PROVEN**

```text
BREAK: tms-store.ts — add testCases/setTestCases/suites/testRuns
→ FAIL not.toHaveProperty("testCases")
RESTORE → PASS
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/dead-ui-state-p3.test.ts
npx vitest run ../testframework/stages/unit/frontend-test/domains/dashboard/a3w9-2-widgets.test.tsx
```

Baseline / final: dead-ui **1 FAIL** (W10-A3-019 comment false-positive) + widgets **PASS**.  
Mutations via `.planning/audits/total-2026-07/_run_mutations_w13_task075.py`.

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| W9-A2-019 deadlines Link | ✅ RTL | PROVEN |
| 8× source-guards + tms-store runtime | ✅ | PROVEN |
| W10-A3-019 cursor-pointer | ❌ baseline broken | comment self-match → NOT_PROVEN |
| W10-A5-016 avatar-only onClick | ❌ kill recipe insufficient | whole-file click() still present → NOT_PROVEN |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** CONFIRMED — 9/11 runnable PROVEN; 2 blockers reopen task.

## Handoff

- Evidence: этот файл · TASK-075 **Открыта** · README sync  
- Vitest: `domains/platform/dead-ui-state-p3.test.ts` + `domains/dashboard/a3w9-2-widgets.test.tsx`  
- Reopen reasons: W10-A3-019 comment/regex; W10-A5-016 kill recipe vs source-guard  
- Коммит не создавался  
