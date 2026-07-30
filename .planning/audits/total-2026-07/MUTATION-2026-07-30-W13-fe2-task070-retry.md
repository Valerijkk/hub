# MUTATION W13 slot 5 refill retry — TASK-070 (fe2) · 2026-07-30

> Checker: **fe2** · Env: local Vitest · Prior: `MUTATION-2026-07-30-W13-fe2-task070.md` (W9-A5-016 NOT_PROVEN)  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**7/7 PROVEN · 2 GAP_NOTED · tsc N/A** · TASK-070 → **Закрыта**

| ID | Verdict | RED |
|----|---------|-----|
| W9-A1-018 | **PROVEN** | (prior W13) `TrashManager tabs use tablist/tab/aria-selected` |
| W9-A4-013 | **PROVEN** | (prior W13) `RouteErrorBoundary fallback has role=alert` |
| W9-A5-013 | **PROVEN** | (prior W13) `CardTitle renders as heading` |
| W9-A5-016 | **PROVEN** | `EmptyState decorative icon is aria-hidden` — wrap `aria-hidden` null |
| W9-A5-012 | **PROVEN** | (prior W13) `Avatar announces online status when data-online` |
| W9-A5-014 | **PROVEN** | (prior W13) `Badge renders as span without focus ring classes` |
| W9-A5-019 | **PROVEN** | (prior W13) `FormField required sets aria-required on control` |
| W9-A5-015 | **GAP_NOTED** | нет dedicated `it` |
| W9-A5-017 | **GAP_NOTED** | нет dedicated `it` |
| `npx tsc --noEmit` | **N/A** | не primary |

Production restored (`empty-state.tsx` `git diff` empty). Leftover MUTATED: **нет**.

## Fix (assert strengthen)

Prior NOT_PROVEN: `container.querySelector("[aria-hidden='true']")` ловил lucide SVG default.

Усиленный assert в `a11y-structure-p3.test.tsx`:

- icon wrap = `container.firstElementChild?.firstElementChild`
- `expect(iconWrap).toHaveAttribute("aria-hidden", "true")`
- SVG child также assert `aria-hidden` (явный prop + lucide)

## Per-ID evidence (retry focus)

### W9-A5-016 — EmptyState explicit wrap aria-hidden — PROVEN

```text
BREAK: empty-state.tsx:24-25 — убрать aria-hidden с wrap и Icon
→ FAIL EmptyState decorative icon is aria-hidden
     expect(iconWrap).toHaveAttribute("aria-hidden", "true")
     Received: null
RESTORE → PASS
```

Guard: `frontend/src/components/ui/empty-state.tsx:24-25`  
Vitest: `domains/platform/a11y-structure-p3.test.tsx` · `it("EmptyState decorative icon is aria-hidden")`

### Prior PROVEN (W13 slot 5) — not re-broken

W9-A1-018, W9-A4-013, W9-A5-013, W9-A5-012, W9-A5-014, W9-A5-019 — evidence в `MUTATION-2026-07-30-W13-fe2-task070.md`. Full file re-run GREEN confirms suite intact.

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs

# RED (после BREAK) / GREEN (после RESTORE / baseline)
npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/a11y-structure-p3.test.tsx -t "EmptyState"
npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/a11y-structure-p3.test.tsx
```

Baseline (strengthened): EmptyState 1 passed · ~1.9s  
RED (mutated): EmptyState FAIL · wrap aria-hidden null · ~1.7s  
Final GREEN: Test Files 1 passed · Tests **7 passed** · ~1.9s

## Gaps (criteria)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| trash / route alert / card / empty wrap / avatar / badge / form-field | ✅ RTL | 7/7 PROVEN |
| W9-A5-015, W9-A5-017 | ❌ GAP | Class source IDs без dedicated `it` |
| `npx tsc --noEmit` | N/A | не primary |

**Confidence:** HIGH — 7/7 runnable PROVEN after assert strengthen; 015/017 remain GAP_NOTED (no kill recipe / no `it`).

## Handoff

- Evidence: этот файл + prior W13 report  
- TASK-070 **Закрыта** · README sync  
- Diff leftover: strengthened Vitest only (`a11y-structure-p3.test.tsx`); UI restored  
- Коммит: не создавался
