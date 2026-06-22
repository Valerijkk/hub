# Phase 60 — UX-1 Gate (a11y + shortcuts + onboarding Zod)

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust` (uncommitted)  
**Depends on:** Phase 59 Coverage + UX Close-out

## Goal

Close UX-1 backlog items: **FB-01** shortcuts E2E, **UX-P1-06** axe calls/TMS, **FB-12** onboarding project step Zod validation.

## Waves

| Wave | Task | Agent | Done |
|------|------|-------|------|
| W1 | a11y E2E: `/calls` + TMS cases | qa | ✓ |
| W2 | E2E `shortcuts.spec.ts` (`?` overlay) + CI | qa + devops | ✓ |
| W3 | `onboarding-project-schema.ts` + step 4 validation | fe | ✓ |
| W4 | Vitest: shortcuts-dialog + schema (5 tests) | fe-ver | ✓ |
| W5 | i18n validation keys + docs sync | docs-writer | ✓ |

## Deliverables

| Item | File |
|------|------|
| Calls a11y | `e2e/a11y.spec.ts` |
| TMS a11y | `e2e/a11y.spec.ts` |
| Shortcuts E2E | `e2e/shortcuts.spec.ts` |
| Onboarding Zod | `onboarding-project-schema.ts` |
| Onboarding UX | `onboarding.lazy.tsx` step 4 errors |
| Vitest | `shortcuts-dialog.test.tsx`, `onboarding-project-schema.test.ts` |
| CI | `ci.yml` + `shortcuts.spec.ts` |

## Verification

```bash
cd frontend && npm run lint && npx tsc --noEmit && npx vitest run
# E2E (stack required):
npx playwright test e2e/shortcuts.spec.ts e2e/a11y.spec.ts --grep "calls|TMS|Shortcuts"
```

## Metrics

- Vitest: **88/88** (35 files)
- a11y specs: **14** authenticated routes (+ calls, TMS)
- E2E specs in CI: **15** files

## Next

- Squash commit `feat/v2.5-trust`
- PG-03 secret rotation (manual)
- Profile step RHF (optional UX-1 remainder)
