# MUTATION W13 slot 2 refill RETRY — TASK-075 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Prior: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-fe1-task075.md` (9/11, 2 NOT_PROVEN)  
> Criteria: `.planning/backlog/archive/legacy/TASK-075-fe-p3-dead-ui-state.md` `[MUTATION CRITERIA]`  
> Коммит: **не создавался**.

## Verdict

**2/2 retry PROVEN → 11/11 overall** · TASK-075 → **Закрыта**

| ID | Prior | Retry | RED |
|----|-------|-------|-----|
| W10-A3-019 | NOT_PROVEN (baseline comment false-positive) | **PROVEN** | `issue-sidebar-meta assignee/reporter are not cursor-pointer` |
| W10-A5-016 | NOT_PROVEN (avatar-only onClick still GREEN) | **PROVEN** | `profile avatar opens file input; no dead department state` |

Assert fixes in `dead-ui-state-p3.test.ts`. Comment hygiene in `issue-sidebar-meta.tsx` (drop self-matching `cursor-pointer` substring). Kill mutations restored; leftover MUTATED: **нет**.

## Assert fixes

### W10-A3-019

**Root cause:** `/Исполнитель[\s\S]{0,200}cursor-pointer/` matched comment  
`{/* Исполнитель — … cursor-pointer … */}` → baseline already RED.

**Fix:**
1. Slice people section (`{/* Исполнитель` … `{/* Метки`) so Status/Priority `cursor-pointer` above do not interfere.
2. Assert `not.toMatch(/className="[^"]*cursor-pointer/)` — comment text cannot self-match.
3. Comment reworded to «без ложного click-affordance».

### W10-A5-016

**Root cause:** whole-file `/fileInputRef\.current\?\.click\(\)/` still matched the «Загрузить фото» button after avatar `onClick` removal.

**Fix:** require avatar wiring specifically:

```ts
/aria-label="Загрузить фото профиля"\s*\n\s*onClick=\{\(\) => fileInputRef\.current\?\.click\(\)\}/
```

Button-only click no longer satisfies the guard.

## Per-ID evidence

### W10-A3-019 — **PROVEN**

```text
BASELINE: PASS (assert + comment fix)
BREAK: assignee row className += cursor-pointer
→ FAIL expected people not to match /className="[^"]*cursor-pointer/
RESTORE → PASS
```

### W10-A5-016 — **PROVEN**

```text
BASELINE: PASS
BREAK: remove onClick from avatar button only (leave «Загрузить фото» button)
→ FAIL expected toMatch /aria-label=…\n\s*onClick={() => fileInputRef…}/
RESTORE → PASS
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/dead-ui-state-p3.test.ts
# → 9/9 PASS

python .planning/audits/total-2026-07/_run_mutations_w13_task075_retry.py
# W10-A3-019 PROVEN · W10-A5-016 PROVEN · full dead-ui exit=0
```

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| W10-A3-019 cursor-pointer on people rows | ✅ source-guard (className in sliced section) | PROVEN |
| W10-A5-016 avatar onClick → fileInput | ✅ source-guard (aria-label+onClick adjacency) | PROVEN |
| Prior 9 IDs | ✅ unchanged | PROVEN (W13 fe1 first pass) |

**Confidence:** CONFIRMED — **11/11** runnable PROVEN after retry.

## Handoff

- Evidence: этот файл · prior `MUTATION-2026-07-30-W13-fe1-task075.md` · TASK-075 **Закрыта** · README sync  
- Vitest: `domains/platform/dead-ui-state-p3.test.ts` (9/9)  
- Product delta kept: comment wording in `issue-sidebar-meta.tsx` only  
- Коммит не создавался  
