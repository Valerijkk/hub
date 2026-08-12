# PLAN: Phase 138 — Python E2E migration 20/53

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Перенести 15 mechanical и 5 complex Playwright TS specs в Python, сохранив
полный behavioral contract 20 specs / 53 cases и dual-run.

## Depends on

- Phase 137 PASS (shared fixtures/POM и 3-spec smoke parity).

## Success criteria

- [ ] 15 mechanical specs сохраняют test IDs, assertions и negative paths.
- [ ] 5 complex specs (`logout`, `onboarding-profile`, `a11y`, `ai-hub`,
  `project-lifecycle`) переписаны на observable waits/isolated data.
- [ ] Python collection соответствует 20 specs / 53 baseline cases либо каждая
  consolidation имеет verifier-approved manifest row.
- [ ] Full TS и Python runs на одном build green; unexpected skips = 0;
  legacy TS sources не удалены.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-195 | Mechanical migration 15 specs с immutable names/IDs и row-level parity | qa-auto1 | qa-auto-ver | `testframework/e2e/{admin,ai,auth,messenger,navigation,platform,tms,v4}/**`, `testframework/reports/**`, `.planning/evidence/test-rebase/138/mechanical/**` | Phase 137 | 1–1.5д |
| TASK-196 | Complex five + full 20 specs/53 cases same-build parity | qa-auto2 | qa-auto-ver | `testframework/e2e/{auth,platform,ai,tracker}/**`, `testframework/reports/**`, `.planning/audits/test-rebase/inventory.json`, `.planning/evidence/test-rebase/138/complex/**` | Phase 137, TASK-195 | 1–1.5д |

## Волны

- Wave 1: TASK-195.
- Wave 2: TASK-196.
- Wave 3: TS/Python full comparison и verifier sign-off.

## Риски

| Риск | Mitigation |
|---|---|
| «Mechanical» скрывает flaky waits | Web-first expectations; no raw timeout; preserve semantics |
| AI provider недетерминирован | Approved provider-double/profile; optional skip не входит required full |
| 53 cases превращаются в меньше assertions | Manifest tracks case IDs и assertion map, не только file count |

## Verification

- `qa-auto-ver`: exact collected/passed/failed/skipped, same-build parity,
  sentinel red/green, trace artifacts и `PARITY_REQUIRED = 0` для E2E.
- `qa-manual-ver`: выборочная browser-проверка complex five.
