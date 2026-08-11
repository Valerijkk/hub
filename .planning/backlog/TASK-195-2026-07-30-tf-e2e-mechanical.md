# TASK-195 — 🎭 Work/realtime/admin E2E journeys

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **138**

> Runtime/traceability gate: the future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. Launcher status remains `NOT_READY` and runtime remains `NOT_RUN` until a separate implementation and independent acceptance. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase `e2e-mechanical` slug is kept
> only for link stability. Current purpose is risk-based work/realtime/admin
> journeys; route by TASK-195, title and Phase 138, not by the filename.

## [OUTCOME]

New Python E2E journeys cover current work/realtime/admin requirements and risks;
this is not mechanical TS migration.

## [ROLE]

- Owner: `qa-auto1`; verifiers: `qa-auto-ver`, `qa-manual-ver`.
- Depends on: TASK-217 and corresponding Phase 134 API suites.

## [OWNED TARGET PATHS]

`testframework/tests/e2e/work/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Happy plus invalid/denied/boundary/conflict/failure-recovery variants.
- Persisted outcome checked beyond toast; role/tenant and applicable a11y covered.
- Stable TC trace, repeated/shuffled Docker runs, sentinel and cleanup PASS.
- No legacy spec/file/case count is an acceptance criterion.
