# Phase 37: PG-02 Biome Lint Gate

**Goal:** `npm run lint` → 0 errors, 0 warnings (PG-02 close).

**Branch:** `feat/v2.5-trust` (uncommitted)

## Tasks

- [x] Biome `--write` safe auto-fixes (21 files formatted/organized)
- [x] Fix 10 lint errors (noArrayIndexKey, useExhaustiveDependencies, noAssignInExpressions)
- [x] Fix 72 warnings across a11y + suspicious + style rules
- [x] Add `@/lib/keyboard` helper (`activateOnEnterOrSpace`)
- [x] `buildQueryString` typed as `object` (no `any`, tsc clean)

## Verification

```bash
cd frontend && npm run lint          # 0 errors, 0 warnings
cd frontend && npx tsc --noEmit      # clean
cd frontend && npx vitest run        # 66 tests
```

## Patterns

| Rule | Fix |
|------|-----|
| noLabelWithoutControl | `<span>` section headers or `htmlFor`+`id` |
| noStaticElementInteractions | `<button type="button">` or keyboard handler |
| useSemanticElements | `div role=button` → native `<button>` |
| noArrayIndexKey | content-based keys / message UUIDs |
| useIterableCallbackReturn | `for...of` instead of `forEach` return |

**Status:** COMPLETE 2026-06-24
