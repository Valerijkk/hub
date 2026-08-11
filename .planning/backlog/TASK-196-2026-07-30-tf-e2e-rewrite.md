# TASK-196 — 🎭 Business/integration/recovery E2E journeys

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **138**

> Runtime/traceability gate: the future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. Launcher status remains `NOT_READY` and runtime remains `NOT_RUN` until a separate implementation and independent acceptance. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase `e2e-rewrite` slug is kept only
> for link stability. Current purpose is greenfield business/integration/
> recovery journeys; route by TASK-196, title and Phase 138.

## [OUTCOME]

New Python E2E journeys cover current business and integration user outcomes,
including visible dependency failure and recovery.

## [ROLE]

- Owner: `qa-auto2`; verifiers: `qa-auto-ver`, `qa-manual-ver`.
- Depends on: TASK-217 and applicable Phase 135 API suites.

## [OWNED TARGET PATHS]

`testframework/tests/e2e/business/**`, `testframework/tests/e2e/recovery/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Current requirement/risk selection; happy/negative/boundary/recovery variants.
- Providers use controlled local doubles; user-visible retry/recovery asserted.
- Persisted oracle, role/tenant/a11y as applicable, stable TC trace.
- Repeated/shuffled Docker runs, sentinel, unexpected skip=0 and cleanup PASS.
